package validatorInfo

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// IsLeavingEligible returns true if the validator was eligible in the epoch but has done an unstake.
func IsLeavingEligible(valInfo *state.ValidatorInfo) bool {
	if valInfo == nil {
		return false
	}

	return valInfo.List == string(core.LeavingList) &&
		(valInfo.LeaderFailure > 0 ||
			valInfo.LeaderSuccess > 0 ||
			valInfo.ValidatorSuccess > 0 ||
			valInfo.ValidatorFailure > 0)
}
