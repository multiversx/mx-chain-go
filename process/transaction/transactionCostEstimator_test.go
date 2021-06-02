package transaction

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
	"github.com/stretchr/testify/require"
)

func TestTransactionCostEstimator_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(nil, &mock.FeeHandlerStub{}, &mock.TransactionSimulatorStub{}, 0)

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestTransactionCostEstimator_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, nil, &mock.TransactionSimulatorStub{}, 0)

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestTransactionCostEstimator_NilTransactionSimulatorShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, &mock.FeeHandlerStub{}, nil, 0)

	require.Nil(t, tce)
	require.Equal(t, txsimulator.ErrNilTxSimulatorProcessor, err)
}

func TestTransactionCostEstimator_Ok(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, &mock.FeeHandlerStub{}, &mock.TransactionSimulatorStub{}, 0)

	require.Nil(t, err)
	require.False(t, check.IfNil(tce))
}

func TestComputeTransactionGasLimit_MoveBalance(t *testing.T) {
	t.Parallel()

	consumedGasUnits := uint64(1000)
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.MoveBalance
		},
	}, &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return math.MaxUint64
		},
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return consumedGasUnits
		},
	}, &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
			return &transaction.SimulationResults{}, nil
		},
	}, 0)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
}

func TestComputeTransactionGasLimit_BuiltInFunction(t *testing.T) {
	consumedGasUnits := uint64(4000)
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}, &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return math.MaxUint64
		},
	},
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
				return &transaction.SimulationResults{
					VMOutput: &vmcommon.VMOutput{
						ReturnCode:   vmcommon.Ok,
						GasRemaining: math.MaxUint64 - 1 - consumedGasUnits,
					},
				}, nil
			},
		}, 0)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
}

func TestComputeTransactionGasLimit_BuiltInFunctionShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}, &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return math.MaxUint64
		},
	},
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
				return nil, localErr
			},
		}, 0)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, localErr.Error(), cost.RetMessage)
}

func TestComputeTransactionGasLimit_NilVMOutput(t *testing.T) {
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}, &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return math.MaxUint64
		},
	},
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
				return &transaction.SimulationResults{}, nil
			},
		}, 0)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, process.ErrNilVMOutput.Error(), cost.RetMessage)
}

func TestComputeTransactionGasLimit_RetCodeNotOk(t *testing.T) {
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}, &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return math.MaxUint64
		},
	},
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
				return &transaction.SimulationResults{
					VMOutput: &vmcommon.VMOutput{
						ReturnCode: vmcommon.UserError,
					},
				}, nil
			},
		}, 0)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.True(t, strings.Contains(cost.RetMessage, vmcommon.UserError.String()))
}

func TestTransactionCostEstimator_RelayedTxShouldErr(t *testing.T) {
	t.Parallel()

	tce, _ := NewTransactionCostEstimator(
		&mock.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.RelayedTx, process.RelayedTx
			},
		},
		&mock.FeeHandlerStub{},
		&mock.TransactionSimulatorStub{}, 0,
	)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, "cannot compute cost of the relayed transaction", cost.RetMessage)
}
