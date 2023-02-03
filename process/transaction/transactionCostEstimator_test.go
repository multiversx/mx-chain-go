package transaction

import (
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/txsimulator"
	txSimData "github.com/multiversx/mx-chain-go/process/txsimulator/data"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestTransactionCostEstimator_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(
		nil,
		&economicsmocks.EconomicsHandlerStub{},
		&mock.TransactionSimulatorStub{},
		&stateMock.AccountsStub{},
		&mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestTransactionCostEstimator_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(
		&testscommon.TxTypeHandlerMock{},
		nil,
		&mock.TransactionSimulatorStub{},
		&stateMock.AccountsStub{},
		&mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestTransactionCostEstimator_NilTransactionSimulatorShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(
		&testscommon.TxTypeHandlerMock{},
		&economicsmocks.EconomicsHandlerStub{},
		nil,
		&stateMock.AccountsStub{},
		&mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	require.Nil(t, tce)
	require.Equal(t, txsimulator.ErrNilTxSimulatorProcessor, err)
}

func TestTransactionCostEstimator_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(
		&testscommon.TxTypeHandlerMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&mock.TransactionSimulatorStub{},
		&stateMock.AccountsStub{},
		&mock.ShardCoordinatorStub{},
		nil)

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestTransactionCostEstimator_Ok(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(
		&testscommon.TxTypeHandlerMock{},
		&economicsmocks.EconomicsHandlerStub{},
		&mock.TransactionSimulatorStub{},
		&stateMock.AccountsStub{},
		&mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	require.Nil(t, err)
	require.False(t, check.IfNil(tce))
}

func TestComputeTransactionGasLimit_MoveBalance(t *testing.T) {
	t.Parallel()

	consumedGasUnits := uint64(1000)
	tce, _ := NewTransactionCostEstimator(&testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.MoveBalance
		},
	}, &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			return consumedGasUnits
		},
	}, &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
			return &txSimData.SimulationResults{}, nil
		},
	}, &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
		},
	}, &mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
}

func TestComputeTransactionGasLimit_MoveBalanceInvalidNonceShouldStillComputeCost(t *testing.T) {
	t.Parallel()

	simulationErr := errors.New("invalid nonce")
	consumedGasUnits := uint64(1000)
	tce, _ := NewTransactionCostEstimator(&testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.MoveBalance
		},
	}, &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			return consumedGasUnits
		},
	}, &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
			return nil, simulationErr
		},
	}, &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
		},
	}, &mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
}

func TestComputeTransactionGasLimit_BuiltInFunction(t *testing.T) {
	consumedGasUnits := uint64(4000)
	tce, _ := NewTransactionCostEstimator(&testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}, &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
	},
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
				return &txSimData.SimulationResults{
					VMOutput: &vmcommon.VMOutput{
						ReturnCode:   vmcommon.Ok,
						GasRemaining: math.MaxUint64 - 1 - consumedGasUnits,
					},
				}, nil
			},
		}, &stateMock.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
			},
		}, &mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
}

func TestComputeTransactionGasLimit_BuiltInFunctionShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	tce, _ := NewTransactionCostEstimator(&testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}, &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
	},
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
				return nil, localErr
			},
		}, &stateMock.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
			},
		}, &mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, localErr.Error(), cost.ReturnMessage)
}

func TestComputeTransactionGasLimit_NilVMOutput(t *testing.T) {
	tce, _ := NewTransactionCostEstimator(&testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}, &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
	},
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
				return &txSimData.SimulationResults{}, nil
			},
		}, &stateMock.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
			},
		}, &mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, process.ErrNilVMOutput.Error(), cost.ReturnMessage)
}

func TestComputeTransactionGasLimit_RetCodeNotOk(t *testing.T) {
	tce, _ := NewTransactionCostEstimator(&testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}, &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
	},
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
				return &txSimData.SimulationResults{
					VMOutput: &vmcommon.VMOutput{
						ReturnCode: vmcommon.UserError,
					},
				}, nil
			},
		}, &stateMock.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
			},
		}, &mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.True(t, strings.Contains(cost.ReturnMessage, vmcommon.UserError.String()))
}

func TestTransactionCostEstimator_RelayedTxShouldErr(t *testing.T) {
	t.Parallel()

	tce, _ := NewTransactionCostEstimator(
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.RelayedTx, process.RelayedTx
			},
		},
		&economicsmocks.EconomicsHandlerStub{},
		&mock.TransactionSimulatorStub{},
		&stateMock.AccountsStub{},
		&mock.ShardCoordinatorStub{},
		&testscommon.EnableEpochsHandlerStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, "cannot compute cost of the relayed transaction", cost.ReturnMessage)
}

func TestExtractGasRemainedFromMessage(t *testing.T) {
	t.Parallel()

	require.Equal(t, uint64(500000000), extractGasRemainedFromMessage("too much gas provided, gas needed = 10000, gas remained = 500000000", gasRemainedSplitString))
	require.Equal(t, uint64(0), extractGasRemainedFromMessage("", gasRemainedSplitString))
	require.Equal(t, uint64(0), extractGasRemainedFromMessage("too much gas provided, gas needed = 10000, gas remained = wrong", gasRemainedSplitString))
}

func TestExtractGasUsedFromMessage(t *testing.T) {
	t.Parallel()

	require.Equal(t, uint64(500000000), extractGasRemainedFromMessage("too much gas provided, gas needed = 10000, gas used = 500000000", gasUsedSlitString))
	require.Equal(t, uint64(0), extractGasRemainedFromMessage("", gasRemainedSplitString))
	require.Equal(t, uint64(0), extractGasRemainedFromMessage("too much gas provided, gas needed = 10000, gas used = wrong", gasUsedSlitString))
}
