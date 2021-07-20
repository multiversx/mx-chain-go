package validatorInfo

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/stretchr/testify/require"
)

func Test_IsLeavingEligible_NilValidatorStatisticsDoesNotErr(t *testing.T) {
	isLeavingEligible := WasLeavingEligibleInCurrentEpoch(nil)

	require.False(t, isLeavingEligible)
}

func Test_IsLeavinEligible_Eligible(t *testing.T) {
	valInfo := &state.ValidatorInfo{
		List:             string(common.EligibleList),
		LeaderSuccess:    0,
		LeaderFailure:    0,
		ValidatorSuccess: 0,
		ValidatorFailure: 0,
	}

	isLeavingEligible := WasLeavingEligibleInCurrentEpoch(valInfo)
	require.False(t, isLeavingEligible)
}

func Test_IsLeavingEligible_NotEligibleNotLeaving(t *testing.T) {
	valInfo := &state.ValidatorInfo{
		List:             string(common.InactiveList),
		LeaderSuccess:    1,
		LeaderFailure:    10,
		ValidatorSuccess: 11,
		ValidatorFailure: 11,
	}

	isLeavingEligible := WasLeavingEligibleInCurrentEpoch(valInfo)
	require.False(t, isLeavingEligible)
}

func Test_IsLeavingEligible_LeavingNoData(t *testing.T) {
	valInfo := &state.ValidatorInfo{
		List:             string(common.LeavingList),
		LeaderSuccess:    0,
		LeaderFailure:    0,
		ValidatorSuccess: 0,
		ValidatorFailure: 0,
	}

	isLeavingEligible := WasLeavingEligibleInCurrentEpoch(valInfo)
	require.False(t, isLeavingEligible)
}

func Test_IsLeavingEligible_LeavingWithData(t *testing.T) {
	// should be considered leaving eligible

	valInfo := &state.ValidatorInfo{
		List:             string(common.LeavingList),
		LeaderSuccess:    1,
		LeaderFailure:    10,
		ValidatorSuccess: 11,
		ValidatorFailure: 11,
	}

	isLeavingEligible := WasLeavingEligibleInCurrentEpoch(valInfo)
	require.True(t, isLeavingEligible)
}
