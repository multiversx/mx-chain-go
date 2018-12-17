package syncValidators_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/syncValidators"
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

func TestNewSyncValidatorsShouldThrowNilRound(t *testing.T) {
	sv, err := syncValidators.NewSyncValidators(
		nil,
		nil,
	)

	assert.Nil(t, sv)
	assert.Equal(t, execution.ErrNilRound, err)
}

func TestNewSyncValidatorsShouldThrowNilAccountsAdapter(t *testing.T) {
	sv, err := syncValidators.NewSyncValidators(
		&mock.RoundStub{},
		nil,
	)

	assert.Nil(t, sv)
	assert.Equal(t, execution.ErrNilAccountsAdapter, err)
}

func TestNewSyncValidatorsShouldWork(t *testing.T) {
	sv, err := syncValidators.NewSyncValidators(
		&mock.RoundStub{},
		&mock.AccountsStub{},
	)

	assert.NotNil(t, sv)
	assert.Nil(t, err)
}

func TestGetEligibleListShouldHaveNoValidatorsAfterOneRegisterRequest(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := syncValidators.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      *big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	assert.Equal(t, 0, len(sv.GetEligibleList()))
}

func TestGetEligibleListShouldHaveOneValidatorAfterOneRegisterRequestAndSomeRounds(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := syncValidators.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      *big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	index := regData.RoundIndex
	rnds.IndexCalled = func() int32 {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 1, len(sv.GetEligibleList()))
}

func TestGetEligibleListShouldHaveOneValidatorWithIncreasedStakeAfterTwoRegisterRequestsAndSomeRounds(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := syncValidators.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      *big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	regData2 := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      *big.NewInt(2),
		RoundIndex: 3,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData2)

	index := regData.RoundIndex
	rnds.IndexCalled = func() int32 {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 1, len(sv.GetEligibleList()))
	assert.Equal(t, *big.NewInt(3), sv.GetEligibleList()["node1"].Stake)
}

func TestGetEligibleListShouldHaveOneValidatorAfterOneRegisterAndUnregisterRequest(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := syncValidators.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      *big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	regData2 := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArUnregister,
		Stake:      *big.NewInt(1),
		RoundIndex: 1,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData2)

	index := regData.RoundIndex
	rnds.IndexCalled = func() int32 {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 1, len(sv.GetEligibleList()))
}

func TestGetEligibleListShouldHaveNoValidatorsAfterOneRegisterAndUnregisterRequestAndSomeRounds(t *testing.T) {
	rnds, acss, jawm := Init()
	sv, _ := syncValidators.NewSyncValidators(
		rnds,
		acss,
	)

	regData := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArRegister,
		Stake:      *big.NewInt(1),
		RoundIndex: 0,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData)

	regData2 := state.RegistrationData{
		NodePubKey: []byte("node1"),
		Action:     state.ArUnregister,
		Stake:      *big.NewInt(1),
		RoundIndex: 1,
	}

	jawm.Account.RegistrationData = append(jawm.Account.RegistrationData, regData2)

	index := regData2.RoundIndex
	rnds.IndexCalled = func() int32 {
		return index + syncValidators.RoundsToWaitToBeEligible + 1
	}

	assert.Equal(t, 0, len(sv.GetEligibleList()))
}
