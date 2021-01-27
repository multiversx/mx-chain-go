package validatorInfo

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// WasActiveInCurrentEpoch returns true if the node was active in current epoch
func WasActiveInCurrentEpoch(valInfo *state.ValidatorInfo) bool {
	active := valInfo.LeaderFailure > 0 || valInfo.LeaderSuccess > 0 || valInfo.ValidatorSuccess > 0 || valInfo.ValidatorFailure > 0
	return active
}

// WasLeavingEligibleInCurrentEpoch returns true if the validator was eligible in the epoch but has done an unstake.
func WasLeavingEligibleInCurrentEpoch(valInfo *state.ValidatorInfo) bool {
	if valInfo == nil {
		return false
	}

	return valInfo.List == string(core.LeavingList) && WasActiveInCurrentEpoch(valInfo)
}

// WasJailedEligibleInCurrentEpoch returns true if the validator was jailed in the epoch but also active/eligible due to not enough
//nodes in shard.
func WasJailedEligibleInCurrentEpoch(valInfo *state.ValidatorInfo) bool {
	if valInfo == nil {
		return false
	}

	return valInfo.List == string(core.JailedList) && WasActiveInCurrentEpoch(valInfo)
}

// WasEligibleInCurrentEpoch returns true if the validator was eligible for consensus in current epoch
func WasEligibleInCurrentEpoch(valInfo *state.ValidatorInfo) bool {
	wasEligibleInShard := valInfo.List == string(core.EligibleList) ||
		WasLeavingEligibleInCurrentEpoch(valInfo) ||
		WasJailedEligibleInCurrentEpoch(valInfo)

	return wasEligibleInShard
}
