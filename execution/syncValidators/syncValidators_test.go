package syncValidators_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/syncValidators"
	"github.com/stretchr/testify/assert"
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
		&mock.RoundMock{},
	)

	assert.NotNil(t, sv)
	assert.Nil(t, err)
}

func TestAddValidatorInWaitListShouldHaveOneValidator(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))
	assert.Equal(t, 1, len(sv.GetWaitList()))
}

func TestAddValidatorTwiceInWaitListShouldHaveOneValidator(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))
	sv.AddValidator("node1", *big.NewInt(0))
	assert.Equal(t, 1, len(sv.GetWaitList()))
}

func TestAddValidatorTwiceInWaitListShouldHaveTwoValidators(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))
	sv.AddValidator("node2", *big.NewInt(0))
	assert.Equal(t, 2, len(sv.GetWaitList()))
}

func TestAddValidatorInUnregisterListShouldHaveNoValidators(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.RemoveValidator("node1")
	assert.Equal(t, 0, len(sv.GetUnregisterList()))
}

func TestAddValidatorInUnregisterListShouldHaveNoValidators2(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))
	sv.RemoveValidator("node1")
	assert.Equal(t, 0, len(sv.GetUnregisterList()))
}

func TestAddValidatorInUnregisterListShouldHaveOneValidator(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.RemoveValidator("node1")

	assert.Equal(t, 1, len(sv.GetUnregisterList()))
}

func TestAddValidatorInUnregisterListShouldHaveNoValidators3(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.RemoveValidator("node1")

	index = rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 0, len(sv.GetUnregisterList()))
}

func TestAddValidatorWithDifferentIdsInUnregisterListShouldHaveTwoValidators(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))
	sv.AddValidator("node2", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.RemoveValidator("node1")
	sv.RemoveValidator("node2")

	assert.Equal(t, 2, len(sv.GetUnregisterList()))
}

func TestGetEligibleListShouldHaveNoValidators(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))
	assert.Equal(t, 0, len(sv.GetEligibleList()))
}

func TestGetEligibleListShouldHaveOneValidatorAfterSomeTime(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 1, len(sv.GetEligibleList()))
}

func TestGetWaitListShouldHaveNoValidatorsAfterSomeTime(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 0, len(sv.GetWaitList()))
}

func TestGetEligibleListShouldHaveOneValidatorFromTwoAfterSomeTime(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.AddValidator("node2", *big.NewInt(0))

	assert.Equal(t, 1, len(sv.GetEligibleList()))
}

func TestGetEligibleListShouldHaveTwoValidatorsAfterSomeTime(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.AddValidator("node2", *big.NewInt(0))

	index = rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 2, len(sv.GetEligibleList()))
}

func TestGetWaitListShouldHaveOneValidatorAfterSomeTime(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.AddValidator("node2", *big.NewInt(0))

	assert.Equal(t, 1, len(sv.GetWaitList()))
}

func TestGetWaitListShouldHaveNoValidatorsFromTwoAfterSomeTime(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.AddValidator("node2", *big.NewInt(0))

	index = rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 0, len(sv.GetWaitList()))
}

func TestGetUnregisterListShouldHaveNoValidators(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))
	sv.RemoveValidator("node1")

	assert.Equal(t, 0, len(sv.GetUnregisterList()))
}

func TestGetUnregisterListShouldHaveOneValidator(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.RemoveValidator("node1")

	assert.Equal(t, 1, len(sv.GetUnregisterList()))
}

func TestGetUnregisterListShouldHaveNoValidatorsAfterSomeTime(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(0))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.RemoveValidator("node1")

	index = rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 0, len(sv.GetUnregisterList()))
}

func TestGetWaitListShouldHaveOneValidatorWithIncreasedStake(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(1))
	sv.AddValidator("node1", *big.NewInt(2))

	wl := sv.GetWaitList()

	vd := wl["node1"]

	assert.Equal(t, 1, len(wl))
	assert.Equal(t, *big.NewInt(3), vd.Stake)
}

func TestGetWaitListShouldHaveNoValidatorsAndEligibleListOneValidatorWithIncreasedStake(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(1))

	index := rndm.Index()
	rndm.IndexCalled = func() int {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	sv.AddValidator("node1", *big.NewInt(2))

	wl := sv.GetWaitList()
	el := sv.GetEligibleList()

	vd := el["node1"]

	assert.Equal(t, 0, len(wl))
	assert.Equal(t, 1, len(el))
	assert.Equal(t, *big.NewInt(3), vd.Stake)
}

func TestStakeShouldNotBeChanged(t *testing.T) {
	rndm := &mock.RoundMock{}
	rndm.IndexCalled = func() int {
		return 0
	}

	sv, _ := syncValidators.NewSyncValidators(
		rndm,
	)

	sv.AddValidator("node1", *big.NewInt(1))

	wl1 := sv.GetWaitList()

	sv.AddValidator("node1", *big.NewInt(3))

	wl2 := sv.GetWaitList()

	assert.Equal(t, *big.NewInt(1), wl1["node1"].Stake)
	assert.Equal(t, *big.NewInt(4), wl2["node1"].Stake)
}
