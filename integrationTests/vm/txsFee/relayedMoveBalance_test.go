package txsFee

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	dataTransaction "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/process"
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
	senderBalance := big.NewInt(0)
	gasPrice := uint64(10)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))

	// gas consumed = 50
	userTx := vm.CreateTransaction(senderNonce, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("aaaa"))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// check relayer balance
	// 3000 - value(100) - gasLimit(275)*gasPrice(10) = 2850
	expectedBalanceRelayer := big.NewInt(150)
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
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, 1, rTxGasLimit, rtxData)

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

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	userTx := vm.CreateTransaction(1, big.NewInt(100), sndAddr, rcvAddr, 1, 100, []byte("aaaa"))

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + userTx.GasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, 1, rTxGasLimit, rtxData)

	retcode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retcode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceRelayer := big.NewInt(2721)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(279), accumulatedFees)
}

func TestRelayedMoveBalanceInvalidUserTxValueShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	userTx := vm.CreateTransaction(0, big.NewInt(150), sndAddr, rcvAddr, 1, 100, []byte("aaaa"))

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + userTx.GasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(100), relayerAddr, sndAddr, 1, rTxGasLimit, rtxData)

	retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceRelayer := big.NewInt(2725)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(275), accumulatedFees)
}

func TestRelayedMoveBalanceHigherNonce(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		RelayedNonceFixEnableEpoch: 0,
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
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr)

		senderAccount := getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce+1, senderAccount.GetNonce())

		rtxDataV2 := integrationTests.PrepareRelayedTxDataV2(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV2, big.NewInt(0), sndAddr)

		senderAccount = getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce+2, senderAccount.GetNonce())
	})
	t.Run("active flag should not increment", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})
		initialSenderNonce := getAccount(t, testContext, sndAddr).GetNonce()

		rtxDataV1 := integrationTests.PrepareRelayedTxDataV1(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr)

		senderAccount := getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce, senderAccount.GetNonce())

		rtxDataV2 := integrationTests.PrepareRelayedTxDataV2(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV2, big.NewInt(0), sndAddr)

		senderAccount = getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce, senderAccount.GetNonce())
	})
}

func TestRelayedMoveBalanceLowerNonce(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		RelayedNonceFixEnableEpoch: 0,
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
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr)

		senderAccount := getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce+1, senderAccount.GetNonce())

		rtxDataV2 := integrationTests.PrepareRelayedTxDataV2(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV2, big.NewInt(0), sndAddr)

		senderAccount = getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce+2, senderAccount.GetNonce())
	})
	t.Run("active flag should not increment", func(t *testing.T) {
		testContext.EpochNotifier.CheckEpoch(&block.Header{Epoch: 1})
		initialSenderNonce := getAccount(t, testContext, sndAddr).GetNonce()

		rtxDataV1 := integrationTests.PrepareRelayedTxDataV1(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV1, big.NewInt(100), sndAddr)

		senderAccount := getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce, senderAccount.GetNonce())

		rtxDataV2 := integrationTests.PrepareRelayedTxDataV2(userTx)
		executeRelayedTransaction(t, testContext, relayerAddr, userTx, rtxDataV2, big.NewInt(0), sndAddr)

		senderAccount = getAccount(t, testContext, sndAddr)
		require.NotNil(t, senderAccount)
		assert.Equal(t, initialSenderNonce, senderAccount.GetNonce())
	})
}

func executeRelayedTransaction(
	tb testing.TB,
	testContext *vm.VMTestContext,
	relayerAddress []byte,
	userTx *dataTransaction.Transaction,
	userTxPrepared []byte,
	value *big.Int,
	senderAddress []byte,
) {
	relayerAccount := getAccount(tb, testContext, relayerAddress)
	gasLimit := 1 + userTx.GasLimit + uint64(len(userTxPrepared))

	relayedTx := vm.CreateTransaction(relayerAccount.GetNonce(), value, relayerAddress, senderAddress, 1, gasLimit, userTxPrepared)
	retCode, _ := testContext.TxProcessor.ProcessTransaction(relayedTx)
	require.Equal(tb, vmcommon.UserError, retCode)

	_, err := testContext.Accounts.Commit()
	require.Nil(tb, err)
}

func getAccount(tb testing.TB, testContext *vm.VMTestContext, address []byte) vmcommon.UserAccountHandler {
	account, err := testContext.Accounts.LoadAccount(address)
	require.Nil(tb, err)

	return account.(vmcommon.UserAccountHandler)
}
