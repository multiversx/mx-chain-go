package state_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"math/big"
	"testing"
)

func TestAccount_MarshalUnmarshal_ShouldWork(t *testing.T) {
	t.Parallel()

	acnt := &state.Account{
		Nonce:    8,
		Balance:  big.NewInt(56),
		CodeHash: nil,
		RootHash: nil,
	}

	marshalizer := mock.MarshalizerMock{}

	buff, _ := marshalizer.Marshal(&acnt)

	acntRecovered, _ := state.NewAccount()
	_ = marshalizer.Unmarshal(acntRecovered, buff)

	assert.Equal(t, acnt, acntRecovered)
}
