package txsFee

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

var esdtToken = []byte("miiutoken")
var egldBalance = big.NewInt(50000000000)
var esdtBalance = big.NewInt(100)

func TestAsyncCallLegacy(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	gasLimit := uint64(5000000)

	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")

	childSCAddress, forwarderSCAddress := deployForwarderAndTestContract(
		testContext,
		"testdata/first/output/first.wasm",
		ownerAddr, senderAddr, t,
		egldBalance,
		esdtBalance,
		gasPrice)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_legacy_async").
		Bytes(childSCAddress).
		Bytes([]byte("callMe"))

	sendTx(0, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)
	sendTx(1, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContext, t)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, childSCAddress, "numCalled")
	require.Equal(t, big.NewInt(1), res)
}

func TestAsyncCallMulti(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	gasLimit := uint64(5000000)

	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")

	childSCAddress, forwarderSCAddress := deployForwarderAndTestContract(
		testContext,
		"testdata/first/output/first.wasm",
		ownerAddr, senderAddr, t,
		egldBalance,
		esdtBalance,
		gasPrice)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_promise").
		Bytes(childSCAddress).
		Int64(50000).
		Bytes([]byte("callMe"))

	sendTx(0, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)
	sendTx(1, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContext, t)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, childSCAddress, "numCalled")
	require.Equal(t, big.NewInt(2), res)

	res = vm.GetIntValueFromSC(nil, testContext.Accounts, forwarderSCAddress, "callback_count")
	require.Equal(t, big.NewInt(2), res)
}

func TestAsyncCallTransferAndExecute(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	gasLimit := uint64(5000000)

	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")

	childSCAddress, forwarderSCAddress := deployForwarderAndTestContract(
		testContext,
		"testdata/first/output/first.wasm",
		ownerAddr, senderAddr, t,
		egldBalance,
		esdtBalance,
		gasPrice)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_transfer_execute").
		Bytes(childSCAddress).
		Int64(50000).
		Bytes([]byte("callMe"))

	sendFirstCall := int64(5)
	sendSecondCall := int64(10)

	sendTxWithValue(0, sendFirstCall, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)
	sendTxWithValue(1, sendSecondCall, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContext, t)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	childAccount, err := testContext.Accounts.GetExistingAccount(childSCAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(sendFirstCall+sendSecondCall), childAccount.(state.UserAccountHandler).GetBalance())

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, childSCAddress, "numCalled")
	require.Equal(t, big.NewInt(2), res)
}

func TestAsyncCallTransferESDTAndExecute_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numberOfCallsFromParent := 3
	numberOfBackTransfers := 2
	transferESDTAndExecute(t, numberOfCallsFromParent, numberOfBackTransfers)
}

func transferESDTAndExecute(t *testing.T, numberOfCallsFromParent int, numberOfBackTransfers int) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	gasLimit := uint64(5000000)

	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")

	childSCAddress, forwarderSCAddress := deployForwarderAndTestContract(
		testContext,
		"testdata/forwarderQueue/vault-promises.wasm",
		ownerAddr, senderAddr, t,
		egldBalance,
		esdtBalance,
		gasPrice)

	utils.CheckESDTNFTBalance(t, testContext, forwarderSCAddress, esdtToken, 0, esdtBalance)

	esdtToTransferFromParent := int64(20)
	esdtToTransferBackFromChild := int64(5)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_transfer_esdt").
		Bytes(childSCAddress).
		Int64(50000).
		Bytes([]byte("retrieve_funds_promises")).
		Bytes(esdtToken).
		Int64(esdtToTransferFromParent).
		Int(numberOfBackTransfers).
		Int64(esdtToTransferBackFromChild)

	nonce := uint64(0)

	for call := 0; call < numberOfCallsFromParent; call++ {
		sendTx(nonce, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, testContext, t)
		nonce++
	}

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(nonce, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderRunQueue, testContext, t)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	utils.CheckESDTNFTBalance(t, testContext, childSCAddress, esdtToken, 0,
		big.NewInt(int64(numberOfCallsFromParent)*esdtToTransferFromParent- /* received */
			int64(numberOfCallsFromParent)*int64(numberOfBackTransfers)*esdtToTransferBackFromChild /* sent back via callback*/))

	utils.CheckESDTNFTBalance(t, testContext, forwarderSCAddress, esdtToken, 0,
		big.NewInt(esdtBalance.Int64()- /* initial */
			int64(numberOfCallsFromParent)*esdtToTransferFromParent+ /* sent via async */
			int64(numberOfCallsFromParent)*int64(numberOfBackTransfers)*esdtToTransferBackFromChild /* received back via callback */))

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, childSCAddress, "num_called_retrieve_funds_promises")
	require.Equal(t, big.NewInt(int64(numberOfCallsFromParent)), res)

	res = vm.GetIntValueFromSC(nil, testContext.Accounts, childSCAddress, "num_async_calls_sent_from_child")
	require.Equal(t, big.NewInt(int64(numberOfCallsFromParent)*int64(numberOfBackTransfers)), res)

	res = vm.GetIntValueFromSC(nil, testContext.Accounts, forwarderSCAddress, "callback_count")
	require.Equal(t, big.NewInt(int64(numberOfCallsFromParent)), res)
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

	childSCAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(0))

	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	pathToForwarder := "testdata/forwarderQueue/forwarder-queue-promises.wasm"

	forwarderSCAddress := utils.DoDeploySecond(t, testContext, pathToForwarder, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(0))

	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, forwarderSCAddress, egldBalance, esdtToken, 0, esdtBalance)

	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
	testContext.TxFeeHandler.CreateBlockStarted(gasAndFees)

	return childSCAddress, forwarderSCAddress
}

func TestAsyncCallMulti_CrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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

	gasLimit := uint64(5000000)
	firstAccount, _ := testContextFirstContract.Accounts.LoadAccount(firstContractOwner)

	pathToContract := "testdata/first/output/first.wasm"
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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	childShard, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer childShard.Close()

	forwarderShard, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer forwarderShard.Close()

	testContextSender, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSender.Close()

	childOwner := []byte("12345678901234567890123456789010")
	require.Equal(t, uint32(0), testContextSender.ShardCoordinator.ComputeId(childOwner))

	forwarderOwner := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContextSender.ShardCoordinator.ComputeId(forwarderOwner))

	senderAddr := []byte("12345678901234567890123456789032")
	require.Equal(t, uint32(2), testContextSender.ShardCoordinator.ComputeId(senderAddr))

	_, _ = vm.CreateAccount(testContextSender.Accounts, senderAddr, 0, egldBalance)
	_, _ = vm.CreateAccount(childShard.Accounts, childOwner, 0, egldBalance)
	_, _ = vm.CreateAccount(forwarderShard.Accounts, forwarderOwner, 0, egldBalance)

	gasLimit := uint64(5000000)
	childOwnerAccount, _ := childShard.Accounts.LoadAccount(childOwner)

	pathToContract := "testdata/first/output/first.wasm"
	childSCAddress := utils.DoDeploySecond(t, childShard, pathToContract, childOwnerAccount, gasPrice, gasLimit, nil, big.NewInt(0))

	forwarderOwnerAccount, _ := forwarderShard.Accounts.LoadAccount(forwarderOwner)
	pathToContract = "testdata/forwarderQueue/forwarder-queue-promises.wasm"
	forwarderSCAddress := utils.DoDeploySecond(t, forwarderShard, pathToContract, forwarderOwnerAccount, gasPrice, gasLimit, nil, big.NewInt(0))

	utils.CleanAccumulatedIntermediateTransactions(t, childShard)
	utils.CleanAccumulatedIntermediateTransactions(t, forwarderShard)

	childShard.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	forwarderShard.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_transfer_execute").
		Bytes(childSCAddress).
		Int64(50000).
		Bytes([]byte("callMe"))

	sendFirstCall := int64(5)
	sendSecondCall := int64(10)

	sendTxWithValue(0, sendFirstCall, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, forwarderShard, t)
	sendTxWithValue(1, sendSecondCall, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, forwarderShard, t)

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(2, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderRunQueue, forwarderShard, t)

	intermediateTxs := forwarderShard.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)
	require.Equal(t, 2, len(intermediateTxs))

	// execute cross shard calls
	scrCall1 := intermediateTxs[0]
	utils.ProcessSCRResult(t, childShard, scrCall1, vmcommon.Ok, nil)
	scrCall2 := intermediateTxs[1]
	utils.ProcessSCRResult(t, childShard, scrCall2, vmcommon.Ok, nil)

	childOwnerAccount, err = childShard.Accounts.GetExistingAccount(childSCAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(sendFirstCall+sendSecondCall), childOwnerAccount.(state.UserAccountHandler).GetBalance())

	res := vm.GetIntValueFromSC(nil, childShard.Accounts, childSCAddress, "numCalled")
	require.Equal(t, big.NewInt(2), res)
}

func TestAsyncCallTransferESDTAndExecute_CrossShard_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numberOfCallsFromParent := 3
	numberOfBackTransfers := 2
	transferESDTAndExecuteCrossShard(t, numberOfCallsFromParent, numberOfBackTransfers)
}

func transferESDTAndExecuteCrossShard(t *testing.T, numberOfCallsFromParent int, numberOfBackTransfers int) {
	vaultShard, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	})
	require.Nil(t, err)
	defer vaultShard.Close()

	forwarderShard, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	})
	require.Nil(t, err)
	defer forwarderShard.Close()

	testContextSender, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	})
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

	gasLimit := uint64(5000000)
	vaultOwnerAccount, _ := vaultShard.Accounts.LoadAccount(vaultOwner)

	pathToContract := "testdata/forwarderQueue/vault-promises.wasm"
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
	esdtToTransferBackFromChild := int64(5)

	txBuilderEnqueue := txDataBuilder.NewBuilder()
	txBuilderEnqueue.
		Clear().
		Func("add_queued_call_transfer_esdt").
		Bytes(vaultSCAddress).
		Int64(50000).
		Bytes([]byte("retrieve_funds_promises")).
		Bytes(esdtToken).
		Int64(esdtToTransferFromParent).
		Int(numberOfBackTransfers).
		Int64(esdtToTransferBackFromChild)

	nonce := uint64(0)

	for call := 0; call < numberOfCallsFromParent; call++ {
		sendTx(nonce, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderEnqueue, forwarderShard, t)
		nonce++
	}

	txBuilderRunQueue := txDataBuilder.NewBuilder()
	txBuilderRunQueue.
		Clear().
		Func("forward_queued_calls")

	gasLimit = uint64(50000000)
	sendTx(nonce, senderAddr, forwarderSCAddress, gasPrice, gasLimit, txBuilderRunQueue, forwarderShard, t)

	intermediateTxs := forwarderShard.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	for call := 0; call < numberOfCallsFromParent; call++ {
		scrCall1 := intermediateTxs[call]
		expectedReturnCode := vmcommon.Ok
		utils.ProcessSCRResult(t, vaultShard, scrCall1, expectedReturnCode, nil)
	}

	vaultIntermediateTxs := vaultShard.GetIntermediateTransactions(t)
	require.NotNil(t, vaultIntermediateTxs)

	for call := 0; call < numberOfCallsFromParent*3; call++ {
		scrCall1 := vaultIntermediateTxs[call]
		if bytes.Equal(scrCall1.GetSndAddr(), scrCall1.GetRcvAddr()) {
			continue
		}
		utils.ProcessSCRResult(t, forwarderShard, scrCall1, vmcommon.Ok, nil)
	}

	utils.CheckESDTNFTBalance(t, vaultShard, vaultSCAddress, esdtToken, 0,
		big.NewInt(int64(numberOfCallsFromParent)*esdtToTransferFromParent- /* received */
			int64(numberOfCallsFromParent)*int64(numberOfBackTransfers)*esdtToTransferBackFromChild /* sent back via callback*/))

	utils.CheckESDTNFTBalance(t, forwarderShard, forwarderSCAddress, esdtToken, 0,
		big.NewInt(esdtBalance.Int64()- /* initial */
			int64(numberOfCallsFromParent)*esdtToTransferFromParent+ /* sent via async */
			int64(numberOfCallsFromParent)*int64(numberOfBackTransfers)*esdtToTransferBackFromChild /* received back via callback */))

	res := vm.GetIntValueFromSC(nil, vaultShard.Accounts, vaultSCAddress, "num_called_retrieve_funds_promises")
	require.Equal(t, big.NewInt(int64(numberOfCallsFromParent)), res)

	res = vm.GetIntValueFromSC(nil, vaultShard.Accounts, vaultSCAddress, "num_async_calls_sent_from_child")
	require.Equal(t, big.NewInt(int64(numberOfCallsFromParent)*int64(numberOfBackTransfers)), res)

	res = vm.GetIntValueFromSC(nil, forwarderShard.Accounts, forwarderSCAddress, "callback_count")
	require.Equal(t, big.NewInt(int64(numberOfCallsFromParent)), res)

	resStr := vm.GetStringValueFromSC(nil, forwarderShard.Accounts, forwarderSCAddress, "callback_payments")
	require.Equal(t, "ESDTTransfer@6d696975746f6b656e@05", resStr)
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
