package state_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestAccount_MarshalUnmarshalNilSlice_ShouldWork(t *testing.T) {
	t.Parallel()

	acnt := &state.Account{
		Nonce:            8,
		Balance:          *big.NewInt(56),
		CodeHash:         nil,
		RootHash:         nil,
		RegistrationData: nil,
	}

	marshalizer := mock.MarshalizerMock{}

	buff, _ := marshalizer.Marshal(&acnt)

	acntRecovered := state.NewAccount()
	_ = marshalizer.Unmarshal(acntRecovered, buff)

	assert.Equal(t, acnt, acntRecovered)
}

func TestAccount_MarshalUnmarshalEmptySlice_ShouldWork(t *testing.T) {
	t.Parallel()

	acnt := &state.Account{
		Nonce:            8,
		Balance:          *big.NewInt(56),
		CodeHash:         nil,
		RootHash:         nil,
		RegistrationData: make([]state.RegistrationData, 0),
	}

	marshalizer := mock.MarshalizerMock{}

	buff, _ := marshalizer.Marshal(acnt)

	acntRecovered := state.NewAccount()
	_ = marshalizer.Unmarshal(acntRecovered, buff)

	assert.Equal(t, acnt, acntRecovered)
}

func TestAccount_MarshalUnmarshalWithRegData_ShouldWork(t *testing.T) {
	t.Parallel()

	acnt := &state.Account{
		Nonce:    8,
		Balance:  *big.NewInt(56),
		CodeHash: nil,
		RootHash: nil,
		RegistrationData: []state.RegistrationData{
			{
				OriginatorPubKey: []byte("a"),
				NodePubKey:       []byte("b"),
				Stake:            *big.NewInt(5),
				Action:           state.ArRegister,
			},
		},
	}

	marshalizer := mock.MarshalizerMock{}

	buff, _ := marshalizer.Marshal(acnt)

	acntRecovered := state.NewAccount()
	_ = marshalizer.Unmarshal(acntRecovered, buff)

	assert.Equal(t, acnt, acntRecovered)
}
