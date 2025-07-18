package txsFee

import (
	"encoding/hex"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationTests/vm/txFee")

func prepareTestContextForEpoch836(tb testing.TB) (*vm.VMTestContext, []byte) {
	unreachableEpoch := uint32(999999)
	db := integrationtests.CreateStorer(tb.TempDir())
	gasScheduleDir := "../../../cmd/node/config/gasSchedules"

	cfg := config.GasScheduleByEpochs{
		StartEpoch: 0,
		FileName:   "gasScheduleV6.toml",
	}

	argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig: config.GasScheduleConfig{
			GasScheduleByEpochs: []config.GasScheduleByEpochs{cfg},
		},
		ConfigDir:          gasScheduleDir,
		EpochNotifier:      forking.NewGenericEpochNotifier(),
		WasmVMChangeLocker: &sync.RWMutex{},
	}
	gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasScheduleNotifier)
	require.Nil(tb, err)

	testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGasAndRoundConfig(
		config.EnableEpochs{
			GovernanceEnableEpoch:                           unreachableEpoch,
			SetSenderInEeiOutputTransferEnableEpoch:         unreachableEpoch,
			RefactorPeersMiniBlocksEnableEpoch:              unreachableEpoch,
			MaxBlockchainHookCountersEnableEpoch:            unreachableEpoch,
			DynamicGasCostForDataTrieStorageLoadEnableEpoch: unreachableEpoch,
		},
		mock.NewMultiShardsCoordinatorMock(2),
		db,
		gasScheduleNotifier,
		testscommon.GetDefaultRoundsConfig(),
		1,
	)
	require.Nil(tb, err)

	senderBalance := big.NewInt(1000000000000000000)
	scAddress, _ := utils.DoColdDeploy(
		tb,
		testContext,
		"../wasm/testdata/distributeRewards/code.wasm",
		senderBalance,
		"0100",
	)
	utils.OverwriteAccountStorageWithHexFileContent(tb, testContext, scAddress, "../wasm/testdata/distributeRewards/data.hex")
	utils.CleanAccumulatedIntermediateTransactions(tb, testContext)

	db.ClearCache()

	return testContext, scAddress
}

func TestScCallShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	}, 1)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9988100, 11900, 399)
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(1000000000)
	gasLimit := uint64(100000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	for idx := uint64(0); idx < 10; idx++ {
		tx := vm.CreateTransaction(idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

		calculatedGasLimit := vm.ComputeGasLimit(nil, testContext, tx)
		require.Equal(t, uint64(15704), calculatedGasLimit)

		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		require.Equal(t, vmcommon.Ok, returnCode)

		_, errCommit := testContext.Accounts.Commit()
		require.Nil(t, errCommit)
	}

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(11), ret)

	expectedBalance := big.NewInt(998429600)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 10, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(1582300), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(157339), developerFees)
}

func TestScCallContractNotFoundShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress := "00000000000000000500dbb53e4b23392b0d6f36cce32deb2d623e9625ab3132"
	scAddrBytes, _ := hex.DecodeString(scAddress)
	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddrBytes, gasPrice, gasLimit, []byte("increment"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(90000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10000), accumulatedFees)
}

func TestScCallInvalidMethodToCallShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9988100, 11900, 399)
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("invalidMethod"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(1), ret)

	expectedBalance := big.NewInt(90000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(21900), accumulatedFees)
}

func TestScCallInsufficientGasLimitShouldNotConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9988100, 11900, 399)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasLimit := uint64(9)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))
	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, process.ErrInsufficientGasLimitInTx, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(1), ret)

	expectedBalance := big.NewInt(100000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 0, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(11900), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(399), developerFees)
}

func TestScCallOutOfGasShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9988100, 11900, 399)
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasLimit := uint64(20)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(1), ret)

	expectedBalance := big.NewInt(99800)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(12100), accumulatedFees)
}

func TestScCallAndGasChangeShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	}, 1)
	require.Nil(t, err)
	defer testContext.Close()

	mockGasSchedule := testContext.GasSchedule.(*mock.GasScheduleNotifierMock)

	scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9988100, 11900, 399)
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(10000000)
	gasLimit := uint64(100000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	numIterations := uint64(10)
	for idx := uint64(0); idx < numIterations; idx++ {
		tx := vm.CreateTransaction(idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		require.Equal(t, vmcommon.Ok, returnCode)

		_, errCommit := testContext.Accounts.Commit()
		require.Nil(t, errCommit)
	}

	newGasSchedule := wasmConfig.MakeGasMapForTests()
	newGasSchedule["WASMOpcodeCost"] = wasmConfig.FillGasMapWASMOpcodeValues(2)
	mockGasSchedule.ChangeGasSchedule(newGasSchedule)

	for idx := uint64(0); idx < numIterations; idx++ {
		tx := vm.CreateTransaction(numIterations+idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		require.Equal(t, vmcommon.Ok, returnCode)

		_, errCommit := testContext.Accounts.Commit()
		require.Nil(t, errCommit)
	}
}

func TestESDTScCallAndGasChangeShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContext.Close()

	owner := []byte("12345678901234567890123456789011")
	senderBalance := big.NewInt(1000000000)
	gasLimit := uint64(2000000)

	_, _ = vm.CreateAccount(testContext.Accounts, owner, 0, senderBalance)
	ownerAccount, _ := testContext.Accounts.LoadAccount(owner)
	scAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/forwarder-raw.wasm", ownerAccount, gasPrice, gasLimit, nil, big.NewInt(0))
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance = big.NewInt(10000000)
	gasLimit = uint64(30000)

	localEsdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, senderBalance, token, 0, localEsdtBalance, uint32(core.Fungible))

	txData := txDataBuilder.NewBuilder()
	valueToSendToSc := int64(1000)
	txData.TransferESDT(string(token), valueToSendToSc).Str("forward_direct_esdt_via_transf_exec").Bytes(sndAddr)
	numIterations := uint64(10)
	for idx := uint64(0); idx < numIterations; idx++ {
		tx := vm.CreateTransaction(idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, txData.ToBytes())

		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		require.Equal(t, vmcommon.Ok, returnCode)

		_, errCommit := testContext.Accounts.Commit()
		require.Nil(t, errCommit)
	}

	mockGasSchedule := testContext.GasSchedule.(*mock.GasScheduleNotifierMock)
	testGasSchedule := wasmConfig.MakeGasMapForTests()
	newGasSchedule := defaults.FillGasMapInternal(testGasSchedule, 1)
	newGasSchedule["BuiltInCost"][core.BuiltInFunctionESDTTransfer] = 2
	newGasSchedule[common.BaseOpsAPICost]["TransferValue"] = 2
	mockGasSchedule.ChangeGasSchedule(newGasSchedule)

	for idx := uint64(0); idx < numIterations; idx++ {
		tx := vm.CreateTransaction(numIterations+idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, txData.ToBytes())

		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		require.Equal(t, vmcommon.Ok, returnCode)

		_, errCommit := testContext.Accounts.Commit()
		require.Nil(t, errCommit)
	}
}

func prepareTestContextForEpoch460(tb testing.TB) (*vm.VMTestContext, []byte) {
	unreachableEpoch := uint32(999999)

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		GovernanceEnableEpoch:                             unreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:                    unreachableEpoch,
		CorrectJailedNotUnstakedEmptyQueueEpoch:           unreachableEpoch,
		OptimizeNFTStoreEnableEpoch:                       unreachableEpoch,
		CreateNFTThroughExecByCallerEnableEpoch:           unreachableEpoch,
		StopDecreasingValidatorRatingWhenStuckEnableEpoch: unreachableEpoch,
		FrontRunningProtectionEnableEpoch:                 unreachableEpoch,
		IsPayableBySCEnableEpoch:                          unreachableEpoch,
		CleanUpInformativeSCRsEnableEpoch:                 unreachableEpoch,
		StorageAPICostOptimizationEnableEpoch:             unreachableEpoch,
		TransformToMultiShardCreateEnableEpoch:            unreachableEpoch,
		ESDTRegisterAndSetAllRolesEnableEpoch:             unreachableEpoch,
		DoNotReturnOldBlockInBlockchainHookEnableEpoch:    unreachableEpoch,
		AddFailedRelayedTxToInvalidMBsDisableEpoch:        unreachableEpoch,
		SCRSizeInvariantOnBuiltInResultEnableEpoch:        unreachableEpoch,
		CheckCorrectTokenIDForTransferRoleEnableEpoch:     unreachableEpoch,
		DisableExecByCallerEnableEpoch:                    unreachableEpoch,
		FailExecutionOnEveryAPIErrorEnableEpoch:           unreachableEpoch,
		ManagedCryptoAPIsEnableEpoch:                      unreachableEpoch,
		RefactorContextEnableEpoch:                        unreachableEpoch,
		CheckFunctionArgumentEnableEpoch:                  unreachableEpoch,
		CheckExecuteOnReadOnlyEnableEpoch:                 unreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:              unreachableEpoch,
		ESDTMetadataContinuousCleanupEnableEpoch:          unreachableEpoch,
		FixAsyncCallBackArgsListEnableEpoch:               unreachableEpoch,
		FixOldTokenLiquidityEnableEpoch:                   unreachableEpoch,
		SetSenderInEeiOutputTransferEnableEpoch:           unreachableEpoch,
		RefactorPeersMiniBlocksEnableEpoch:                unreachableEpoch,
		RuntimeMemStoreLimitEnableEpoch:                   unreachableEpoch,
		MaxBlockchainHookCountersEnableEpoch:              unreachableEpoch,
	}, 1)
	require.Nil(tb, err)

	senderBalance := big.NewInt(1000000000000000000)
	gasLimit := uint64(100000)
	params := []string{"01"}
	scAddress, _ := utils.DoDeployWithCustomParams(
		tb,
		testContext,
		"../wasm/testdata/buyNFTCall/code.wasm",
		senderBalance,
		gasLimit,
		params,
	)
	utils.OverwriteAccountStorageWithHexFileContent(tb, testContext, scAddress, "../wasm/testdata/buyNFTCall/data.hex")
	utils.CleanAccumulatedIntermediateTransactions(tb, testContext)

	return testContext, scAddress
}

func TestScCallBuyNFT_OneFailedTxAndOneOkTx(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, scAddress := prepareTestContextForEpoch460(t)
	defer testContext.Close()

	sndAddr1 := []byte("12345678901234567890123456789112")
	sndAddr2 := []byte("12345678901234567890123456789113")
	senderBalance := big.NewInt(1000000000000000000)
	gasLimit := uint64(1000000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr1, 0, senderBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr2, 0, senderBalance)

	blockChainHook := testContext.BlockchainHook.(process.BlockChainHookHandler)
	t.Run("transaction that fails", func(t *testing.T) {
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)
		err := blockChainHook.SetCurrentHeader(&block.Header{
			TimeStamp: 1635880560,
		})
		require.Nil(t, err)

		txData, errDecode := hex.DecodeString("6275794e6674406338403435353035353465346235333264333433363632333133383336406533")
		require.Nil(t, errDecode)
		tx := vm.CreateTransaction(0, big.NewInt(250000000000000000), sndAddr1, scAddress, gasPrice, gasLimit, txData)

		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		require.Equal(t, vmcommon.UserError, returnCode)

		_, errCommit := testContext.Accounts.Commit()
		require.Nil(t, errCommit)

		logs := testContext.TxsLogsProcessor.GetAllCurrentLogs()
		assert.Equal(t, 2, len(logs))

		logEvents := logs[1].GetLogEvents()
		assert.Equal(t, 2, len(logEvents))

		topics := logEvents[0].GetTopics()
		assert.Equal(t, 2, len(topics))
		assert.Equal(t, vmhost.ErrInvalidTokenIndex.Error(), string(topics[1]))
	})
	t.Run("transaction that succeed", func(t *testing.T) {
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)
		err := blockChainHook.SetCurrentHeader(&block.Header{
			TimeStamp: 1635880566, // next timestamp
		})
		require.Nil(t, err)

		txData, errDecode := hex.DecodeString("6275794e6674403264403435353035353465346235333264333433363632333133383336403337")
		require.Nil(t, errDecode)
		tx := vm.CreateTransaction(0, big.NewInt(250000000000000000), sndAddr2, scAddress, gasPrice, gasLimit, txData)

		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		assert.Equal(t, vmcommon.UserError, returnCode)

		_, errCommit := testContext.Accounts.Commit()
		require.Nil(t, errCommit)

		logs := testContext.TxsLogsProcessor.GetAllCurrentLogs()
		assert.Equal(t, 3, len(logs))

		logEvents := logs[1].GetLogEvents()
		assert.Equal(t, 2, len(logEvents))

		topics := logEvents[0].GetTopics()
		assert.Equal(t, 2, len(topics))
		assert.Equal(t, vmhost.ErrInvalidTokenIndex.Error(), string(topics[1]))
	})
}

func TestScCallBuyNFT_TwoOkTxs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, scAddress := prepareTestContextForEpoch460(t)
	defer testContext.Close()

	sndAddr1 := []byte("12345678901234567890123456789112")
	sndAddr2 := []byte("12345678901234567890123456789113")
	senderBalance := big.NewInt(1000000000000000000)
	gasLimit := uint64(1000000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr1, 0, senderBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr2, 0, senderBalance)

	blockChainHook := testContext.BlockchainHook.(process.BlockChainHookHandler)
	t.Run("first transaction that succeed", func(t *testing.T) {
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)
		err := blockChainHook.SetCurrentHeader(&block.Header{
			TimeStamp: 1635880566, // next timestamp
		})
		require.Nil(t, err)

		txData, errDecode := hex.DecodeString("6275794e6674403264403435353035353465346235333264333433363632333133383336403337")
		require.Nil(t, errDecode)
		tx := vm.CreateTransaction(0, big.NewInt(250000000000000000), sndAddr1, scAddress, gasPrice, gasLimit, txData)

		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		assert.Equal(t, vmcommon.UserError, returnCode)

		_, errCommit := testContext.Accounts.Commit()
		require.Nil(t, errCommit)

		logs := testContext.TxsLogsProcessor.GetAllCurrentLogs()

		logEvents := logs[1].GetLogEvents()
		assert.Equal(t, 2, len(logEvents))

		topics := logEvents[0].GetTopics()
		assert.Equal(t, 2, len(topics))
		assert.Equal(t, string(sndAddr1), string(topics[0]))
		assert.Equal(t, vmhost.ErrInvalidTokenIndex.Error(), string(topics[1]))
	})
	t.Run("second transaction that succeed", func(t *testing.T) {
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)
		err := blockChainHook.SetCurrentHeader(&block.Header{
			TimeStamp: 1635880572, // next timestamp
		})
		require.Nil(t, err)

		txData, errDecode := hex.DecodeString("6275794e6674403434403435353035353465346235333264333433363632333133383336403531")
		require.Nil(t, errDecode)
		tx := vm.CreateTransaction(0, big.NewInt(250000000000000000), sndAddr2, scAddress, gasPrice, gasLimit, txData)

		returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, errProcess)
		assert.Equal(t, vmcommon.UserError, returnCode)

		_, errCommit := testContext.Accounts.Commit()
		require.Nil(t, errCommit)

		logs := testContext.TxsLogsProcessor.GetAllCurrentLogs()

		logEvents := logs[1].GetLogEvents()
		assert.Equal(t, 2, len(logEvents))

		topics := logEvents[0].GetTopics()
		assert.Equal(t, 2, len(topics))
		assert.Equal(t, string(sndAddr1), string(topics[0]))
		assert.Equal(t, vmhost.ErrInvalidTokenIndex.Error(), string(topics[1]))
	})
}

func TestScCallDistributeStakingRewards_ShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, scAddress := prepareTestContextForEpoch836(t)
	defer testContext.Close()

	pkConv, err := pubkeyConverter.NewBech32PubkeyConverter(32, integrationTests.AddressHrp)
	require.NoError(t, err)
	sndAddr1, err := pkConv.Decode("erd1rkhyj0ne054upekymjafwas44v2trdykd22vcg27ap8x2hpg5u7q0296ne")
	require.Nil(t, err)

	senderBalance := big.NewInt(1000000000000000000)
	gasLimit := uint64(600000000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr1, 0, senderBalance)

	blockChainHook := testContext.BlockchainHook.(process.BlockChainHookHandler)

	utils.CleanAccumulatedIntermediateTransactions(t, testContext)
	err = blockChainHook.SetCurrentHeader(&block.Header{
		TimeStamp: 1668430842,
	})
	require.Nil(t, err)

	txData, errDecode := hex.DecodeString("646973747269627574655f7374616b696e675f72657761726473")
	require.Nil(t, errDecode)
	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr1, scAddress, gasPrice, gasLimit, txData)

	startTime := time.Now()
	returnCode, errProcess := testContext.TxProcessor.ProcessTransaction(tx)
	endTime := time.Now()
	assert.Nil(t, errProcess)
	assert.Equal(t, vmcommon.Ok, returnCode)

	_, errCommit := testContext.Accounts.Commit()
	require.Nil(t, errCommit)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediateTxs))
	log.Info(integrationtests.TransactionHandlerToString(pkConv, intermediateTxs...))
	log.Info("transaction took", "time", endTime.Sub(startTime))

	scr := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	assert.Equal(t, "", string(scr.ReturnMessage))
	assert.Equal(t, sndAddr1, intermediateTxs[0].(*smartContractResult.SmartContractResult).RcvAddr)
}
