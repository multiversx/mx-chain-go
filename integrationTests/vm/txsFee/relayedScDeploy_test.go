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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789012")

	senderNonce := uint64(0)
	senderBalance := big.NewInt(0)
	gasLimit := uint64(1962)

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

	expectedBalanceRelayer := big.NewInt(2530)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check balance inner tx sender
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(47470), accumulatedFees)
}

func TestRelayedScDeployInvalidCodeShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
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

	expectedBalanceRelayer := big.NewInt(17030)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check balance inner tx sender
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(32970), accumulatedFees)
}

func TestRelayedScDeployInsufficientGasLimitShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
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

	expectedBalanceRelayer := big.NewInt(17130)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check balance inner tx sender
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(32870), accumulatedFees)
}

func TestRelayedScDeployOutOfGasShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
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

	expectedBalanceRelayer := big.NewInt(16430)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	// check balance inner tx sender
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(33570), accumulatedFees)
}
