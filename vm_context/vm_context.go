package vm_context

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/trie"
	"github.com/vitelabs/go-vite/vm/contracts/abi"
	"github.com/vitelabs/go-vite/vm_context/vmctxt_interface"
	"math/big"
	"time"
)

var (
	STORAGE_KEY_BALANCE = []byte("$balance")
	STORAGE_KEY_CODE    = []byte("$code")
)

func BalanceKey(tokenTypeId *types.TokenTypeId) []byte {
	return append(STORAGE_KEY_BALANCE, tokenTypeId.Bytes()...)
}

type VmContext struct {
	chain   Chain
	address *types.Address

	currentSnapshotBlock *ledger.SnapshotBlock
	prevAccountBlock     *ledger.AccountBlock
	trie                 *trie.Trie

	unsavedCache *UnsavedCache
	frozen       bool

	log log15.Logger
}

func NewEmptyVmContextByTrie(t *trie.Trie) vmctxt_interface.VmDatabase {
	if t == nil {
		t = trie.NewTrie(nil, nil, nil)
	}

	vmContext := &VmContext{
		frozen: false,
		trie:   t,
	}
	vmContext.unsavedCache = NewUnsavedCache(t)
	return vmContext
}

func NewVmContext(chain Chain, snapshotBlockHash *types.Hash, prevAccountBlockHash *types.Hash, addr *types.Address) (vmctxt_interface.VmDatabase, error) {
	vmContext := &VmContext{
		chain:   chain,
		address: addr,

		frozen: false,
		log:    log15.New("module", "vmContext"),
	}

	var currentSnapshotBlock *ledger.SnapshotBlock
	if snapshotBlockHash == nil {
		currentSnapshotBlock = chain.GetLatestSnapshotBlock()
	} else {
		var err error
		currentSnapshotBlock, err = chain.GetSnapshotBlockByHash(snapshotBlockHash)
		if err != nil {
			return nil, err
		}

		if currentSnapshotBlock == nil {
			err := errors.New("currentSnapshotBlock is nil")
			vmContext.log.Error(err.Error(), "method", "NewVmContext")
			return nil, err
		}
	}

	vmContext.currentSnapshotBlock = currentSnapshotBlock

	var prevAccountBlock *ledger.AccountBlock
	if prevAccountBlockHash == nil {
		if addr != nil {
			var err error
			prevAccountBlock, err = chain.GetLatestAccountBlock(addr)
			if err != nil {
				return nil, err
			}
		}
	} else {
		var err error
		prevAccountBlock, err = chain.GetAccountBlockByHash(prevAccountBlockHash)
		if err != nil {
			return nil, err
		}

		if prevAccountBlock == nil {
			err := errors.New("prevAccountBlock is nil")
			vmContext.log.Error(err.Error(), "method", "NewVmContext")
			return nil, err
		}
	}

	if prevAccountBlock != nil {
		vmContext.prevAccountBlock = prevAccountBlock
		vmContext.trie = chain.GetStateTrie(&prevAccountBlock.StateHash)
	}

	if vmContext.trie == nil {
		vmContext.trie = chain.NewStateTrie()
	}

	vmContext.unsavedCache = NewUnsavedCache(vmContext.trie)

	return vmContext, nil
}

func (context *VmContext) CopyAndFreeze() vmctxt_interface.VmDatabase {
	copyTrie := context.unsavedCache.Trie().Copy()
	context.frozen = true
	return &VmContext{
		chain:                context.chain,
		address:              context.address,
		currentSnapshotBlock: context.currentSnapshotBlock,

		trie:         copyTrie,
		unsavedCache: NewUnsavedCache(copyTrie),
		frozen:       false,
	}
}

func (context *VmContext) Address() *types.Address {
	return context.address
}

func (context *VmContext) CurrentSnapshotBlock() *ledger.SnapshotBlock {
	return context.currentSnapshotBlock
}
func (context *VmContext) PrevAccountBlock() *ledger.AccountBlock {
	return context.prevAccountBlock
}

func (context *VmContext) UnsavedCache() vmctxt_interface.UnsavedCache {
	return context.unsavedCache
}

func (context *VmContext) isSelf(addr *types.Address) bool {
	return context.Address() != nil && (addr == nil || bytes.Equal(addr.Bytes(), context.Address().Bytes()))
}

func (context *VmContext) codeKey() []byte {
	return STORAGE_KEY_CODE
}

func (context *VmContext) GetBalance(addr *types.Address, tokenTypeId *types.TokenTypeId) *big.Int {
	var balance = big.NewInt(0)
	if balanceBytes := context.GetStorage(addr, BalanceKey(tokenTypeId)); balanceBytes != nil {
		balance.SetBytes(balanceBytes)
	}
	return balance
}

func (context *VmContext) AddBalance(tokenTypeId *types.TokenTypeId, amount *big.Int) {
	if context.frozen {
		return
	}
	currentBalance := context.GetBalance(context.address, tokenTypeId)

	currentBalance.Add(currentBalance, amount)

	context.SetStorage(BalanceKey(tokenTypeId), currentBalance.Bytes())
}

func (context *VmContext) SubBalance(tokenTypeId *types.TokenTypeId, amount *big.Int) {
	if context.frozen {
		return
	}
	currentBalance := context.GetBalance(context.address, tokenTypeId)
	currentBalance.Sub(currentBalance, amount)

	if currentBalance.Sign() < 0 {
		return
	}

	context.SetStorage(BalanceKey(tokenTypeId), currentBalance.Bytes())
}

func (context *VmContext) GetSnapshotBlocks(startHeight, count uint64, forward, containSnapshotContent bool) []*ledger.SnapshotBlock {
	// For NewEmptyVmContextByTrie
	if context.chain == nil {
		context.log.Error("context.chain is nil", "method", "GetSnapshotBlocks")
		return nil
	}

	if startHeight > context.currentSnapshotBlock.Height {
		return nil
	}

	if forward {
		maxCount := context.currentSnapshotBlock.Height - startHeight + 1
		if count > maxCount {
			count = maxCount
		}
	}

	snapshotBlocks, _ := context.chain.GetSnapshotBlocksByHeight(startHeight, count, forward, containSnapshotContent)
	return snapshotBlocks
}

func (context *VmContext) GetSnapshotBlockByHeight(height uint64) (*ledger.SnapshotBlock, error) {
	// For NewEmptyVmContextByTrie
	if context.chain == nil {
		context.log.Error("context.chain is nil", "method", "GetSnapshotBlockByHeight")
		return nil, errors.New("context.chain is nil")
	}

	if height > context.currentSnapshotBlock.Height {
		return nil, nil
	}
	snapshotBlock, _ := context.chain.GetSnapshotBlockByHeight(height)

	return snapshotBlock, nil
}

func (context *VmContext) GetSnapshotBlockByHash(hash *types.Hash) *ledger.SnapshotBlock {
	// For NewEmptyVmContextByTrie
	if context.chain == nil {
		context.log.Error("context.chain is nil", "method", "GetSnapshotBlockByHash")
		return nil
	}

	snapshotBlock, _ := context.chain.GetSnapshotBlockByHash(hash)
	if snapshotBlock != nil && snapshotBlock.Height > context.currentSnapshotBlock.Height {
		return nil
	}
	return snapshotBlock
}

func (context *VmContext) Reset() {
	context.unsavedCache = NewUnsavedCache(context.trie)
}

func (context *VmContext) SetContractGid(gid *types.Gid, addr *types.Address) {
	if context.frozen {
		return
	}

	contractGid := &ContractGid{
		gid:  gid,
		addr: addr,
	}
	context.unsavedCache.contractGidList = append(context.unsavedCache.contractGidList, contractGid)
}

func (context *VmContext) SetContractCode(code []byte) {
	if context.frozen {
		return
	}

	context.SetStorage(context.codeKey(), code)
}

func (context *VmContext) GetContractCode(addr *types.Address) []byte {
	return context.GetStorage(addr, context.codeKey())
}

func (context *VmContext) SetStorage(key []byte, value []byte) {
	if context.frozen {
		return
	}

	// For unsaved judge.
	if value == nil {
		value = make([]byte, 0)
	}
	context.unsavedCache.SetStorage(key, value)
}

func (context *VmContext) GetStorage(addr *types.Address, key []byte) []byte {
	if context.isSelf(addr) {
		if value := context.unsavedCache.GetStorage(key); value != nil {
			return value
		}
		if value := context.unsavedCache.trie.GetValue(key); value != nil {
			return value
		}
		return context.trie.GetValue(key)
	} else if context.chain != nil {
		latestAccountBlock, _ := context.chain.GetConfirmAccountBlock(context.currentSnapshotBlock.Height, addr)
		if latestAccountBlock != nil {
			trie := context.chain.GetStateTrie(&latestAccountBlock.StateHash)
			return trie.GetValue(key)
		}
	}
	return nil
}

func (context *VmContext) GetStorageHash() *types.Hash {
	return context.unsavedCache.Trie().Hash()
}

func (context *VmContext) GetGid() *types.Gid {
	if context.chain == nil {
		context.log.Error("context.chain is nil", "method", "GetGid")
		return nil
	}
	gid, _ := context.chain.GetContractGid(context.address)
	return gid
}

func (context *VmContext) AddLog(log *ledger.VmLog) {
	if context.frozen {
		return
	}
	context.unsavedCache.logList = append(context.unsavedCache.logList, log)
}

func (context *VmContext) GetLogListHash() *types.Hash {
	return context.unsavedCache.logList.Hash()
}

func (context *VmContext) IsAddressExisted(addr *types.Address) bool {
	if context.chain == nil {
		context.log.Error("context.chain is nil", "method", "IsAddressExisted")
		return false
	}

	account, _ := context.chain.GetAccount(addr)
	if account == nil {
		return false
	}
	return true
}

func (context *VmContext) GetAccountBlockByHash(hash *types.Hash) *ledger.AccountBlock {
	if context.chain == nil {
		context.log.Error("context.chain is nil", "method", "GetAccountBlockByHash")
		return nil
	}

	accountBlock, _ := context.chain.GetAccountBlockByHash(hash)
	return accountBlock
}

func (context *VmContext) NewStorageIterator(addr *types.Address, prefix []byte) vmctxt_interface.StorageIterator {
	if context.isSelf(addr) {
		return NewStorageIterator(context.unsavedCache.Trie(), prefix)
	} else {
		latestAccountBlock, _ := context.chain.GetConfirmAccountBlock(context.currentSnapshotBlock.Height, addr)
		if latestAccountBlock != nil {
			trie := context.chain.GetStateTrie(&latestAccountBlock.StateHash)
			return NewStorageIterator(trie, prefix)
		}
	}
	return nil
}

func (context *VmContext) getBalanceBySnapshotHash(addr *types.Address, tokenTypeId *types.TokenTypeId, snapshotHash *types.Hash) *big.Int {
	var balance = big.NewInt(0)
	if balanceBytes := context.GetStorageBySnapshotHash(addr, BalanceKey(tokenTypeId), snapshotHash); balanceBytes != nil {
		balance.SetBytes(balanceBytes)
	}
	return balance
}
func (context *VmContext) GetStorageBySnapshotHash(addr *types.Address, key []byte, snapshotHash *types.Hash) []byte {
	if snapshotHash == nil || *snapshotHash == context.currentSnapshotBlock.Hash {
		return context.GetStorage(addr, key)
	}
	if snapshotBlock := context.GetSnapshotBlockByHash(snapshotHash); snapshotBlock != nil {
		if latestAccountBlock, _ := context.chain.GetConfirmAccountBlock(snapshotBlock.Height, addr); latestAccountBlock != nil {
			trie := context.chain.GetStateTrie(&latestAccountBlock.StateHash)
			return trie.GetValue(key)
		}
	}
	return nil
}
func (context *VmContext) NewStorageIteratorBySnapshotHash(addr *types.Address, prefix []byte, snapshotHash *types.Hash) vmctxt_interface.StorageIterator {
	if snapshotHash == nil || *snapshotHash == context.currentSnapshotBlock.Hash {
		return context.NewStorageIterator(addr, prefix)
	}
	if snapshotBlock := context.GetSnapshotBlockByHash(snapshotHash); snapshotBlock != nil {
		if latestAccountBlock, _ := context.chain.GetConfirmAccountBlock(snapshotBlock.Height, addr); latestAccountBlock != nil {
			trie := context.chain.GetStateTrie(&latestAccountBlock.StateHash)
			return NewStorageIterator(trie, prefix)
		}
	}
	return nil
}

func (context *VmContext) GetConsensusGroupList(snapshotHash types.Hash) ([]*types.ConsensusGroupInfo, error) {
	return abi.GetActiveConsensusGroupList(context, &snapshotHash), nil
}
func (context *VmContext) GetRegisterList(snapshotHash types.Hash, gid types.Gid) ([]*types.Registration, error) {
	return abi.GetCandidateList(context, gid, &snapshotHash), nil
}
func (context *VmContext) GetVoteMap(snapshotHash types.Hash, gid types.Gid) ([]*types.VoteInfo, error) {
	return abi.GetVoteList(context, gid, &snapshotHash), nil
}
func (context *VmContext) GetBalanceList(snapshotHash types.Hash, tokenTypeId types.TokenTypeId, addressList []types.Address) (map[types.Address]*big.Int, error) {
	balanceList := make(map[types.Address]*big.Int)
	for _, addr := range addressList {
		balanceList[addr] = context.getBalanceBySnapshotHash(&addr, &tokenTypeId, &snapshotHash)
	}
	return balanceList, nil
}
func (context *VmContext) GetSnapshotBlockBeforeTime(timestamp *time.Time) (*ledger.SnapshotBlock, error) {
	return context.chain.GetSnapshotBlockBeforeTime(timestamp)
}

func (context *VmContext) GetGenesisSnapshotBlock() *ledger.SnapshotBlock {
	return context.chain.GetGenesisSnapshotBlock()
}
