package validatorInfo

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/stretchr/testify/assert"
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

func TestWasActiveInCurrentEpoch(t *testing.T) {
	t.Parallel()

	assert.False(t, WasActiveInCurrentEpoch(nil))
}

func TestWasJailedEligibleInCurrentEpoch(t *testing.T) {
	t.Parallel()

	assert.False(t, WasJailedEligibleInCurrentEpoch(nil))

	// leaving not jailed should be false
	valInfo := &state.ValidatorInfo{
		List: string(common.WaitingList),
	}
	assert.False(t, WasJailedEligibleInCurrentEpoch(valInfo))

	// leaving jailed but not in current epoch should be false
	valInfo = &state.ValidatorInfo{
		List:             string(common.JailedList),
		LeaderSuccess:    0,
		LeaderFailure:    0,
		ValidatorSuccess: 0,
		ValidatorFailure: 0,
	}
	assert.False(t, WasJailedEligibleInCurrentEpoch(valInfo))

	// leaving jailed and active in current epoch should be true
	valInfo = &state.ValidatorInfo{
		List:             string(common.JailedList),
		LeaderSuccess:    1,
		LeaderFailure:    10,
		ValidatorSuccess: 11,
		ValidatorFailure: 11,
	}
	assert.True(t, WasJailedEligibleInCurrentEpoch(valInfo))
}

func TestWasEligibleInCurrentEpoch(t *testing.T) {
	t.Parallel()

	assert.False(t, WasEligibleInCurrentEpoch(nil))

	// jailed but not active should be false
	valInfo := &state.ValidatorInfo{
		List: string(common.JailedList),
	}
	assert.False(t, WasEligibleInCurrentEpoch(valInfo))

	// eligible should be true
	valInfo = &state.ValidatorInfo{
		List: string(common.EligibleList),
	}
	assert.True(t, WasEligibleInCurrentEpoch(valInfo))

	// jailed and active in current epoch should be true
	valInfo = &state.ValidatorInfo{
		List:             string(common.JailedList),
		LeaderSuccess:    1,
		LeaderFailure:    10,
		ValidatorSuccess: 11,
		ValidatorFailure: 11,
	}
	assert.True(t, WasEligibleInCurrentEpoch(valInfo))

	// leaving and active in current epoch should be true
	valInfo = &state.ValidatorInfo{
		List:             string(common.LeavingList),
		LeaderSuccess:    1,
		LeaderFailure:    10,
		ValidatorSuccess: 11,
		ValidatorFailure: 11,
	}
	assert.True(t, WasEligibleInCurrentEpoch(valInfo))
}
