package fee

import (
	"encoding/hex"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockFeeComputerArgs() ArgsNewFeeComputer {
	return ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsPenalizedTooMuchGasFlagEnabledInEpochCalled: func(epoch uint32) bool {
				return epoch >= 124
			},
			IsGasPriceModifierFlagEnabledInEpochCalled: func(epoch uint32) bool {
				return epoch >= 180
			},
		},
		TxVersionChecker: &testscommon.TxVersionCheckerStub{},
	}
}

func TestNewFeeComputer(t *testing.T) {
	t.Run("nil builtin function cost handler should error", func(t *testing.T) {
		args := createMockFeeComputerArgs()
		args.BuiltInFunctionsCostHandler = nil
		computer, err := NewFeeComputer(args)
		require.Equal(t, process.ErrNilBuiltInFunctionsCostHandler, err)
		require.Nil(t, computer)
	})
	t.Run("nil tx version checker should error", func(t *testing.T) {
		args := createMockFeeComputerArgs()
		args.TxVersionChecker = nil
		computer, err := NewFeeComputer(args)
		require.Equal(t, process.ErrNilTransactionVersionChecker, err)
		require.Nil(t, computer)
	})
	t.Run("nil enable epochs handler should error", func(t *testing.T) {
		args := createMockFeeComputerArgs()
		args.EnableEpochsHandler = nil
		computer, err := NewFeeComputer(args)
		require.Equal(t, process.ErrNilEnableEpochsHandler, err)
		require.Nil(t, computer)
	})
	t.Run("AllArgumentsProvided", func(t *testing.T) {
		args := createMockFeeComputerArgs()
		computer, err := NewFeeComputer(args)
		require.Nil(t, err)
		require.NotNil(t, computer)
	})
}

func TestFeeComputer_ComputeGasUsedAndFeeBasedOnRefundValue(t *testing.T) {
	args := createMockFeeComputerArgs()
	computer, _ := NewFeeComputer(args)

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
	args := createMockFeeComputerArgs()
	computer, _ := NewFeeComputer(args)

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
	args := createMockFeeComputerArgs()
	computer, _ := NewFeeComputer(args)

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
	args := createMockFeeComputerArgs()
	contract, _ := hex.DecodeString("000000000000000000010000000000000000000000000000000000000000abba")

	computer, _ := NewFeeComputer(args)

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
	args := createMockFeeComputerArgs()
	computer, _ := NewFeeComputer(args)

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

	args := createMockFeeComputerArgs()
	fc, _ = NewFeeComputer(args)
	require.False(t, fc.IsInterfaceNil())
}
