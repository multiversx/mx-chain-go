package defaults

import (
	"github.com/ElrondNetwork/elrond-go/common"
)

// FillGasMapInternal -
func FillGasMapInternal(gasMap map[string]map[string]uint64, value uint64) map[string]map[string]uint64 {
	gasMap[common.BaseOperationCost] = FillGasMapBaseOperationCosts(value)
	gasMap[common.BuiltInCost] = FillGasMapBuiltInCosts(value)
	gasMap[common.MetaChainSystemSCsCost] = FillGasMapMetaChainSystemSCsCosts(value)

	return gasMap
}

// FillGasMapBaseOperationCosts -
func FillGasMapBaseOperationCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["StorePerByte"] = value
	gasMap["DataCopyPerByte"] = value
	gasMap["ReleasePerByte"] = value
	gasMap["PersistPerByte"] = value
	gasMap["CompilePerByte"] = value
	gasMap["AoTPreparePerByte"] = value
	gasMap["GetCode"] = value
	return gasMap
}

// FillGasMapBuiltInCosts -
func FillGasMapBuiltInCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["ClaimDeveloperRewards"] = value
	gasMap["ChangeOwnerAddress"] = value
	gasMap["SaveUserName"] = value
	gasMap["SaveKeyValue"] = value
	gasMap["ESDTTransfer"] = value
	gasMap["ESDTBurn"] = value
	gasMap["ESDTLocalMint"] = value
	gasMap["ESDTLocalBurn"] = value
	gasMap["ESDTNFTCreate"] = value
	gasMap["ESDTNFTAddQuantity"] = value
	gasMap["ESDTNFTBurn"] = value
	gasMap["ESDTNFTTransfer"] = value
	gasMap["ESDTNFTChangeCreateOwner"] = value
	gasMap["ESDTNFTAddUri"] = value
	gasMap["ESDTNFTUpdateAttributes"] = value
	gasMap["ESDTNFTMultiTransfer"] = value

	return gasMap
}

// FillGasMapMetaChainSystemSCsCosts -
func FillGasMapMetaChainSystemSCsCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["Stake"] = value
	gasMap["UnStake"] = value
	gasMap["UnBond"] = value
	gasMap["Claim"] = value
	gasMap["Get"] = value
	gasMap["ChangeRewardAddress"] = value
	gasMap["ChangeValidatorKeys"] = value
	gasMap["UnJail"] = value
	gasMap["ESDTIssue"] = value
	gasMap["ESDTOperations"] = value
	gasMap["Proposal"] = value
	gasMap["Vote"] = value
	gasMap["DelegateVote"] = value
	gasMap["RevokeVote"] = value
	gasMap["CloseProposal"] = value
	gasMap["DelegationOps"] = value
	gasMap["UnStakeTokens"] = value
	gasMap["UnBondTokens"] = value
	gasMap["DelegationMgrOps"] = value
	gasMap["GetAllNodeStates"] = value
	gasMap["ValidatorToDelegation"] = value

	return gasMap
}
