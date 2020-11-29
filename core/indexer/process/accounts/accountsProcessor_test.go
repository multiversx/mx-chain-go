package accounts

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountsProcessor_ComputeBalanceAsFloat(t *testing.T) {
	t.Parallel()

	ap := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})
	require.NotNil(t, ap)

	tests := []struct {
		input  *big.Int
		output float64
	}{
		{
			input:  big.NewInt(200000000000000000),
			output: float64(20000000),
		},
		{
			input:  big.NewInt(57777777777),
			output: 5.7777777777,
		},
		{
			input:  big.NewInt(5777779),
			output: 0.0005777779,
		},
		{
			input:  big.NewInt(7),
			output: 0.0000000007,
		},
		{
			input:  big.NewInt(-7),
			output: 0.0,
		},
	}

	for _, tt := range tests {
		out := ap.computeBalanceAsFloat(tt.input)
		assert.Equal(t, tt.output, out)
	}
}

func TestGetESDTInfo(t *testing.T) {
	ap := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})
	require.NotNil(t, ap)

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
	}

	tokenIdentifier := "token-001"
	wrapAccount := &AccountESDT{
		Account: &mock.UserAccountStub{
			DataTrieTrackerCalled: func() state.DataTrieTracker {
				return &mock.DataTrieTrackerStub{
					RetrieveValueCalled: func(key []byte) ([]byte, error) {
						return json.Marshal(esdtToken)
					},
				}
			},
		},
		TokenIdentifier: tokenIdentifier,
	}
	balance, prop, err := ap.getESDTInfo(wrapAccount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1000), balance)
	require.Equal(t, hex.EncodeToString([]byte("ok")), prop)
}
