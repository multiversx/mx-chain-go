package syncValidators_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/syncValidators"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestNewSyncValidatorsShouldThrowNilRound(t *testing.T) {
	sv, err := syncValidators.NewSyncValidators(
		nil,
	)

	assert.Nil(t, sv)
	assert.Equal(t, execution.ErrNilRound, err)
}

func TestNewSyncValidatorsShouldWork(t *testing.T) {
	sv, err := syncValidators.NewSyncValidators(
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
	)

	assert.NotNil(t, sv)
	assert.Nil(t, err)
}

func TestAddValidatorInWaitListShouldHaveOneValidator(t *testing.T) {
	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))
	assert.Equal(t, 1, len(sv.GetWaitList()))
}

func TestAddValidatorTwiceInWaitListShouldHaveOneValidator(t *testing.T) {
	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))
	sv.AddValidator("node1", *big.NewInt(0))
	assert.Equal(t, 1, len(sv.GetWaitList()))
}

func TestAddValidatorTwiceInWaitListShouldHaveTwoValidators(t *testing.T) {
	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))
	sv.AddValidator("node2", *big.NewInt(0))
	assert.Equal(t, 2, len(sv.GetWaitList()))
}

func TestAddValidatorInUnregisterListShouldHaveOneValidator(t *testing.T) {
	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
	)

	sv.RemoveValidator("node1")
	assert.Equal(t, 1, len(sv.GetUnregisterList()))
}

func TestAddValidatorTwiceInUnregisterListShouldHaveOneValidator(t *testing.T) {
	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
	)

	sv.RemoveValidator("node1")
	sv.RemoveValidator("node1")
	assert.Equal(t, 1, len(sv.GetUnregisterList()))
}

func TestAddValidatorTwiceInUnregisterListShouldHaveTwoValidator(t *testing.T) {
	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
	)

	sv.RemoveValidator("node1")
	sv.RemoveValidator("node2")
	assert.Equal(t, 2, len(sv.GetUnregisterList()))
}

func TestGetEligibleListShouldHaveNoValidators(t *testing.T) {
	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))
	assert.Equal(t, 0, len(sv.GetEligibleList()))
}

func TestGetEligibleListShouldHaveOneValidatorAfterSomeTime(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	assert.Equal(t, 1, len(sv.GetEligibleList()))
}

func TestGetWaitListShouldHaveNoValidatorsAfterSomeTime(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	assert.Equal(t, 0, len(sv.GetWaitList()))
}

func TestGetEligibleListShouldHaveOneValidatorFromTwoAfterSomeTime(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.AddValidator("node2", *big.NewInt(0))

	assert.Equal(t, 1, len(sv.GetEligibleList()))
}

func TestGetEligibleListShouldHaveTwoValidatorsAfterSomeTime(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.AddValidator("node2", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	assert.Equal(t, 2, len(sv.GetEligibleList()))
}

func TestGetWaitListShouldHaveOneValidatorAfterSomeTime(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.AddValidator("node2", *big.NewInt(0))

	assert.Equal(t, 1, len(sv.GetWaitList()))
}

func TestGetWaitListShouldHaveNoValidatorsFromTwoAfterSomeTime(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.AddValidator("node2", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	assert.Equal(t, 0, len(sv.GetWaitList()))
}

func TestGetUnregisterListShouldHaveNoValidators(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))
	sv.RemoveValidator("node1")

	assert.Equal(t, 0, len(sv.GetUnregisterList()))
}

func TestGetUnregisterListShouldHaveOneValidator(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.Refresh()

	sv.RemoveValidator("node1")

	assert.Equal(t, 1, len(sv.GetUnregisterList()))
}

func TestGetUnregisterListShouldHaveNoValidatorsAfterSomeTime(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(0))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.Refresh()

	sv.RemoveValidator("node1")

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	assert.Equal(t, 0, len(sv.GetUnregisterList()))
}

func TestGetWaitListShouldHaveOneValidatorWithIncreasedStake(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(1))
	sv.AddValidator("node1", *big.NewInt(2))

	wl := sv.GetWaitList()

	vd := wl["node1"]

	assert.Equal(t, 1, len(wl))
	assert.Equal(t, *big.NewInt(3), vd.Stake)
}

func TestGetWaitListShouldHaveOneValidatorWithoutIncreasedStake(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(1))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.Refresh()

	sv.AddValidator("node1", *big.NewInt(2))

	wl := sv.GetWaitList()

	vd := wl["node1"]

	assert.Equal(t, 1, len(sv.GetEligibleList()))
	assert.Equal(t, 1, len(wl))
	assert.Equal(t, *big.NewInt(2), vd.Stake)
}

func TestGetEligibleListShouldHaveOneValidatorWithoutIncreasedStake(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(1))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.Refresh()

	sv.AddValidator("node1", *big.NewInt(2))

	el := sv.GetEligibleList()

	vd := el["node1"]

	assert.Equal(t, 1, len(sv.GetWaitList()))
	assert.Equal(t, 1, len(el))
	assert.Equal(t, *big.NewInt(1), vd.Stake)
}

func TestGetEligibleListShouldHaveOneValidatorWithIncreasedStake(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(1))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.Refresh()

	sv.AddValidator("node1", *big.NewInt(2))

	currentTime = currentTime.Add((syncValidators.RoundsToWait + 1) * 100 * time.Millisecond)
	sv.Round().UpdateRound(genesisTime, currentTime)

	sv.Refresh()

	el := sv.GetEligibleList()

	vd := el["node1"]

	assert.Equal(t, 0, len(sv.GetWaitList()))
	assert.Equal(t, 1, len(el))
	assert.Equal(t, *big.NewInt(3), vd.Stake)
}

func TestStakeShouldNotBeChanged(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	sv, _ := syncValidators.NewSyncValidators(
		chronology.NewRound(genesisTime, currentTime, time.Duration(100*time.Millisecond)),
	)

	sv.AddValidator("node1", *big.NewInt(1))

	wl1 := sv.GetWaitList()

	sv.AddValidator("node1", *big.NewInt(3))

	wl2 := sv.GetWaitList()

	assert.Equal(t, *big.NewInt(1), wl1["node1"].Stake)
	assert.Equal(t, *big.NewInt(4), wl2["node1"].Stake)
}
