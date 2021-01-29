package smartContract

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxEpoch = math.MaxUint32

func generateEmptyByteSlice(size int) []byte {
	buff := make([]byte, size)

	return buff
}

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func createAccounts(tx data.TransactionHandler) (state.UserAccountHandler, state.UserAccountHandler) {
	acntSrc, _ := state.NewUserAccount(tx.GetSndAddr())
	acntSrc.Balance = acntSrc.Balance.Add(acntSrc.Balance, tx.GetValue())
	totalFee := big.NewInt(0)
	totalFee = totalFee.Mul(big.NewInt(int64(tx.GetGasLimit())), big.NewInt(int64(tx.GetGasPrice())))
	acntSrc.Balance.Set(acntSrc.Balance.Add(acntSrc.Balance, totalFee))

	acntDst, _ := state.NewUserAccount(tx.GetRcvAddr())

	return acntSrc, acntDst
}

func createMockSmartContractProcessorArguments() ArgsNewSmartContractProcessor {
	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[core.ElrondAPICost] = make(map[string]uint64)
	gasSchedule[core.ElrondAPICost][core.AsyncCallStepField] = 1000
	gasSchedule[core.ElrondAPICost][core.AsyncCallbackGasLockField] = 3000
	gasSchedule[core.BuiltInCost] = make(map[string]uint64)
	gasSchedule[core.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 2000

	return ArgsNewSmartContractProcessor{
		VmContainer: &mock.VMContainerMock{},
		ArgsParser:  &mock.ArgumentParserMock{},
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
		AccountsDB: &mock.AccountsStub{
			RevertToSnapshotCalled: func(snapshot int) error {
				return nil
			},
		},
		BlockChainHook:   &mock.BlockChainHookHandlerMock{},
		PubkeyConv:       createMockPubkeyConverter(),
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(5),
		ScrForwarder:     &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:   &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler:     &mock.FeeAccumulatorStub{},
		TxLogsProcessor:  &mock.TxLogsProcessorStub{},
		EconomicsFee: &mock.FeeHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.0
			},
			ComputeTxFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
				return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
			},
			ComputeFeeForProcessingCalled: func(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
				return core.SafeMul(tx.GetGasPrice(), gasToUse)
			},
		},
		TxTypeHandler: &mock.TxTypeHandlerMock{},
		GasHandler: &mock.GasHandlerMock{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		GasSchedule:      mock.NewGasScheduleNotifierMock(gasSchedule),
		BuiltInFunctions: builtInFunctions.NewBuiltInFunctionContainer(),
		EpochNotifier:    &mock.EpochNotifierStub{},
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

func TestNewSmartContractProcessor_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.EpochNotifier = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilEpochNotifier, err)
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
	arguments.GasSchedule = mock.NewGasScheduleNotifierMock(nil)
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilGasSchedule, err)
}

func TestNewSmartContractProcessor_NilBuiltInFunctionsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.BuiltInFunctions = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilBuiltInFunction, err)
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

func TestNewSmartContractProcessor_ShouldRegisterNotifiers(t *testing.T) {
	t.Parallel()

	epochNotifierRegisterCalled := false
	gasScheduleRegisterCalled := false

	arguments := createMockSmartContractProcessorArguments()
	arguments.EpochNotifier = &mock.EpochNotifierStub{
		RegisterNotifyHandlerCalled: func(handler core.EpochSubscriberHandler) {
			epochNotifierRegisterCalled = true
		},
	}
	gasSchedule := mock.NewGasScheduleNotifierMock(make(map[string]map[string]uint64))
	gasSchedule.RegisterNotifyHandlerCalled = func(handler core.GasScheduleSubscribeHandler) {
		gasScheduleRegisterCalled = true
	}
	arguments.GasSchedule = gasSchedule

	_, _ = NewSmartContractProcessor(arguments)

	require.True(t, epochNotifierRegisterCalled)
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

	apiCosts := arguments.GasSchedule.LatestGasSchedule()[core.ElrondAPICost]
	builtInFuncCost := arguments.GasSchedule.LatestGasSchedule()[core.BuiltInCost]

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
	assert.Equal(t, apiCosts[core.AsyncCallStepField], sc.asyncCallStepCost)
	assert.Equal(t, apiCosts[core.AsyncCallbackGasLockField], sc.asyncCallbackGasLock)
	assert.Equal(t, builtInFuncCost[core.BuiltInFunctionESDTTransfer], sc.esdtTransferCost)
	assert.Equal(t, arguments.BuiltInFunctions, sc.builtInFunctions)
	assert.Equal(t, arguments.TxLogsProcessor, sc.txLogsProcessor)
	assert.Equal(t, arguments.BadTxForwarder, sc.badTxForwarder)
	assert.Equal(t, arguments.DeployEnableEpoch, sc.deployEnableEpoch)
	assert.Equal(t, arguments.BuiltinEnableEpoch, sc.builtinEnableEpoch)
	assert.Equal(t, arguments.PenalizedTooMuchGasEnableEpoch, sc.penalizedTooMuchGasEnableEpoch)
}

// ===================== TestGasScheduleChange =====================

func TestGasScheduleChangeNoApiCostShouldNotChange(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessor(arguments)

	apiCosts := arguments.GasSchedule.LatestGasSchedule()[core.ElrondAPICost]
	builtInFuncCost := arguments.GasSchedule.LatestGasSchedule()[core.BuiltInCost]

	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[core.BuiltInCost] = make(map[string]uint64)
	gasSchedule[core.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 2000

	sc.GasScheduleChange(gasSchedule)
	require.Equal(t, apiCosts[core.AsyncCallStepField], sc.asyncCallStepCost)
	require.Equal(t, apiCosts[core.AsyncCallbackGasLockField], sc.asyncCallbackGasLock)
	require.Equal(t, builtInFuncCost[core.BuiltInFunctionESDTTransfer], sc.esdtTransferCost)

	gasSchedule[core.ElrondAPICost] = nil
	gasSchedule[core.BuiltInCost] = make(map[string]uint64)
	gasSchedule[core.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 2000
	sc.GasScheduleChange(gasSchedule)

	require.Equal(t, apiCosts[core.AsyncCallStepField], sc.asyncCallStepCost)
	require.Equal(t, apiCosts[core.AsyncCallbackGasLockField], sc.asyncCallbackGasLock)
	require.Equal(t, builtInFuncCost[core.BuiltInFunctionESDTTransfer], sc.esdtTransferCost)

	gasSchedule[core.ElrondAPICost] = make(map[string]uint64)
	sc.GasScheduleChange(gasSchedule)

	require.Equal(t, apiCosts[core.AsyncCallStepField], sc.asyncCallStepCost)
	require.Equal(t, apiCosts[core.AsyncCallbackGasLockField], sc.asyncCallbackGasLock)
	require.Equal(t, builtInFuncCost[core.BuiltInFunctionESDTTransfer], sc.esdtTransferCost)
}

func TestGasScheduleChangeNoApiCostNoAsyncCallStepFieldShouldNotChange(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessor(arguments)

	apiCosts := arguments.GasSchedule.LatestGasSchedule()[core.ElrondAPICost]
	builtInFuncCost := arguments.GasSchedule.LatestGasSchedule()[core.BuiltInCost]

	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[core.ElrondAPICost] = make(map[string]uint64)
	gasSchedule[core.ElrondAPICost][core.AsyncCallbackGasLockField] = 3000
	gasSchedule[core.BuiltInCost] = make(map[string]uint64)
	gasSchedule[core.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 2000
	sc.GasScheduleChange(gasSchedule)

	require.Equal(t, apiCosts[core.AsyncCallStepField], sc.asyncCallStepCost)
	require.Equal(t, apiCosts[core.AsyncCallbackGasLockField], sc.asyncCallbackGasLock)
	require.Equal(t, builtInFuncCost[core.BuiltInFunctionESDTTransfer], sc.esdtTransferCost)
}

func TestGasScheduleChangeNoApiCostNoAsyncCallbackGasLockFieldShouldNotChange(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessor(arguments)

	apiCosts := arguments.GasSchedule.LatestGasSchedule()[core.ElrondAPICost]
	builtInFuncCost := arguments.GasSchedule.LatestGasSchedule()[core.BuiltInCost]

	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[core.ElrondAPICost] = make(map[string]uint64)
	gasSchedule[core.ElrondAPICost][core.AsyncCallStepField] = 1000
	gasSchedule[core.BuiltInCost] = make(map[string]uint64)
	gasSchedule[core.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 2000
	sc.GasScheduleChange(gasSchedule)

	require.Equal(t, apiCosts[core.AsyncCallStepField], sc.asyncCallStepCost)
	require.Equal(t, apiCosts[core.AsyncCallbackGasLockField], sc.asyncCallbackGasLock)
	require.Equal(t, builtInFuncCost[core.BuiltInFunctionESDTTransfer], sc.esdtTransferCost)
}

func TestGasScheduleChangeNoBuiltInCostShouldNotChange(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessor(arguments)

	apiCosts := arguments.GasSchedule.LatestGasSchedule()[core.ElrondAPICost]
	builtInFuncCost := arguments.GasSchedule.LatestGasSchedule()[core.BuiltInCost]

	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[core.ElrondAPICost] = make(map[string]uint64)
	gasSchedule[core.ElrondAPICost][core.AsyncCallStepField] = 1000
	gasSchedule[core.ElrondAPICost][core.AsyncCallbackGasLockField] = 3000

	sc.GasScheduleChange(gasSchedule)

	require.Equal(t, apiCosts[core.AsyncCallStepField], sc.asyncCallStepCost)
	require.Equal(t, apiCosts[core.AsyncCallbackGasLockField], sc.asyncCallbackGasLock)
	require.Equal(t, builtInFuncCost[core.BuiltInFunctionESDTTransfer], sc.esdtTransferCost)

	gasSchedule[core.BuiltInCost] = make(map[string]uint64)

	sc.GasScheduleChange(gasSchedule)

	require.Equal(t, apiCosts[core.AsyncCallStepField], sc.asyncCallStepCost)
	require.Equal(t, apiCosts[core.AsyncCallbackGasLockField], sc.asyncCallbackGasLock)
	require.Equal(t, builtInFuncCost[core.BuiltInFunctionESDTTransfer], sc.esdtTransferCost)
}

func TestGasScheduleChangeShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessor(arguments)

	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[core.ElrondAPICost] = make(map[string]uint64)
	gasSchedule[core.ElrondAPICost][core.AsyncCallStepField] = 10
	gasSchedule[core.ElrondAPICost][core.AsyncCallbackGasLockField] = 30
	gasSchedule[core.BuiltInCost] = make(map[string]uint64)
	gasSchedule[core.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 20

	sc.GasScheduleChange(gasSchedule)

	apiCosts := gasSchedule[core.ElrondAPICost]
	builtInFuncCost := gasSchedule[core.BuiltInCost]
	require.Equal(t, apiCosts[core.AsyncCallStepField], sc.asyncCallStepCost)
	require.Equal(t, apiCosts[core.AsyncCallbackGasLockField], sc.asyncCallbackGasLock)
	require.Equal(t, builtInFuncCost[core.BuiltInFunctionESDTTransfer], sc.esdtTransferCost)
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

	arguments.AccountsDB = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
		LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
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
	arguments.AccountsDB = &mock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
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
	arguments.AccountsDB = &mock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	arguments.DeployEnableEpoch = maxEpoch
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

	_, _ = sc.DeploySmartContract(tx, acntSrc)
	tsc := NewTestScProcessor(sc)
	require.Equal(t, process.ErrSmartContractDeploymentIsDisabled, tsc.GetLatestTestError())
}

func TestScProcessor_BuiltInCallSmartContractDisabled(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &mock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	arguments.BuiltinEnableEpoch = maxEpoch
	funcName := "builtIn"
	_ = arguments.BuiltInFunctions.Add(funcName, &mock.BuiltInFunctionStub{})
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
	arguments.AccountsDB = &mock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	arguments.BuiltinEnableEpoch = maxEpoch
	funcName := "builtIn"
	localError := errors.New("failed built in call")
	_ = arguments.BuiltInFunctions.Add(funcName, &mock.BuiltInFunctionStub{
		ProcessBuiltinFunctionCalled: func(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			return nil, localError
		},
	})

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
	accountState := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	arguments.AccountsDB = accountState
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	arguments.BuiltinEnableEpoch = maxEpoch
	funcName := "builtIn"
	_ = arguments.BuiltInFunctions.Add(funcName, &mock.BuiltInFunctionStub{})
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &smartContractResult.SmartContractResult{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, arguments.PubkeyConv.Len())
	tx.Data = []byte(funcName + "@0500@0000")
	tx.Value = big.NewInt(0)
	tx.CallType = vmcommon.AsynchronousCallBack
	acntSrc, actDst := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}
	accountState.LoadAccountCalled = func(address []byte) (state.AccountHandler, error) {
		if bytes.Equal(tx.SndAddr, address) {
			return acntSrc, nil
		}
		if bytes.Equal(tx.RcvAddr, address) {
			return actDst, nil
		}
		return nil, nil
	}

	sc.flagBuiltin.Set()
	retCode, err := sc.ExecuteBuiltInFunction(tx, acntSrc, actDst)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteBuiltInFunction(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	accountState := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	arguments.AccountsDB = accountState
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	arguments.BuiltinEnableEpoch = maxEpoch
	funcName := "builtIn"
	_ = arguments.BuiltInFunctions.Add(funcName, &mock.BuiltInFunctionStub{})
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
	accountState.LoadAccountCalled = func(address []byte) (state.AccountHandler, error) {
		return acntSrc, nil
	}

	sc.flagBuiltin.Set()
	retCode, err := sc.ExecuteBuiltInFunction(tx, acntSrc, nil)
	require.Equal(t, vmcommon.Ok, retCode)
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
	arguments.PubkeyConv = mock.NewPubkeyConverterMock(3)
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

	arguments.EconomicsFee = &mock.FeeHandlerStub{
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
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
	arguments.AccountsDB = &mock.AccountsStub{
		SaveAccountCalled: func(account state.AccountHandler) error {
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

	sc.flagPenalizedTooMuchGas.Toggle(false)

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
	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	economicsFee := &mock.FeeHandlerStub{
		DeveloperPercentageCalled: func() float64 {
			return 0.0
		},
		ComputeTxFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
		},
		ComputeFeeForProcessingCalled: func(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			return core.SafeMul(tx.GetGasPrice(), gasToUse)
		},
		// second call is in rewards and will fail
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
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

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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
	accountsDb := &mock.AccountsStub{
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

	accountsDb.LoadAccountCalled = func(address []byte) (state.AccountHandler, error) {
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
	accntState := &mock.AccountsStub{}
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

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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
	accntState := &mock.AccountsStub{}
	arguments.AccountsDB = accntState
	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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
	accntState := &mock.AccountsStub{}
	arguments.AccountsDB = accntState
	sc, _ := NewSmartContractProcessor(arguments)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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
	accntState := &mock.AccountsStub{}
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

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acntSrc, nil
	}

	_, err = sc.DeploySmartContract(tx, acntSrc)
	tsp := NewTestScProcessor(sc)
	require.Nil(t, err)
	require.Nil(t, tsp.GetLatestTestError())
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
	accntState := &mock.AccountsStub{}
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

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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
	accntState := &mock.AccountsStub{}
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

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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
	accntState := &mock.AccountsStub{}
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

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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
	accntState := &mock.AccountsStub{}
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

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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
	require.Equal(t, vmcommon.DirectCall, input.CallType)
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
	arguments.EconomicsFee = &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
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

	return acntSrc.(state.UserAccountHandler), acntDst.(state.UserAccountHandler), tx
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
	_, err = sc.processVMOutput(vmOutput, txHash, tx, vmcommon.DirectCall, 0)
	require.Nil(t, err)
}

func TestScProcessor_processVMOutputNilDstAcc(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &mock.AccountsStub{}
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

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acntSnd, nil
	}

	tx.Value = big.NewInt(0)
	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, err = sc.processVMOutput(vmOutput, txHash, tx, vmcommon.DirectCall, 0)
	require.Nil(t, err)
}

func TestScProcessor_GetAccountFromAddressAccNotFound(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		getCalled++
		acc, _ := state.NewUserAccount(address)
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

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
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

	accountsDB := &mock.AccountsStub{}
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

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		acc, _ := state.NewUserAccount(address)
		return acc, nil
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
		ComputeTxFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
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

	acntSrc, _ := state.NewUserAccount(tx.SndAddr)
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
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()
	modifiedBalance := currBalance - tx.Value.Uint64() - tx.GasLimit*tx.GasLimit

	err = sc.processSCPayment(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, modifiedBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
}

func TestScProcessor_ProcessSCPaymentWithNewFlags(t *testing.T) {
	t.Parallel()

	txFee := big.NewInt(25)

	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &mock.FeeHandlerStub{
		DeveloperPercentageCalled: func() float64 {
			return 0.0
		},
		ComputeTxFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			return txFee
		},
		ComputeFeeForProcessingCalled: func(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			return core.SafeMul(tx.GetGasPrice(), gasToUse)
		},
	}
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
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()
	modifiedBalance := currBalance - tx.Value.Uint64() - txFee.Uint64()
	err = sc.processSCPayment(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, modifiedBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())

	acntSrc, _ = createAccounts(tx)
	modifiedBalance = currBalance - tx.Value.Uint64() - tx.GasLimit*tx.GasLimit
	sc.flagPenalizedTooMuchGas.Toggle(false)
	err = sc.processSCPayment(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, modifiedBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
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
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(0)}
	_, _ = sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmcommon.DirectCall,
	)
	require.Nil(t, err)
	require.Equal(t, currBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
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
		vmcommon.DirectCall,
	)
	require.Nil(t, err)
	require.NotNil(t, sctx)

	vmOutput = &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(10)}
	sctx, _ = sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmcommon.DirectCall,
	)
	require.Nil(t, err)
	require.NotNil(t, sctx)
}

func TestScProcessor_RefundGasToSender(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(10)
	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &mock.FeeHandlerStub{MinGasPriceCalled: func() uint64 {
		return minGasPrice
	}}
	arguments.DeployEnableEpoch = 10
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
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()

	refundGas := big.NewInt(10)
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: refundGas}
	scr, _ := sc.createSCRForSenderAndRelayer(
		vmOutput,
		tx,
		txHash,
		vmcommon.DirectCall,
	)
	require.Nil(t, err)

	finalValue := big.NewInt(0).Mul(refundGas, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))
	require.Equal(t, currBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
	require.Equal(t, scr.Value.Cmp(finalValue), 0)
}

func TestScProcessor_DoNotRefundGasToSenderForAsyncCall(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(10)
	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &mock.FeeHandlerStub{MinGasPriceCalled: func() uint64 {
		return minGasPrice
	}}
	arguments.DeployEnableEpoch = 10
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
		vmcommon.AsynchronousCall,
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

	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acntSrc, nil
	}

	tx.Value = big.NewInt(0)
	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, err = sc.processVMOutput(vmOutput, txHash, tx, vmcommon.DirectCall, 0)
	require.Nil(t, err)
}

func TestScProcessor_processSCOutputAccounts(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}

	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}
	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMOutput{}, vmcommon.DirectCall, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Code = []byte("contract-code")
	outacc1.Nonce = 5
	outacc1.BalanceDelta = big.NewInt(int64(5))
	outputAccounts = append(outputAccounts, outacc1)

	testAddr := outaddress
	testAcc, _ := state.NewUserAccount(testAddr)

	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		if bytes.Equal(address, testAddr) {
			return testAcc, nil
		}
		return nil, state.ErrAccNotFound
	}

	accountsDB.SaveAccountCalled = func(accountHandler state.AccountHandler) error {
		return nil
	}

	tx.Value = big.NewInt(int64(5))
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMOutput{}, vmcommon.DirectCall, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outacc1.BalanceDelta = nil
	outacc1.Nonce++
	tx.Value = big.NewInt(0)
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMOutput{}, vmcommon.DirectCall, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outacc1.Nonce++
	outacc1.BalanceDelta = big.NewInt(int64(10))
	tx.Value = big.NewInt(int64(10))

	currentBalance := testAcc.Balance.Uint64()
	vmOutBalance := outacc1.BalanceDelta.Uint64()
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMOutput{}, vmcommon.DirectCall, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)
	require.Equal(t, currentBalance+vmOutBalance, testAcc.Balance.Uint64())
}

func TestScProcessor_processSCOutputAccountsNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.ShardCoordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}
	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMOutput{}, vmcommon.DirectCall, outputAccounts, tx, []byte("hash"))
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

	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMOutput{}, vmcommon.DirectCall, outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outacc1.BalanceDelta = big.NewInt(-50)
	_, _, err = sc.processSCOutputAccounts(&vmcommon.VMOutput{}, vmcommon.DirectCall, outputAccounts, tx, []byte("hash"))
	require.Equal(t, err, process.ErrNegativeBalanceDeltaOnCrossShardAccount)
}

func TestScProcessor_CreateCrossShardTransactions(t *testing.T) {
	t.Parallel()

	testAccounts, _ := state.NewUserAccount([]byte("address"))
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, err error) {
			return testAccounts, nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
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

	createdAsyncSCR, scTxs, err := sc.processSCOutputAccounts(&vmcommon.VMOutput{}, vmcommon.DirectCall, outputAccounts, tx, txHash)
	require.Nil(t, err)
	require.Equal(t, len(outputAccounts), len(scTxs))
	require.False(t, createdAsyncSCR)
}

func TestScProcessor_CreateCrossShardTransactionsWithAsyncCalls(t *testing.T) {
	t.Parallel()

	testAccounts, _ := state.NewUserAccount([]byte("address"))
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, err error) {
			return testAccounts, nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
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

	createdAsyncSCR, scTxs, err := sc.processSCOutputAccounts(&vmcommon.VMOutput{GasRemaining: 1000}, vmcommon.AsynchronousCall, outputAccounts, tx, txHash)
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
	createdAsyncSCR, scTxs, err = sc.processSCOutputAccounts(&vmcommon.VMOutput{GasRemaining: 1000}, vmcommon.AsynchronousCall, outputAccounts, tx, txHash)
	require.Nil(t, err)
	require.Equal(t, len(outputAccounts), len(scTxs))
	require.True(t, createdAsyncSCR)

	lastScTx := scTxs[len(scTxs)-1].(*smartContractResult.SmartContractResult)
	require.Equal(t, vmcommon.AsynchronousCallBack, lastScTx.CallType)

	tx.Value = big.NewInt(0)
	scTxs, err = sc.processVMOutput(&vmcommon.VMOutput{GasRemaining: 1000}, txHash, tx, vmcommon.AsynchronousCall, 10000)
	require.Nil(t, err)
	require.Equal(t, 1, len(scTxs))
	lastScTx = scTxs[len(scTxs)-1].(*smartContractResult.SmartContractResult)
	require.Equal(t, vmcommon.AsynchronousCallBack, lastScTx.CallType)
}

func TestScProcessor_ProcessSmartContractResultNilScr(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
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
	accountsDB := &mock.AccountsStub{LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
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

	accountsDB := &mock.AccountsStub{}
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

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return &mock.AccountWrapMock{}, nil
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

	userAcc, _ := state.NewUserAccount([]byte("recv address"))
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			if bytes.Equal(address, userAcc.Address) {
				return userAcc, nil
			}
			return state.NewEmptyUserAccount(), nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
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
	arguments.BlockChainHook = &mock.BlockChainHookHandlerMock{
		IsPayableCalled: func(address []byte) (bool, error) {
			return false, nil
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		RcvAddr: userAcc.Address,
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

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return state.NewUserAccount(address)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
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
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return state.NewUserAccount(address)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
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
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return state.NewUserAccount(address)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
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

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return state.NewUserAccount(address)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
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
	arguments.TxTypeHandler = &mock.TxTypeHandlerMock{
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
	dstScAddress, _ := state.NewUserAccount(scAddress)
	dstScAddress.SetCode([]byte("code"))
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			if bytes.Equal(scAddress, address) {
				return dstScAddress, nil
			}
			return nil, nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
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
	arguments.TxTypeHandler = &mock.TxTypeHandlerMock{
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

func TestScProcessor_ProcessRelayedSCRValueBackToRelayer(t *testing.T) {
	t.Parallel()

	scAddress := []byte("000000000001234567890123456789012")
	dstScAddress, _ := state.NewUserAccount(scAddress)
	dstScAddress.SetCode([]byte("code"))

	baseValue := big.NewInt(100)
	userAddress := []byte("111111111111234567890123456789012")
	userAcc, _ := state.NewUserAccount(userAddress)
	_ = userAcc.AddToBalance(baseValue)
	relayedAddress := []byte("211111111111234567890123456789012")
	relayedAcc, _ := state.NewUserAccount(relayedAddress)

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
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
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
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
	arguments.TxTypeHandler = &mock.TxTypeHandlerMock{
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
	contract, err := state.NewUserAccount([]byte("contract"))
	require.Nil(t, err)
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
	sc, _ := NewSmartContractProcessor(arguments)

	gasProvided := uint64(1000)
	maxGasToRemain := gasProvided - (gasProvided / process.MaxGasFeeHigherFactorAccepted)

	callType := vmcommon.DirectCall
	vmOutput := &vmcommon.VMOutput{
		GasRemaining: maxGasToRemain,
	}
	sc.penalizeUserIfNeeded(&transaction.Transaction{}, []byte("txHash"), callType, gasProvided, vmOutput)
	assert.Equal(t, maxGasToRemain, vmOutput.GasRemaining)

	callType = vmcommon.AsynchronousCall
	vmOutput = &vmcommon.VMOutput{
		GasRemaining: maxGasToRemain + 1,
	}
	sc.penalizeUserIfNeeded(&transaction.Transaction{}, []byte("txHash"), callType, gasProvided, vmOutput)
	assert.Equal(t, maxGasToRemain+1, vmOutput.GasRemaining)

	callType = vmcommon.DirectCall
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
	sc, _ := NewSmartContractProcessor(arguments)

	gasProvided := uint64(1000)
	maxGasToRemain := gasProvided - (gasProvided / process.MaxGasFeeHigherFactorAccepted)

	callType := vmcommon.DirectCall
	vmOutput := &vmcommon.VMOutput{
		GasRemaining: maxGasToRemain + 1,
	}

	sc.penalizedTooMuchGasEnableEpoch = 1

	sc.EpochConfirmed(0)
	sc.penalizeUserIfNeeded(&transaction.Transaction{}, []byte("txHash"), callType, gasProvided, vmOutput)
	assert.Equal(t, maxGasToRemain+1, vmOutput.GasRemaining)

	sc.EpochConfirmed(1)
	sc.penalizeUserIfNeeded(&transaction.Transaction{}, []byte("txHash"), callType, gasProvided, vmOutput)
	assert.Equal(t, uint64(0), vmOutput.GasRemaining)
}

func TestSCProcessor_createSCRWhenError(t *testing.T) {
	arguments := createMockSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessor(arguments)

	acntSnd := &mock.UserAccountStub{}
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
		&smartContractResult.SmartContractResult{CallType: vmcommon.AsynchronousCall},
		"string",
		[]byte("msg"),
		0)
	assert.Equal(t, uint64(0), scr.GasLimit)
	assert.Equal(t, consumedFee.Cmp(big.NewInt(0)), 0)
	assert.Equal(t, "@04@6d7367", string(scr.Data))

	sc.asyncCallbackGasLock = 10
	sc.asyncCallStepCost = 10
	scr, consumedFee = sc.createSCRsWhenError(
		acntSnd,
		[]byte("txHash"),
		&smartContractResult.SmartContractResult{CallType: vmcommon.AsynchronousCall, GasPrice: 1, GasLimit: 100},
		"string",
		[]byte("msg"),
		20)
	assert.Equal(t, uint64(1), scr.GasPrice)
	assert.Equal(t, consumedFee.Cmp(big.NewInt(80)), 0)
	assert.Equal(t, "@04@6d7367", string(scr.Data))
	assert.Equal(t, uint64(20), scr.GasLimit)

	sc.asyncCallbackGasLock = 100
	sc.asyncCallStepCost = 100
	scr, consumedFee = sc.createSCRsWhenError(
		acntSnd,
		[]byte("txHash"),
		&smartContractResult.SmartContractResult{CallType: vmcommon.AsynchronousCall, GasPrice: 1, GasLimit: 100},
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
	sc, _ := NewSmartContractProcessor(arguments)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Nonce = 0
	outacc1.Balance = big.NewInt(5)
	outacc1.BalanceDelta = big.NewInt(15)
	outTransfer := vmcommon.OutputTransfer{
		Value:     big.NewInt(5),
		CallType:  vmcommon.AsynchronousCall,
		GasLocked: 100,
		GasLimit:  100,
		Data:      []byte("functionCall"),
	}
	outacc1.OutputTransfers = append(outacc1.OutputTransfers, outTransfer)
	vmOutput := &vmcommon.VMOutput{
		OutputAccounts: make(map[string]*vmcommon.OutputAccount),
	}
	vmOutput.OutputAccounts[string(outaddress)] = outacc1

	asyncCallback, results := sc.createSmartContractResults(vmOutput, vmcommon.DirectCall, outacc1, &transaction.Transaction{}, []byte("hash"))
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
}

func TestSmartContractProcessor_computeTotalConsumedFeeAndDevRwd(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = NewArgumentParser()
	shardCoordinator := &mock.CoordinatorStub{ComputeIdCalled: func(address []byte) uint32 {
		return 0
	}}
	feeHandler := &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return 0
		},
		ComputeTxFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
		},
		ComputeFeeForProcessingCalled: func(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
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

	feeHandler.ComputeGasLimitCalled = func(tx process.TransactionWithFeeHandler) uint64 {
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
	acc, err := state.NewUserAccount(scAccountAddress)
	require.Nil(t, err)
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
	arguments.AccountsDB = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
		LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return acc, nil
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
	expectedTotalFee, expectedDevFees := computeExpectedResults(args, tx, builtInGasUsed, vmoutput)
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
			acc, err := state.NewUserAccount(scAccountAddress)
			require.Nil(t, err)
			require.NotNil(t, acc)

			arguments := createMockSmartContractProcessorArguments()
			arguments.ArgsParser = NewArgumentParser()
			shardCoordinator := &mock.CoordinatorStub{ComputeIdCalled: func(address []byte) uint32 {
				return 0
			}}

			// use a real fee handler
			args := createRealEconomicsDataArgs()
			arguments.EconomicsFee, err = economics.NewEconomicsData(*args)
			require.Nil(t, err)

			arguments.TxFeeHandler, err = postprocess.NewFeeAccumulator()
			require.Nil(t, err)

			arguments.ShardCoordinator = shardCoordinator
			arguments.AccountsDB = &mock.AccountsStub{
				RevertToSnapshotCalled: func(snapshot int) error {
					return nil
				},
				LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
					return acc, nil
				},
			}

			sc, err := NewSmartContractProcessor(arguments)
			require.Nil(t, err)
			require.NotNil(t, sc)

			expectedTotalFee, expectedDevFees := computeExpectedResults(args, test.tx, test.builtInGasUsed, test.vmOutput)

			retcode, err := sc.finishSCExecution(nil, []byte("txhash"), test.tx, test.vmOutput, test.builtInGasUsed)
			require.Nil(t, err)
			require.Equal(t, retcode, vmcommon.Ok)
			require.Nil(t, err)
			require.Equal(t, expectedDevFees, acc.DeveloperReward)
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
	arguments.EconomicsFee = &mock.FeeHandlerStub{ComputeFeeForProcessingCalled: func(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
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
		CallType:       vmcommon.DirectCall,
	}

	vmOutput := &vmcommon.VMOutput{GasRemaining: 1000}
	_, relayerRefund := sc.createSCRForSenderAndRelayer(vmOutput, scrWithRelayed, []byte("txhash"), vmcommon.DirectCall)
	assert.NotNil(t, relayerRefund)

	senderID := sc.shardCoordinator.ComputeId(relayerRefund.SndAddr)
	assert.Equal(t, sc.shardCoordinator.SelfId(), senderID)
}

func TestScProcessor_ProcessIfErrorRevertAccountFails(t *testing.T) {
	t.Parallel()
	expectedError := errors.New("expected error")
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return expectedError
		},
	}

	sc, _ := NewSmartContractProcessor(arguments)

	sndAccount := &mock.UserAccountStub{}
	err := sc.ProcessIfError(sndAccount, []byte("txHash"), nil, "0", []byte("message"), 1, 100)
	require.NotNil(t, err)
	require.Equal(t, expectedError, err)
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
	arguments.EconomicsFee = &mock.FeeHandlerStub{
		ComputeFeeForProcessingCalled: func(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
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

	sndAccount := &mock.UserAccountStub{}
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
	arguments.EconomicsFee = &mock.FeeHandlerStub{
		ComputeFeeForProcessingCalled: func(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
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

	sc.EpochConfirmed(100)

	tx := &transaction.Transaction{
		SndAddr: []byte("snd"),
		RcvAddr: []byte("rcv"),
		Value:   big.NewInt(15),
	}

	err := sc.ProcessIfError(nil, []byte("txHash"), tx, "0", []byte("message"), 1, 100)
	require.Nil(t, err)
	require.False(t, called)
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
				LeaderPercentage:                 0.1,
				DeveloperPercentage:              0.3,
				ProtocolSustainabilityPercentage: 0.1,
				ProtocolSustainabilityAddress:    "erd1j25xk97yf820rgdp3mj5scavhjkn6tjyn0t63pmv5qyjj7wxlcfqqe2rw5",
				TopUpGradientPoint:               "3000000000000000000000000",
				TopUpFactor:                      0.25,
			},
			FeeSettings: config.FeeSettings{
				MaxGasLimitPerBlock:     "1500000000",
				MaxGasLimitPerMetaBlock: "15000000000",
				GasPerDataByte:          "1500",
				MinGasPrice:             "1000000000",
				MinGasLimit:             "50000",
				GasPriceModifier:        0.01,
			},
		},
		EpochNotifier:                  &mock.EpochNotifierStub{},
		PenalizedTooMuchGasEnableEpoch: 0,
		GasPriceModifierEnableEpoch:    0,
	}
}

func computeExpectedResults(args *economics.ArgsNewEconomicsData, tx *transaction.Transaction, builtInGasUsed uint64, vmoutput *vmcommon.VMOutput) (*big.Int, *big.Int) {
	minGasLimitBigInt, _ := big.NewInt(0).SetString(args.Economics.FeeSettings.MinGasLimit, 10)
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
	expectedDevFees := core.GetPercentageOfValue(processFee, args.Economics.RewardsSettings.DeveloperPercentage)
	return expectedTotalFee, expectedDevFees
}
