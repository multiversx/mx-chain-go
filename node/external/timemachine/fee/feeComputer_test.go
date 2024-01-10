package fee

import (
	"encoding/hex"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createEconomicsData() process.EconomicsDataHandler {
	economicsConfig := testscommon.GetEconomicsConfig()
	economicsData, _ := economics.NewEconomicsData(economics.ArgsNewEconomicsData{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		Economics:                   &economicsConfig,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				if flag == common.PenalizedTooMuchGasFlag {
					return epoch >= 124
				}
				if flag == common.GasPriceModifierFlag {
					return epoch >= 180
				}
				return false
			},
		},
		TxVersionChecker: &testscommon.TxVersionCheckerStub{},
		EpochNotifier:    &epochNotifier.EpochNotifierStub{},
	})

	return economicsData
}

func TestNewFeeComputer(t *testing.T) {
	t.Parallel()

	t.Run("nil economics data should error", func(t *testing.T) {
		t.Parallel()

		computer, err := NewFeeComputer(nil)
		require.Equal(t, errNilEconomicsData, err)
		require.Nil(t, computer)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		computer, err := NewFeeComputer(createEconomicsData())
		require.Nil(t, err)
		require.NotNil(t, computer)
	})
}

func TestFeeComputer_ComputeGasUsedAndFeeBasedOnRefundValue(t *testing.T) {
	computer, _ := NewFeeComputer(createEconomicsData())

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
	computer, _ := NewFeeComputer(createEconomicsData())

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
	computer, _ := NewFeeComputer(createEconomicsData())

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
	contract, _ := hex.DecodeString("000000000000000000010000000000000000000000000000000000000000abba")

	computer, _ := NewFeeComputer(createEconomicsData())

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
	computer, _ := NewFeeComputer(createEconomicsData())

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

func TestFeeComputer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var fc *feeComputer
	require.True(t, fc.IsInterfaceNil())

	fc, _ = NewFeeComputer(createEconomicsData())
	require.False(t, fc.IsInterfaceNil())
}
