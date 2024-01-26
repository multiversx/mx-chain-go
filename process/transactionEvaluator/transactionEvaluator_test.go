package transactionEvaluator

import (
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsApiTransactionEvaluator {
	return ArgsApiTransactionEvaluator{
		TxTypeHandler:       &testscommon.TxTypeHandlerMock{},
		FeeHandler:          &economicsmocks.EconomicsHandlerStub{},
		TxSimulator:         &mock.TransactionSimulatorStub{},
		Accounts:            &stateMock.AccountsStub{},
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		BlockChain:          &testscommon.ChainHandlerMock{},
	}
}

func TestTransactionEvaluator_NilTxTypeHandler(t *testing.T) {
	t.Parallel()
	args := createArgs()
	args.TxTypeHandler = nil
	tce, err := NewAPITransactionEvaluator(args)

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestTransactionEvaluator_NilBlockChain(t *testing.T) {
	t.Parallel()
	args := createArgs()
	args.BlockChain = nil
	tce, err := NewAPITransactionEvaluator(args)

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilBlockChain, err)
}

func TestTransactionEvaluator_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.FeeHandler = nil
	tce, err := NewAPITransactionEvaluator(args)
	require.Nil(t, tce)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestTransactionEvaluator_NilTransactionSimulatorShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.TxSimulator = nil
	tce, err := NewAPITransactionEvaluator(args)
	require.Nil(t, tce)
	require.Equal(t, ErrNilTxSimulatorProcessor, err)
}

func TestTransactionEvaluator_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.EnableEpochsHandler = nil
	tce, err := NewAPITransactionEvaluator(args)

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestTransactionEvaluator_InvalidEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()
	tce, err := NewAPITransactionEvaluator(args)

	require.Nil(t, tce)
	require.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestTransactionEvaluator_Ok(t *testing.T) {
	t.Parallel()

	args := createArgs()
	tce, err := NewAPITransactionEvaluator(args)

	require.Nil(t, err)
	require.False(t, check.IfNil(tce))
}

func TestComputeTransactionGasLimit_MoveBalance(t *testing.T) {
	t.Parallel()

	consumedGasUnits := uint64(1000)

	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.MoveBalance
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			return consumedGasUnits
		},
	}
	args.TxSimulator = &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(tx *transaction.Transaction, currentHeader data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
			return &txSimData.SimulationResultsWithVMOutput{}, nil
		},
	}
	args.Accounts = &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
		},
	}
	tce, err := NewAPITransactionEvaluator(args)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
}

func TestComputeTransactionGasLimit_MoveBalanceInvalidNonceShouldStillComputeCost(t *testing.T) {
	t.Parallel()

	simulationErr := errors.New("invalid nonce")
	consumedGasUnits := uint64(1000)

	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.MoveBalance
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			return consumedGasUnits
		},
	}
	args.TxSimulator = &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(tx *transaction.Transaction, currentHeader data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
			return nil, simulationErr
		},
	}
	args.Accounts = &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
		},
	}
	tce, _ := NewAPITransactionEvaluator(args)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
}

func TestComputeTransactionGasLimit_BuiltInFunction(t *testing.T) {
	consumedGasUnits := uint64(4000)
	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
	}
	args.TxSimulator = &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(tx *transaction.Transaction, currentHeader data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
			return &txSimData.SimulationResultsWithVMOutput{
				VMOutput: &vmcommon.VMOutput{
					ReturnCode:   vmcommon.Ok,
					GasRemaining: math.MaxUint64 - 1 - consumedGasUnits,
				},
			}, nil
		},
	}
	args.Accounts = &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
		},
	}
	tce, _ := NewAPITransactionEvaluator(args)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
}

func TestComputeTransactionGasLimit_BuiltInFunctionShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
	}
	args.TxSimulator = &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(tx *transaction.Transaction, currentHeader data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
			return nil, localErr
		},
	}
	args.Accounts = &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
		},
	}
	tce, _ := NewAPITransactionEvaluator(args)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, localErr.Error(), cost.ReturnMessage)
}

func TestComputeTransactionGasLimit_NilVMOutput(t *testing.T) {
	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
	}
	args.TxSimulator = &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(tx *transaction.Transaction, currentHeader data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
			return &txSimData.SimulationResultsWithVMOutput{}, nil
		},
	}
	args.Accounts = &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
		},
	}
	tce, err := NewAPITransactionEvaluator(args)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, process.ErrNilVMOutput.Error(), cost.ReturnMessage)
}

func TestComputeTransactionGasLimit_RetCodeNotOk(t *testing.T) {
	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerStub{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return math.MaxUint64
		},
	}
	args.TxSimulator = &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(tx *transaction.Transaction, _ data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
			return &txSimData.SimulationResultsWithVMOutput{
				VMOutput: &vmcommon.VMOutput{
					ReturnCode: vmcommon.UserError,
				},
			}, nil
		},
	}
	args.Accounts = &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.UserAccountStub{Balance: big.NewInt(100000)}, nil
		},
	}

	tce, _ := NewAPITransactionEvaluator(args)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.True(t, strings.Contains(cost.ReturnMessage, vmcommon.UserError.String()))
}

func TestTransactionEvaluator_RelayedTxShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.RelayedTx, process.RelayedTx
		},
	}
	tce, _ := NewAPITransactionEvaluator(args)

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

func TestApiTransactionEvaluator_SimulateTransactionExecution(t *testing.T) {
	t.Parallel()

	called := false
	expectedNonce := uint64(1000)
	args := createArgs()
	args.BlockChain = &testscommon.ChainHandlerMock{}
	_ = args.BlockChain.SetCurrentBlockHeaderAndRootHash(&block.Header{Nonce: expectedNonce}, []byte("test"))

	args.TxSimulator = &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(_ *transaction.Transaction, currentHeader data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
			called = true
			require.Equal(t, expectedNonce, currentHeader.GetNonce())
			return nil, nil
		},
	}

	tce, err := NewAPITransactionEvaluator(args)
	require.Nil(t, err)

	tx := &transaction.Transaction{}

	_, err = tce.SimulateTransactionExecution(tx)
	require.Nil(t, err)
	require.True(t, called)
}

func TestApiTransactionEvaluator_ComputeTransactionGasLimit(t *testing.T) {
	t.Parallel()

	called := false
	expectedNonce := uint64(1000)
	args := createArgs()
	args.BlockChain = &testscommon.ChainHandlerMock{}
	_ = args.BlockChain.SetCurrentBlockHeaderAndRootHash(&block.Header{Nonce: expectedNonce}, []byte("test"))

	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		},
	}
	args.TxSimulator = &mock.TransactionSimulatorStub{
		ProcessTxCalled: func(_ *transaction.Transaction, currentHeader data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
			called = true
			require.Equal(t, expectedNonce, currentHeader.GetNonce())
			return &txSimData.SimulationResultsWithVMOutput{}, nil
		},
	}

	tce, err := NewAPITransactionEvaluator(args)
	require.Nil(t, err)

	tx := &transaction.Transaction{}

	_, err = tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.True(t, called)
}
