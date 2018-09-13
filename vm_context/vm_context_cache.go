package vm_context

import (
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/trie"
)

type ContractGid struct {
	gid  *types.Gid
	addr *types.Address
	open bool
}

func (contractGid *ContractGid) Gid() *types.Gid {
	return contractGid.gid
}

func (contractGid *ContractGid) Addr() *types.Address {
	return contractGid.addr
}

func (contractGid *ContractGid) Open() bool {
	return contractGid.open
}

type UnsavedCache struct {
	contractGidList []*ContractGid

	logList    ledger.VmLogList
	storage    map[string][]byte
	sendBlocks []*ledger.AccountBlock

	trie *trie.Trie

	trieDirty bool
}

func NewUnsavedCache(trie *trie.Trie) *UnsavedCache {
	return &UnsavedCache{
		storage: make(map[string][]byte),

		trie:      trie.Copy(),
		trieDirty: false,
	}
}

func (cache *UnsavedCache) SendBlocks() []*ledger.AccountBlock {
	return cache.sendBlocks
}

func (cache *UnsavedCache) Trie() *trie.Trie {
	if cache.trieDirty {
		for key, value := range cache.storage {
			cache.trie.SetValue([]byte(key), value)
		}
	}
	return cache.trie
}

func (cache *UnsavedCache) SetStorage(key []byte, value []byte) {
	cache.storage[string(key)] = value
	cache.trieDirty = true
}

func (cache *UnsavedCache) GetStorage(key []byte) []byte {
	return cache.storage[string(key)]
}

func (cache *UnsavedCache) ContractGidList() []*ContractGid {
	return cache.contractGidList
}

func (cache *UnsavedCache) LogList() ledger.VmLogList {
	return cache.logList
}

func (cache *UnsavedCache) Storage() map[string][]byte {
	return cache.storage
}
