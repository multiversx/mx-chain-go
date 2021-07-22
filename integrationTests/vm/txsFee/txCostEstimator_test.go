package txsFee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	"github.com/stretchr/testify/require"
)

func TestSCCallCostTransactionCost(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeployNoChecks(t, testContext, "../arwen/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

	res, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(387), res.GasUnits)
}

func TestScDeployTransactionCost(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(1000000000))

	scCode := arwen.GetSCCode("../arwen/testdata/misc/fib_arwen/output/fib_arwen.wasm")
	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), 0, 0, []byte(arwen.CreateDeployTxData(scCode)))

	res, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(849), res.GasUnits)
}

func TestAsyncCallsTransactionCost(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(100000000)
	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, senderAddr, 0, egldBalance)

	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(2000)

	pathToContract := "testdata/first/first.wasm"
	firstScAddress, _ := utils.DoDeployNoChecks(t, testContext, pathToContract)

	args := [][]byte{[]byte(hex.EncodeToString(firstScAddress))}
	pathToContract = "testdata/second/output/async.wasm"
	secondSCAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(50))

	tx := vm.CreateTransaction(1, big.NewInt(0), senderAddr, secondSCAddress, 0, 0, []byte("doSomething"))
	resWithCost, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(99991601), resWithCost.GasUnits)
}

func TestBuiltInFunctionTransactionCost(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(
		vm.ArgEnableEpoch{
			PenalizedTooMuchGasEnableEpoch: 100,
		})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, owner := utils.DoDeployNoChecks(t, testContext, "../arwen/testdata/counter/output/counter.wasm")
	testContext.TxFeeHandler.CreateBlockStarted()
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	newOwner := []byte("12345678901234567890123456789112")

	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	tx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddress, 0, 0, txData)
	res, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(85), res.GasUnits)
}

func TestESDTTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	egldBalance := big.NewInt(100000000)
	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, esdtBalance)

	tx := utils.CreateESDTTransferTx(0, sndAddr, rcvAddr, token, big.NewInt(100), 0, 0)
	res, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(36), res.GasUnits)
}

func TestAsyncESDTTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Arwen fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(100000000)
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)

	// create an address with ESDT token
	sndAddr := []byte("12345678901234567890123456789012")

	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, esdtBalance)

	// deploy 2 contracts
	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	argsSecond := [][]byte{[]byte(hex.EncodeToString(token))}
	secondSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/second-contract.wasm", ownerAccount, gasPrice, deployGasLimit, argsSecond, big.NewInt(0))

	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	firstSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/first-contract.wasm", ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(0))

	testContext.TxFeeHandler.CreateBlockStarted()
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	tx := utils.CreateESDTTransferTx(0, sndAddr, firstSCAddress, token, big.NewInt(5000), 0, 0)
	tx.Data = []byte(string(tx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractHalf")))

	res, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(34207), res.GasUnits)
}
