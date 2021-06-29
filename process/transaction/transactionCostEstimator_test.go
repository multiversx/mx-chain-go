package transaction

import (
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestTransactionCostEstimator_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(nil, &mock.FeeHandlerStub{}, &mock.TransactionSimulatorStub{}, &testscommon.AccountsStub{}, &mock.ShardCoordinatorStub{})

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestTransactionCostEstimator_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, nil, &mock.TransactionSimulatorStub{}, &testscommon.AccountsStub{}, &mock.ShardCoordinatorStub{})

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestTransactionCostEstimator_NilTransactionSimulatorShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, &mock.FeeHandlerStub{}, nil, &testscommon.AccountsStub{}, &mock.ShardCoordinatorStub{})

	require.Nil(t, tce)
	require.Equal(t, txsimulator.ErrNilTxSimulatorProcessor, err)
}

func TestTransactionCostEstimator_Ok(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, &mock.FeeHandlerStub{}, &mock.TransactionSimulatorStub{}, &testscommon.AccountsStub{}, &mock.ShardCoordinatorStub{})

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
	}, &testscommon.AccountsStub{
		LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return &mock.UserAccountStub{Balance: big.NewInt(100000)}, nil
		},
	}, &mock.ShardCoordinatorStub{})

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
		}, &testscommon.AccountsStub{
			LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
				return &mock.UserAccountStub{Balance: big.NewInt(100000)}, nil
			},
		}, &mock.ShardCoordinatorStub{})

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
		}, &testscommon.AccountsStub{
			LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
				return &mock.UserAccountStub{Balance: big.NewInt(100000)}, nil
			},
		}, &mock.ShardCoordinatorStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, localErr.Error(), cost.ReturnMessage)
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
		}, &testscommon.AccountsStub{
			LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
				return &mock.UserAccountStub{Balance: big.NewInt(100000)}, nil
			},
		}, &mock.ShardCoordinatorStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, process.ErrNilVMOutput.Error(), cost.ReturnMessage)
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
		}, &testscommon.AccountsStub{
			LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
				return &mock.UserAccountStub{Balance: big.NewInt(100000)}, nil
			},
		}, &mock.ShardCoordinatorStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.True(t, strings.Contains(cost.ReturnMessage, vmcommon.UserError.String()))
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
		&mock.TransactionSimulatorStub{}, &testscommon.AccountsStub{}, &mock.ShardCoordinatorStub{},
	)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, "cannot compute cost of the relayed transaction", cost.ReturnMessage)
}
func TestExtractGasNeededFromMessage(t *testing.T) {
	t.Parallel()

	require.Equal(t, uint64(500000000), extractGasRemainedFromMessage("too much gas provided, gas needed = 10000, gas remained = 500000000"))
	require.Equal(t, uint64(0), extractGasRemainedFromMessage(""))
	require.Equal(t, uint64(0), extractGasRemainedFromMessage("too much gas provided, gas needed = 10000, gas remained = wrong"))
}
