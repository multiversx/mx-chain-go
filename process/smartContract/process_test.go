package smartContract

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/testscommon/vmcommonMocks"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const setGuardianCost = 250000

func generateEmptyByteSlice(size int) []byte {
	buff := make([]byte, size)

	return buff
}

func createMockPubkeyConverter() *testscommon.PubkeyConverterMock {
	return testscommon.NewPubkeyConverterMock(32)
}

func createAccount(address []byte) state.UserAccountHandler {
	acc, _ := accounts.NewUserAccount(address, &trie.DataTrieTrackerStub{}, &trie.TrieLeafParserStub{})
	return acc
}

func createAccounts(tx data.TransactionHandler) (state.UserAccountHandler, state.UserAccountHandler) {
	acntSrc := createAccount(tx.GetSndAddr())
	_ = acntSrc.AddToBalance(tx.GetValue())
	totalFee := big.NewInt(0)
	totalFee = totalFee.Mul(big.NewInt(int64(tx.GetGasLimit())), big.NewInt(int64(tx.GetGasPrice())))
	_ = acntSrc.AddToBalance(totalFee)

	acntDst := createAccount(tx.GetRcvAddr())

	return acntSrc, acntDst
}

func createMockSmartContractProcessorArguments() scrCommon.ArgsNewSmartContractProcessor {
	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[common.BaseOpsAPICost] = make(map[string]uint64)
	gasSchedule[common.BaseOpsAPICost][common.AsyncCallStepField] = 1000
	gasSchedule[common.BaseOpsAPICost][common.AsyncCallbackGasLockField] = 3000
	gasSchedule[common.BuiltInCost] = make(map[string]uint64)
	gasSchedule[common.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 2000
	gasSchedule[common.BuiltInCost][core.BuiltInFunctionSetGuardian] = setGuardianCost

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
		GasSchedule:         testscommon.NewGasScheduleNotifierMock(gasSchedule),
		EnableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.SCDeployFlag
			},
		},
		WasmVMChangeLocker: &sync.RWMutex{},
		VMOutputCacher:     txcache.NewDisabledCache(),
	}
}

// ===================== TestNewSmartContractProcessor =====================
func TestNewSmartContractProcessorNilVM(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Nil(t, sc)
	require.Equal(t, process.ErrNoVM, err)
}

func TestNewSmartContractProcessorNilVMOutputCacher(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.VMOutputCacher = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Nil(t, sc)
	require.Equal(t, process.ErrNilCacher, err)
}

func TestNewSmartContractProcessorNilBuiltInFunctions(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.BuiltInFunctions = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Nil(t, sc)
	require.Equal(t, process.ErrNilBuiltInFunction, err)
}

func TestNewSmartContractProcessorNilArgsParser(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilArgumentParser, err)
}

func TestNewSmartContractProcessorNilHasher(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.Hasher = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilHasher, err)
}

func TestNewSmartContractProcessorNilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.Marshalizer = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewSmartContractProcessorNilAccountsDB(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewSmartContractProcessorNilAdrConv(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.PubkeyConv = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewSmartContractProcessorNilShardCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ShardCoordinator = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewSmartContractProcessorNilFakeAccountsHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.BlockChainHook = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilTemporaryAccountsHandler, err)
}

func TestNewSmartContractProcessor_NilIntermediateMock(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ScrForwarder = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilIntermediateTransactionHandler, err)
}

func TestNewSmartContractProcessor_ErrNilUnsignedTxHandlerMock(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.TxFeeHandler = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilUnsignedTxHandler, err)
}

func TestNewSmartContractProcessor_ErrNilGasHandlerMock(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.GasHandler = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilGasHandler, err)
}

func TestNewSmartContractProcessor_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.EnableEpochsHandler = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewSmartContractProcessor_InvalidEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestNewSmartContractProcessor_NilEconomicsFeeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewSmartContractProcessor_NilTxTypeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.TxTypeHandler = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestNewSmartContractProcessor_NilGasScheduleShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.GasSchedule = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilGasSchedule, err)
}

func TestNewSmartContractProcessor_NilLatestGasScheduleShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.GasSchedule = testscommon.NewGasScheduleNotifierMock(nil)
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilGasSchedule, err)
}

func TestNewSmartContractProcessor_NilTxLogsProcessorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.TxLogsProcessor = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilTxLogsProcessor, err)
}

func TestNewSmartContractProcessor_NilBadTxForwarderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.BadTxForwarder = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilBadTxHandler, err)
}

func TestNewSmartContractProcessor_NilLockerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.WasmVMChangeLocker = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilLocker, err)
}

func TestNewSmartContractProcessor_ShouldRegisterNotifiers(t *testing.T) {
	t.Parallel()

	gasScheduleRegisterCalled := false

	arguments := createMockSmartContractProcessorArguments()
	gasSchedule := testscommon.NewGasScheduleNotifierMock(make(map[string]map[string]uint64))
	gasSchedule.RegisterNotifyHandlerCalled = func(handler core.GasScheduleSubscribeHandler) {
		gasScheduleRegisterCalled = true
	}
	arguments.GasSchedule = gasSchedule

	_, _ = NewSmartContractProcessor(arguments)

	require.True(t, gasScheduleRegisterCalled)
}

func TestNewSmartContractProcessor(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)
	require.False(t, sc.IsInterfaceNil())
}

func TestNewSmartContractProcessorVerifyAllMembers(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessor(arguments)

	assert.Equal(t, arguments.VmContainer, sc.vmContainer)
	assert.Equal(t, arguments.ArgsParser, sc.argsParser)
	assert.Equal(t, arguments.Hasher, sc.hasher)
	assert.Equal(t, arguments.AccountsDB, sc.accounts)
	assert.Equal(t, arguments.BlockChainHook, sc.blockChainHook)
	assert.Equal(t, arguments.PubkeyConv, sc.pubkeyConv)
	assert.Equal(t, arguments.ShardCoordinator, sc.shardCoordinator)
	assert.Equal(t, arguments.ScrForwarder, sc.scrForwarder)
	assert.Equal(t, arguments.TxFeeHandler, sc.txFeeHandler)
	assert.Equal(t, arguments.EconomicsFee, sc.economicsFee)
	assert.Equal(t, arguments.TxTypeHandler, sc.txTypeHandler)
	assert.Equal(t, arguments.TxLogsProcessor, sc.txLogsProcessor)
	assert.Equal(t, arguments.BadTxForwarder, sc.badTxForwarder)
}

// ===================== TestGasScheduleChange =====================

func TestGasScheduleChangeNoApiCostShouldNotChange(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessor(arguments)

	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[common.BuiltInCost] = nil

	sc.GasScheduleChange(gasSchedule)
	require.Equal(t, sc.builtInGasCosts[core.BuiltInFunctionESDTNFTTransfer], uint64(0))

	gasSchedule[common.BuiltInCost] = make(map[string]uint64)
	gasSchedule[common.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 2000
	sc.GasScheduleChange(gasSchedule)
	require.Equal(t, sc.builtInGasCosts[core.BuiltInFunctionESDTTransfer], uint64(2000))
}

func TestGasScheduleChangeShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessor(arguments)

	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[common.BuiltInCost] = make(map[string]uint64)
	gasSchedule[common.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 20

	sc.GasScheduleChange(gasSchedule)

	require.Equal(t, sc.builtInGasCosts[core.BuiltInFunctionESDTTransfer], uint64(20))
}

// ===================== TestDeploySmartContract =====================

func TestScProcessor_DeploySmartContractBadParse(t *testing.T) {
	t.Parallel()

	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = &mock.VMContainerMock{}
	arguments.ArgsParser = argParser

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	parseError := fmt.Errorf("fooError")
	argParser.ParseDeployDataCalled = func(data string) (*parsers.DeployArgs, error) {
		return nil, parseError
	}

	arguments.AccountsDB = &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return acntSrc, nil
		},
	}

	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	returnCode, _ := sc.DeploySmartContract(tx, acntSrc)

	tsc := NewTestScProcessor(sc)

	scrs := tsc.GetAllSCRs()
	expectedError := "@" + hex.EncodeToString([]byte(parseError.Error()))
	require.Equal(t, expectedError, string(scrs[0].GetData()))
	require.Equal(t, vmcommon.UserError, returnCode)
	require.Equal(t, uint64(1), acntSrc.GetNonce())
	require.True(t, acntSrc.GetBalance().Cmp(tx.Value) == 0)
}

func TestScProcessor_DeploySmartContractRunError(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &stateMock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}

	createError := fmt.Errorf("fooError")
	vm.RunSmartContractCreateCalled = func(input *vmcommon.ContractCreateInput) (output *vmcommon.VMOutput, e error) {
		return nil, createError
	}

	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	_, _ = sc.DeploySmartContract(tx, acntSrc)
	tsc := NewTestScProcessor(sc)
	scrs := tsc.GetAllSCRs()
	expectedError := "@" + hex.EncodeToString([]byte(createError.Error()))
	require.Equal(t, expectedError, string(scrs[0].GetData()))
}

func TestScProcessor_DeploySmartContractDisabled(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &stateMock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.BuiltInFunctionsFlag
		},
	}

	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	returnCode, err := sc.DeploySmartContract(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, returnCode)
}

func TestScProcessor_BuiltInCallSmartContractDisabled(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &stateMock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	funcName := "builtIn"
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte(funcName + "@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	_, err = sc.ExecuteBuiltInFunction(tx, acntSrc, nil)
	require.Equal(t, process.ErrFailedTransaction, err)
}

func TestScProcessor_BuiltInCallSmartContractSenderFailed(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &stateMock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	funcName := "builtIn"
	localError := errors.New("failed built in call")
	arguments.BlockChainHook = &testscommon.BlockChainHookStub{
		ProcessBuiltInFunctionCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			return nil, localError
		},
	}

	scrAdded := false
	badTxAdded := false
	arguments.BadTxForwarder = &mock.IntermediateTransactionHandlerMock{
		AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
			badTxAdded = true
			return nil
		},
	}
	arguments.ScrForwarder = &mock.IntermediateTransactionHandlerMock{
		AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
			scrAdded = true
			return nil
		},
	}

	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte(funcName + "@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	_, err = sc.ExecuteBuiltInFunction(tx, acntSrc, nil)
	require.Equal(t, process.ErrFailedTransaction, err)
	require.True(t, scrAdded)
	require.True(t, badTxAdded)

	_, err = sc.ExecuteSmartContractTransaction(tx, nil, acntSrc)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteBuiltInFunctionSCResultCallSelfShard(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	accountState := &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	arguments.AccountsDB = accountState
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub
	funcName := "builtIn"
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &smartContractResult.SmartContractResult{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, arguments.PubkeyConv.Len())
	tx.Data = []byte(funcName + "@@@0500@0000")
	tx.Value = big.NewInt(0)
	tx.CallType = vmData.AsynchronousCallBack
	acntSrc, actDst := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}
	accountState.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(tx.SndAddr, address) {
			return acntSrc, nil
		}
		if bytes.Equal(tx.RcvAddr, address) {
			return actDst, nil
		}
		return nil, nil
	}

	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.BuiltInFunctionsFlag
	}
	retCode, err := sc.ExecuteBuiltInFunction(tx, acntSrc, actDst)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteBuiltInFunctionSCResultCallSelfShardCannotSaveLog(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	accountState := &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}

	called := false
	localErr := errors.New("local err")
	arguments.TxLogsProcessor = &mock.TxLogsProcessorStub{
		SaveLogCalled: func(_ []byte, tx data.TransactionHandler, _ []*vmcommon.LogEntry) error {
			called = true
			return localErr
		},
	}

	arguments.AccountsDB = accountState
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub
	funcName := "builtIn"
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &smartContractResult.SmartContractResult{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, arguments.PubkeyConv.Len())
	tx.Data = []byte(funcName + "@@@0500@0000")
	tx.Value = big.NewInt(0)
	tx.CallType = vmData.AsynchronousCallBack
	acntSrc, actDst := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}
	accountState.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(tx.SndAddr, address) {
			return acntSrc, nil
		}
		if bytes.Equal(tx.RcvAddr, address) {
			return actDst, nil
		}
		return nil, nil
	}

	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.BuiltInFunctionsFlag
	}
	retCode, err := sc.ExecuteBuiltInFunction(tx, acntSrc, actDst)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
	require.True(t, called)
}

func TestScProcessor_ExecuteBuiltInFunction(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	accountState := &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	arguments.AccountsDB = accountState
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub
	funcName := "builtIn"
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte(funcName + "@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}
	accountState.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		return acntSrc, nil
	}

	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.BuiltInFunctionsFlag
	}
	retCode, err := sc.ExecuteBuiltInFunction(tx, acntSrc, nil)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteBuiltInFunctionSCRTooBig(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	accountState := &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	arguments.AccountsDB = accountState
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.BuiltInFunctionsFlag
		},
	}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub
	funcName := "builtIn"
	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte(funcName + "@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)
	userAcc, _ := acntSrc.(vmcommon.UserAccountHandler)

	builtInFunc := &mock.BuiltInFunctionStub{ProcessBuiltinFunctionCalled: func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
		longData := bytes.Repeat([]byte{1}, 1<<21)

		return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok, ReturnData: [][]byte{longData}}, nil
	}}
	_ = arguments.BuiltInFunctions.Add(funcName, builtInFunc)
	arguments.BlockChainHook = &testscommon.BlockChainHookStub{
		ProcessBuiltInFunctionCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			return builtInFunc.ProcessBuiltinFunction(userAcc, nil, input)
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}
	accountState.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		return acntSrc, nil
	}

	retCode, err := sc.ExecuteBuiltInFunction(tx, acntSrc, nil)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_ = acntSrc.AddToBalance(big.NewInt(100))
	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.BuiltInFunctionsFlag || flag == common.SCRSizeInvariantOnBuiltInResultFlag || flag == common.SCRSizeInvariantCheckFlag
	}
	retCode, err = sc.ExecuteBuiltInFunction(tx, acntSrc, nil)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)
}

func TestScProcessor_DeploySmartContractWrongTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err := sc.DeploySmartContract(tx, acntSrc)
	require.Equal(t, process.ErrWrongTransaction, err)
}

func TestScProcessor_DeploySmartContractNilTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err := sc.DeploySmartContract(nil, acntSrc)
	require.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_DeploySmartContractNotEmptyDestinationAddress(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.PubkeyConv = testscommon.NewPubkeyConverterMock(3)
	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err := sc.DeploySmartContract(tx, acntSrc)
	require.Equal(t, process.ErrWrongTransaction, err)
}

func TestScProcessor_DeploySmartContractCalculateHashFails(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser

	arguments.Marshalizer = &mock.MarshalizerMock{
		Fail: true,
	}

	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err = sc.DeploySmartContract(tx, acntSrc)
	require.NotNil(t, err)
	require.Equal(t, "MarshalizerMock generic error", err.Error())
}

func TestScProcessor_DeploySmartContractEconomicsFeeValidateFails(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser

	arguments.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return expectedError
		},
	}

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err := sc.DeploySmartContract(tx, acntSrc)
	require.Equal(t, expectedError, err)
}

func TestScProcessor_DeploySmartContractEconomicsFeeSaveAccountsFails(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = &stateMock.AccountsStub{
		SaveAccountCalled: func(account vmcommon.AccountHandler) error {
			return expectedError
		},
	}

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err := sc.DeploySmartContract(tx, acntSrc)
	require.Equal(t, expectedError, err)
}

func TestScProcessor_DeploySmartContractEconomicsWithFlagPenalizeTooMuchGasEnabled(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err := sc.DeploySmartContract(tx, acntSrc)
	require.Equal(t, nil, err)
}

func TestScProcessor_DeploySmartContractVmContainerGetFails(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	argParser := NewArgumentParser()
	vm := &mock.VMContainerMock{}
	vm.GetCalled = func(key []byte) (vmcommon.VMExecutionHandler, error) {
		return nil, expectedError
	}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	errCode, err := sc.DeploySmartContract(tx, acntSrc)
	// TODO: check why nil error here
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, errCode)
}

func TestScProcessor_DeploySmartContractVmExecuteCreateSCFails(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	argParser := NewArgumentParser()
	vm := &mock.VMContainerMock{}
	vmExecutor := &mock.VMExecutionHandlerStub{
		RunSmartContractCreateCalled: func(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
			return nil, expectedError
		},
	}
	vm.GetCalled = func(key []byte) (vmcommon.VMExecutionHandler, error) {
		return vmExecutor, nil
	}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	errCode, err := sc.DeploySmartContract(tx, acntSrc)
	// TODO: check why nil error here
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, errCode)
}

func TestScProcessor_DeploySmartContractVmExecuteCreateSCVMOutputNil(t *testing.T) {
	t.Parallel()

	argParser := NewArgumentParser()
	vm := &mock.VMContainerMock{}
	vmExecutor := &mock.VMExecutionHandlerStub{
		RunSmartContractCreateCalled: func(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
			return nil, nil
		},
	}
	vm.GetCalled = func(key []byte) (vmcommon.VMExecutionHandler, error) {
		return vmExecutor, nil
	}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	errCode, err := sc.DeploySmartContract(tx, acntSrc)
	// TODO: check why nil error here
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, errCode)
}

func TestScProcessor_DeploySmartContractVmOutputReturnCodeNotOk(t *testing.T) {
	t.Parallel()

	argParser := NewArgumentParser()
	vm := &mock.VMContainerMock{}
	vmExecutor := &mock.VMExecutionHandlerStub{
		RunSmartContractCreateCalled: func(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
			return &vmcommon.VMOutput{ReturnCode: vmcommon.CallStackOverFlow, ReturnMessage: "returnMessage"}, nil
		},
	}
	vm.GetCalled = func(key []byte) (vmcommon.VMExecutionHandler, error) {
		return vmExecutor, nil
	}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	errCode, err := sc.DeploySmartContract(tx, acntSrc)
	require.Nil(t, err)
	// TODO: check why userError here and not executionException? The run failed with callStackOverflow
	require.Equal(t, vmcommon.UserError, errCode)
}

func TestScProcessor_DeploySmartContractUpdateDeveloperRewardsFails(t *testing.T) {
	t.Parallel()

	nrCalls := 0

	vm := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	accntState := &stateMock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	economicsFee := &economicsmocks.EconomicsHandlerStub{
		DeveloperPercentageCalled: func() float64 {
			return 0.0
		},
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
		},
		ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			return core.SafeMul(tx.GetGasPrice(), gasToUse)
		},
		// second call is in rewards and will fail
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			if nrCalls == 0 {
				nrCalls++
				return 0
			}
			return 1000000
		},
	}
	arguments.EconomicsFee = economicsFee
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSrc, nil
	}

	errCode, err := sc.DeploySmartContract(tx, acntSrc)
	require.NotNil(t, err)
	require.Equal(t, core.ErrSubtractionOverflow, err)
	require.Equal(t, vmcommon.Ok, errCode)
}

func TestScProcessor_DeploySmartContractVmOutputGasRemainingNotOk(t *testing.T) {
	t.Parallel()

	gasRemainingOnVmOutput := uint64(100)
	gasLimit := uint64(50)

	argParser := NewArgumentParser()
	vm := &mock.VMContainerMock{}
	vmExecutor := &mock.VMExecutionHandlerStub{
		RunSmartContractCreateCalled: func(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
			return &vmcommon.VMOutput{
				GasRemaining: gasRemainingOnVmOutput,
				ReturnCode:   vmcommon.Ok}, nil
		},
	}
	vm.GetCalled = func(key []byte) (vmcommon.VMExecutionHandler, error) {
		return vmExecutor, nil
	}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)
	tx.GasLimit = gasLimit
	acntSrc, _ := createAccounts(tx)

	errCode, err := sc.DeploySmartContract(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, vmcommon.ExecutionFailed, errCode)
}

func TestScProcessor_DeploySmartContractVmOutputProcessDeleteAccountsFails(t *testing.T) {
	t.Parallel()

	argParser := NewArgumentParser()
	vm := &mock.VMContainerMock{}
	vmExecutor := &mock.VMExecutionHandlerStub{
		RunSmartContractCreateCalled: func(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
			return &vmcommon.VMOutput{
				DeletedAccounts: [][]byte{[]byte("deletedAccount")},
				ReturnCode:      vmcommon.Ok,
			}, nil
		},
	}
	vm.GetCalled = func(key []byte) (vmcommon.VMExecutionHandler, error) {
		return vmExecutor, nil
	}
	accountsDb := &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
		RemoveAccountCalled: func(address []byte) error {
			return errors.New("remove failed")
		},
	}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDb

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accountsDb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		return acntSrc, nil
	}

	errCode, err := sc.DeploySmartContract(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, vmcommon.ExecutionFailed, errCode)
}

func TestScProcessor_DeploySmartContractReloadAccountFails(t *testing.T) {
	t.Parallel()

	calledNr := 0

	vm := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	accntState := &stateMock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		if calledNr == 0 {
			calledNr++
			return acntSrc, nil
		}
		return nil, process.ErrWrongTypeAssertion
	}

	errCode, err := sc.DeploySmartContract(tx, acntSrc)

	require.Equal(t, process.ErrWrongTypeAssertion, err)
	require.Equal(t, vmcommon.Ok, errCode)
}

func TestScProcessor_DeploySmartContractAddIntermediateTxFails(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")

	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = argParser
	arguments.ScrForwarder = &mock.IntermediateTransactionHandlerMock{
		AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
			return expectedError
		},
	}
	accntState := &stateMock.AccountsStub{}
	arguments.AccountsDB = accntState
	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSrc, nil
	}

	returnCode, err := sc.DeploySmartContract(tx, acntSrc)
	// TODO: check OK?
	require.Equal(t, vmcommon.Ok, returnCode)
	require.Equal(t, expectedError, err)
}

func TestScProcessor_DeploySmartContractComputeRewardsFails(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")

	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = argParser
	arguments.ScrForwarder = &mock.IntermediateTransactionHandlerMock{
		AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
			return expectedError
		},
	}
	accntState := &stateMock.AccountsStub{}
	arguments.AccountsDB = accntState
	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSrc, nil
	}

	returnCode, err := sc.DeploySmartContract(tx, acntSrc)
	// TODO: check OK?
	require.Equal(t, vmcommon.Ok, returnCode)
	require.Equal(t, expectedError, err)
}

func TestScProcessor_DeploySmartContract(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	accntState := &stateMock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSrc, nil
	}

	returnCode, err := sc.DeploySmartContract(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
}

func TestScProcessor_ExecuteSmartContractTransactionNilTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	_, err = sc.ExecuteSmartContractTransaction(nil, acntSrc, acntDst)
	require.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_ExecuteSmartContractTransactionNilAccount(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, nil)
	require.Equal(t, process.ErrNilSCDestAccount, err)

	acntSrc, acntDst := createAccounts(tx)
	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.Nil(t, err)

	acntSrc, _ = createAccounts(tx)
	acntDst = nil
	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.Equal(t, process.ErrNilSCDestAccount, err)
}

func TestScProcessor_ExecuteSmartContractTransactionBadParser(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	acntDst.SetCode([]byte("code"))
	tmpError := errors.New("error")
	called := false
	argParser.ParseCallDataCalled = func(data string) (string, [][]byte, error) {
		called = true
		return "", nil, tmpError
	}
	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.True(t, called)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteSmartContractTransactionVMRunError(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	acntDst.SetCode([]byte("code"))
	tmpError := errors.New("error")
	vm := &mock.VMExecutionHandlerStub{}
	called := false
	vm.RunSmartContractCallCalled = func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
		called = true
		return nil, tmpError
	}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.True(t, called)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteSmartContractTransactionGasConsumedChecksError(t *testing.T) {
	t.Parallel()

	gasRemainingOnVmOutput := uint64(100)

	argParser := NewArgumentParser()
	vm := &mock.VMContainerMock{}
	vmExecutor := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			return &vmcommon.VMOutput{
				GasRemaining: gasRemainingOnVmOutput,
				ReturnCode:   vmcommon.Ok}, nil
		},
	}
	vm.GetCalled = func(key []byte) (vmcommon.VMExecutionHandler, error) {
		return vmExecutor, nil
	}
	accntState := &stateMock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	arguments.VMOutputCacher, _ = storageunit.NewCache(storageunit.CacheConfig{
		Type:     storageunit.LRUCache,
		Capacity: 10000,
	})

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(0)
	acntSrc, acntDst := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSrc, nil
	}
	accntState.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}

	acntDst.SetCode([]byte("code"))
	errCode, err := sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.Nil(t, err)
	require.Equal(t, vmcommon.ExecutionFailed, errCode)
}

func TestScProcessor_ExecuteSmartContractTransactionVmOutputError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")

	argParser := NewArgumentParser()
	vm := &mock.VMContainerMock{}
	vmExecutor := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			return &vmcommon.VMOutput{
				GasRemaining:    0,
				DeletedAccounts: [][]byte{[]byte("DEL")},
				ReturnCode:      vmcommon.Ok}, nil
		},
	}
	vm.GetCalled = func(key []byte) (vmcommon.VMExecutionHandler, error) {
		return vmExecutor, nil
	}
	accntState := &stateMock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(0)
	acntSrc, acntDst := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSrc, nil
	}
	accntState.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}
	accntState.RemoveAccountCalled = func(address []byte) error {
		return expectedError
	}

	acntDst.SetCode([]byte("code"))
	errCode, err := sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.Nil(t, err)
	require.Equal(t, vmcommon.ExecutionFailed, errCode)
}

func TestScProcessor_ExecuteSmartContractTransaction(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &stateMock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(0)
	acntSrc, acntDst := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSrc, nil
	}

	acntDst.SetCode([]byte("code"))
	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteSmartContractTransactionSaveLogCalled(t *testing.T) {
	t.Parallel()

	slCalled := false

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &stateMock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	arguments.TxLogsProcessor = &mock.TxLogsProcessorStub{
		SaveLogCalled: func(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error {
			slCalled = true
			return nil
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(0)
	acntSrc, acntDst := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSrc, nil
	}

	acntDst.SetCode([]byte("code"))
	_, _ = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.True(t, slCalled)
}

func TestScProcessor_CreateVMCallInputWrongCode(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	tmpError := errors.New("error")
	argParser.ParseCallDataCalled = func(data string) (string, [][]byte, error) {
		return "", nil, tmpError
	}
	input, err := sc.createVMCallInput(tx, []byte{}, false)
	require.Nil(t, input)
	require.Equal(t, tmpError, err)
}

func TestScProcessor_CreateVMCallInput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	input, err := sc.createVMCallInput(tx, []byte{}, false)
	require.NotNil(t, input)
	require.Nil(t, err)
}

func TestScProcessor_CreateVMDeployBadCode(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Data = nil
	tx.Value = big.NewInt(0)

	badCodeError := errors.New("fooError")
	argParser.ParseDeployDataCalled = func(data string) (*parsers.DeployArgs, error) {
		return nil, badCodeError
	}

	input, vmType, err := sc.createVMDeployInput(tx)
	require.Nil(t, vmType)
	require.Nil(t, input)
	require.Equal(t, badCodeError, err)
}

func TestScProcessor_CreateVMDeployInput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("foobar")
	tx.Value = big.NewInt(45)

	expectedVMType := []byte{5, 6}
	expectedCodeMetadata := vmcommon.CodeMetadata{Upgradeable: true}
	argParser.ParseDeployDataCalled = func(data string) (*parsers.DeployArgs, error) {
		return &parsers.DeployArgs{
			Code:         []byte("code"),
			VMType:       expectedVMType,
			CodeMetadata: expectedCodeMetadata,
			Arguments:    nil,
		}, nil
	}

	input, vmType, err := sc.createVMDeployInput(tx)
	require.NotNil(t, input)
	require.Nil(t, err)
	require.Equal(t, vmData.DirectCall, input.CallType)
	require.True(t, bytes.Equal(expectedVMType, vmType))
	require.Equal(t, expectedCodeMetadata.ToBytes(), input.ContractCodeMetadata)
	require.Nil(t, err)
}

func TestScProcessor_CreateVMDeployInputNotEnoughArguments(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data@0000")
	tx.Value = big.NewInt(45)

	input, vmType, err := sc.createVMDeployInput(tx)
	require.Nil(t, input)
	require.Nil(t, vmType)
	require.Equal(t, parsers.ErrInvalidDeployArguments, err)
}

func TestScProcessor_CreateVMDeployInputWrongArgument(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	tmpError := errors.New("fooError")
	argParser.ParseDeployDataCalled = func(data string) (*parsers.DeployArgs, error) {
		return nil, tmpError
	}
	input, vmType, err := sc.createVMDeployInput(tx)
	require.Nil(t, input)
	require.Nil(t, vmType)
	require.Equal(t, tmpError, err)
}

func TestScProcessor_InitializeVMInputFromTx_ShouldErrNotEnoughGas(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			return 1000
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	tx.GasLimit = 100

	vmInput := &vmcommon.VMInput{}
	err = sc.initializeVMInputFromTx(vmInput, tx)
	require.Equal(t, process.ErrNotEnoughGas, err)
}

func TestScProcessor_InitializeVMInputFromTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	vmInput := &vmcommon.VMInput{}
	err = sc.initializeVMInputFromTx(vmInput, tx)
	require.Nil(t, err)
}

func createAccountsAndTransaction() (state.UserAccountHandler, state.UserAccountHandler, *transaction.Transaction) {
	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	acntSrc, acntDst := createAccounts(tx)

	return acntSrc, acntDst, tx
}

func TestScProcessor_processVMOutputNilSndAcc(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.ShardCoordinator = &mock.CoordinatorStub{ComputeIdCalled: func(address []byte) uint32 {
		return 5
	}}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}
	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, err = sc.processVMOutput(&vmcommon.VMInput{
		CallType:    vmData.DirectCall,
		GasProvided: 0,
	}, vmOutput, txHash, tx)
	require.Nil(t, err)
}

func TestScProcessor_processVMOutputNilDstAcc(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &stateMock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acntSnd, _, tx := createAccountsAndTransaction()

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSnd, nil
	}

	tx.Value = big.NewInt(0)
	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, err = sc.processVMOutput(&vmcommon.VMInput{
		CallType:    vmData.DirectCall,
		GasProvided: 0,
	}, vmOutput, txHash, tx)
	require.Nil(t, err)
}

func TestScProcessor_GetAccountFromAddressAccNotFound(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	accountsDB.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return nil, state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("SRC"))
	require.Nil(t, acc)
	require.Equal(t, state.ErrAccNotFound, err)
}

func TestScProcessor_GetAccountFromAddrFailedGetExistingAccount(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		getCalled++
		return nil, state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	require.Nil(t, acc)
	require.Equal(t, state.ErrAccNotFound, err)
	require.Equal(t, 1, getCalled)
}

func TestScProcessor_GetAccountFromAddrAccNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		getCalled++
		return nil, state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId() + 1
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	require.Nil(t, acc)
	require.Nil(t, err)
	require.Equal(t, 0, getCalled)
}

func TestScProcessor_GetAccountFromAddr(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		getCalled++
		acc := createAccount(address)
		return acc, nil
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	require.NotNil(t, acc)
	require.Nil(t, err)
	require.Equal(t, 1, getCalled)
}

func TestScProcessor_DeleteAccountsFailedAtRemove(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	removeCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return nil, state.ErrAccNotFound
	}
	accountsDB.RemoveAccountCalled = func(address []byte) error {
		removeCalled++
		return state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	deletedAccounts := make([][]byte, 0)
	deletedAccounts = append(deletedAccounts, []byte("acc1"), []byte("acc2"), []byte("acc3"))
	err = sc.deleteAccounts(deletedAccounts)
	require.Equal(t, state.ErrAccNotFound, err)
	require.Equal(t, 0, removeCalled)
}

func TestScProcessor_DeleteAccountsNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	removeCalled := 0
	accountsDB.RemoveAccountCalled = func(address []byte) error {
		removeCalled++
		return state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	computeIdCalled := 0
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		computeIdCalled++
		return shardCoordinator.SelfId() + 1
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	deletedAccounts := make([][]byte, 0)
	deletedAccounts = append(deletedAccounts, []byte("acc1"), []byte("acc2"), []byte("acc3"))
	err = sc.deleteAccounts(deletedAccounts)
	require.Nil(t, err)
	require.Equal(t, 0, removeCalled)
	require.Equal(t, len(deletedAccounts), computeIdCalled)
}

func TestScProcessor_DeleteAccountsInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	removeCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return createAccount(address), nil
	}
	accountsDB.RemoveAccountCalled = func(address []byte) error {
		removeCalled++
		return nil
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	computeIdCalled := 0
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		computeIdCalled++
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	deletedAccounts := make([][]byte, 0)
	deletedAccounts = append(deletedAccounts, []byte("acc1"), []byte("acc2"), []byte("acc3"))
	err = sc.deleteAccounts(deletedAccounts)
	require.Nil(t, err)
	require.Equal(t, len(deletedAccounts), removeCalled)
	require.Equal(t, len(deletedAccounts), computeIdCalled)
}

func TestScProcessor_ProcessSCPaymentAccNotInShardShouldNotReturnError(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	err = sc.processSCPayment(tx, nil)
	require.Nil(t, err)
}

func TestScProcessor_ProcessSCPaymentNotEnoughBalance(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return core.SafeMul(tx.GetGasPrice(), tx.GetGasLimit())
		}}
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 15

	acntSrc := createAccount(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(45))

	currBalance := acntSrc.GetBalance().Uint64()

	err = sc.processSCPayment(tx, acntSrc)
	require.Equal(t, process.ErrInsufficientFunds, err)
	require.Equal(t, currBalance, acntSrc.GetBalance().Uint64())
}

func TestScProcessor_ProcessSCPayment(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.GetBalance().Uint64()
	modifiedBalance := currBalance - tx.Value.Uint64() - tx.GasLimit*tx.GasLimit

	err = sc.processSCPayment(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, modifiedBalance, acntSrc.GetBalance().Uint64())
}

func TestScProcessor_ProcessSCPaymentWithNewFlags(t *testing.T) {
	t.Parallel()

	txFee := big.NewInt(25)

	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		DeveloperPercentageCalled: func() float64 {
			return 0.0
		},
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return txFee
		},
		ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			return core.SafeMul(tx.GetGasPrice(), gasToUse)
		},
	}
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.PenalizedTooMuchGasFlag
		},
	}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.GetBalance().Uint64()
	modifiedBalance := currBalance - tx.Value.Uint64() - txFee.Uint64()
	err = sc.processSCPayment(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, modifiedBalance, acntSrc.GetBalance().Uint64())

	acntSrc, _ = createAccounts(tx)
	modifiedBalance = currBalance - tx.Value.Uint64() - tx.GasLimit*tx.GasLimit
	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return false
	}
	err = sc.processSCPayment(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, modifiedBalance, acntSrc.GetBalance().Uint64())
}

func TestScProcessor_RefundGasToSenderNilAndZeroRefund(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	txHash := []byte("txHash")

	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.GetBalance().Uint64()
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(0)}
	_, _ = sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmData.DirectCall,
	)
	require.Nil(t, err)
	require.Equal(t, currBalance, acntSrc.GetBalance().Uint64())
}

func TestScProcessor_RefundGasToSenderAccNotInShard(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10
	txHash := []byte("txHash")
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(10)}
	sctx, _ := sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmData.DirectCall,
	)
	require.Nil(t, err)
	require.NotNil(t, sctx)

	vmOutput = &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(10)}
	sctx, _ = sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmData.DirectCall,
	)
	require.Nil(t, err)
	require.NotNil(t, sctx)
}

func TestScProcessor_RefundGasToSender(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(10)
	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &economicsmocks.EconomicsHandlerStub{MinGasPriceCalled: func() uint64 {
		return minGasPrice
	}}
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 15
	tx.GasLimit = 15
	txHash := []byte("txHash")
	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.GetBalance().Uint64()

	refundGas := big.NewInt(10)
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: refundGas}
	scr, _ := sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmData.DirectCall,
	)
	require.Nil(t, err)

	finalValue := big.NewInt(0).Mul(refundGas, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))
	require.Equal(t, currBalance, acntSrc.GetBalance().Uint64())
	require.Equal(t, scr.Value.Cmp(finalValue), 0)
}

func TestScProcessor_DoNotRefundGasToSenderForAsyncCall(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(10)
	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &economicsmocks.EconomicsHandlerStub{MinGasPriceCalled: func() uint64 {
		return minGasPrice
	}}
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &smartContractResult.SmartContractResult{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 15
	tx.GasLimit = 15
	txHash := []byte("txHash")

	refundGas := big.NewInt(10)
	vmOutput := &vmcommon.VMOutput{GasRemaining: 10, GasRefund: refundGas}
	scr, _ := sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmData.AsynchronousCall,
	)
	require.Nil(t, err)

	scrValue := big.NewInt(0).Mul(vmOutput.GasRefund, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))
	require.Equal(t, scr.Value.Cmp(scrValue), 0)
	require.Equal(t, scr.GasLimit, vmOutput.GasRemaining)
	require.Equal(t, scr.GasPrice, tx.GasPrice)
}

func TestScProcessor_processVMOutput(t *testing.T) {
	t.Parallel()

	acntSrc, _, tx := createAccountsAndTransaction()

	accntState := &stateMock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}

	accntState.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		return acntSrc, nil
	}

	tx.Value = big.NewInt(0)
	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, err = sc.processVMOutput(&vmcommon.VMInput{
		CallType:    vmData.DirectCall,
		GasProvided: 0,
	}, vmOutput, txHash, tx)
	require.Nil(t, err)
}

func TestScProcessor_processSCOutputAccounts(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}

	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}
	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.DirectCall},
		&vmcommon.VMOutput{}, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Code = []byte("contract-code")
	outacc1.Nonce = 5
	outacc1.BalanceDelta = big.NewInt(int64(5))
	outputAccounts = append(outputAccounts, outacc1)

	testAddr := outaddress
	testAcc := createAccount(testAddr)

	accountsDB.LoadAccountCalled = func(address []byte) (handler vmcommon.AccountHandler, e error) {
		if bytes.Equal(address, testAddr) {
			return testAcc, nil
		}
		return nil, state.ErrAccNotFound
	}

	accountsDB.SaveAccountCalled = func(accountHandler vmcommon.AccountHandler) error {
		return nil
	}

	tx.Value = big.NewInt(int64(5))
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.DirectCall},
		&vmcommon.VMOutput{}, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outacc1.BalanceDelta = nil
	outacc1.Nonce++
	tx.Value = big.NewInt(0)
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.DirectCall},
		&vmcommon.VMOutput{}, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outacc1.Nonce++
	outacc1.BalanceDelta = big.NewInt(int64(10))
	tx.Value = big.NewInt(int64(10))

	currentBalance := testAcc.GetBalance().Uint64()
	vmOutBalance := outacc1.BalanceDelta.Uint64()
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.DirectCall},
		&vmcommon.VMOutput{}, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)
	require.Equal(t, currentBalance+vmOutBalance, testAcc.GetBalance().Uint64())
}

func TestScProcessor_processSCOutputAccountsNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}
	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.DirectCall},
		&vmcommon.VMOutput{}, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Code = []byte("contract-code")
	outacc1.Nonce = 5
	outputAccounts = append(outputAccounts, outacc1)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId() + 1
	}

	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.DirectCall},
		&vmcommon.VMOutput{}, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outacc1.BalanceDelta = big.NewInt(-50)
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.DirectCall},
		&vmcommon.VMOutput{}, outputAccounts, tx, []byte("hash"))
	require.Equal(t, err, process.ErrNegativeBalanceDeltaOnCrossShardAccount)
}

func TestScProcessor_CreateCrossShardTransactions(t *testing.T) {
	t.Parallel()

	testAccounts := createAccount([]byte("address"))
	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, err error) {
			return testAccounts, nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Nonce = 0
	outacc1.Balance = big.NewInt(5)
	outacc1.BalanceDelta = big.NewInt(15)
	outTransfer := vmcommon.OutputTransfer{Value: big.NewInt(5)}
	outacc1.OutputTransfers = append(outacc1.OutputTransfers, outTransfer)
	outputAccounts = append(outputAccounts, outacc1, outacc1, outacc1)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 15
	txHash := []byte("txHash")

	createdAsyncSCR, scTxs, err := sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.DirectCall},
		&vmcommon.VMOutput{}, outputAccounts, tx, txHash)
	require.Nil(t, err)
	require.Equal(t, len(outputAccounts), len(scTxs))
	require.False(t, createdAsyncSCR)
}

func TestScProcessor_CreateCrossShardTransactionsWithAsyncCalls(t *testing.T) {
	t.Parallel()

	testAccounts := createAccount([]byte("address"))
	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, err error) {
			return testAccounts, nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return false
		},
	}
	arguments.EnableEpochsHandler = enableEpochsHandler
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Nonce = 0
	outacc1.Balance = big.NewInt(5)
	outacc1.BalanceDelta = big.NewInt(15)
	outTransfer := vmcommon.OutputTransfer{Value: big.NewInt(5)}
	outacc1.OutputTransfers = append(outacc1.OutputTransfers, outTransfer)
	outputAccounts = append(outputAccounts, outacc1, outacc1, outacc1)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 15
	txHash := []byte("txHash")

	createdAsyncSCR, scTxs, err := sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.AsynchronousCall},
		&vmcommon.VMOutput{GasRemaining: 1000}, outputAccounts, tx, txHash)
	require.Nil(t, err)
	require.Equal(t, len(outputAccounts), len(scTxs))
	require.False(t, createdAsyncSCR)

	outAccBackTransfer := &vmcommon.OutputAccount{
		Address:         tx.SndAddr,
		Nonce:           0,
		Balance:         big.NewInt(0),
		BalanceDelta:    big.NewInt(0),
		OutputTransfers: []vmcommon.OutputTransfer{outTransfer},
		GasUsed:         0,
	}
	outputAccounts = append(outputAccounts, outAccBackTransfer)
	shardCoordinator.ComputeIdCalled = func(_ []byte) uint32 {
		return shardCoordinator.SelfId() + 1
	}
	createdAsyncSCR, scTxs, err = sc.processSCOutputAccounts(
		&vmcommon.VMInput{
			Arguments: [][]byte{{}, {}},
			CallType:  vmData.AsynchronousCall},
		&vmcommon.VMOutput{GasRemaining: 1000},
		outputAccounts,
		tx,
		txHash)
	require.Nil(t, err)
	require.Equal(t, len(outputAccounts), len(scTxs))
	require.True(t, createdAsyncSCR)
	lastScTx := scTxs[len(scTxs)-1].(*smartContractResult.SmartContractResult)

	t.Run("backwards compatibility", func(t *testing.T) {
		require.Equal(t, vmData.AsynchronousCallBack, lastScTx.CallType)
		require.Equal(t, []byte(nil), lastScTx.Data)
	})
	enableEpochsHandler.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.FixAsyncCallBackArgsListFlag
	}

	_, scTxs, err = sc.processSCOutputAccounts(
		&vmcommon.VMInput{CallType: vmData.AsynchronousCall},
		&vmcommon.VMOutput{GasRemaining: 1000},
		outputAccounts,
		tx,
		txHash)
	require.Nil(t, err)
	require.Equal(t, len(outputAccounts), len(scTxs))
	require.True(t, createdAsyncSCR)
	lastScTx = scTxs[len(scTxs)-1].(*smartContractResult.SmartContractResult)
	t.Run("fix enabled, data field is correctly populated", func(t *testing.T) {
		require.Equal(t, vmData.AsynchronousCallBack, lastScTx.CallType)
		require.Equal(t, []byte("@"+core.ConvertToEvenHex(int(vmcommon.Ok))), lastScTx.Data)
	})

	tx.Value = big.NewInt(0)
	scTxs, err = sc.processVMOutput(&vmcommon.VMInput{
		Arguments:   [][]byte{{}, {}},
		CallType:    vmData.AsynchronousCall,
		GasProvided: 10000,
	}, &vmcommon.VMOutput{GasRemaining: 1000}, txHash, tx)
	require.Nil(t, err)
	require.Equal(t, 1, len(scTxs))
	lastScTx = scTxs[len(scTxs)-1].(*smartContractResult.SmartContractResult)
	require.Equal(t, vmData.AsynchronousCallBack, lastScTx.CallType)
}

func TestScProcessor_CreateIntraShardTransactionsWithAsyncCalls(t *testing.T) {
	t.Parallel()

	testAccounts := createAccount([]byte("address"))
	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, err error) {
			return testAccounts, nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.MultiESDTTransferFixOnCallBackFlag
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Nonce = 0
	outacc1.Balance = big.NewInt(5)
	outacc1.BalanceDelta = big.NewInt(15)
	outTransfer := vmcommon.OutputTransfer{Value: big.NewInt(5)}
	outacc1.OutputTransfers = append(outacc1.OutputTransfers, outTransfer)
	outputAccounts = append(outputAccounts, outacc1, outacc1, outacc1)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 15
	txHash := []byte("txHash")

	outAccBackTransfer := &vmcommon.OutputAccount{
		Address:         tx.SndAddr,
		Nonce:           0,
		Balance:         big.NewInt(0),
		BalanceDelta:    big.NewInt(0),
		OutputTransfers: []vmcommon.OutputTransfer{outTransfer},
		GasUsed:         0,
	}
	outputAccounts = append(outputAccounts, outAccBackTransfer)
	shardCoordinator.ComputeIdCalled = func(_ []byte) uint32 {
		return shardCoordinator.SelfId()
	}
	createdAsyncSCR, scTxs, err := sc.processSCOutputAccounts(&vmcommon.VMInput{CallType: vmData.AsynchronousCall},
		&vmcommon.VMOutput{GasRemaining: 1000}, outputAccounts, tx, txHash)
	require.Nil(t, err)
	require.Equal(t, len(outputAccounts), len(scTxs))
	require.False(t, createdAsyncSCR)
}

func TestScProcessor_ProcessSmartContractResultNilScr(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	_, err = sc.ProcessSmartContractResult(nil)
	require.Equal(t, process.ErrNilSmartContractResult, err)
}

func TestScProcessor_ProcessSmartContractResultErrGetAccount(t *testing.T) {
	t.Parallel()

	accError := errors.New("account get error")
	called := false
	accountsDB := &stateMock.AccountsStub{LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
		called = true
		return nil, accError
	}}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{RcvAddr: []byte("recv address")}
	_, _ = sc.ProcessSmartContractResult(&scr)
	require.True(t, called)
}

func TestScProcessor_ProcessSmartContractResultAccNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.CurrentShard + 1
	}
	scr := smartContractResult.SmartContractResult{RcvAddr: []byte("recv address")}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Equal(t, err, process.ErrNilSCDestAccount)
}

func TestScProcessor_ProcessSmartContractResultBadAccType(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return &stateMock.AccountWrapMock{}, nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{RcvAddr: []byte("recv address"), Value: big.NewInt(0)}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultNotPayable(t *testing.T) {
	t.Parallel()

	userAcc := createAccount([]byte("recv address"))
	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			if bytes.Equal(address, userAcc.AddressBytes()) {
				return userAcc, nil
			}
			return &stateMock.AccountWrapMock{
				Balance: big.NewInt(0),
			}, nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	arguments.BlockChainHook = &testscommon.BlockChainHookStub{
		IsPayableCalled: func(_ []byte, _ []byte) (bool, error) {
			return false, nil
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		RcvAddr: userAcc.AddressBytes(),
		SndAddr: []byte("snd addr"),
		Value:   big.NewInt(0),
	}
	returnCode, err := sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)

	scr.Value = big.NewInt(10)
	returnCode, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, returnCode)
	require.True(t, userAcc.GetBalance().Cmp(zero) == 0)

	scr.OriginalSender = scr.RcvAddr
	returnCode, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	require.True(t, userAcc.GetBalance().Cmp(scr.Value) == 0)
}

func TestScProcessor_ProcessSmartContractResultOutputBalanceNil(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return createAccount(address), nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		RcvAddr: []byte("recv address"),
		SndAddr: []byte("snd addr"),
		Value:   big.NewInt(0),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultWithCode(t *testing.T) {
	t.Parallel()

	putCodeCalled := 0
	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return createAccount(address), nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			putCodeCalled++
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		SndAddr: []byte("snd addr"),
		RcvAddr: []byte("recv address"),
		Code:    []byte("code"),
		Value:   big.NewInt(15),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, 1, putCodeCalled)
}

func TestScProcessor_ProcessSmartContractResultWithData(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0
	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return createAccount(address), nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	test := "test"
	result := ""
	sep := "@"
	for i := 0; i < 6; i++ {
		result += test
		result += sep
	}

	scr := smartContractResult.SmartContractResult{
		SndAddr: []byte("snd addr"),
		RcvAddr: []byte("recv address"),
		Data:    []byte(result),
		Value:   big.NewInt(15),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, 1, saveAccountCalled)
}

func TestScProcessor_ProcessSmartContractResultDeploySCShouldError(t *testing.T) {
	t.Parallel()

	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return createAccount(address), nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	arguments.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCDeployment, process.SCDeployment
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		SndAddr: []byte("snd addr"),
		RcvAddr: []byte("recv address"),
		Data:    []byte("code@06"),
		Value:   big.NewInt(15),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultExecuteSC(t *testing.T) {
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
		return shardCoordinator.SelfId() + 1
	}

	executeCalled := false
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	arguments.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return &mock.VMExecutionHandlerStub{
				RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
					executeCalled = true
					return nil, nil
				},
			}, nil
		},
	}
	arguments.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
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
}

func TestScProcessor_ProcessSmartContractResultExecuteSCIfMetaAndBuiltIn(t *testing.T) {
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
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	arguments.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return &mock.VMExecutionHandlerStub{
				RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
					executeCalled = true
					return nil, nil
				},
			}, nil
		},
	}
	arguments.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.SCDeployFlag
		},
	}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub

	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		SndAddr: []byte("snd addr"),
		RcvAddr: scAddress,
		Data:    []byte("code@06"),
		Value:   big.NewInt(0),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.True(t, executeCalled)

	executeCalled = false
	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.BuiltInFunctionsFlag || flag == common.SCDeployFlag || flag == common.BuiltInFunctionOnMetaFlag
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.False(t, executeCalled)
}

func TestScProcessor_ProcessRelayedSCRValueBackToRelayer(t *testing.T) {
	t.Parallel()

	scAddress := []byte("000000000001234567890123456789012")
	dstScAddress := createAccount(scAddress)
	dstScAddress.SetCode([]byte("code"))

	baseValue := big.NewInt(100)
	userAddress := []byte("111111111111234567890123456789012")
	userAcc := createAccount(userAddress)
	_ = userAcc.AddToBalance(baseValue)
	relayedAddress := []byte("211111111111234567890123456789012")
	relayedAcc := createAccount(relayedAddress)

	accountsDB := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			if bytes.Equal(scAddress, address) {
				return dstScAddress, nil
			}
			if bytes.Equal(userAddress, address) {
				return userAcc, nil
			}
			if bytes.Equal(relayedAddress, address) {
				return relayedAcc, nil
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
	executeCalled := false
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	arguments.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return &mock.VMExecutionHandlerStub{
				RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
					executeCalled = true
					return &vmcommon.VMOutput{ReturnCode: vmcommon.UserError}, nil
				},
			}, nil
		},
	}
	arguments.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		SndAddr:      userAddress,
		RcvAddr:      scAddress,
		RelayerAddr:  relayedAddress,
		RelayedValue: big.NewInt(10),
		Data:         []byte("code@06"),
		Value:        big.NewInt(15),
	}
	returnCode, err := sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.True(t, executeCalled)
	require.Equal(t, returnCode, vmcommon.UserError)

	require.True(t, relayedAcc.GetBalance().Cmp(scr.RelayedValue) == 0)
	userReturnValue := big.NewInt(0).Sub(scr.Value, scr.RelayedValue)
	userFinalValue := baseValue.Sub(baseValue, scr.Value)
	userFinalValue.Add(userFinalValue, userReturnValue)
	require.True(t, userAcc.GetBalance().Cmp(userFinalValue) == 0)
}

func TestScProcessor_checkUpgradePermission(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	// Not an upgrade
	err = sc.checkUpgradePermission(nil, &vmcommon.ContractCallInput{})
	require.Nil(t, err)

	// Upgrade, nil contract passed (quite impossible though)
	err = sc.checkUpgradePermission(nil, &vmcommon.ContractCallInput{Function: "upgradeContract"})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)

	// Create a contract, owned by Alice
	contract := createAccount([]byte("contract"))
	contract.SetOwnerAddress([]byte("alice"))
	// Not yet upgradeable
	contract.SetCodeMetadata([]byte{0, 0})

	// Upgrade as Alice, but not allowed (not upgradeable)
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: []byte("alice")}})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)

	// Upgrade as Bob, but not allowed
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: []byte("bob")}})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)

	// Mark as upgradeable
	contract.SetCodeMetadata([]byte{1, 0})

	// Upgrade as Alice, allowed
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: []byte("alice")}})
	require.Nil(t, err)

	// Upgrade as Bob, but not allowed
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: []byte("bob")}})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)

	// Upgrade as nobody, not allowed
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: nil}})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)
}

func TestScProcessor_penalizeUserIfNeededShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.PenalizedTooMuchGasFlag
		},
	}
	sc, _ := NewSmartContractProcessor(arguments)

	gasProvided := uint64(1000)
	maxGasToRemain := gasProvided - (gasProvided / process.MaxGasFeeHigherFactorAccepted)

	callType := vmData.DirectCall
	vmOutput := &vmcommon.VMOutput{
		GasRemaining: maxGasToRemain,
	}
	sc.penalizeUserIfNeeded(&transaction.Transaction{}, []byte("txHash"), callType, gasProvided, vmOutput)
	assert.Equal(t, maxGasToRemain, vmOutput.GasRemaining)

	callType = vmData.AsynchronousCall
	vmOutput = &vmcommon.VMOutput{
		GasRemaining: maxGasToRemain + 1,
	}
	sc.penalizeUserIfNeeded(&transaction.Transaction{}, []byte("txHash"), callType, gasProvided, vmOutput)
	assert.Equal(t, maxGasToRemain+1, vmOutput.GasRemaining)

	callType = vmData.DirectCall
	vmOutput = &vmcommon.VMOutput{
		GasRemaining: maxGasToRemain + 1,
	}
	sc.penalizeUserIfNeeded(&transaction.Transaction{}, []byte("txHash"), callType, gasProvided, vmOutput)
	assert.Equal(t, uint64(0), vmOutput.GasRemaining)
}

func TestScProcessor_isTooMuchGasProvidedShouldWork(t *testing.T) {
	t.Parallel()

	gasProvided := uint64(100)
	maxGasToRemain := gasProvided - (gasProvided / process.MaxGasFeeHigherFactorAccepted)

	isTooMuchGas := isTooMuchGasProvided(gasProvided, gasProvided)
	assert.False(t, isTooMuchGas)

	isTooMuchGas = isTooMuchGasProvided(gasProvided, maxGasToRemain-1)
	assert.False(t, isTooMuchGas)

	isTooMuchGas = isTooMuchGasProvided(gasProvided, maxGasToRemain)
	assert.False(t, isTooMuchGas)

	isTooMuchGas = isTooMuchGasProvided(gasProvided, maxGasToRemain+1)
	assert.True(t, isTooMuchGas)
}

func TestScProcessor_penalizeUserIfNeededShouldWorkOnFlagActivation(t *testing.T) {
	arguments := createMockSmartContractProcessorArguments()
	genericEpochNotifier := forking.NewGenericEpochNotifier()
	cfg := config.EnableEpochs{
		PenalizedTooMuchGasEnableEpoch: 1,
	}
	arguments.EnableEpochsHandler, _ = enablers.NewEnableEpochsHandler(cfg, genericEpochNotifier)
	sc, _ := NewSmartContractProcessor(arguments)

	gasProvided := uint64(1000)
	maxGasToRemain := gasProvided - (gasProvided / process.MaxGasFeeHigherFactorAccepted)

	callType := vmData.DirectCall
	vmOutput := &vmcommon.VMOutput{
		GasRemaining: maxGasToRemain + 1,
	}

	genericEpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 0,
	})
	sc.penalizeUserIfNeeded(&transaction.Transaction{}, []byte("txHash"), callType, gasProvided, vmOutput)
	assert.Equal(t, maxGasToRemain+1, vmOutput.GasRemaining)

	genericEpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: 1,
	})
	sc.penalizeUserIfNeeded(&transaction.Transaction{}, []byte("txHash"), callType, gasProvided, vmOutput)
	assert.Equal(t, uint64(0), vmOutput.GasRemaining)
}

func TestSCProcessor_createSCRWhenError(t *testing.T) {
	arguments := createMockSmartContractProcessorArguments()
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.SCDeployFlag || flag == common.PenalizedTooMuchGasFlag || flag == common.RepairCallbackFlag
		},
	}
	sc, _ := NewSmartContractProcessor(arguments)

	acntSnd := &stateMock.UserAccountStub{}
	scr, consumedFee := sc.createSCRsWhenError(nil, []byte("txHash"), &transaction.Transaction{}, "string", []byte("msg"), 0)
	assert.Equal(t, uint64(0), scr.GasLimit)
	assert.Equal(t, consumedFee.Cmp(big.NewInt(0)), 0)
	expectedError := "@" + hex.EncodeToString([]byte("string"))
	assert.Equal(t, expectedError, string(scr.Data))

	scr, consumedFee = sc.createSCRsWhenError(acntSnd, []byte("txHash"), &transaction.Transaction{}, "string", []byte("msg"), 0)
	assert.Equal(t, uint64(0), scr.GasLimit)
	assert.Equal(t, consumedFee.Cmp(big.NewInt(0)), 0)
	assert.Equal(t, expectedError, string(scr.Data))

	scr, consumedFee = sc.createSCRsWhenError(
		acntSnd,
		[]byte("txHash"),
		&smartContractResult.SmartContractResult{CallType: vmData.AsynchronousCall},
		"string",
		[]byte("msg"),
		0)
	assert.Equal(t, uint64(0), scr.GasLimit)
	assert.Equal(t, consumedFee.Cmp(big.NewInt(0)), 0)
	assert.Equal(t, "@04@6d7367", string(scr.Data))

	scr, consumedFee = sc.createSCRsWhenError(
		acntSnd,
		[]byte("txHash"),
		&smartContractResult.SmartContractResult{CallType: vmData.AsynchronousCall, GasPrice: 1, GasLimit: 100},
		"string",
		[]byte("msg"),
		20)
	assert.Equal(t, uint64(1), scr.GasPrice)
	assert.Equal(t, consumedFee.Cmp(big.NewInt(80)), 0)
	assert.Equal(t, "@04@6d7367", string(scr.Data))
	assert.Equal(t, uint64(20), scr.GasLimit)

	scr, consumedFee = sc.createSCRsWhenError(
		acntSnd,
		[]byte("txHash"),
		&smartContractResult.SmartContractResult{CallType: vmData.AsynchronousCall, GasPrice: 1, GasLimit: 100},
		"string",
		[]byte("msg"),
		0)
	assert.Equal(t, uint64(1), scr.GasPrice)
	assert.Equal(t, consumedFee.Cmp(big.NewInt(100)), 0)
	assert.Equal(t, "@04@6d7367", string(scr.Data))
	assert.Equal(t, uint64(0), scr.GasLimit)
}

func TestGasLockedInSmartContractProcessor(t *testing.T) {
	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = NewArgumentParser()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	shardCoordinator.ComputeIdCalled = func(_ []byte) uint32 {
		return shardCoordinator.SelfId() + 1
	}
	arguments.ShardCoordinator = shardCoordinator
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.MultiESDTTransferFixOnCallBackFlag
		},
	}
	sc, _ := NewSmartContractProcessor(arguments)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Nonce = 0
	outacc1.Balance = big.NewInt(5)
	outacc1.BalanceDelta = big.NewInt(15)
	outTransfer := vmcommon.OutputTransfer{
		Value:     big.NewInt(5),
		CallType:  vmData.AsynchronousCall,
		GasLocked: 100,
		GasLimit:  100,
		Data:      []byte("functionCall"),
	}
	outacc1.OutputTransfers = append(outacc1.OutputTransfers, outTransfer)
	vmOutput := &vmcommon.VMOutput{
		OutputAccounts: make(map[string]*vmcommon.OutputAccount),
	}
	vmOutput.OutputAccounts[string(outaddress)] = outacc1

	asyncCallback, results, err := sc.createSmartContractResults(&vmcommon.VMInput{CallType: vmData.DirectCall},
		vmOutput, outacc1, &transaction.Transaction{}, []byte("hash"))
	require.Nil(t, err)
	require.False(t, asyncCallback)
	require.Equal(t, 1, len(results))

	scr := results[0].(*smartContractResult.SmartContractResult)
	gasLocked := sc.getGasLockedFromSCR(scr)
	require.Equal(t, gasLocked, outTransfer.GasLocked)

	_, args, err := sc.argsParser.ParseCallData(string(scr.Data))
	require.Nil(t, err)
	require.Equal(t, 1, len(args))

	finalArguments, gasLocked := sc.getAsyncCallGasLockFromTxData(scr.CallType, args)
	require.Equal(t, 0, len(finalArguments))
	require.Equal(t, gasLocked, outTransfer.GasLocked)

	shardCoordinator.ComputeIdCalled = func(_ []byte) uint32 {
		return shardCoordinator.SelfId()
	}
	asyncCallback, results, err = sc.createSmartContractResults(&vmcommon.VMInput{CallType: vmData.DirectCall},
		vmOutput, outacc1, &transaction.Transaction{}, []byte("hash"))
	require.Nil(t, err)
	require.False(t, asyncCallback)
	require.Equal(t, 1, len(results))

	scr = results[0].(*smartContractResult.SmartContractResult)
	gasLocked = sc.getGasLockedFromSCR(scr)
	require.Equal(t, gasLocked, uint64(0))
}

func TestSmartContractProcessor_computeTotalConsumedFeeAndDevRwd(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = NewArgumentParser()
	shardCoordinator := &mock.CoordinatorStub{ComputeIdCalled: func(address []byte) uint32 {
		return 0
	}}
	feeHandler := &economicsmocks.EconomicsHandlerStub{
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			return 0
		},
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
		},
		ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			return core.SafeMul(tx.GetGasPrice(), gasToUse)
		},
	}
	arguments.EconomicsFee = feeHandler
	arguments.ShardCoordinator = shardCoordinator
	sc, _ := NewSmartContractProcessor(arguments)

	totalFee, devFees := sc.computeTotalConsumedFeeAndDevRwd(&transaction.Transaction{GasPrice: 1}, &vmcommon.VMOutput{}, 0)
	assert.Equal(t, totalFee.Int64(), int64(0))
	assert.Equal(t, devFees.Int64(), int64(0))

	totalFee, devFees = sc.computeTotalConsumedFeeAndDevRwd(&transaction.Transaction{GasLimit: 100, GasPrice: 1}, &vmcommon.VMOutput{GasRemaining: 200}, 0)
	assert.Equal(t, totalFee.Int64(), int64(0))
	assert.Equal(t, devFees.Int64(), int64(0))

	totalFee, _ = sc.computeTotalConsumedFeeAndDevRwd(&transaction.Transaction{GasLimit: 100, GasPrice: 1}, &vmcommon.VMOutput{GasRemaining: 50}, 0)
	assert.Equal(t, totalFee.Int64(), int64(50))

	feeHandler.DeveloperPercentageCalled = func() float64 {
		return 0.5
	}
	totalFee, devFees = sc.computeTotalConsumedFeeAndDevRwd(&transaction.Transaction{GasLimit: 100, GasPrice: 1}, &vmcommon.VMOutput{GasRemaining: 50}, 10)
	assert.Equal(t, totalFee.Int64(), int64(50))
	assert.Equal(t, devFees.Int64(), int64(20))

	feeHandler.ComputeGasLimitCalled = func(tx data.TransactionWithFeeHandler) uint64 {
		return 10
	}
	shardCoordinator.SelfIdCalled = func() uint32 {
		return 1
	}
	totalFee, devFees = sc.computeTotalConsumedFeeAndDevRwd(&transaction.Transaction{GasLimit: 100, GasPrice: 1}, &vmcommon.VMOutput{GasRemaining: 50}, 10)
	assert.Equal(t, totalFee.Int64(), int64(30))
	assert.Equal(t, devFees.Int64(), int64(15))

	vmOutput := &vmcommon.VMOutput{GasRemaining: 50}
	vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	vmOutput.OutputAccounts["address"] = &vmcommon.OutputAccount{OutputTransfers: []vmcommon.OutputTransfer{{GasLimit: 10}}}
	totalFee, devFees = sc.computeTotalConsumedFeeAndDevRwd(
		&transaction.Transaction{GasLimit: 100, GasPrice: 1},
		vmOutput,
		10)
	assert.Equal(t, totalFee.Int64(), int64(20))
	assert.Equal(t, devFees.Int64(), int64(10))
}

func TestSmartContractProcessor_computeTotalConsumedFeeAndDevRwdWithDifferentSCCallPrice(t *testing.T) {
	t.Parallel()

	scAccountAddress := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x1e, 0x2e, 0x61, 0x1a, 0x9c, 0xe1, 0xe0, 0xc8, 0xe3, 0x28, 0x3c, 0xcc, 0x7c, 0x1b, 0x0f, 0x46, 0x61, 0x91, 0x70, 0x79, 0xa7, 0x5c}
	acc := createAccount(scAccountAddress)
	require.NotNil(t, acc)

	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = NewArgumentParser()
	shardCoordinator := &mock.CoordinatorStub{ComputeIdCalled: func(address []byte) uint32 {
		return 0
	}}

	// use a real fee handler
	args := createRealEconomicsDataArgs()
	feeHandler, err := economics.NewEconomicsData(*args)
	require.Nil(t, err)
	require.NotNil(t, feeHandler)
	arguments.TxFeeHandler, _ = postprocess.NewFeeAccumulator()

	arguments.EconomicsFee = feeHandler
	arguments.ShardCoordinator = shardCoordinator
	arguments.AccountsDB = &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return acc, nil
		},
	}
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.SCDeployFlag || flag == common.StakingV2FlagAfterEpoch
		},
	}

	sc, err := NewSmartContractProcessor(arguments)
	require.Nil(t, err)
	require.NotNil(t, sc)

	tx := &transaction.Transaction{
		RcvAddr:  scAccountAddress,
		GasPrice: 1000000000,
		GasLimit: 30000000,
		Data:     make([]byte, 100),
	}
	vmoutput := &vmcommon.VMOutput{
		GasRemaining: 10000000,
		GasRefund:    big.NewInt(0),
	}
	builtInGasUsed := uint64(1000000)

	totalFee, devFees := sc.computeTotalConsumedFeeAndDevRwd(tx, vmoutput, builtInGasUsed)
	expectedTotalFee, expectedDevFees := computeExpectedResults(args, tx, builtInGasUsed, vmoutput, true)
	require.Equal(t, expectedTotalFee, totalFee)
	require.Equal(t, expectedDevFees, devFees)
}

func TestSmartContractProcessor_finishSCExecutionV2(t *testing.T) {
	scAccountAddress := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x1e, 0x2e, 0x61, 0x1a, 0x9c, 0xe1, 0xe0, 0xc8, 0xe3, 0x28, 0x3c, 0xcc, 0x7c, 0x1b, 0x0f, 0x46, 0x61, 0x91, 0x70, 0x79, 0xa7, 0x5c}
	tests := []struct {
		name           string
		tx             *transaction.Transaction
		vmOutput       *vmcommon.VMOutput
		builtInGasUsed uint64
	}{
		{
			name:           "intra shard smart contract execution with builtin gas and remaining gas",
			tx:             &transaction.Transaction{RcvAddr: scAccountAddress, GasPrice: 1000000000, GasLimit: 30000000, Data: make([]byte, 100)},
			vmOutput:       &vmcommon.VMOutput{GasRemaining: 10000000, GasRefund: big.NewInt(0)},
			builtInGasUsed: uint64(1000000),
		},
		{
			name:           "intra shard smart contract execution with no builtin gas and remaining gas",
			tx:             &transaction.Transaction{RcvAddr: scAccountAddress, GasPrice: 1000000000, GasLimit: 30000000, Data: make([]byte, 100)},
			vmOutput:       &vmcommon.VMOutput{GasRemaining: 10000000, GasRefund: big.NewInt(0)},
			builtInGasUsed: uint64(1000000),
		},
		{
			name:           "intra shard smart contract execution with builtin gas and no remaining gas",
			tx:             &transaction.Transaction{RcvAddr: scAccountAddress, GasPrice: 2000000000, GasLimit: 20000000, Data: make([]byte, 100)},
			vmOutput:       &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(0)},
			builtInGasUsed: uint64(1000000),
		},
		{
			name:           "intra shard smart contract execution with no builtin gas and no remaining gas",
			tx:             &transaction.Transaction{RcvAddr: scAccountAddress, GasPrice: 2000000000, GasLimit: 20000000, Data: make([]byte, 100)},
			vmOutput:       &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(0)},
			builtInGasUsed: uint64(0),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			acc := createAccount(scAccountAddress)
			require.NotNil(t, acc)

			arguments := createMockSmartContractProcessorArguments()
			arguments.ArgsParser = NewArgumentParser()
			shardCoordinator := &mock.CoordinatorStub{ComputeIdCalled: func(address []byte) uint32 {
				return 0
			}}

			// use a real fee handler
			args := createRealEconomicsDataArgs()
			var err error
			arguments.EconomicsFee, err = economics.NewEconomicsData(*args)
			require.Nil(t, err)

			arguments.TxFeeHandler, err = postprocess.NewFeeAccumulator()
			require.Nil(t, err)

			arguments.ShardCoordinator = shardCoordinator
			arguments.AccountsDB = &stateMock.AccountsStub{
				RevertToSnapshotCalled: func(snapshot int) error {
					return nil
				},
				LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
					return acc, nil
				},
			}
			arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.SCDeployFlag || flag == common.StakingV2FlagAfterEpoch
				},
			}

			sc, err := NewSmartContractProcessor(arguments)
			require.Nil(t, err)
			require.NotNil(t, sc)

			expectedTotalFee, expectedDevFees := computeExpectedResults(args, test.tx, test.builtInGasUsed, test.vmOutput, true)

			retcode, err := sc.finishSCExecution(nil, []byte("txhash"), test.tx, test.vmOutput, test.builtInGasUsed)
			require.Nil(t, err)
			require.Equal(t, retcode, vmcommon.Ok)
			require.Nil(t, err)
			require.Equal(t, expectedDevFees, acc.GetDeveloperReward())
			require.Equal(t, expectedTotalFee, sc.txFeeHandler.GetAccumulatedFees())
			require.Equal(t, expectedDevFees, sc.txFeeHandler.GetDeveloperFees())
		})
	}
}

func TestScProcessor_CreateRefundForRelayerFromAnotherShard(t *testing.T) {
	arguments := createMockSmartContractProcessorArguments()
	sndAddress := []byte("sender11")
	rcvAddress := []byte("receiver")

	shardCoordinator := &mock.CoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			if bytes.Equal(address, sndAddress) {
				return 1
			}
			if bytes.Equal(address, rcvAddress) {
				return 0
			}
			return 2
		},
		SelfIdCalled: func() uint32 {
			return 0
		}}
	arguments.ShardCoordinator = shardCoordinator
	arguments.EconomicsFee = &economicsmocks.EconomicsHandlerStub{ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
		return big.NewInt(100)
	}}
	sc, _ := NewSmartContractProcessor(arguments)

	scrWithRelayed := &smartContractResult.SmartContractResult{
		Nonce:          0,
		Value:          big.NewInt(0),
		RcvAddr:        rcvAddress,
		SndAddr:        sndAddress,
		RelayerAddr:    []byte("relayer1"),
		RelayedValue:   big.NewInt(0),
		PrevTxHash:     []byte("someHash"),
		OriginalTxHash: []byte("someHash"),
		GasLimit:       10000,
		GasPrice:       10,
		CallType:       vmData.DirectCall,
	}

	vmOutput := &vmcommon.VMOutput{GasRemaining: 1000}
	_, relayerRefund := sc.createSCRForSenderAndRelayer(vmOutput, scrWithRelayed, []byte("txhash"), vmData.DirectCall)
	assert.NotNil(t, relayerRefund)

	senderID := sc.shardCoordinator.ComputeId(relayerRefund.SndAddr)
	assert.Equal(t, sc.shardCoordinator.SelfId(), senderID)
}

func TestScProcessor_ProcessIfErrorRevertAccountFails(t *testing.T) {
	t.Parallel()
	expectedError := errors.New("expected error")
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return expectedError
		},
	}

	sc, _ := NewSmartContractProcessor(arguments)

	sndAccount := &stateMock.UserAccountStub{}
	err := sc.ProcessIfError(sndAccount, []byte("txHash"), nil, "0", []byte("message"), 1, 100)
	require.NotNil(t, err)
	require.Equal(t, expectedError, err)
}

func TestScProcessor_ProcessIfErrorAsyncCallBack(t *testing.T) {
	t.Parallel()
	arguments := createMockSmartContractProcessorArguments()
	scr := &smartContractResult.SmartContractResult{
		CallType: vmData.AsynchronousCallBack,
		Value:    big.NewInt(10),
		SndAddr:  make([]byte, 32),
	}

	dstAccount := &stateMock.UserAccountStub{
		AddToBalanceCalled: func(value *big.Int) error {
			assert.Equal(t, value, scr.Value)
			return nil
		},
	}
	arguments.AccountsDB = &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return dstAccount, nil
		},
		SaveAccountCalled: func(account vmcommon.AccountHandler) error {
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}

	sc, _ := NewSmartContractProcessor(arguments)

	err := sc.ProcessIfError(nil, []byte("txHash"), scr, "0", []byte("message"), 1, 100)
	require.Nil(t, err)
}

func TestProcessIfErrorCheckBackwardsCompatibilityProcessTransactionFeeCalledShouldBeCalled(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	shardCoordinator := &mock.CoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			return 1
		},
		SelfIdCalled: func() uint32 {
			return 0
		}}
	arguments.ShardCoordinator = shardCoordinator
	arguments.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			return big.NewInt(100)
		},
	}

	called := false
	arguments.TxFeeHandler = &mock.FeeAccumulatorStub{
		ProcessTransactionFeeCalled: func(cost *big.Int, devFee *big.Int, hash []byte) {
			called = true
		},
	}

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{
		SndAddr: []byte("snd"),
		RcvAddr: []byte("rcv"),
		Value:   big.NewInt(15),
	}

	sndAccount := &stateMock.UserAccountStub{}
	err := sc.ProcessIfError(sndAccount, []byte("txHash"), tx, "0", []byte("message"), 1, 100)
	require.Nil(t, err)
	require.True(t, called)
}

func TestProcessIfErrorCheckBackwardsCompatibilityProcessTransactionFeeCalledShouldNOTBeCalled(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	shardCoordinator := &mock.CoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			return 1
		},
		SelfIdCalled: func() uint32 {
			return 0
		}}
	arguments.ShardCoordinator = shardCoordinator
	arguments.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			return big.NewInt(100)
		},
	}

	called := false
	arguments.TxFeeHandler = &mock.FeeAccumulatorStub{
		ProcessTransactionFeeCalled: func(cost *big.Int, devFee *big.Int, hash []byte) {
			called = true
		},
	}

	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.SCDeployFlag || flag == common.CleanUpInformativeSCRsFlag || flag == common.OptimizeGasUsedInCrossMiniBlocksFlag
		},
	}

	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{
		SndAddr: []byte("snd"),
		RcvAddr: []byte("rcv"),
		Value:   big.NewInt(15),
	}

	err := sc.ProcessIfError(nil, []byte("txHash"), tx, "0", []byte("message"), 1, 100)
	require.Nil(t, err)
	require.False(t, called)
}

func TestProcessSCRSizeTooBig(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub
	sc, _ := NewSmartContractProcessor(arguments)

	scrTooBig := &smartContractResult.SmartContractResult{Data: bytes.Repeat([]byte{1}, core.MegabyteSize)}
	scrs := make([]data.TransactionHandler, 0)
	scrs = append(scrs, scrTooBig)

	err := sc.checkSCRSizeInvariant(scrs)
	assert.Nil(t, err)

	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.SCRSizeInvariantCheckFlag
	}
	err = sc.checkSCRSizeInvariant(scrs)
	assert.Equal(t, err, process.ErrResultingSCRIsTooBig)
}

func TestProcessIsInformativeSCR(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	builtInFuncs := builtInFunctions.NewBuiltInFunctionContainer()
	arguments.BuiltInFunctions = builtInFuncs
	arguments.ArgsParser = NewArgumentParser()
	sc, _ := NewSmartContractProcessor(arguments)

	scr := &smartContractResult.SmartContractResult{Value: big.NewInt(1)}
	assert.False(t, sc.isInformativeTxHandler(scr))

	scr.Value = big.NewInt(0)
	scr.CallType = vmData.AsynchronousCallBack
	assert.False(t, sc.isInformativeTxHandler(scr))

	scr.CallType = vmData.DirectCall
	scr.Data = []byte("@abab")
	assert.True(t, sc.isInformativeTxHandler(scr))

	scr.Data = []byte("ab@ab")
	scr.RcvAddr = make([]byte, 32)
	assert.False(t, sc.isInformativeTxHandler(scr))

	scr.RcvAddr = []byte("address")
	assert.True(t, sc.isInformativeTxHandler(scr))

	_ = builtInFuncs.Add("ab", &mock.BuiltInFunctionStub{})
	assert.False(t, sc.isInformativeTxHandler(scr))
}

func TestCleanInformativeOnlySCRs(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	builtInFuncs := builtInFunctions.NewBuiltInFunctionContainer()
	arguments.BuiltInFunctions = builtInFuncs
	arguments.ArgsParser = NewArgumentParser()
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub
	sc, _ := NewSmartContractProcessor(arguments)

	scrs := make([]data.TransactionHandler, 0)
	scrs = append(scrs, &smartContractResult.SmartContractResult{Value: big.NewInt(1)})
	scrs = append(scrs, &smartContractResult.SmartContractResult{Value: big.NewInt(0), Data: []byte("@6b6f")})

	finalSCRs, logs := sc.cleanInformativeOnlySCRs(scrs)
	assert.Equal(t, len(finalSCRs), len(scrs))
	assert.Equal(t, 1, len(logs))

	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.CleanUpInformativeSCRsFlag
	}
	finalSCRs, logs = sc.cleanInformativeOnlySCRs(scrs)
	assert.Equal(t, 1, len(finalSCRs))
	assert.Equal(t, 1, len(logs))
}

func TestProcessGetOriginalTxHashForRelayedIntraShard(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = NewArgumentParser()
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(2, 0)
	arguments.ShardCoordinator = shardCoordinator
	sc, _ := NewSmartContractProcessor(arguments)

	scr := &smartContractResult.SmartContractResult{Value: big.NewInt(1), SndAddr: bytes.Repeat([]byte{1}, 32)}
	scrHash := []byte("hash")

	logHash := sc.getOriginalTxHashIfIntraShardRelayedSCR(scr, scrHash)
	assert.Equal(t, scrHash, logHash)

	scr.OriginalTxHash = []byte("originalHash")
	scr.RelayerAddr = bytes.Repeat([]byte{1}, 32)
	scr.SndAddr = bytes.Repeat([]byte{1}, 32)
	scr.RcvAddr = bytes.Repeat([]byte{1}, 32)
	logHash = sc.getOriginalTxHashIfIntraShardRelayedSCR(scr, scrHash)
	assert.Equal(t, scr.OriginalTxHash, logHash)

	scr.RcvAddr = bytes.Repeat([]byte{2}, 32)
	logHash = sc.getOriginalTxHashIfIntraShardRelayedSCR(scr, scrHash)
	assert.Equal(t, scrHash, logHash)
}

func TestProcess_createCompletedTxEvent(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = NewArgumentParser()
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(2, 0)
	arguments.ShardCoordinator = shardCoordinator
	completedLogSaved := false
	arguments.TxLogsProcessor = &mock.TxLogsProcessorStub{SaveLogCalled: func(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error {
		for _, vmLog := range vmLogs {
			if string(vmLog.Identifier) == core.CompletedTxEventIdentifier {
				completedLogSaved = true
			}
		}
		return nil
	}}
	sc, _ := NewSmartContractProcessor(arguments)

	scAddress := bytes.Repeat([]byte{0}, 32)
	scAddress[31] = 2
	userAddress := bytes.Repeat([]byte{1}, 32)

	scr := &smartContractResult.SmartContractResult{
		Value:      big.NewInt(1),
		SndAddr:    userAddress,
		RcvAddr:    scAddress,
		PrevTxHash: []byte("prevTxHash"),
		GasLimit:   1000,
		GasPrice:   1000,
	}
	scrHash := []byte("hash")

	completeTxEvent := sc.createCompleteEventLogIfNoMoreAction(scr, scrHash, nil)
	assert.NotNil(t, completeTxEvent)

	scrWithTransfer := &smartContractResult.SmartContractResult{
		Value:   big.NewInt(1),
		SndAddr: scAddress,
		RcvAddr: userAddress,
		Data:    []byte("transfer"),
	}
	completeTxEvent = sc.createCompleteEventLogIfNoMoreAction(scr, scrHash, []data.TransactionHandler{scrWithTransfer})
	assert.Nil(t, completeTxEvent)

	scrWithTransfer.Value = big.NewInt(0)
	completeTxEvent = sc.createCompleteEventLogIfNoMoreAction(scr, scrHash, []data.TransactionHandler{scrWithTransfer})
	assert.NotNil(t, completeTxEvent)
	assert.Equal(t, completeTxEvent.Identifier, []byte(core.CompletedTxEventIdentifier))
	assert.Equal(t, completeTxEvent.Topics[0], scr.PrevTxHash)

	scrWithRefund := &smartContractResult.SmartContractResult{Value: big.NewInt(10), PrevTxHash: scrHash, Data: []byte("@6f6b@aaffaa")}
	completedLogSaved = false

	acntDst := createAccount(userAddress)
	err := sc.processSimpleSCR(scrWithRefund, []byte("scrHash"), acntDst)
	assert.Nil(t, err)
	assert.True(t, completedLogSaved)
}

func createRealEconomicsDataArgs() *economics.ArgsNewEconomicsData {
	return &economics.ArgsNewEconomicsData{
		Economics: &config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply: "20000000000000000000000000",
				MinimumInflation:   0.0,
				YearSettings: []*config.YearSetting{
					{Year: 1, MaximumInflation: 0.10845130},
				},
				Denomination: 18,
			},
			RewardsSettings: config.RewardsSettings{
				RewardsConfigByEpoch: []config.EpochRewardSettings{
					{
						LeaderPercentage:                 0.1,
						DeveloperPercentage:              0.3,
						ProtocolSustainabilityPercentage: 0.1,
						ProtocolSustainabilityAddress:    "erd1j25xk97yf820rgdp3mj5scavhjkn6tjyn0t63pmv5qyjj7wxlcfqqe2rw5",
						TopUpGradientPoint:               "300000000000000000000",
						TopUpFactor:                      0.25,
					},
				},
			},
			FeeSettings: config.FeeSettings{
				GasLimitSettings: []config.GasLimitSetting{
					{
						MaxGasLimitPerBlock:         "1500000000",
						MaxGasLimitPerMiniBlock:     "1500000000",
						MaxGasLimitPerMetaBlock:     "15000000000",
						MaxGasLimitPerMetaMiniBlock: "15000000000",
						MaxGasLimitPerTx:            "1500000000",
						MinGasLimit:                 "50000",
						ExtraGasLimitGuardedTx:      "50000",
					},
				},
				GasPerDataByte:         "1500",
				MinGasPrice:            "1000000000",
				GasPriceModifier:       0.01,
				MaxGasPriceSetGuardian: "100000",
			},
		},
		EpochNotifier: &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.GasPriceModifierFlag
			},
		},
		BuiltInFunctionsCostHandler: &mock.BuiltInCostHandlerStub{},
		TxVersionChecker:            &testscommon.TxVersionCheckerStub{},
	}
}

func computeExpectedResults(
	args *economics.ArgsNewEconomicsData,
	tx *transaction.Transaction,
	builtInGasUsed uint64,
	vmoutput *vmcommon.VMOutput,
	stakingV2Enabled bool,
) (*big.Int, *big.Int) {
	minGasLimitBigInt, _ := big.NewInt(0).SetString(args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit, 10)
	gasPerByteBigInt, _ := big.NewInt(0).SetString(args.Economics.FeeSettings.GasPerDataByte, 10)
	minGasLimit := minGasLimitBigInt.Uint64()
	gasPerByte := gasPerByteBigInt.Uint64()

	moveGas := uint64(len(tx.Data))*gasPerByte + minGasLimit
	moveFee := big.NewInt(0).SetUint64(moveGas)
	moveFee = big.NewInt(0).Mul(moveFee, big.NewInt(0).SetUint64(tx.GasPrice))

	processGas := tx.GasLimit - builtInGasUsed - moveGas - vmoutput.GasRemaining
	processPrice := big.NewInt(0).SetUint64(uint64(float64(tx.GasPrice) * args.Economics.FeeSettings.GasPriceModifier))
	processFee := big.NewInt(0).SetUint64(processGas)
	processFee = big.NewInt(0).Mul(processFee, processPrice)

	builtInFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(builtInGasUsed), processPrice)

	expectedTotalFee := big.NewInt(0).Add(moveFee, processFee)
	expectedTotalFee.Add(expectedTotalFee, builtInFee)
	var expectedDevFees *big.Int
	if stakingV2Enabled {
		expectedDevFees = core.GetIntTrimmedPercentageOfValue(processFee, args.Economics.RewardsSettings.RewardsConfigByEpoch[0].DeveloperPercentage)
	} else {
		expectedDevFees = core.GetApproximatePercentageOfValue(processFee, args.Economics.RewardsSettings.RewardsConfigByEpoch[0].DeveloperPercentage)
	}
	return expectedTotalFee, expectedDevFees
}

func TestMergeVmOutputLogs(t *testing.T) {
	t.Parallel()

	vmOutput1 := &vmcommon.VMOutput{
		Logs: nil,
	}

	vmOutput2 := &vmcommon.VMOutput{
		Logs: nil,
	}

	mergeVMOutputLogs(vmOutput1, vmOutput2)
	require.Nil(t, vmOutput1.Logs)

	vmOutput1 = &vmcommon.VMOutput{
		Logs: nil,
	}

	vmOutput2 = &vmcommon.VMOutput{
		Logs: []*vmcommon.LogEntry{
			{},
		},
	}

	mergeVMOutputLogs(vmOutput1, vmOutput2)
	require.Len(t, vmOutput1.Logs, 1)

	vmOutput1 = &vmcommon.VMOutput{
		Logs: []*vmcommon.LogEntry{
			{},
		},
	}

	vmOutput2 = &vmcommon.VMOutput{
		Logs: []*vmcommon.LogEntry{
			{},
		},
	}

	mergeVMOutputLogs(vmOutput1, vmOutput2)
	require.Len(t, vmOutput1.Logs, 2)

	vmOutput1 = &vmcommon.VMOutput{
		Logs: []*vmcommon.LogEntry{
			{
				Identifier: []byte("identifier2"),
			},
		},
	}

	vmOutput2 = &vmcommon.VMOutput{
		Logs: []*vmcommon.LogEntry{
			{
				Identifier: []byte("identifier1"),
			},
		},
	}

	mergeVMOutputLogs(vmOutput1, vmOutput2)
	require.Len(t, vmOutput1.Logs, 2)
	require.Equal(t, []byte("identifier1"), vmOutput1.Logs[0].Identifier)
	require.Equal(t, []byte("identifier2"), vmOutput1.Logs[1].Identifier)
}

func TestScProcessor_TooMuchGasProvidedMessage(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.SCDeployFlag || flag == common.PenalizedTooMuchGasFlag
		},
	}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub
	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, sc.pubkeyConv.Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)

	vmOutput := &vmcommon.VMOutput{GasRemaining: 10}
	sc.penalizeUserIfNeeded(tx, []byte("txHash"), vmData.DirectCall, 11, vmOutput)
	returnMessage := "@" + fmt.Sprintf("%s: gas needed = %d, gas remained = %d",
		TooMuchGasProvidedMessage, 1, 10)
	assert.Equal(t, vmOutput.ReturnMessage, returnMessage)

	enableEpochsHandlerStub.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
		return flag == common.SCDeployFlag || flag == common.PenalizedTooMuchGasFlag || flag == common.CleanUpInformativeSCRsFlag
	}
	vmOutput = &vmcommon.VMOutput{GasRemaining: 10}
	sc.penalizeUserIfNeeded(tx, []byte("txHash"), vmData.DirectCall, 11, vmOutput)
	returnMessage = "@" + fmt.Sprintf("%s for processing: gas provided = %d, gas used = %d",
		TooMuchGasProvidedMessage, 11, 1)
	assert.Equal(t, vmOutput.ReturnMessage, returnMessage)
}

func TestScProcessor_CheckBuiltinFunctionIsExecutable(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	expectedErr := errors.New("expected error")

	t.Run("nil transaction should error", func(t *testing.T) {
		sc, _ := NewSmartContractProcessor(arguments)
		err := sc.CheckBuiltinFunctionIsExecutable("SetGuardian", nil)
		require.Equal(t, process.ErrNilTransaction, err)
	})
	t.Run("args parser error should error", func(t *testing.T) {
		tx := &transaction.Transaction{}
		sc, _ := NewSmartContractProcessor(arguments)
		err := sc.CheckBuiltinFunctionIsExecutable("SetGuardian", tx)
		require.Error(t, err)
	})
	t.Run("", func(t *testing.T) {
		argsCopy := arguments
		argsCopy.ArgsParser = &mock.ArgumentParserMock{
			ParseCallDataCalled: func(data string) (string, [][]byte, error) {
				return "", nil, expectedErr
			},
		}
		sc, _ := NewSmartContractProcessor(argsCopy)
		err := sc.CheckBuiltinFunctionIsExecutable("SetGuardian", &transaction.Transaction{})
		require.Equal(t, expectedErr, err)
	})
	t.Run("expected builtin function different than the parsed function name should return error", func(t *testing.T) {
		argsCopy := arguments
		argsCopy.ArgsParser = &mock.ArgumentParserMock{
			ParseCallDataCalled: func(data string) (string, [][]byte, error) {
				return "differentFunction", nil, nil
			},
		}
		sc, _ := NewSmartContractProcessor(argsCopy)
		err := sc.CheckBuiltinFunctionIsExecutable("SetGuardian", &transaction.Transaction{})
		require.Equal(t, process.ErrBuiltinFunctionMismatch, err)
	})
	t.Run("prepare gas provided with error should error", func(t *testing.T) {
		argsCopy := arguments
		argsCopy.ArgsParser = &mock.ArgumentParserMock{
			ParseCallDataCalled: func(data string) (string, [][]byte, error) {
				return "SetGuardian", nil, nil
			},
		}
		argsCopy.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return setGuardianCost
			},
		}
		sc, _ := NewSmartContractProcessor(argsCopy)
		err := sc.CheckBuiltinFunctionIsExecutable("SetGuardian", &transaction.Transaction{
			GasLimit: setGuardianCost - 100,
		})
		require.Equal(t, process.ErrNotEnoughGas, err)
	})
	t.Run("builtin function not found should error", func(t *testing.T) {
		argsCopy := arguments
		argsCopy.ArgsParser = &mock.ArgumentParserMock{
			ParseCallDataCalled: func(data string) (string, [][]byte, error) {
				return "SetGuardian", nil, nil
			},
		}
		argsCopy.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return setGuardianCost
			},
		}
		sc, _ := NewSmartContractProcessor(argsCopy)
		err := sc.CheckBuiltinFunctionIsExecutable(
			"SetGuardian",
			&transaction.Transaction{
				GasLimit: setGuardianCost + 1000,
			})

		require.Equal(t, process.ErrBuiltinFunctionNotExecutable, err)
	})
	t.Run("builtin function not supporting executable check should error", func(t *testing.T) {
		argsCopy := arguments
		argsCopy.ArgsParser = &mock.ArgumentParserMock{
			ParseCallDataCalled: func(data string) (string, [][]byte, error) {
				return "SetGuardian", nil, nil
			},
		}
		argsCopy.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return setGuardianCost
			},
		}
		// BuiltInFunctionStub is not supporting executable check
		_ = argsCopy.BuiltInFunctions.Add("SetGuardian", &mock.BuiltInFunctionStub{})
		sc, _ := NewSmartContractProcessor(argsCopy)
		err := sc.CheckBuiltinFunctionIsExecutable("SetGuardian", &transaction.Transaction{
			GasLimit: setGuardianCost + 1000,
		})
		require.Equal(t, process.ErrBuiltinFunctionNotExecutable, err)
	})
	t.Run("OK", func(t *testing.T) {
		argsCopy := arguments
		argsCopy.ArgsParser = &mock.ArgumentParserMock{
			ParseCallDataCalled: func(data string) (string, [][]byte, error) {
				return "SetGuardian", nil, nil
			},
		}
		argsCopy.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return setGuardianCost
			},
		}
		argsCopy.BuiltInFunctions = builtInFunctions.NewBuiltInFunctionContainer()
		// BuiltInFunctionExecutableStub is supporting executable check
		_ = argsCopy.BuiltInFunctions.Add("SetGuardian", &vmcommonMocks.BuiltInFunctionExecutableStub{})
		sc, _ := NewSmartContractProcessor(argsCopy)
		err := sc.CheckBuiltinFunctionIsExecutable("SetGuardian", &transaction.Transaction{
			GasLimit: setGuardianCost + 1000,
		})
		require.Nil(t, err)
	})
}

func Test_createExecutableCheckersMap(t *testing.T) {
	t.Parallel()

	t.Run("empty builtInFunctions should return empty map", func(t *testing.T) {
		arguments := createMockSmartContractProcessorArguments()
		builtinFuncs := arguments.BuiltInFunctions
		executableCheckersMap := scrCommon.CreateExecutableCheckersMap(builtinFuncs)
		require.NotNil(t, executableCheckersMap)
		require.Equal(t, 0, len(executableCheckersMap))
	})
	t.Run("no builtinFunctions implementing ExecutableChecker interface should return empty map", func(t *testing.T) {
		arguments := createMockSmartContractProcessorArguments()
		builtinFuncs := arguments.BuiltInFunctions
		_ = builtinFuncs.Add("SetGuardian", &mock.BuiltInFunctionStub{})
		executableCheckersMap := scrCommon.CreateExecutableCheckersMap(builtinFuncs)
		require.NotNil(t, executableCheckersMap)
		require.Equal(t, 0, len(executableCheckersMap))
	})
	t.Run("one builtinFunctions implementing ExecutableChecker interface should return map with one entry that builtin func", func(t *testing.T) {
		arguments := createMockSmartContractProcessorArguments()
		expectedExecutableChecker := &vmcommonMocks.BuiltInFunctionExecutableStub{}
		builtinFuncs := arguments.BuiltInFunctions
		_ = builtinFuncs.Add("SetGuardian", expectedExecutableChecker)
		_ = builtinFuncs.Add("SetGuardian2", &mock.BuiltInFunctionStub{})
		executableCheckersMap := scrCommon.CreateExecutableCheckersMap(builtinFuncs)
		require.NotNil(t, executableCheckersMap)
		require.Equal(t, 1, len(executableCheckersMap))
		require.Equal(t, expectedExecutableChecker, executableCheckersMap["SetGuardian"])
	})
}

func TestScProcessor_DisableAsyncCalls(t *testing.T) {
	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = NewArgumentParser()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	shardCoordinator.ComputeIdCalled = func(_ []byte) uint32 {
		return shardCoordinator.SelfId() + 1
	}
	arguments.ShardCoordinator = shardCoordinator
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	arguments.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
		IsDisableAsyncCallV1EnabledCalled: func() bool {
			return false
		},
	}
	sc, _ := NewSmartContractProcessor(arguments)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Nonce = 0
	outacc1.Balance = big.NewInt(5)
	outacc1.BalanceDelta = big.NewInt(15)
	outTransfer := vmcommon.OutputTransfer{
		Value:     big.NewInt(5),
		CallType:  vmData.AsynchronousCall,
		GasLocked: 100,
		GasLimit:  100,
		Data:      []byte("functionCall"),
	}
	outacc1.OutputTransfers = append(outacc1.OutputTransfers, outTransfer)
	vmOutput := &vmcommon.VMOutput{
		OutputAccounts: make(map[string]*vmcommon.OutputAccount),
	}
	vmOutput.OutputAccounts[string(outaddress)] = outacc1

	_, scResults, err := sc.createSmartContractResults(&vmcommon.VMInput{CallType: vmData.DirectCall},
		vmOutput, outacc1, &transaction.Transaction{}, []byte("hash"))
	require.Nil(t, err)
	require.NotNil(t, scResults)

	arguments.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
		IsDisableAsyncCallV1EnabledCalled: func() bool {
			return true
		},
	}
	sc, _ = NewSmartContractProcessor(arguments)

	_, scResults, err = sc.createSmartContractResults(&vmcommon.VMInput{CallType: vmData.DirectCall},
		vmOutput, outacc1, &transaction.Transaction{}, []byte("hash"))
	require.Equal(t, process.ErrAsyncCallsDisabled.Error(), err.Error())
	require.Nil(t, scResults)
}
