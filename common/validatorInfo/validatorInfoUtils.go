package validatorInfo

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// WasActiveInCurrentEpoch returns true if the node was active in current epoch
func WasActiveInCurrentEpoch(valInfo state.ValidatorInfoHandler) bool {
	if valInfo == nil {
		return false
	}

	active := valInfo.GetLeaderFailure() > 0 || valInfo.GetLeaderSuccess() > 0 || valInfo.GetValidatorSuccess() > 0 || valInfo.GetValidatorFailure() > 0
	return active
}

// WasLeavingEligibleInCurrentEpoch returns true if the validator was eligible in the epoch but has done an unstake.
func WasLeavingEligibleInCurrentEpoch(valInfo state.ValidatorInfoHandler) bool {
	if valInfo == nil {
		return false
	}

	return valInfo.GetList() == string(common.LeavingList) && WasActiveInCurrentEpoch(valInfo)
}

// WasJailedEligibleInCurrentEpoch returns true if the validator was jailed in the epoch but also active/eligible due to not enough
// nodes in shard.
func WasJailedEligibleInCurrentEpoch(valInfo state.ValidatorInfoHandler) bool {
	if valInfo == nil {
		return false
	}

	return valInfo.GetList() == string(common.JailedList) && WasActiveInCurrentEpoch(valInfo)
}

// WasEligibleInCurrentEpoch returns true if the validator was eligible for consensus in current epoch
func WasEligibleInCurrentEpoch(valInfo state.ValidatorInfoHandler) bool {
	if valInfo == nil {
		return false
	}

	wasEligibleInShard := valInfo.GetList() == string(common.EligibleList) ||
		WasLeavingEligibleInCurrentEpoch(valInfo) ||
		WasJailedEligibleInCurrentEpoch(valInfo)

	return wasEligibleInShard
}
