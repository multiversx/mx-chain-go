//go:build !race
// +build !race

package txsFee

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

var esdtToken = []byte("miiutoken")
var egldBalance = big.NewInt(50000000000)
var esdtBalance = big.NewInt(100)

func TestAsyncCallLegacy(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	gasPrice := uint64(10)
	gasLimit := uint64(5000000)

	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")

	firstSCAddress, secondSCAddress := deployForwarderAndTestContract(
		testContext,
		"testdata/first/first.wasm",
		ownerAddr, senderAddr, t,
		egldBalance,
		esdtBalance,
		gasPrice)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_legacy_async").
		Bytes(firstSCAddress).
		Bytes([]byte("callMe"))

	sendTx(0, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)
	sendTx(1, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContext, t)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, firstSCAddress, "numCalled")
	require.Equal(t, big.NewInt(1), res)
}

func TestAsyncCallMulti(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	gasPrice := uint64(10)
	gasLimit := uint64(5000000)

	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")

	firstSCAddress, secondSCAddress := deployForwarderAndTestContract(
		testContext,
		"testdata/first/first.wasm",
		ownerAddr, senderAddr, t,
		egldBalance,
		esdtBalance,
		gasPrice)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_promise").
		Bytes(firstSCAddress).
		Int64(50000).
		Bytes([]byte("callMe"))

	sendTx(0, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)
	sendTx(1, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContext, t)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, firstSCAddress, "numCalled")
	require.Equal(t, big.NewInt(2), res)

	res = vm.GetIntValueFromSC(nil, testContext.Accounts, secondSCAddress, "callback_count")
	require.Equal(t, big.NewInt(2), res)
}

func TestAsyncCallTransferAndExecute(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	gasPrice := uint64(10)
	gasLimit := uint64(5000000)

	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")

	firstSCAddress, secondSCAddress := deployForwarderAndTestContract(
		testContext,
		"testdata/first/first.wasm",
		ownerAddr, senderAddr, t,
		egldBalance,
		esdtBalance,
		gasPrice)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_transfer_execute").
		Bytes(firstSCAddress).
		Int64(50000).
		Bytes([]byte("callMe"))

	sendFirstCall := int64(5)
	sendSecondCall := int64(10)

	sendTxWithValue(0, sendFirstCall, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)
	sendTxWithValue(1, sendSecondCall, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContext, t)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	firstAccount, err := testContext.Accounts.GetExistingAccount(firstSCAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(sendFirstCall+sendSecondCall), firstAccount.(state.UserAccountHandler).GetBalance())

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, firstSCAddress, "numCalled")
	require.Equal(t, big.NewInt(2), res)
}

func TestAsyncCallTransferESDTAndExecute(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	gasPrice := uint64(10)
	gasLimit := uint64(5000000)

	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")

	firstSCAddress, secondSCAddress := deployForwarderAndTestContract(
		testContext,
		"testdata/forwarderQueue/vault.wasm",
		ownerAddr, senderAddr, t,
		egldBalance,
		esdtBalance,
		gasPrice)

	utils.CheckESDTNFTBalance(t, testContext, secondSCAddress, esdtToken, 0, esdtBalance)

	esdtToTransferFromParent := int64(20)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_transfer_esdt").
		Bytes(firstSCAddress).
		Int64(50000).
		Bytes([]byte("retrieve_funds_promises")).
		Bytes(esdtToken).
		Int64(esdtToTransferFromParent)

	sendTx(0, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)
	sendTx(1, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContext, t)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	utils.CheckESDTNFTBalance(t, testContext, firstSCAddress, esdtToken, 0,
		big.NewInt(2*esdtToTransferFromParent- /* received */
			esdtToTransferFromParent /* sent back via callback*/))
	utils.CheckESDTNFTBalance(t, testContext, secondSCAddress, esdtToken, 0,
		big.NewInt((esdtBalance.Int64() - /* initial */
			2*esdtToTransferFromParent + /* sent via async */
			esdtToTransferFromParent /* received back via callback */)))

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, firstSCAddress, "num_called_retrieve_funds_promises")
	require.Equal(t, big.NewInt(2), res)

	res = vm.GetIntValueFromSC(nil, testContext.Accounts, secondSCAddress, "callback_count")
	require.Equal(t, big.NewInt(2), res)
}

func deployForwarderAndTestContract(
	testContext *vm.VMTestContext,
	pathToContract string,
	ownerAddr []byte,
	senderAddr []byte,
	t *testing.T,
	egldBalance *big.Int,
	esdtBalance *big.Int,
	gasPrice uint64) ([]byte, []byte) {
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, senderAddr, 0, egldBalance)

	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	firstSCAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(0))

	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	pathToForwarder := "testdata/forwarderQueue/forwarder-queue-promises.wasm"

	secondSCAddress := utils.DoDeploySecond(t, testContext, pathToForwarder, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(0))

	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, secondSCAddress, egldBalance, esdtToken, 0, esdtBalance)

	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
	testContext.TxFeeHandler.CreateBlockStarted(gasAndFees)

	return firstSCAddress, secondSCAddress
}

func TestAsyncCallMulti_CrossShard(t *testing.T) {
	testContextFirstContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextFirstContract.Close()

	testContextSecondContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSecondContract.Close()

	testContextSender, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSender.Close()

	firstContractOwner := []byte("12345678901234567890123456789010")
	require.Equal(t, uint32(0), testContextSender.ShardCoordinator.ComputeId(firstContractOwner))

	secondContractOwner := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContextSender.ShardCoordinator.ComputeId(secondContractOwner))

	senderAddr := []byte("12345678901234567890123456789032")
	require.Equal(t, uint32(2), testContextSender.ShardCoordinator.ComputeId(senderAddr))

	_, _ = vm.CreateAccount(testContextSender.Accounts, senderAddr, 0, egldBalance)
	_, _ = vm.CreateAccount(testContextFirstContract.Accounts, firstContractOwner, 0, egldBalance)
	_, _ = vm.CreateAccount(testContextSecondContract.Accounts, secondContractOwner, 0, egldBalance)

	gasPrice := uint64(10)
	gasLimit := uint64(5000000)
	firstAccount, _ := testContextFirstContract.Accounts.LoadAccount(firstContractOwner)

	pathToContract := "testdata/first/first.wasm"
	firstScAddress := utils.DoDeploySecond(t, testContextFirstContract, pathToContract, firstAccount, gasPrice, gasLimit, nil, big.NewInt(50))

	secondAccount, _ := testContextSecondContract.Accounts.LoadAccount(secondContractOwner)
	pathToContract = "testdata/forwarderQueue/forwarder-queue-promises.wasm"
	secondSCAddress := utils.DoDeploySecond(t, testContextSecondContract, pathToContract, secondAccount, gasPrice, gasLimit, nil, big.NewInt(0))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextFirstContract)
	utils.CleanAccumulatedIntermediateTransactions(t, testContextSecondContract)

	testContextFirstContract.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	testContextSecondContract.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_promise").
		Bytes(firstScAddress).
		Int64(50000).
		Bytes([]byte("callMe"))

	sendTx(0, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContextSecondContract, t)
	sendTx(1, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContextSecondContract, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContextSecondContract, t)

	intermediateTxs := testContextSecondContract.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	// execute cross shard calls
	scrCall1 := intermediateTxs[0]
	utils.ProcessSCRResult(t, testContextFirstContract, scrCall1, vmcommon.Ok, nil)
	scrCall2 := intermediateTxs[1]
	utils.ProcessSCRResult(t, testContextFirstContract, scrCall2, vmcommon.Ok, nil)

	intermediateTxs = testContextFirstContract.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	// execute cross shard callbacks
	scrCallBack1 := intermediateTxs[0]
	utils.ProcessSCRResult(t, testContextSecondContract, scrCallBack1, vmcommon.Ok, nil)
	scrCallBack2 := intermediateTxs[1]
	utils.ProcessSCRResult(t, testContextSecondContract, scrCallBack2, vmcommon.Ok, nil)

	res := vm.GetIntValueFromSC(nil, testContextFirstContract.Accounts, firstScAddress, "numCalled")
	require.Equal(t, big.NewInt(2), res)

	res = vm.GetIntValueFromSC(nil, testContextSecondContract.Accounts, secondSCAddress, "callback_count")
	require.Equal(t, big.NewInt(2), res)
}

func TestAsyncCallTransferAndExecute_CrossShard(t *testing.T) {
	testContextFirstContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextFirstContract.Close()

	testContextSecondContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSecondContract.Close()

	testContextSender, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSender.Close()

	firstContractOwner := []byte("12345678901234567890123456789010")
	require.Equal(t, uint32(0), testContextSender.ShardCoordinator.ComputeId(firstContractOwner))

	secondContractOwner := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContextSender.ShardCoordinator.ComputeId(secondContractOwner))

	senderAddr := []byte("12345678901234567890123456789032")
	require.Equal(t, uint32(2), testContextSender.ShardCoordinator.ComputeId(senderAddr))

	_, _ = vm.CreateAccount(testContextSender.Accounts, senderAddr, 0, egldBalance)
	_, _ = vm.CreateAccount(testContextFirstContract.Accounts, firstContractOwner, 0, egldBalance)
	_, _ = vm.CreateAccount(testContextSecondContract.Accounts, secondContractOwner, 0, egldBalance)

	gasPrice := uint64(10)
	gasLimit := uint64(5000000)
	firstSCAccount, _ := testContextFirstContract.Accounts.LoadAccount(firstContractOwner)

	pathToContract := "testdata/first/first.wasm"
	firstSCAddress := utils.DoDeploySecond(t, testContextFirstContract, pathToContract, firstSCAccount, gasPrice, gasLimit, nil, big.NewInt(0))

	secondAccount, _ := testContextSecondContract.Accounts.LoadAccount(secondContractOwner)
	pathToContract = "testdata/forwarderQueue/forwarder-queue-promises.wasm"
	secondSCAddress := utils.DoDeploySecond(t, testContextSecondContract, pathToContract, secondAccount, gasPrice, gasLimit, nil, big.NewInt(0))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextFirstContract)
	utils.CleanAccumulatedIntermediateTransactions(t, testContextSecondContract)

	testContextFirstContract.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	testContextSecondContract.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_transfer_execute").
		Bytes(firstSCAddress).
		Int64(50000).
		Bytes([]byte("callMe"))

	sendFirstCall := int64(5)
	sendSecondCall := int64(10)

	sendTxWithValue(0, sendFirstCall, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContextSecondContract, t)
	sendTxWithValue(1, sendSecondCall, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContextSecondContract, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContextSecondContract, t)

	intermediateTxs := testContextSecondContract.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	// execute cross shard calls
	scrCall1 := intermediateTxs[0]
	utils.ProcessSCRResult(t, testContextFirstContract, scrCall1, vmcommon.Ok, nil)
	scrCall2 := intermediateTxs[1]
	utils.ProcessSCRResult(t, testContextFirstContract, scrCall2, vmcommon.Ok, nil)

	firstSCAccount, err = testContextFirstContract.Accounts.GetExistingAccount(firstSCAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(sendFirstCall+sendSecondCall), firstSCAccount.(state.UserAccountHandler).GetBalance())

	res := vm.GetIntValueFromSC(nil, testContextFirstContract.Accounts, firstSCAddress, "numCalled")
	require.Equal(t, big.NewInt(2), res)
}

func TestAsyncCallTransferESDTAndExecute_CrossShard(t *testing.T) {

	logger.SetLogLevel("*:TRACE")
	// logger.SetLogLevel("vm/runtime:TRACE,vm/async:TRACE")
	// logger.SetLogLevel("vm:TRACE")

	vaultShard, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer vaultShard.Close()

	forwarderShard, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer forwarderShard.Close()

	testContextSender, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSender.Close()

	vaultOwner := []byte("12345678901234567890123456789010")
	require.Equal(t, uint32(0), testContextSender.ShardCoordinator.ComputeId(vaultOwner))

	forwarderOwner := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContextSender.ShardCoordinator.ComputeId(forwarderOwner))

	senderAddr := []byte("12345678901234567890123456789032")
	require.Equal(t, uint32(2), testContextSender.ShardCoordinator.ComputeId(senderAddr))

	_, _ = vm.CreateAccount(testContextSender.Accounts, senderAddr, 0, egldBalance)
	_, _ = vm.CreateAccount(vaultShard.Accounts, vaultOwner, 0, egldBalance)
	_, _ = vm.CreateAccount(forwarderShard.Accounts, forwarderOwner, 0, egldBalance)

	gasPrice := uint64(10)
	gasLimit := uint64(5000000)
	vaultOwnerAccount, _ := vaultShard.Accounts.LoadAccount(vaultOwner)

	pathToContract := "testdata/forwarderQueue/vault.wasm"
	vaultSCAddress := utils.DoDeploySecond(t, vaultShard, pathToContract, vaultOwnerAccount, gasPrice, gasLimit, nil, big.NewInt(0))

	forwarderOwnerAccount, _ := forwarderShard.Accounts.LoadAccount(forwarderOwner)
	pathToContract = "testdata/forwarderQueue/forwarder-queue-promises.wasm"
	forwarderSCAddress := utils.DoDeploySecond(t, forwarderShard, pathToContract, forwarderOwnerAccount, gasPrice, gasLimit, nil, big.NewInt(0))

	utils.CreateAccountWithESDTBalance(t, forwarderShard.Accounts, forwarderSCAddress, egldBalance, esdtToken, 0, esdtBalance)

	utils.CheckESDTNFTBalance(t, forwarderShard, forwarderSCAddress, esdtToken, 0, esdtBalance)

	utils.CleanAccumulatedIntermediateTransactions(t, vaultShard)
	utils.CleanAccumulatedIntermediateTransactions(t, forwarderShard)

	vaultShard.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	forwarderShard.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())

	esdtToTransferFromParent := int64(20)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_transfer_esdt").
		Bytes(vaultSCAddress).
		Int64(50000).
		Bytes([]byte("retrieve_funds_promises")).
		Bytes(esdtToken).
		Int64(esdtToTransferFromParent)

	sendTx(0, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, forwarderShard, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderRunQueue, forwarderShard, t)

	intermediateTxs := forwarderShard.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	scrCall1 := intermediateTxs[0]
	utils.ProcessSCRResult(t, vaultShard, scrCall1, vmcommon.Ok, nil)

	vaultIntermediateTxs := vaultShard.GetIntermediateTransactions(t)
	require.NotNil(t, vaultIntermediateTxs)

	scrCall := vaultIntermediateTxs[1]
	utils.ProcessSCRResult(t, forwarderShard, scrCall, vmcommon.Ok, nil)

	numberOfCalls := int64(1) // TODO update to 2 !!!

	utils.CheckESDTNFTBalance(t, vaultShard, vaultSCAddress, esdtToken, 0,
		big.NewInt(numberOfCalls*esdtToTransferFromParent- /* received */
			esdtToTransferFromParent /* sent back via callback*/))

	utils.CheckESDTNFTBalance(t, forwarderShard, forwarderSCAddress, esdtToken, 0,
		big.NewInt((esdtBalance.Int64() - /* initial */
			numberOfCalls*esdtToTransferFromParent + /* sent via async */
			esdtToTransferFromParent /* received back via callback */)))

	res := vm.GetIntValueFromSC(nil, vaultShard.Accounts, vaultSCAddress, "num_called_retrieve_funds_promises")
	require.Equal(t, big.NewInt(numberOfCalls), res)

	res = vm.GetIntValueFromSC(nil, forwarderShard.Accounts, forwarderSCAddress, "callback_count")
	require.Equal(t, big.NewInt(numberOfCalls), res)
}

func sendTx(nonce uint64, senderAddr []byte, secondSCAddress []byte, gasPrice uint64, gasLimit uint64, txBuilder *txDataBuilder.TxDataBuilder, testContext *vm.VMTestContext, t *testing.T) {
	sendTxWithValue(nonce, 0, senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilder, testContext, t)
}

func sendTxWithValue(nonce uint64, value int64, senderAddr []byte, secondSCAddress []byte, gasPrice uint64, gasLimit uint64, txBuilder *txDataBuilder.TxDataBuilder, testContext *vm.VMTestContext, t *testing.T) {
	tx := vm.CreateTransaction(nonce, big.NewInt(value), senderAddr, secondSCAddress, gasPrice, gasLimit, txBuilder.ToBytes())
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)
}
