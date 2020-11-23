package accounts

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/mock"
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
