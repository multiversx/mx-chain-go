//go:build !race
// +build !race

// TODO remove build condition above to allow -race -short, after Arwen fix

package txsFee

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestScCallShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../arwen/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	for idx := uint64(0); idx < 10; idx++ {
		tx := vm.CreateTransaction(idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

		calculatedGasLimit := vm.ComputeGasLimit(nil, testContext, tx)
		require.Equal(t, uint64(387), calculatedGasLimit)

		_, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Nil(t, testContext.GetLatestError())

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		intermediateTxs := testContext.GetIntermediateTransactions(t)
		testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true)
		testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

		indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
		require.Equal(t, uint64(387), indexerTx.GasUsed)
		require.Equal(t, "3870", indexerTx.Fee)
	}

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(11), ret)

	expectedBalance := big.NewInt(61300)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 10, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(49670), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(4138), developerFees)
}

func TestScCallContractNotFoundShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress := "00000000000000000500dbb53e4b23392b0d6f36cce32deb2d623e9625ab3132"
	scAddrBytes, _ := hex.DecodeString(scAddress)
	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddrBytes, gasPrice, gasLimit, []byte("increment"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)
	require.Equal(t, fmt.Errorf("contract not found"), testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(90000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10000), accumulatedFees)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true)
	testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

	indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
	require.Equal(t, tx.GasLimit, indexerTx.GasUsed)
	require.Equal(t, "10000", indexerTx.Fee)
}

func TestScCallInvalidMethodToCallShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../arwen/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("invalidMethod"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)
	require.Equal(t, fmt.Errorf(vmcommon.FunctionNotFound.String()), testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(1), ret)

	expectedBalance := big.NewInt(90000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(20970), accumulatedFees)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true)
	testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

	indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
	require.Equal(t, tx.GasLimit, indexerTx.GasUsed)
	require.Equal(t, "10000", indexerTx.Fee)
}

func TestScCallInsufficientGasLimitShouldNotConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../arwen/testdata/counter/output/counter.wasm")

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(9)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))
	_, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, process.ErrInsufficientGasLimitInTx, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(1), ret)

	expectedBalance := big.NewInt(100000)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 0, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10970), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(368), developerFees)
}

func TestScCallOutOfGasShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../arwen/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(20)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)

	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)
	require.Equal(t, fmt.Errorf("out of gas"), testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(1), ret)

	expectedBalance := big.NewInt(99800)
	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(11170), accumulatedFees)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true)
	testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

	indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
	require.Equal(t, tx.GasLimit, indexerTx.GasUsed)
	require.Equal(t, "200", indexerTx.Fee)
}

func TestScCallAndGasChangeShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	mockGasSchedule := testContext.GasSchedule.(*mock.GasScheduleNotifierMock)

	scAddress, _ := utils.DoDeploy(t, testContext, "../arwen/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(10000000)
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	numIterations := uint64(10)
	for idx := uint64(0); idx < numIterations; idx++ {
		tx := vm.CreateTransaction(idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

		_, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Nil(t, testContext.GetLatestError())

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		intermediateTxs := testContext.GetIntermediateTransactions(t)
		testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true)
		testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

		indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
		require.Equal(t, uint64(387), indexerTx.GasUsed)
	}

	newGasSchedule := arwenConfig.MakeGasMapForTests()
	newGasSchedule["WASMOpcodeCost"] = arwenConfig.FillGasMap_WASMOpcodeValues(2)
	mockGasSchedule.ChangeGasSchedule(newGasSchedule)

	for idx := uint64(0); idx < numIterations; idx++ {
		tx := vm.CreateTransaction(numIterations+idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

		_, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Nil(t, testContext.GetLatestError())

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		intermediateTxs := testContext.GetIntermediateTransactions(t)
		testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true)
		testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

		indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
		require.Equal(t, uint64(400), indexerTx.GasUsed)
	}
}

func TestESDTScCallAndGasChangeShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(vm.ArgEnableEpoch{})
	require.Nil(t, err)
	defer testContext.Close()

	owner := []byte("12345678901234567890123456789011")
	senderBalance := big.NewInt(1000000000)
	gasPrice := uint64(10)
	gasLimit := uint64(2000000)

	_, _ = vm.CreateAccount(testContext.Accounts, owner, 0, senderBalance)
	ownerAccount, _ := testContext.Accounts.LoadAccount(owner)
	scAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/forwarder-raw.wasm", ownerAccount, gasPrice, gasLimit, nil, big.NewInt(0))
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance = big.NewInt(10000000)
	gasPrice = uint64(10)
	gasLimit = uint64(30000)

	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, senderBalance, token, 0, esdtBalance)

	txData := txDataBuilder.NewBuilder()
	valueToSendToSc := int64(1000)
	txData.TransferESDT(string(token), valueToSendToSc).Str("forward_direct_esdt_via_transf_exec").Bytes(sndAddr)
	numIterations := uint64(10)
	for idx := uint64(0); idx < numIterations; idx++ {
		tx := vm.CreateTransaction(idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, txData.ToBytes())

		_, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Nil(t, testContext.GetLatestError())

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		intermediateTxs := testContext.GetIntermediateTransactions(t)
		testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true)
		testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

		indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
		require.Equal(t, uint64(25092), indexerTx.GasUsed)
	}

	mockGasSchedule := testContext.GasSchedule.(*mock.GasScheduleNotifierMock)
	testGasSchedule := arwenConfig.MakeGasMapForTests()
	newGasSchedule := defaults.FillGasMapInternal(testGasSchedule, 1)
	newGasSchedule["BuiltInCost"][core.BuiltInFunctionESDTTransfer] = 2
	newGasSchedule["ElrondAPICost"]["TransferValue"] = 2
	mockGasSchedule.ChangeGasSchedule(newGasSchedule)

	for idx := uint64(0); idx < numIterations; idx++ {
		tx := vm.CreateTransaction(numIterations+idx, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, txData.ToBytes())

		_, err = testContext.TxProcessor.ProcessTransaction(tx)
		require.Nil(t, err)
		require.Nil(t, testContext.GetLatestError())

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		intermediateTxs := testContext.GetIntermediateTransactions(t)
		testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true)
		testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

		indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
		require.Equal(t, uint64(25095), indexerTx.GasUsed)
	}
}
