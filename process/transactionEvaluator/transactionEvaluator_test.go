package transactionEvaluator

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
	stateAccounts "github.com/multiversx/mx-chain-go/state/accounts"
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
		FeeHandler:          &economicsmocks.EconomicsHandlerMock{},
		TxSimulator:         &mock.TransactionSimulatorStub{},
		Accounts:            &stateMock.AccountsStub{},
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		BlockChain:          &testscommon.ChainHandlerMock{},
	}
}

func TestTransactionEvaluator_MetaRequiresPeerAccounts(t *testing.T) {
	t.Parallel()
	args := createArgs()
	args.ShardCoordinator = &mock.ShardCoordinatorStub{
		SelfIdCalled: func() uint32 { return core.MetachainShardId },
	}
	evaluator, err := NewAPITransactionEvaluator(args)
	require.Nil(t, evaluator)
	require.ErrorIs(t, err, process.ErrNilPeerAccountsAdapter)
}

func TestTransactionEvaluator_CleansBothSimulationAdapters(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"simulate", "cost", "scr cost"} {
		for _, fail := range []bool{false, true} {
			t.Run(operation+fmt.Sprint(fail), func(t *testing.T) {
				args := createArgs()
				reads := 0
				original := &stateMock.AccountsStub{
					GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
						reads++
						return stateAccounts.NewPeerAccount(address)
					},
				}
				user, err := NewSimulationAccountsDB(original)
				require.NoError(t, err)
				peer, err := NewSimulationAccountsDB(original)
				require.NoError(t, err)
				args.Accounts = user
				args.PeerAccounts = peer
				args.ShardCoordinator = &mock.ShardCoordinatorStub{
					SelfIdCalled: func() uint32 { return core.MetachainShardId },
				}
				args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
					ComputeTransactionTypeCalled: func(data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
						return process.SCInvoking, process.SCInvoking, false
					},
				}
				run := func() (*txSimData.SimulationResultsWithVMOutput, error) {
					_, errRead := user.GetExistingAccount([]byte("user"))
					require.NoError(t, errRead)
					_, errRead = peer.GetExistingAccount([]byte("peer"))
					require.NoError(t, errRead)
					if fail {
						return nil, errors.New("execution failed")
					}
					return &txSimData.SimulationResultsWithVMOutput{
						VMOutput: &vmcommon.VMOutput{ReturnCode: vmcommon.UserError},
					}, nil
				}
				args.TxSimulator = &mock.TransactionSimulatorStub{
					ProcessTxCalled: func(*transaction.Transaction, data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
						return run()
					},
					ProcessSCRCalled: func(*smartContractResult.SmartContractResult, data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
						return run()
					},
				}
				evaluator, err := NewAPITransactionEvaluator(args)
				require.NoError(t, err)
				for iteration := 1; iteration <= 2; iteration++ {
					switch operation {
					case "simulate":
						_, err = evaluator.SimulateTransactionExecution(&transaction.Transaction{})
						require.Equal(t, fail, err != nil)
					case "cost":
						_, err = evaluator.ComputeTransactionGasLimit(&transaction.Transaction{GasLimit: 100})
						require.NoError(t, err)
					case "scr cost":
						_, err = evaluator.SimulateSCRExecutionCost(&smartContractResult.SmartContractResult{})
						require.NoError(t, err)
					}
					require.Empty(t, user.cachedAccounts)
					require.Empty(t, peer.cachedAccounts)
					require.Equal(t, 2*iteration, reads)
				}
			})
		}
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
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
			return process.MoveBalance, process.MoveBalance, false
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerMock{
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
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
			return process.MoveBalance, process.MoveBalance, false
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerMock{
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

func TestSimulateSCRCost(t *testing.T) {
	t.Parallel()

	consumedGasUnits := uint64(4001)
	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall, false
		},
	}

	args.TxSimulator = &mock.TransactionSimulatorStub{
		ProcessSCRCalled: func(tx *smartContractResult.SmartContractResult, currentHeader data.HeaderHandler) (*txSimData.SimulationResultsWithVMOutput, error) {
			return &txSimData.SimulationResultsWithVMOutput{
				VMOutput: &vmcommon.VMOutput{
					ReturnCode:   vmcommon.Ok,
					GasRemaining: 1000,
				},
			}, nil
		},
	}

	tce, _ := NewAPITransactionEvaluator(args)

	scr := &smartContractResult.SmartContractResult{
		GasLimit: 5000,
	}
	cost, err := tce.SimulateSCRExecutionCost(scr)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost.GasUnits)
}

func TestComputeTransactionGasLimit_BuiltInFunction(t *testing.T) {
	consumedGasUnits := uint64(4000)
	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall, false
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerMock{
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
	require.Equal(t, consumedGasUnits+1, cost.GasUnits)
}

func TestComputeTransactionGasLimit_BuiltInFunctionShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	args := createArgs()
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall, false
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerMock{
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
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall, false
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerMock{
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
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall, false
		},
	}
	args.FeeHandler = &economicsmocks.EconomicsHandlerMock{
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
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
			return process.RelayedTx, process.RelayedTx, false
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
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
			return process.SCInvoking, process.SCInvoking, false
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

func TestApiTransactionEvaluator_GetCurrentHeader(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.BlockChain = &testscommon.ChainHandlerMock{}
	_ = args.BlockChain.SetGenesisHeader(&block.Header{Nonce: 0})

	tce, err := NewAPITransactionEvaluator(args)
	require.Nil(t, err)

	currentHeader := tce.getCurrentBlockHeader()
	require.Equal(t, uint64(0), currentHeader.GetNonce())

	expectedNonce := uint64(100)
	_ = args.BlockChain.SetCurrentBlockHeaderAndRootHash(&block.Header{Nonce: expectedNonce}, []byte("root"))

	currentHeader = tce.getCurrentBlockHeader()
	require.Equal(t, expectedNonce, currentHeader.GetNonce())
}
