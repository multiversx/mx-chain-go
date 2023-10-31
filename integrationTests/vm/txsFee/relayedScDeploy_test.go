//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestRelayedScDeployShouldWork(t *testing.T) {
	t.Run("before relayed fix", testRelayedScDeployShouldWork(integrationTests.UnreachableEpoch))
	t.Run("after relayed fix", testRelayedScDeployShouldWork(0))
}

func testRelayedScDeployShouldWork(relayedFixActivationEpoch uint32) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedMoveBalanceEnableEpoch: relayedFixActivationEpoch,
		})
		require.Nil(t, err)
		defer testContext.Close()

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789012")

		senderNonce := uint64(0)
		senderBalance := big.NewInt(0)
		gasLimit := uint64(1000)

		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(50000))

		scCode := wasm.GetSCCode("../wasm/testdata/misc/fib_wasm/output/fib_wasm.wasm")
		userTx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(wasm.CreateDeployTxData(scCode)))

		rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
		rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.Ok, retCode)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		expectedBalanceRelayer := big.NewInt(28440)
		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

		// check balance inner tx sender
		vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, big.NewInt(21560), accumulatedFees)
	}
}

func TestRelayedScDeployInvalidCodeShouldConsumeGas(t *testing.T) {
	t.Run("before relayed fix", testRelayedScDeployInvalidCodeShouldConsumeGas(integrationTests.UnreachableEpoch))
	t.Run("after relayed fix", testRelayedScDeployInvalidCodeShouldConsumeGas(0))
}

func testRelayedScDeployInvalidCodeShouldConsumeGas(relayedFixActivationEpoch uint32) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedMoveBalanceEnableEpoch: relayedFixActivationEpoch,
		})
		require.Nil(t, err)
		defer testContext.Close()

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789012")

		senderNonce := uint64(0)
		senderBalance := big.NewInt(0)
		gasLimit := uint64(574)

		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(50000))

		scCode := wasm.GetSCCode("../wasm/testdata/misc/fib_wasm/output/fib_wasm.wasm")
		scCodeBytes := []byte(wasm.CreateDeployTxData(scCode))
		scCodeBytes = append(scCodeBytes, []byte("aaaaa")...)
		userTx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, scCodeBytes)

		rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
		rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.UserError, retCode)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		expectedBalanceRelayer := big.NewInt(31090)
		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

		// check balance inner tx sender
		vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, big.NewInt(18910), accumulatedFees)
	}
}

func TestRelayedScDeployInsufficientGasLimitShouldConsumeGas(t *testing.T) {
	t.Run("before relayed fix", testRelayedScDeployInsufficientGasLimitShouldConsumeGas(integrationTests.UnreachableEpoch, big.NewInt(31930), big.NewInt(18070)))
	t.Run("after relayed fix", testRelayedScDeployInsufficientGasLimitShouldConsumeGas(0, big.NewInt(31240), big.NewInt(18760)))
}

func testRelayedScDeployInsufficientGasLimitShouldConsumeGas(relayedFixActivationEpoch uint32, expectedBalance *big.Int, expectedAccumulatedFees *big.Int) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedMoveBalanceEnableEpoch: relayedFixActivationEpoch,
		})
		require.Nil(t, err)
		defer testContext.Close()

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789012")

		senderNonce := uint64(0)
		senderBalance := big.NewInt(0)
		gasLimit := uint64(500)

		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(50000))

		scCode := wasm.GetSCCode("../wasm/testdata/misc/fib_wasm/output/fib_wasm.wasm")
		userTx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(wasm.CreateDeployTxData(scCode)))

		rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
		rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.UserError, retCode)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

		// check balance inner tx sender
		vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, expectedAccumulatedFees, accumulatedFees)
	}
}

func TestRelayedScDeployOutOfGasShouldConsumeGas(t *testing.T) {
	t.Run("before relayed fix", testRelayedScDeployOutOfGasShouldConsumeGas(integrationTests.UnreachableEpoch))
	t.Run("after relayed fix", testRelayedScDeployOutOfGasShouldConsumeGas(0))
}

func testRelayedScDeployOutOfGasShouldConsumeGas(relayedFixActivationEpoch uint32) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedMoveBalanceEnableEpoch: relayedFixActivationEpoch,
		})
		require.Nil(t, err)
		defer testContext.Close()

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789012")

		senderNonce := uint64(0)
		senderBalance := big.NewInt(0)
		gasLimit := uint64(570)

		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(50000))

		scCode := wasm.GetSCCode("../wasm/testdata/misc/fib_wasm/output/fib_wasm.wasm")
		userTx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(wasm.CreateDeployTxData(scCode)))

		rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
		rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		code, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.UserError, code)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		expectedBalanceRelayer := big.NewInt(31230)
		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

		// check balance inner tx sender
		vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, big.NewInt(18770), accumulatedFees)
	}
}
