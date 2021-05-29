package transaction

import (
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
)

func TestTransactionCostEstimator_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(nil, &mock.FeeHandlerStub{}, &mock.TransactionSimulatorStub{})

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestTransactionCostEstimator_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, nil, &mock.TransactionSimulatorStub{})

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestTransactionCostEstimator_Ok(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, &mock.FeeHandlerStub{}, &mock.TransactionSimulatorStub{})

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
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return consumedGasUnits
		},
	}, &mock.TransactionSimulatorStub{})

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
	}, &mock.FeeHandlerStub{},
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
				return &transaction.SimulationResults{
					VMOutput: &vmcommon.VMOutput{
						ReturnCode:   vmcommon.Ok,
						GasRemaining: math.MaxUint64 - consumedGasUnits,
					},
				}, nil
			},
		})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
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
		&mock.TransactionSimulatorStub{},
	)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, "cannot compute cost of the relayed transaction", cost.RetMessage)
}
