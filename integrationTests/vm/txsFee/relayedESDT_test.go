//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestRelayedESDTTransferShouldWork(t *testing.T) {
	t.Run("before relayed move balance fix", testRelayedESDTTransferShouldWork(integrationTests.UnreachableEpoch))
	t.Run("after relayed move balance fix", testRelayedESDTTransferShouldWork(0))
}

func testRelayedESDTTransferShouldWork(relayedFixActivationEpoch uint32) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedMoveBalanceEnableEpoch: relayedFixActivationEpoch,
		})
		require.Nil(t, err)
		defer testContext.Close()

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789012")
		rcvAddr := []byte("12345678901234567890123456789022")

		relayerBalance := big.NewInt(10000000)
		localEsdtBalance := big.NewInt(100000000)
		token := []byte("miiutoken")
		utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, big.NewInt(0), token, 0, localEsdtBalance)
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, relayerBalance)

		gasLimit := uint64(40)
		innerTx := utils.CreateESDTTransferTx(0, sndAddr, rcvAddr, token, big.NewInt(100), gasPrice, gasLimit)

		rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
		rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.Ok, retCode)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		expectedBalanceSnd := big.NewInt(99999900)
		utils.CheckESDTBalance(t, testContext, sndAddr, token, expectedBalanceSnd)

		expectedReceiverBalance := big.NewInt(100)
		utils.CheckESDTBalance(t, testContext, rcvAddr, token, expectedReceiverBalance)

		expectedEGLDBalance := big.NewInt(0)
		utils.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedEGLDBalance)

		utils.TestAccount(t, testContext.Accounts, relayerAddr, 1, big.NewInt(9997290))

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, big.NewInt(2710), accumulatedFees)
	}
}

func TestRelayedESTTransferNotEnoughESTValueShouldConsumeGas(t *testing.T) {
	t.Run("before relayed move balance fix", testRelayedESTTransferNotEnoughESTValueShouldConsumeGas(integrationTests.UnreachableEpoch))
	t.Run("after relayed move balance fix", testRelayedESTTransferNotEnoughESTValueShouldConsumeGas(0))
}

func testRelayedESTTransferNotEnoughESTValueShouldConsumeGas(relayedFixActivationEpoch uint32) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedMoveBalanceEnableEpoch: relayedFixActivationEpoch,
		})
		require.Nil(t, err)
		defer testContext.Close()

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789012")
		rcvAddr := []byte("12345678901234567890123456789022")

		relayerBalance := big.NewInt(10000000)
		localEsdtBalance := big.NewInt(100000000)
		token := []byte("miiutoken")
		utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, big.NewInt(0), token, 0, localEsdtBalance)
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, relayerBalance)

		gasLimit := uint64(42)
		innerTx := utils.CreateESDTTransferTx(0, sndAddr, rcvAddr, token, big.NewInt(100000001), gasPrice, gasLimit)

		rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
		rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.ExecutionFailed, retCode)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		expectedBalanceSnd := big.NewInt(100000000)
		utils.CheckESDTBalance(t, testContext, sndAddr, token, expectedBalanceSnd)

		expectedReceiverBalance := big.NewInt(0)
		utils.CheckESDTBalance(t, testContext, rcvAddr, token, expectedReceiverBalance)

		expectedEGLDBalance := big.NewInt(0)
		utils.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedEGLDBalance)

		utils.TestAccount(t, testContext.Accounts, relayerAddr, 1, big.NewInt(9997110))

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, big.NewInt(2890), accumulatedFees)
	}
}
