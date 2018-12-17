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

	acnt := state.Account{
		Nonce:            8,
		Balance:          *big.NewInt(56),
		CodeHash:         nil,
		RootHash:         nil,
		RegistrationData: nil,
	}

	marshalizer := mock.MarshalizerMock{}

	buff, err := marshalizer.Marshal(&acnt)
	assert.Nil(t, err)

	acntRecovered := state.NewAccount()
	err = marshalizer.Unmarshal(acntRecovered, buff)
	assert.Nil(t, err)
	assert.Equal(t, uint64(8), acntRecovered.Nonce)
	assert.Equal(t, *big.NewInt(56), acntRecovered.Balance)
	assert.Nil(t, acntRecovered.CodeHash)
	assert.Nil(t, acntRecovered.RootHash)
	assert.Nil(t, acntRecovered.RegistrationData)

}

func TestAccount_MarshalUnmarshalEmptySlice_ShouldWork(t *testing.T) {
	t.Parallel()

	acnt := state.Account{
		Nonce:            8,
		Balance:          *big.NewInt(56),
		CodeHash:         nil,
		RootHash:         nil,
		RegistrationData: make([]state.RegistrationData, 0),
	}

	marshalizer := mock.MarshalizerMock{}

	buff, err := marshalizer.Marshal(&acnt)
	assert.Nil(t, err)

	acntRecovered := state.NewAccount()
	err = marshalizer.Unmarshal(acntRecovered, buff)
	assert.Nil(t, err)
	assert.Equal(t, uint64(8), acntRecovered.Nonce)
	assert.Equal(t, *big.NewInt(56), acntRecovered.Balance)
	assert.Nil(t, acntRecovered.CodeHash)
	assert.Nil(t, acntRecovered.RootHash)
	assert.Equal(t, 0, len(acntRecovered.RegistrationData))
}

func TestAccount_MarshalUnmarshalWithRegData_ShouldWork(t *testing.T) {
	t.Parallel()

	acnt := state.Account{
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

	buff, err := marshalizer.Marshal(&acnt)
	assert.Nil(t, err)

	acntRecovered := state.NewAccount()
	err = marshalizer.Unmarshal(acntRecovered, buff)
	assert.Nil(t, err)
	assert.Equal(t, uint64(8), acntRecovered.Nonce)
	assert.Equal(t, *big.NewInt(56), acntRecovered.Balance)
	assert.Nil(t, acntRecovered.CodeHash)
	assert.Nil(t, acntRecovered.RootHash)
	assert.Equal(t, 1, len(acntRecovered.RegistrationData))
	assert.Equal(t, []byte("a"), acntRecovered.RegistrationData[0].OriginatorPubKey)
	assert.Equal(t, []byte("b"), acntRecovered.RegistrationData[0].NodePubKey)
	assert.Equal(t, *big.NewInt(5), acntRecovered.RegistrationData[0].Stake)
	assert.Equal(t, state.ArRegister, acntRecovered.RegistrationData[0].Action)
}
