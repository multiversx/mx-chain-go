package processorV2

import (
	"bytes"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/stretchr/testify/require"
)

func createRealSmartContractProcessorArguments() scrCommon.ArgsNewSmartContractProcessor {
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
		PubkeyConv:       createMockPubkeyConverter(),
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
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsSCDeployFlagEnabledField: true,
		},
		GasSchedule:        testscommon.NewGasScheduleNotifierMock(gasSchedule),
		WasmVMChangeLocker: &sync.RWMutex{},
		VMOutputCacher:     txcache.NewDisabledCache(),
	}
}

func TestSovereignSCProcessor_ProcessSmartContractResultExecuteSCIfMetaAndBuiltIn(t *testing.T) {
	t.Parallel()

	scAddress := []byte("000000000001234567890123456789012")
	dstScAddress := createAccount(scAddress)
	dstScAddress.SetCode([]byte("code"))
	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			if bytes.Equal(scAddress, address) {
				return dstScAddress, nil
			}
			return nil, nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(scAddress, address) {
			return shardCoordinator.SelfId()
		}
		return 0
	}
	shardCoordinator.CurrentShard = core.MetachainShardId

	executeCalled := false
	arguments := createRealSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	arguments.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return &mock.VMExecutionHandlerStub{
				RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
					executeCalled = true
					return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
				},
			}, nil
		},
	}
	arguments.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub

	sc, err := NewSmartContractProcessorV2(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		SndAddr: []byte("snd addr"),
		RcvAddr: scAddress,
		Data:    []byte("code@06"),
		Value:   big.NewInt(15),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.True(t, executeCalled)

	executeCalled = false
	enableEpochsHandlerStub.IsBuiltInFunctionOnMetaFlagEnabledField = true
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.False(t, executeCalled)
}
