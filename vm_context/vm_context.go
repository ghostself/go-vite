package vm_context

import (
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"math/big"
)

type Gid [10]byte

const (
	ACTION_ADD_BALANCE = iota
	ACTION_SUB_BALANCE
	ACTION_SET_GID
	ACTION_SET_CONTRACT
	ACTION_SET_TOKEN
	ACTION_SET_STORAGE
	ACTION_ADD_LOG
)

type Action struct {
	ActionType int32
	Params     []interface{}
}

func NewAction(actionType int32, params []interface{}) *Action {
	return &Action{
		ActionType: actionType,
		Params:     params,
	}
}

type VmContext struct {
	currentSnapshotBlockHash *types.Hash
	prevAccountBlockHash     *types.Hash
	address                  *types.Address

	actionList []*Action
	cache      *VmContextCache
}

func NewVmContext(snapshotBlockHash *types.Hash, prevAccountBlockHash *types.Hash, addr *types.Address) *VmContext {
	return &VmContext{
		currentSnapshotBlockHash: snapshotBlockHash,
		prevAccountBlockHash:     prevAccountBlockHash,
		address:                  addr,

		cache: NewVmContextCache(),
	}
}

func (context *VmContext) addAction(actionType int32, params []interface{}) {
	context.actionList = append(context.actionList, NewAction(actionType, params))
}

func (context *VmContext) Address() *types.Address {
	return context.address
}

func (context *VmContext) ActionList() []*Action {
	return context.actionList
}

func (context *VmContext) GetBalance(addr *types.Address, tokenTypeId *types.TokenTypeId) *big.Int {
	return big.NewInt(0)
}

func (context *VmContext) AddBalance(tokenTypeId *types.TokenTypeId, amount *big.Int) {
	context.addAction(ACTION_ADD_BALANCE, []interface{}{tokenTypeId, amount})
}

func (context *VmContext) SubBalance(tokenTypeId *types.TokenTypeId, amount *big.Int) {
	context.addAction(ACTION_SUB_BALANCE, []interface{}{tokenTypeId, amount})
}

func (context *VmContext) GetSnapshotBlock(hash *types.Hash) (*ledger.SnapshotBlock, error) {

	return nil, nil
}

func (context *VmContext) GetSnapshotBlockByHeight(height *big.Int) (*ledger.SnapshotBlock, error) {

	return nil, nil
}

func (context *VmContext) Reset() {

}

func (context *VmContext) SetContractGid(gid *Gid, open bool) {

}

func (context *VmContext) SetContractCode(gid *Gid, code []byte) {

}

func (context *VmContext) GetContractCode() []byte {
	return nil
}

func (context *VmContext) SetToken() {

}

func (context *VmContext) GetToken(id *types.TokenTypeId) {

}

func (context *VmContext) SetStorage(key []byte, value []byte) {

}

func (context *VmContext) GetStorage(key []byte) {

}

func (context *VmContext) GetStorageHash() *types.Hash {
	return nil
}

func (context *VmContext) AddLog(log *ledger.VmLog) {

}

func (context *VmContext) GetLogListHash() {

}

func (context *VmContext) IsAddressExisted() {

}

func (context *VmContext) GetAccoutBlockByHash() {

}
