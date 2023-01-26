package fee

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewFeeComputer(t *testing.T) {
	t.Run("NilBuiltInFunctionsCostHandler", func(t *testing.T) {
		arguments := ArgsNewFeeComputer{
			BuiltInFunctionsCostHandler: nil,
			EconomicsConfig:             testscommon.GetEconomicsConfig(),
		}

		computer, err := NewFeeComputer(arguments)
		require.Equal(t, process.ErrNilBuiltInFunctionsCostHandler, err)
		require.True(t, check.IfNil(computer))
	})

	t.Run("AllArgumentsProvided", func(t *testing.T) {
		arguments := ArgsNewFeeComputer{
			BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
			EconomicsConfig:             testscommon.GetEconomicsConfig(),
		}

		computer, err := NewFeeComputer(arguments)
		require.Nil(t, err)
		require.NotNil(t, computer)
	})
}

func TestFeeComputer_ComputeTransactionFeeShouldWorkForDifferentEpochs(t *testing.T) {
	arguments := ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
		EnableEpochsConfig: config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: 124,
			GasPriceModifierEnableEpoch:    180,
		},
	}

	contract, _ := hex.DecodeString("000000000000000000010000000000000000000000000000000000000000abba")

	computer, _ := NewFeeComputer(arguments)

	checkComputedFee(t, "50000000000000", computer, 0, 80000, 1000000000, "", nil)
	checkComputedFee(t, "57500000000000", computer, 0, 80000, 1000000000, "hello", nil)
	checkComputedFee(t, "57500000000000", computer, 0, 1000000, 1000000000, "hello", contract)
	checkComputedFee(t, "80000000000000", computer, 124, 80000, 1000000000, "", nil)
	checkComputedFee(t, "1000000000000000", computer, 124, 1000000, 1000000000, "hello", contract)
	checkComputedFee(t, "404000000000000", computer, 124, 404000, 1000000000, "ESDTTransfer@464f4f2d653962643235@0a", contract)
	checkComputedFee(t, "57500010000000", computer, 180, 57501, 1000000000, "hello", contract)
	checkComputedFee(t, "66925000000000", computer, 180, 1000000, 1000000000, "hello", contract)
	checkComputedFee(t, "66925000000000", computer, 180, 1000000, 1000000000, "hello", contract)
	checkComputedFee(t, "107000000000000", computer, 180, 404000, 1000000000, "ESDTTransfer@464f4f2d653962643235@0a", contract)
	// TODO: Add tests for guarded transactions, when enabled.
}

func checkComputedFee(t *testing.T, expectedFee string, computer *feeComputer, epoch int, gasLimit uint64, gasPrice uint64, data string, receiver []byte) {
	tx := &transaction.ApiTransactionResult{
		Epoch: uint32(epoch),
		Tx: &transaction.Transaction{
			RcvAddr:  receiver,
			Data:     []byte(data),
			GasLimit: gasLimit,
			GasPrice: gasPrice,
		},
	}

	fee := computer.ComputeTransactionFee(tx)
	require.Equal(t, expectedFee, fee.String())
}
