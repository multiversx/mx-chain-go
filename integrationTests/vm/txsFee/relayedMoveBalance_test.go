package txsFee

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	dataTransaction "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayedMoveBalanceShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	senderNonce := uint64(0)
	senderBalance := big.NewInt(100)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))

	// gas consumed = 50
	userTx := vm.CreateTransaction(senderNonce, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("aaaa"))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := gasLimit + minGasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// check relayer balance
	// 3000 - rTxFee(175)*gasPrice(10) + txFeeInner(1000) = 2750
	expectedBalanceRelayer := big.NewInt(250)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check balance inner tx sender
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

	// check balance inner tx receiver
	vm.TestAccount(t, testContext.Accounts, rcvAddr, 0, big.NewInt(100))

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(2750), accumulatedFees)
}

func TestRelayedMoveBalanceInvalidGasLimitShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, rcvAddr, 1, 100, []byte("aaaa"))

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 2 + userTx.GasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, 1, rTxGasLimit, rtxData)

	_, err = testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, process.ErrFailedTransaction, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceRelayer := big.NewInt(2724)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(276), accumulatedFees)
}

func TestRelayedMoveBalanceInvalidUserTxShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(100))
	userTx := vm.CreateTransaction(1, big.NewInt(100), sndAddr, rcvAddr, 1, 100, []byte("aaaa"))

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := minGasLimit + userTx.GasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, 1, rTxGasLimit, rtxData)

	retcode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retcode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// 3000 - rTxFee(179)*gasPrice(1) - innerTxFee(100) = 2721
	expectedBalanceRelayer := big.NewInt(2721)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(279), accumulatedFees)
}

func TestRelayedMoveBalanceInvalidUserTxValueShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		RelayedNonceFixEnableEpoch: 1,
	})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(100))
	userTx := vm.CreateTransaction(0, big.NewInt(150), sndAddr, rcvAddr, 1, 100, []byte("aaaa"))

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := minGasLimit + userTx.GasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, 1, rTxGasLimit, rtxData)

	retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// 3000 - rTxFee(175)*gasPrice(1) - innerTxFee(100) = 2750
	expectedBalanceRelayer := big.NewInt(2725)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(275), accumulatedFees)
}

func TestRelayedMoveBalanceHigherNonce(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		RelayedNonceFixEnableEpoch: 1,
	})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))
	userTx := vm.CreateTransaction(100, big.NewInt(150), sndAddr, rcvAddr, 1, 100, nil)

	t.Run("inactive flag should increment", func(t *testing.T) {
		initialSenderNonce := getAccount(t, testContext, sndAddr).GetNonce()

		rtxDataV1 := integrationTests.PrepareRelayedTxDataV1(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr, vmcommon.UserError)

		senderAccount := getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce+1, senderAccount.GetNonce())

		rtxDataV2 := integrationTests.PrepareRelayedTxDataV2(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV2, big.NewInt(0), sndAddr, vmcommon.UserError)

		senderAccount = getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce+2, senderAccount.GetNonce())
	})
	t.Run("active flag should not increment", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})
		initialSenderNonce := getAccount(t, testContext, sndAddr).GetNonce()

		rtxDataV1 := integrationTests.PrepareRelayedTxDataV1(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr, vmcommon.UserError)

		senderAccount := getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce, senderAccount.GetNonce())

		rtxDataV2 := integrationTests.PrepareRelayedTxDataV2(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV2, big.NewInt(0), sndAddr, vmcommon.UserError)

		senderAccount = getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce, senderAccount.GetNonce())
	})
}

func TestRelayedMoveBalanceLowerNonce(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		RelayedNonceFixEnableEpoch: 1,
	})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 5, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))
	userTx := vm.CreateTransaction(4, big.NewInt(150), sndAddr, rcvAddr, 1, 100, nil)

	t.Run("inactive flag should increment", func(t *testing.T) {
		initialSenderNonce := getAccount(t, testContext, sndAddr).GetNonce()

		rtxDataV1 := integrationTests.PrepareRelayedTxDataV1(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr, vmcommon.UserError)

		senderAccount := getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce+1, senderAccount.GetNonce())

		rtxDataV2 := integrationTests.PrepareRelayedTxDataV2(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV2, big.NewInt(0), sndAddr, vmcommon.UserError)

		senderAccount = getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce+2, senderAccount.GetNonce())
	})
	t.Run("active flag should not increment", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})
		initialSenderNonce := getAccount(t, testContext, sndAddr).GetNonce()

		rtxDataV1 := integrationTests.PrepareRelayedTxDataV1(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr, vmcommon.UserError)

		senderAccount := getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce, senderAccount.GetNonce())

		rtxDataV2 := integrationTests.PrepareRelayedTxDataV2(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV2, big.NewInt(0), sndAddr, vmcommon.UserError)

		senderAccount = getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce, senderAccount.GetNonce())
	})
}

func TestRelayedMoveBalanceHigherNonceWithActivatedFixCrossShard(t *testing.T) {
	enableEpochs := config.EnableEpochs{
		RelayedNonceFixEnableEpoch: 0,
	}

	shardCoordinator0, _ := sharding.NewMultiShardCoordinator(2, 0)
	testContext0, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(
		enableEpochs,
		shardCoordinator0,
		integrationtests.CreateMemUnit(),
		vm.CreateMockGasScheduleNotifier(),
	)
	require.Nil(t, err)

	shardCoordinator1, _ := sharding.NewMultiShardCoordinator(2, 1)
	testContext1, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(
		enableEpochs,
		shardCoordinator1,
		integrationtests.CreateMemUnit(),
		vm.CreateMockGasScheduleNotifier(),
	)
	require.Nil(t, err)
	defer testContext0.Close()
	defer testContext1.Close()

	relayerAddr := []byte("relayer-000000000000000000000000")
	assert.Equal(t, uint32(0), shardCoordinator0.ComputeId(relayerAddr)) // shard 0
	sndAddr := []byte("sender-1111111111111111111111111")
	assert.Equal(t, uint32(1), shardCoordinator0.ComputeId(sndAddr)) // shard 1
	rcvAddr := []byte("receiver-22222222222222222222222")
	assert.Equal(t, uint32(0), shardCoordinator0.ComputeId(rcvAddr)) // shard 0

	_, _ = vm.CreateAccount(testContext0.Accounts, relayerAddr, 0, big.NewInt(3000)) // create relayer in shard 0
	_, _ = vm.CreateAccount(testContext1.Accounts, sndAddr, 0, big.NewInt(0))        // create sender in shard 1

	userTx := vm.CreateTransaction(1, big.NewInt(150), sndAddr, rcvAddr, 1, 100, nil)
	initialSenderNonce := getAccount(t, testContext1, sndAddr).GetNonce()

	rtxDataV1 := integrationTests.PrepareRelayedTxDataV1(userTx)
	executeRelayedTransaction(t, testContext0, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr, vmcommon.Ok)

	results := testContext0.GetIntermediateTransactions(t)
	assert.Equal(t, 0, len(results)) // no scrs, the exact relayed tx will be executed on the receiver shard

	executeRelayedTransaction(t, testContext1, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr, vmcommon.UserError)

	senderAccount := getAccount(t, testContext1, sndAddr)
	require.NotNil(t, senderAccount)
	assert.Equal(t, initialSenderNonce, senderAccount.GetNonce())
}

func executeRelayedTransaction(
	tb testing.TB,
	testContext *vm.VMTestContext,
	relayerAddress []byte,
	userTx *dataTransaction.Transaction,
	userTxPrepared []byte,
	value *big.Int,
	senderAddress []byte,
	expectedReturnCode vmcommon.ReturnCode,
) {
	testContext.TxsLogsProcessor.Clean()
	relayerAccount := getAccount(tb, testContext, relayerAddress)
	gasLimit := minGasLimit + userTx.GasLimit + uint64(len(userTxPrepared))

	relayedTx := vm.CreateTransaction(relayerAccount.GetNonce(), value, relayerAddress, senderAddress, 1, gasLimit, userTxPrepared)
	retCode, _ := testContext.TxProcessor.ProcessTransaction(relayedTx)
	require.Equal(tb, expectedReturnCode, retCode)

	_, err := testContext.Accounts.Commit()
	require.Nil(tb, err)

	relayedTxHash, _ := core.CalculateHash(testContext.Marshalizer, integrationtests.TestHasher, relayedTx)

	if expectedReturnCode == vmcommon.Ok {
		return
	}

	logs, err := testContext.TxsLogsProcessor.GetLog(relayedTxHash)
	assert.Nil(tb, err)
	events := logs.GetLogEvents()
	assert.Equal(tb, 1, len(events))
	assert.Equal(tb, core.SignalErrorOperation, string(events[0].GetIdentifier()))
}
