package processProxy

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/stretchr/testify/assert"
)

func createMockSmartContractProcessorArguments() scrCommon.ArgsNewSmartContractProcessor {
	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[common.BaseOpsAPICost] = make(map[string]uint64)
	gasSchedule[common.BaseOpsAPICost][common.AsyncCallStepField] = 1000
	gasSchedule[common.BaseOpsAPICost][common.AsyncCallbackGasLockField] = 3000
	gasSchedule[common.BuiltInCost] = make(map[string]uint64)
	gasSchedule[common.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 2000

	return scrCommon.ArgsNewSmartContractProcessor{
		VmContainer: &mock.VMContainerMock{},
		ArgsParser:  &mock.ArgumentParserMock{},
		Hasher:      &hashingMocks.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
		AccountsDB: &stateMock.AccountsStub{
			RevertToSnapshotCalled: func(snapshot int) error {
				return nil
			},
		},
		BlockChainHook:   &testscommon.BlockChainHookStub{},
		BuiltInFunctions: builtInFunctions.NewBuiltInFunctionContainer(),
		PubkeyConv:       testscommon.NewPubkeyConverterMock(32),
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(5),
		ScrForwarder:     &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:   &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler:     &mock.FeeAccumulatorStub{},
		TxLogsProcessor:  &mock.TxLogsProcessorStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.0
			},
			ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
				return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
			},
			ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
				return core.SafeMul(tx.GetGasPrice(), gasToUse)
			},
		},
		TxTypeHandler: &testscommon.TxTypeHandlerMock{},
		GasHandler: &testscommon.GasHandlerStub{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		GasSchedule: testscommon.NewGasScheduleNotifierMock(gasSchedule),
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.SCDeployFlag
			},
		},
		EnableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
		WasmVMChangeLocker:  &sync.RWMutex{},
		VMOutputCacher:      txcache.NewDisabledCache(),
	}
}

func TestNewSmartContractProcessorProxy(t *testing.T) {
	t.Parallel()

	t.Run("missing component from arguments should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSmartContractProcessorArguments()
		args.ArgsParser = nil

		proxy, err := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
		assert.True(t, check.IfNil(proxy))
		assert.NotNil(t, err)
		assert.Equal(t, "argument parser is nil", err.Error())
	})
	t.Run("invalid enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSmartContractProcessorArguments()
		args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()

		proxy, err := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
		assert.True(t, check.IfNil(proxy))
		assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
	})
	t.Run("nil epoch notifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSmartContractProcessorArguments()

		proxy, err := NewSmartContractProcessorProxy(args, nil)
		assert.True(t, check.IfNil(proxy))
		assert.Equal(t, process.ErrNilEpochNotifier, err)
	})
	t.Run("should work with sc processor v1", func(t *testing.T) {
		t.Parallel()

		args := createMockSmartContractProcessorArguments()

		proxy, err := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
		assert.False(t, check.IfNil(proxy))
		assert.Nil(t, err)
		assert.Equal(t, "*smartContract.scProcessor", fmt.Sprintf("%T", proxy.processor))
	})
	t.Run("should work with sc processor v2", func(t *testing.T) {
		t.Parallel()

		args := createMockSmartContractProcessorArguments()
		args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.SCProcessorV2Flag)

		proxy, err := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
		assert.False(t, check.IfNil(proxy))
		assert.Nil(t, err)
		assert.Equal(t, "*processorV2.scProcessor", fmt.Sprintf("%T", proxy.processor))
	})
}

func TestSCProcessorProxy_ExecuteSmartContractTransaction(t *testing.T) {
	t.Parallel()

	methodWasCalled := false
	args := createMockSmartContractProcessorArguments()
	proxy, _ := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
	proxy.processor = &testscommon.SCProcessorMock{
		ExecuteSmartContractTransactionCalled: func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
			methodWasCalled = true
			return 0, nil
		},
	}

	code, err := proxy.ExecuteSmartContractTransaction(nil, nil, nil)
	assert.Equal(t, 0, int(code))
	assert.Nil(t, err)
	assert.True(t, methodWasCalled)
}

func TestSCProcessorProxy_ExecuteBuiltInFunction(t *testing.T) {
	t.Parallel()

	methodWasCalled := false
	args := createMockSmartContractProcessorArguments()
	proxy, _ := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
	proxy.processor = &testscommon.SCProcessorMock{
		ExecuteBuiltInFunctionCalled: func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
			methodWasCalled = true
			return 0, nil
		},
	}

	code, err := proxy.ExecuteBuiltInFunction(nil, nil, nil)
	assert.Equal(t, 0, int(code))
	assert.Nil(t, err)
	assert.True(t, methodWasCalled)
}

func TestSCProcessorProxy_DeploySmartContract(t *testing.T) {
	t.Parallel()

	methodWasCalled := false
	args := createMockSmartContractProcessorArguments()
	proxy, _ := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
	proxy.processor = &testscommon.SCProcessorMock{
		DeploySmartContractCalled: func(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error) {
			methodWasCalled = true
			return 0, nil
		},
	}

	code, err := proxy.DeploySmartContract(nil, nil)
	assert.Equal(t, 0, int(code))
	assert.Nil(t, err)
	assert.True(t, methodWasCalled)
}

func TestSCProcessorProxy_ProcessIfError(t *testing.T) {
	t.Parallel()

	methodWasCalled := false
	args := createMockSmartContractProcessorArguments()
	proxy, _ := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
	proxy.processor = &testscommon.SCProcessorMock{
		ProcessIfErrorCalled: func(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
			methodWasCalled = true
			return nil
		},
	}

	err := proxy.ProcessIfError(nil, nil, nil, "", nil, 0, 0)
	assert.Nil(t, err)
	assert.True(t, methodWasCalled)
}

func TestSCProcessorProxy_IsPayable(t *testing.T) {
	t.Parallel()

	methodWasCalled := false
	args := createMockSmartContractProcessorArguments()
	proxy, _ := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
	proxy.processor = &testscommon.SCProcessorMock{
		IsPayableCalled: func(sndAddress, recvAddress []byte) (bool, error) {
			methodWasCalled = true
			return false, nil
		},
	}

	result, err := proxy.IsPayable(nil, nil)
	assert.False(t, result)
	assert.Nil(t, err)
	assert.True(t, methodWasCalled)
}

func TestSCProcessorProxy_ProcessSmartContractResult(t *testing.T) {
	t.Parallel()

	methodWasCalled := false
	args := createMockSmartContractProcessorArguments()
	proxy, _ := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
	proxy.processor = &testscommon.SCProcessorMock{
		ProcessSmartContractResultCalled: func(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
			methodWasCalled = true
			return 0, nil
		},
	}

	result, err := proxy.ProcessSmartContractResult(nil)
	assert.Equal(t, 0, int(result))
	assert.Nil(t, err)
	assert.True(t, methodWasCalled)
}

func TestSCProcessorProxy_ParallelRunOnExportedMethods(t *testing.T) {
	numGoRoutines := 1000

	args := createMockSmartContractProcessorArguments()
	proxy, _ := NewSmartContractProcessorProxy(args, &epochNotifierMock.EpochNotifierStub{})
	proxy.processor = &testscommon.SCProcessorMock{}
	proxy.processorsCache[procV1] = &testscommon.SCProcessorMock{}
	proxy.processorsCache[procV2] = &testscommon.SCProcessorMock{}

	wg := sync.WaitGroup{}
	wg.Add(numGoRoutines)
	for i := 0; i < numGoRoutines; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)

			switch idx {
			case 0:
				proxy.EpochConfirmed(0, 0)
			case 1:
				_, _ = proxy.ExecuteSmartContractTransaction(nil, nil, nil)
			case 2:
				_, _ = proxy.ExecuteBuiltInFunction(nil, nil, nil)
			case 3:
				_, _ = proxy.DeploySmartContract(nil, nil)
			case 4:
				_ = proxy.ProcessIfError(nil, nil, nil, "", nil, 0, 0)
			case 5:
				_, _ = proxy.IsPayable(nil, nil)
			case 6:
				_, _ = proxy.ProcessSmartContractResult(nil)
			default:
				assert.Fail(t, fmt.Sprintf("error in test, got index %d", idx))
			}

			wg.Done()
		}(i % 7)
	}

	wg.Wait()
}
