package contracts

import (
	"github.com/vitelabs/go-vite/abi"
	"github.com/vitelabs/go-vite/common/helper"
	"github.com/vitelabs/go-vite/common/types"
	"math/big"
	"strings"
)

const (
	MethodNameRegister           = "Register"
	MethodNameCancelRegister     = "CancelRegister"
	MethodNameReward             = "Reward"
	MethodNameUpdateRegistration = "UpdateRegistration"
	VariableNameRegistration     = "registration"

	MethodNameVote         = "Vote"
	MethodNameCancelVote   = "CancelVote"
	VariableNameVoteStatus = "voteStatus"

	MethodNamePledge             = "Pledge"
	MethodNameCancelPledge       = "CancelPledge"
	VariableNamePledgeInfo       = "pledgeInfo"
	VariableNamePledgeBeneficial = "pledgeBeneficial"

	MethodNameCreateConsensusGroup         = "CreateConsensusGroup"
	VariableNameConsensusGroupInfo         = "consensusGroupInfo"
	VariableNameConditionCountingOfBalance = "countingOfBalance"
	VariableNameConditionRegisterOfPledge  = "registerOfPledge"
	VariableNameConditionVoteOfKeepToken   = "voteOfKeepToken"

	MethodNameMintage             = "Mintage"
	MethodNameMintageCancelPledge = "CancelPledge"
	VariableNameMintage           = "mintage"
)

const json_register = `
[
	{"type":"function","name":"Register", "inputs":[{"name":"gid","type":"gid"},{"name":"name","type":"string"},{"name":"NodeAddr","type":"address"},{"name":"beneficialAddr","type":"address"}]},
	{"type":"function","name":"UpdateRegistration", "inputs":[{"name":"gid","type":"gid"},{"name":"name","type":"string"},{"name":"NodeAddr","type":"address"},{"name":"beneficialAddr","type":"address"}]},
	{"type":"function","name":"CancelRegister","inputs":[{"name":"gid","type":"gid"}, {"name":"name","type":"string"}]},
	{"type":"function","name":"Reward","inputs":[{"name":"gid","type":"gid"},{"name":"name","type":"string"},{"name":"endHeight","type":"uint64"},{"name":"startHeight","type":"uint64"},{"name":"amount","type":"uint256"}]},
	{"type":"variable","name":"registration","inputs":[{"name":"name","type":"string"},{"name":"NodeAddr","type":"address"},{"name":"pledgeAddr","type":"address"},{"name":"beneficialAddr","type":"address"},{"name":"amount","type":"uint256"},{"name":"timestamp","type":"int64"},{"name":"rewardHeight","type":"uint64"},{"name":"cancelHeight","type":"uint64"}]}
]`
const json_vote = `
[
	{"type":"function","name":"Vote", "inputs":[{"name":"gid","type":"gid"},{"name":"nodeName","type":"string"}]},
	{"type":"function","name":"CancelVote","inputs":[{"name":"gid","type":"gid"}]},
	{"type":"variable","name":"voteStatus","inputs":[{"name":"nodeName","type":"string"}]}
]`
const json_pledge = `
[
	{"type":"function","name":"Pledge", "inputs":[{"name":"beneficial","type":"address"},{"name":"withdrawTime","type":"int64"}]},
	{"type":"function","name":"CancelPledge","inputs":[{"name":"beneficial","type":"address"},{"name":"amount","type":"uint256"}]},
	{"type":"variable","name":"pledgeInfo","inputs":[{"name":"amount","type":"uint256"},{"name":"withdrawTime","type":"int64"}]},
	{"type":"variable","name":"pledgeBeneficial","inputs":[{"name":"amount","type":"uint256"}]}
]`
const json_consensusGroup = `
[
	{"type":"function","name":"CreateConsensusGroup", "inputs":[{"name":"gid","type":"gid"},{"name":"nodeCount","type":"uint8"},{"name":"interval","type":"int64"},{"name":"countingRuleId","type":"uint8"},{"name":"countingRuleParam","type":"bytes"},{"name":"registerConditionId","type":"uint8"},{"name":"registerConditionParam","type":"bytes"},{"name":"voteConditionId","type":"uint8"},{"name":"voteConditionParam","type":"bytes"}]},
	{"type":"variable","name":"consensusGroupInfo","inputs":[{"name":"nodeCount","type":"uint8"},{"name":"interval","type":"int64"},{"name":"countingRuleId","type":"uint8"},{"name":"countingRuleParam","type":"bytes"},{"name":"registerConditionId","type":"uint8"},{"name":"registerConditionParam","type":"bytes"},{"name":"voteConditionId","type":"uint8"},{"name":"voteConditionParam","type":"bytes"}]},
	{"type":"variable","name":"countingOfBalance","inputs":[{"name":"tokenId","type":"tokenId"}]},
	{"type":"variable","name":"registerOfPledge","inputs":[{"name":"pledgeAmount","type":"uint256"},{"name":"pledgeToken","type":"tokenId"},{"name":"pledgeTime","type":"int64"}]},
	{"type":"variable","name":"voteOfKeepToken","inputs":[{"name":"keepAmount","type":"uint256"},{"name":"keepToken","type":"tokenId"}]}
]`
const json_mintage = `
[
	{"type":"function","name":"Mintage","inputs":[{"name":"tokenId","type":"tokenId"},{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"}]},
	{"type":"function","name":"CancelPledge","inputs":[{"name":"tokenId","type":"tokenId"}]},
	{"type":"variable","name":"mintage","inputs":[{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"owner","type":"address"},{"name":"pledgeAmount","type":"uint256"},{"name":"timestamp","type":"int64"}]}
]
`

var (
	ABI_register, _       = abi.JSONToABIContract(strings.NewReader(json_register))
	ABI_vote, _           = abi.JSONToABIContract(strings.NewReader(json_vote))
	ABI_pledge, _         = abi.JSONToABIContract(strings.NewReader(json_pledge))
	ABI_consensusGroup, _ = abi.JSONToABIContract(strings.NewReader(json_consensusGroup))
	ABI_mintage, _        = abi.JSONToABIContract(strings.NewReader(json_mintage))
)

type ParamRegister struct {
	Gid            types.Gid
	Name           string
	NodeAddr       types.Address
	BeneficialAddr types.Address
}
type ParamCancelRegister struct {
	Gid  types.Gid
	Name string
}
type ParamReward struct {
	Gid         types.Gid
	Name        string
	EndHeight   uint64
	StartHeight uint64
	Amount      *big.Int
}
type ParamVote struct {
	Gid      types.Gid
	NodeName string
}
type VariablePledgeInfo struct {
	Amount       *big.Int
	WithdrawTime int64
}
type VariablePledgeBeneficial struct {
	Amount *big.Int
}
type ParamPledge struct {
	Beneficial   types.Address
	WithdrawTime int64
}
type ParamCancelPledge struct {
	Beneficial types.Address
	Amount     *big.Int
}
type VariableConditionRegisterOfPledge struct {
	PledgeAmount *big.Int
	PledgeToken  types.TokenTypeId
	PledgeTime   int64
}
type VariableConditionVoteOfKeepToken struct {
	KeepAmount *big.Int
	KeepToken  types.TokenTypeId
}
type ParamMintage struct {
	TokenId     types.TokenTypeId
	TokenName   string
	TokenSymbol string
	TotalSupply *big.Int
	Decimals    uint8
}

func GetRegisterKey(name string, gid types.Gid) []byte {
	var data = make([]byte, types.HashSize)
	copy(data[0:10], gid[:])
	copy(data[10:], types.DataHash([]byte(name)).Bytes()[10:])
	return data
}
func GetVoteKey(addr types.Address, gid types.Gid) []byte {
	var data = make([]byte, types.HashSize)
	copy(data[0:10], gid[:])
	copy(data[12:], addr[:])
	return data
}
func GetAddrFromVoteKey(key []byte) types.Address {
	addr, _ := types.BytesToAddress(key[12:])
	return addr
}
func GetConsensusGroupKey(gid types.Gid) []byte {
	return helper.LeftPadBytes(gid.Bytes(), types.HashSize)
}
func GetGidFromConsensusGroupKey(key []byte) types.Gid {
	gid, _ := types.BytesToGid(key[types.HashSize-types.GidSize:])
	return gid
}