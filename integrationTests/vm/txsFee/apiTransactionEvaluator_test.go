package txsFee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/stretchr/testify/require"
)

func getZeroGasAndFees() scheduled.GasAndFees {
	return scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
}

func TestSCCallCostTransactionCost(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeployNoChecks(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

	res, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(387), res.GasUnits)
}

func TestScDeployTransactionCost(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(1000000000))

	scCode := wasm.GetSCCode("../wasm/testdata/misc/fib_wasm/output/fib_wasm.wasm")
	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, vm.CreateEmptyAddress(), 0, 0, []byte(wasm.CreateDeployTxData(scCode)))

	res, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(849), res.GasUnits)
}

func TestAsyncCallsTransactionCost(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(100000000)
	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, senderAddr, 0, egldBalance)

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
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(
		config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: integrationTests.UnreachableEpoch,
		})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, owner := utils.DoDeployNoChecks(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
	gasAndFees := getZeroGasAndFees()
	testContext.TxFeeHandler.CreateBlockStarted(gasAndFees)
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
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	egldBalance := big.NewInt(100000000)
	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)

	tx := utils.CreateESDTTransferTx(0, sndAddr, rcvAddr, token, big.NewInt(100), 0, 0)
	res, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(36), res.GasUnits)
}

func TestAsyncESDTTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(100000000)
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)

	// create an address with ESDT token
	sndAddr := []byte("12345678901234567890123456789012")

	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)

	// deploy 2 contracts
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	argsSecond := [][]byte{[]byte(hex.EncodeToString(token))}
	secondSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/second-contract.wasm", ownerAccount, gasPrice, deployGasLimit, argsSecond, big.NewInt(0))

	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	firstSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/first-contract.wasm", ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(0))

	gasAndFees := getZeroGasAndFees()
	testContext.TxFeeHandler.CreateBlockStarted(gasAndFees)
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	tx := utils.CreateESDTTransferTx(0, sndAddr, firstSCAddress, token, big.NewInt(5000), 0, 0)
	tx.Data = []byte(string(tx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractHalf")))

	res, err := testContext.TxCostHandler.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, uint64(34156), res.GasUnits)
}
