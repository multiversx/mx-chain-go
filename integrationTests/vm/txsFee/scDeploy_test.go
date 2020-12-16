package txsFee

import (
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/process"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/stretchr/testify/require"
)

func TestScDeployShouldWork(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	scCode := arwen.GetSCCode("../arwen/testdata/misc/fib_arwen/output/fib_arwen.wasm")
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(arwen.CreateDeployTxData(scCode)))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// 8490 gas units the sc deploy consumed
	expectedBalance := big.NewInt(91510)
	vm.TestAccount(t, testContext.Accounts, sndAddr, senderNonce+1, expectedBalance)
}

func TestScDeployInvalidContractCodeShouldConsumeGas(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	scCode := arwen.GetSCCode("../arwen/testdata/misc/fib_arwen/output/fib_arwen.wasm")
	scCodeBytes := []byte(scCode)
	scCodeBytes = append(scCodeBytes, []byte("aaa")...)
	txDeployData := []byte(arwen.CreateDeployTxData(string(scCodeBytes)))
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, txDeployData)

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, parsers.ErrInvalidCode, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(90000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, senderNonce+1, expectedBalance)
}

func TestScDeployInsufficientGasLimitShouldNotConsumeGas(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(568)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	scCode := arwen.GetSCCode("../arwen/testdata/misc/fib_arwen/output/fib_arwen.wasm")
	txDeployData := []byte(arwen.CreateDeployTxData(scCode))
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, txDeployData)

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, process.ErrInsufficientGasLimitInTx, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(0).Set(senderBalance)
	vm.TestAccount(t, testContext.Accounts, sndAddr, senderNonce, expectedBalance)
}

func TestScDeployOutOfGasShouldConsumeGas(t *testing.T) {
	t.Parallel()

	testContext := vm.CreatePreparedTxProcessorWithVMs(true)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(570)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	scCode := arwen.GetSCCode("../arwen/testdata/misc/fib_arwen/output/fib_arwen.wasm")
	txDeployData := []byte(arwen.CreateDeployTxData(scCode))
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), gasPrice, gasLimit, txDeployData)

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.OutOfGas.String(), testContext.GetLatestError().Error())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(94300)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)
}
