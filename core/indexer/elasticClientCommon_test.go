package indexer

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadResponseBody_NilBodyNilDest(t *testing.T) {
	t.Parallel()

	err := loadResponseBody(nil, nil)
	require.NoError(t, err)
}

func TestLoadResponseBody_NilBodyNotNilDest(t *testing.T) {
	t.Parallel()

	err := loadResponseBody(nil, struct{}{})
	require.NoError(t, err)
}

func TestComputeBalanceAsFloat(t *testing.T) {
	t.Parallel()

	type input struct {
		balance      *big.Int
		denomination int
		numDecimals  int
	}
	type testStr struct {
		input  input
		output float64
	}
	tests := []testStr{
		{
			input: input{
				balance:      big.NewInt(1000),
				denomination: 2,
				numDecimals:  1,
			},
			output: float64(10),
		},
		{
			input: input{
				balance:      big.NewInt(1111),
				denomination: 4,
				numDecimals:  3,
			},
			output: 0.111,
		},
		{
			input: input{
				balance:      big.NewInt(10),
				denomination: 0,
				numDecimals:  0,
			},
			output: float64(10),
		},
		{
			input: input{
				balance:      big.NewInt(555),
				denomination: 5,
				numDecimals:  3,
			},
			output: 0.006,
		},
		{
			input: input{
				balance:      big.NewInt(555),
				denomination: 5,
				numDecimals:  4,
			},
			output: 0.0056,
		},
		{
			input: input{
				balance:      big.NewInt(-555),
				denomination: 3,
				numDecimals:  3,
			},
			output: 0,
		},
	}

	// real case scenarios
	balance, _ := big.NewInt(0).SetString("2220555555555555555555555", 10)
	inp := input{
		balance:      balance,
		denomination: 18,
		numDecimals:  10,
	}
	out := 2220555.5555555556
	tests = append(tests, testStr{
		input:  inp,
		output: out,
	})

	balance, _ = big.NewInt(0).SetString("20000000000000000000000000", 10) // 20 mil
	inp = input{
		balance:      balance,
		denomination: 18,
		numDecimals:  10,
	}
	out = 20000000
	tests = append(tests, testStr{
		input:  inp,
		output: out,
	})

	for _, tt := range tests {
		output := computeBalanceAsFloat(tt.input.balance, tt.input.denomination, tt.input.numDecimals)
		assert.Equal(t, tt.output, output)
	}
}
