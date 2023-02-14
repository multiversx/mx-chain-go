package fee

import (
	"encoding/hex"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
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

func TestFeeComputer_ComputeGasUsedAndFeeBasedOnRefundValue(t *testing.T) {
	arguments := ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
		EnableEpochsConfig: config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: 124,
			GasPriceModifierEnableEpoch:    180,
		},
	}

	computer, _ := NewFeeComputer(arguments)

	contract, _ := hex.DecodeString("000000000000000000010000000000000000000000000000000000000000abba")

	// epoch 0
	tx := &transaction.ApiTransactionResult{
		Epoch: uint32(0),
		Tx: &transaction.Transaction{
			RcvAddr:  contract,
			Data:     []byte("something"),
			GasLimit: 10_000_000,
			GasPrice: 1000000000,
		},
	}

	refundValue := big.NewInt(66350000000000)
	gasUsed, fee := computer.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, uint64(9_933_650), gasUsed)
	require.Equal(t, "9933650000000000", fee.String())

	tx.Epoch = 200
	gasUsed, fee = computer.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, uint64(3_365_000), gasUsed)
	require.Equal(t, "96515000000000", fee.String())
}

func TestFeeComputer_ComputeFeeBasedOnGasUsed(t *testing.T) {
	arguments := ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
		EnableEpochsConfig: config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: 124,
			GasPriceModifierEnableEpoch:    180,
		},
	}

	computer, _ := NewFeeComputer(arguments)

	contract, _ := hex.DecodeString("000000000000000000010000000000000000000000000000000000000000abba")

	// epoch 0
	tx := &transaction.ApiTransactionResult{
		Epoch: uint32(0),
		Tx: &transaction.Transaction{
			RcvAddr:  contract,
			Data:     []byte("something"),
			GasLimit: 10_000_000,
			GasPrice: 1000000000,
		},
	}

	gasUsed := uint64(2_000_00)
	fee := computer.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)
	require.Equal(t, "200000000000000", fee.String())

	tx.Epoch = 200
	fee = computer.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)
	require.Equal(t, "64865000000000", fee.String())
}

func TestFeeComputer_ComputeGasLimit(t *testing.T) {
	arguments := ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
		EnableEpochsConfig: config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: 124,
			GasPriceModifierEnableEpoch:    180,
		},
	}

	computer, _ := NewFeeComputer(arguments)

	contract, _ := hex.DecodeString("000000000000000000010000000000000000000000000000000000000000abba")

	// epoch 0
	tx := &transaction.ApiTransactionResult{
		Epoch: uint32(0),
		Tx: &transaction.Transaction{
			RcvAddr:  contract,
			Data:     []byte("something"),
			GasLimit: 10_000_000,
			GasPrice: 1000000000,
		},
	}

	gasLimit := computer.ComputeGasLimit(tx)
	require.Equal(t, uint64(63_500), gasLimit)

	tx.Epoch = 200
	gasLimit = computer.ComputeGasLimit(tx)
	require.Equal(t, uint64(63_500), gasLimit)
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
	assert.Equal(t, expectedFee, fee.String())
}

func TestFeeComputer_InHighConcurrency(t *testing.T) {
	arguments := ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
		EnableEpochsConfig: config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: 124,
			GasPriceModifierEnableEpoch:    180,
		},
	}

	computer, _ := NewFeeComputer(arguments)

	n := 1000
	wg := sync.WaitGroup{}
	wg.Add(n * 2)

	for i := 0; i < n; i++ {
		go func() {
			checkComputedFee(t, "50000000000000", computer, 0, 80000, 1000000000, "", nil)
			wg.Done()
		}()
	}

	for i := 0; i < n; i++ {
		go func() {
			checkComputedFee(t, "80000000000000", computer, 125, 80000, 1000000000, "", nil)
			wg.Done()
		}()
	}

	wg.Wait()
}
