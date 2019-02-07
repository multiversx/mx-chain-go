package sync_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/stretchr/testify/assert"
)

func Init() (*mock.RoundStub, *mock.AccountsStub, *mock.JournalizedAccountWrapMock) {
	rnds := &mock.RoundStub{}
	rnds.IndexCalled = func() int32 {
		return 0
	}

	acss := &mock.AccountsStub{}

	jawm := mock.NewJournalizedAccountWrapMock(state.RegistrationAddress)

	acss.GetJournalizedAccountCalled = func(
		addressContainer state.AddressContainer,
	) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), state.RegistrationAddress.Bytes()) {
			return jawm, nil
		}

		return nil, errors.New("failure")
	}

	return rnds, acss, jawm
}

func TestNewSyncValidators_ShouldThrowNilRound(t *testing.T) {
	sv, err := sync.NewSyncValidators(
		nil,
		nil,
	)

	assert.Nil(t, sv)
	assert.Equal(t, process.ErrNilRound, err)
}

func TestNewSyncValidators_ShouldThrowNilAccountsAdapter(t *testing.T) {
	sv, err := sync.NewSyncValidators(
		&mock.RoundStub{},
		nil,
	)

	assert.Nil(t, sv)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewSyncValidators_ShouldWork(t *testing.T) {
	sv, err := sync.NewSyncValidators(
		&mock.RoundStub{},
		&mock.AccountsStub{},
	)

	assert.NotNil(t, sv)
	assert.Nil(t, err)
}

func TestGetEligibleList_ShouldHaveNoValidatorsAfterOneRegisterRequest(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := sync.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	assert.Equal(t, 0, len(sv.GetEligibleList()))
}

func TestGetEligibleList_ShouldHaveOneValidatorAfterOneRegisterRequestAndSomeRounds(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := sync.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	index := regData.RoundIndex
	rnds.IndexCalled = func() int32 {
		return index + sync.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 1, len(sv.GetEligibleList()))
}

func TestGetEligibleList_ShouldHaveOneValidatorWithIncreasedStakeAfterTwoRegisterRequestsAndSomeRounds(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := sync.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	regData2 := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      big.NewInt(2),
		RoundIndex: 3,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData2)

	index := regData.RoundIndex
	rnds.IndexCalled = func() int32 {
		return index + sync.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 1, len(sv.GetEligibleList()))
	assert.Equal(t, big.NewInt(3), sv.GetEligibleList()["node1"].Stake)
}

func TestGetEligibleList_ShouldHaveOneValidatorAfterOneRegisterAndUnregisterRequest(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := sync.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	regData2 := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArUnregister,
		Stake:      big.NewInt(1),
		RoundIndex: 1,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData2)

	index := regData.RoundIndex
	rnds.IndexCalled = func() int32 {
		return index + sync.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 1, len(sv.GetEligibleList()))
}

func TestGetEligibleList_ShouldHaveNoValidatorsAfterOneRegisterAndUnregisterRequestAndSomeRounds(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := sync.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	regData2 := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArUnregister,
		Stake:      big.NewInt(1),
		RoundIndex: 1,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData2)

	index := regData2.RoundIndex
	rnds.IndexCalled = func() int32 {
		return index + sync.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 0, len(sv.GetEligibleList()))
}
