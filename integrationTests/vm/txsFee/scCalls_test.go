package txsFee

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/require"
)

func TestScCallShouldWork(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress := doDeploy(t, &testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	for idx := uint64(0); idx < 10; idx++ {
		tx := vm.CreateTransaction(idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))
		_, err := testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Nil(t, testContext.GetLatestError())

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)
	}

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(11), ret)

	expectedBalance := big.NewInt(61400)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 10, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(49570), accumulatedFee)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(4128), developerFees)
}

func TestScCallContractNotFoundShouldConsumeGas(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress := "00000000000000000500dbb53e4b23392b0d6f36cce32deb2d623e9625ab3132"
	scAddrBytes, _ := hex.DecodeString(scAddress)
	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddrBytes, gasPrice, gasLimit, []byte("increment"))
	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, fmt.Errorf("invalid contract code (not found)"), testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(90000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10000), accumulatedFee)
}

func TestScCallInvalidMethodToCallShouldConsumeGas(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress := doDeploy(t, &testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("invalidMethod"))
	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, fmt.Errorf("invalid function (not found)"), testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(1), ret)

	expectedBalance := big.NewInt(90000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(20970), accumulatedFee)
}

func TestScCallInsufficientGasLimitShouldNotConsumeGas(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress := doDeploy(t, &testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(9)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))
	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, process.ErrInsufficientGasLimitInTx, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(1), ret)

	expectedBalance := big.NewInt(100000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 0, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10970), accumulatedFee)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(368), developerFees)
}

func TestScCallOutOfGasShouldConsumeGas(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	scAddress := doDeploy(t, &testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(20)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)
	require.Equal(t, fmt.Errorf("not enough gas"), testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(1), ret)

	expectedBalance := big.NewInt(99800)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(11170), accumulatedFee)
}

func doDeploy(t *testing.T, testContext *vm.VMTestContext) []byte {
	sndAddr := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(2000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	scCode := arwen.GetSCCode("../arwen/testdata/counter/output/counter.wasm")
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(arwen.CreateDeployTxData(scCode)))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(89030)
	vm.TestAccount(t, testContext.Accounts, sndAddr, senderNonce+1, expectedBalance)

	// check accumulated fees
	accumulatedFee := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10970), accumulatedFee)

	scAddress, _ := testContext.BlockchainHook.NewAddress(sndAddr, 0, factory.ArwenVirtualMachine)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(368), developerFees)

	return scAddress
}
