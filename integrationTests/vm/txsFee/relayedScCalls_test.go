package txsFee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func TestRelayedScCallShouldWork(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress, _ := doDeploy(t, &testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

	rtxData := prepareRelayerTxData(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	_, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(2), ret)

	expectedBalance := big.NewInt(24170)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(16700), accumulatedFee)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(744), developerFees)
}

func TestRelayedScCallContractNotFoundShouldConsumeGas(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress := "00000000000000000500dbb53e4b23392b0d6f36cce32deb2d623e9625ab3132"
	scAddrBytes, _ := hex.DecodeString(scAddress)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddrBytes, gasPrice, gasLimit, []byte("increment"))

	rtxData := prepareRelayerTxData(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(18130)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(11870), accumulatedFee)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}

func TestRelayedScCallInvalidMethodShouldConsumeGas(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress, _ := doDeploy(t, &testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("invalidMethod"))

	rtxData := prepareRelayerTxData(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(18050)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(22920), accumulatedFee)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(368), developerFees)
}

func TestRelayedScCallInsufficientGasLimitShouldConsumeGas(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress, _ := doDeploy(t, &testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	gasPrice := uint64(10)
	gasLimit := uint64(5)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

	rtxData := prepareRelayerTxData(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Equal(t, process.ErrFailedTransaction, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(28100)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(12870), accumulatedFee)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(368), developerFees)
}

func TestRelayedScCallOutOfGasShouldConsumeGas(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress, _ := doDeploy(t, &testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	gasPrice := uint64(10)
	gasLimit := uint64(20)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

	rtxData := prepareRelayerTxData(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(27950)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(13020), accumulatedFee)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(368), developerFees)
}
