//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const upgradeContractFunction = "upgradeContract"

func TestAsyncCallShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	balance := big.NewInt(100000000)
	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, balance)
	_, _ = vm.CreateAccount(testContext.Accounts, senderAddr, 0, balance)

	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	pathToContract := "testdata/first/output/first.wasm"
	firstScAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(50))

	gasLimit := uint64(5000000)
	args := [][]byte{[]byte(hex.EncodeToString(firstScAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	pathToContract = "testdata/second/output/async.wasm"
	secondSCAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(50))

	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
	testContext.TxFeeHandler.CreateBlockStarted(gasAndFees)

	tx := vm.CreateTransaction(0, big.NewInt(0), senderAddr, secondSCAddress, gasPrice, gasLimit, []byte("doSomething"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, firstScAddress, "numCalled")
	require.Equal(t, big.NewInt(1), res)

	require.Equal(t, big.NewInt(50000000), testContext.TxFeeHandler.GetAccumulatedFees())
	require.Equal(t, big.NewInt(4999988), testContext.TxFeeHandler.GetDeveloperFees())
}

func TestMinterContractWithAsyncCalls(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMsAndCustomGasSchedule(config.EnableEpochs{}, func(gasMap wasmConfig.GasScheduleMap) {
		// if `MaxBuiltInCallsPerTx` is 200 test will fail
		gasMap[common.MaxPerTransaction]["MaxBuiltInCallsPerTx"] = 199
		gasMap[common.MaxPerTransaction]["MaxNumberOfTransfersPerTx"] = 100000
		gasMap[common.MaxPerTransaction]["MaxNumberOfTrieReadsPerTx"] = 100000
	})
	require.Nil(t, err)
	defer testContext.Close()

	balance := big.NewInt(1000000000000)
	ownerAddr := []byte("12345678901234567890123456789011")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, balance)

	token := []byte("miiutoken")
	roles := [][]byte{[]byte(core.ESDTRoleNFTCreate)}

	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(500000)
	pathToContract := "testdata/minter/minter.wasm"
	firstScAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(0))

	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	secondContractAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(0))

	utils.SetESDTRoles(t, testContext.Accounts, firstScAddress, token, roles)
	utils.SetESDTRoles(t, testContext.Accounts, secondContractAddress, token, roles)

	// DO call
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	dataField := []byte(fmt.Sprintf("setTokenID@%s", hex.EncodeToString(token)))
	tx := vm.CreateTransaction(ownerAccount.GetNonce(), big.NewInt(0), ownerAddr, firstScAddress, gasPrice, deployGasLimit, dataField)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	tx = vm.CreateTransaction(ownerAccount.GetNonce(), big.NewInt(0), ownerAddr, secondContractAddress, gasPrice, deployGasLimit, dataField)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	dataField = []byte(fmt.Sprintf("tryMore@%s", hex.EncodeToString(firstScAddress)))
	gasLimit := 600000000
	tx = vm.CreateTransaction(ownerAccount.GetNonce(), big.NewInt(0), ownerAddr, secondContractAddress, gasPrice, uint64(gasLimit), dataField)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	logs := testContext.TxsLogsProcessor.GetAllCurrentLogs()
	event := logs[4].GetLogEvents()[1]
	require.Equal(t, "internalVMErrors", string(event.GetIdentifier()))
	require.Contains(t, string(event.GetData()), process.ErrMaxCallsReached.Error())
}

func TestAsyncCallsOnInitFunctionOnUpgrade(t *testing.T) {
	firstContractCode := wasm.GetSCCode("./testdata/first/output/first.wasm")
	newContractCode := wasm.GetSCCode("./testdata/asyncOnInit/asyncOnInitAndUpgrade.wasm")

	t.Run("backwards compatibility for unset flag", func(t *testing.T) {
		gasScheduleNotifier := vm.CreateMockGasScheduleNotifier()

		expectedGasLimit := gasScheduleNotifier.LatestGasSchedule()[common.BaseOpsAPICost][common.AsyncCallbackGasLockField] +
			gasScheduleNotifier.LatestGasSchedule()[common.BaseOpsAPICost][common.AsyncCallStepField] +
			gasScheduleNotifier.LatestGasSchedule()[common.BaseOperationCost]["AoTPreparePerByte"]*uint64(len(firstContractCode))/2

		enableEpoch := config.EnableEpochs{
			RuntimeCodeSizeFixEnableEpoch:                   100000, // fix not activated
			DynamicGasCostForDataTrieStorageLoadEnableEpoch: 100000,
			SCProcessorV2EnableEpoch:                        integrationTests.UnreachableEpoch,
		}

		testAsyncCallsOnInitFunctionOnUpgrade(t, enableEpoch, expectedGasLimit, gasScheduleNotifier, newContractCode)
	})
	t.Run("fix activated", func(t *testing.T) {
		gasScheduleNotifier := vm.CreateMockGasScheduleNotifier()

		expectedGasLimit := gasScheduleNotifier.LatestGasSchedule()[common.BaseOpsAPICost][common.AsyncCallbackGasLockField] +
			gasScheduleNotifier.LatestGasSchedule()[common.BaseOpsAPICost][common.AsyncCallStepField] +
			gasScheduleNotifier.LatestGasSchedule()[common.BaseOperationCost]["AoTPreparePerByte"]*uint64(len(newContractCode))/2

		enableEpoch := config.EnableEpochs{
			RuntimeCodeSizeFixEnableEpoch:                   0, // fix activated
			DynamicGasCostForDataTrieStorageLoadEnableEpoch: 100000,
			SCProcessorV2EnableEpoch:                        integrationTests.UnreachableEpoch,
		}

		testAsyncCallsOnInitFunctionOnUpgrade(t, enableEpoch, expectedGasLimit, gasScheduleNotifier, newContractCode)
	})
}

func testAsyncCallsOnInitFunctionOnUpgrade(
	t *testing.T,
	enableEpochs config.EnableEpochs,
	expectedGasLimit uint64,
	gasScheduleNotifier core.GasScheduleNotifier,
	newScCode string,
) {

	shardCoordinatorForShard1, _ := sharding.NewMultiShardCoordinator(3, 1)
	shardCoordinatorForShardMeta, _ := sharding.NewMultiShardCoordinator(3, core.MetachainShardId)

	testContextShard1, err := vm.CreatePreparedTxProcessorWithVMConfigWithShardCoordinatorDBAndGasAndRoundConfig(
		enableEpochs,
		shardCoordinatorForShard1,
		integrationtests.CreateMemUnit(),
		gasScheduleNotifier,
		testscommon.GetDefaultRoundsConfig(),
		vm.CreateVMConfigWithVersion("v1.4"),
	)
	require.Nil(t, err)
	testContextShardMeta, err := vm.CreatePreparedTxProcessorWithVMConfigWithShardCoordinatorDBAndGasAndRoundConfig(
		enableEpochs,
		shardCoordinatorForShardMeta,
		integrationtests.CreateMemUnit(),
		gasScheduleNotifier,
		testscommon.GetDefaultRoundsConfig(),
		vm.CreateVMConfigWithVersion("v1.4"),
	)
	require.Nil(t, err)

	// step 1. deploy the first contract
	scAddress, owner := utils.DoDeployWithCustomParams(
		t,
		testContextShard1,
		"./testdata/first/output/first.wasm",
		big.NewInt(100000000000),
		2000,
		nil,
	)
	assert.Equal(t, 32, len(owner))
	assert.Equal(t, 32, len(scAddress))

	intermediates := testContextShard1.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediates))
	testContextShard1.CleanIntermediateTransactions(t)

	// step 2. call a dummy function on the first version of the contract

	tx := utils.CreateSmartContractCall(1, owner, scAddress, 10, 2000, "callMe", nil)
	code, err := testContextShard1.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, code)

	intermediates = testContextShard1.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediates))
	testContextShard1.CleanIntermediateTransactions(t)

	// step 3. upgrade to the second contract

	txData := strings.Join([]string{
		upgradeContractFunction,
		newScCode,
		wasm.VMTypeHex,
		hex.EncodeToString(core.ESDTSCAddress),
		hex.EncodeToString([]byte("nonExistentFunction")),
		hex.EncodeToString([]byte("dummyArg")),
	}, "@")
	tx = utils.CreateSmartContractCall(2, owner, scAddress, 10, 10000000, txData, nil)
	code, err = testContextShard1.TxProcessor.ProcessTransaction(tx)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, code)

	intermediates = testContextShard1.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediates))
	testContextShard1.CleanIntermediateTransactions(t)

	// step 4. execute scr on metachain, should fail

	scr := intermediates[0].(*smartContractResult.SmartContractResult)
	code, err = testContextShardMeta.ScProcessor.ProcessSmartContractResult(scr)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.UserError, code)

	intermediates = testContextShardMeta.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediates))
	testContextShardMeta.CleanIntermediateTransactions(t)

	// step 5. execute generated metachain scr on the contract
	scr = intermediates[0].(*smartContractResult.SmartContractResult)
	code, err = testContextShard1.ScProcessor.ProcessSmartContractResult(scr)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, code)

	assert.Equal(t, 1, len(intermediates))
	testContextShardMeta.CleanIntermediateTransactions(t)

	assert.Equal(t, expectedGasLimit, intermediates[0].GetGasLimit())
}

func TestAsyncCallsOnInitFunctionOnDeploy(t *testing.T) {
	firstSCCode := wasm.GetSCCode("./testdata/first/output/first.wasm")
	pathToSecondSC := "./testdata/asyncOnInit/asyncOnInitAndUpgrade.wasm"
	secondSCCode := wasm.GetSCCode(pathToSecondSC)

	t.Run("backwards compatibility for unset flag", func(t *testing.T) {
		gasScheduleNotifier := vm.CreateMockGasScheduleNotifier()

		expectedGasLimit := gasScheduleNotifier.LatestGasSchedule()[common.BaseOpsAPICost][common.AsyncCallbackGasLockField] +
			gasScheduleNotifier.LatestGasSchedule()[common.BaseOpsAPICost][common.AsyncCallStepField] +
			gasScheduleNotifier.LatestGasSchedule()[common.BaseOperationCost]["AoTPreparePerByte"]*uint64(len(firstSCCode))/2

		enableEpoch := config.EnableEpochs{
			RuntimeCodeSizeFixEnableEpoch:                   100000, // fix not activated
			DynamicGasCostForDataTrieStorageLoadEnableEpoch: 100000,
			SCProcessorV2EnableEpoch:                        integrationTests.UnreachableEpoch,
		}

		testAsyncCallsOnInitFunctionOnDeploy(t, enableEpoch, expectedGasLimit, gasScheduleNotifier, pathToSecondSC)
	})
	t.Run("fix activated", func(t *testing.T) {
		gasScheduleNotifier := vm.CreateMockGasScheduleNotifier()

		expectedGasLimit := gasScheduleNotifier.LatestGasSchedule()[common.BaseOpsAPICost][common.AsyncCallbackGasLockField] +
			gasScheduleNotifier.LatestGasSchedule()[common.BaseOpsAPICost][common.AsyncCallStepField] +
			gasScheduleNotifier.LatestGasSchedule()[common.BaseOperationCost]["AoTPreparePerByte"]*uint64(len(secondSCCode))/2

		enableEpoch := config.EnableEpochs{
			RuntimeCodeSizeFixEnableEpoch:                   0, // fix activated
			DynamicGasCostForDataTrieStorageLoadEnableEpoch: 100000,
			SCProcessorV2EnableEpoch:                        integrationTests.UnreachableEpoch,
		}

		testAsyncCallsOnInitFunctionOnDeploy(t, enableEpoch, expectedGasLimit, gasScheduleNotifier, pathToSecondSC)
	})
}

func testAsyncCallsOnInitFunctionOnDeploy(t *testing.T,
	enableEpochs config.EnableEpochs,
	expectedGasLimit uint64,
	gasScheduleNotifier core.GasScheduleNotifier,
	pathToSecondSC string,
) {
	shardCoordinatorForShard1, _ := sharding.NewMultiShardCoordinator(3, 1)
	shardCoordinatorForShardMeta, _ := sharding.NewMultiShardCoordinator(3, core.MetachainShardId)

	testContextShard1, err := vm.CreatePreparedTxProcessorWithVMConfigWithShardCoordinatorDBAndGasAndRoundConfig(
		enableEpochs,
		shardCoordinatorForShard1,
		integrationtests.CreateMemUnit(),
		gasScheduleNotifier,
		testscommon.GetDefaultRoundsConfig(),
		vm.CreateVMConfigWithVersion("v1.4"),
	)
	require.Nil(t, err)
	testContextShardMeta, err := vm.CreatePreparedTxProcessorWithVMConfigWithShardCoordinatorDBAndGasAndRoundConfig(
		enableEpochs,
		shardCoordinatorForShardMeta,
		integrationtests.CreateMemUnit(),
		gasScheduleNotifier,
		testscommon.GetDefaultRoundsConfig(),
		vm.CreateVMConfigWithVersion("v1.4"),
	)
	require.Nil(t, err)

	// step 1. deploy the first contract
	scAddressFirst, firstOwner := utils.DoDeployWithCustomParams(
		t,
		testContextShard1,
		"./testdata/first/output/first.wasm",
		big.NewInt(100000000000),
		2000,
		nil,
	)
	assert.Equal(t, 32, len(firstOwner))
	assert.Equal(t, 32, len(scAddressFirst))

	intermediates := testContextShard1.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediates))
	testContextShard1.CleanIntermediateTransactions(t)

	// step 2. call a dummy function on the first contract

	tx := utils.CreateSmartContractCall(1, firstOwner, scAddressFirst, 10, 2000, "callMe", nil)
	code, err := testContextShard1.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, code)

	intermediates = testContextShard1.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediates))
	testContextShard1.CleanIntermediateTransactions(t)

	// step 3. deploy the second contract that does an async on init function

	scAddressSecond, secondOwner := utils.DoDeployWithCustomParams(
		t,
		testContextShard1,
		pathToSecondSC,
		big.NewInt(100000000000),
		10000000,
		[]string{
			hex.EncodeToString(core.ESDTSCAddress),
			hex.EncodeToString([]byte("nonExistentFunction")),
			hex.EncodeToString([]byte("dummyArg")),
		},
	)
	assert.Equal(t, 32, len(secondOwner))
	assert.Equal(t, 32, len(scAddressSecond))

	intermediates = testContextShard1.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediates))
	testContextShard1.CleanIntermediateTransactions(t)

	// step 4. execute scr on metachain, should fail

	scr := intermediates[0].(*smartContractResult.SmartContractResult)
	code, err = testContextShardMeta.ScProcessor.ProcessSmartContractResult(scr)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.UserError, code)

	intermediates = testContextShardMeta.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediates))
	testContextShardMeta.CleanIntermediateTransactions(t)

	// step 5. execute generated metachain scr on the contract
	scr = intermediates[0].(*smartContractResult.SmartContractResult)
	code, err = testContextShard1.ScProcessor.ProcessSmartContractResult(scr)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, code)

	assert.Equal(t, 1, len(intermediates))
	testContextShardMeta.CleanIntermediateTransactions(t)

	assert.Equal(t, expectedGasLimit, intermediates[0].GetGasLimit())
}
