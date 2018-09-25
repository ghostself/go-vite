package vm_context

import (
	"bytes"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/trie"
	"github.com/vitelabs/go-vite/vm_context/vmctxt_interface"
	"math/big"
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
}

func NewVmContext(chain Chain, snapshotBlockHash *types.Hash, prevAccountBlockHash *types.Hash, addr *types.Address) (vmctxt_interface.VmDatabase, error) {
	vmContext := &VmContext{
		chain:   chain,
		address: addr,

		frozen: false,
	}

	currentSnapshotBlock, err := chain.GetSnapshotBlockByHash(snapshotBlockHash)
	if err != nil {
		return nil, err
	}

	vmContext.currentSnapshotBlock = currentSnapshotBlock

	var prevAccountBlock *ledger.AccountBlock
	if prevAccountBlockHash == nil {
		var err error
		prevAccountBlock, err = chain.GetConfirmAccountBlock(currentSnapshotBlock.Height, addr)
		if err != nil {
			return nil, err
		}

	} else {
		var err error
		prevAccountBlock, err = chain.GetAccountBlockByHash(prevAccountBlockHash)
		if err != nil {
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
	return addr == nil || bytes.Equal(addr.Bytes(), context.Address().Bytes())
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

func (context *VmContext) GetSnapshotBlockByHeight(height uint64) *ledger.SnapshotBlock {
	if height > context.currentSnapshotBlock.Height {
		return nil
	}
	snapshotBlock, _ := context.chain.GetSnapshotBlockByHeight(height)

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

		return context.trie.GetValue(key)
	} else {
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
	account, _ := context.chain.GetAccount(addr)
	if account == nil {
		return false
	}
	return true
}

func (context *VmContext) GetAccountBlockByHash(hash *types.Hash) *ledger.AccountBlock {
	accountBlock, _ := context.chain.GetAccountBlockByHash(hash)
	return accountBlock
}

func (context *VmContext) NewStorageIterator(prefix []byte) vmctxt_interface.StorageIterator {
	return NewStorageIterator(context.unsavedCache.Trie(), prefix)
}