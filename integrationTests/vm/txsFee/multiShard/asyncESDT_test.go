package multiShard

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestAsyncESDTTransferWithSCCallShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochs := config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	}

	testContextSender, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, enableEpochs)
	require.Nil(t, err)
	defer testContextSender.Close()

	testContextFirstContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, enableEpochs)
	require.Nil(t, err)
	defer testContextFirstContract.Close()

	testContextSecondContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, enableEpochs)
	require.Nil(t, err)
	defer testContextSecondContract.Close()

	senderAddr := []byte("12345678901234567890123456789030")
	require.Equal(t, uint32(0), testContextSender.ShardCoordinator.ComputeId(senderAddr))

	firstContractOwner := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContextSender.ShardCoordinator.ComputeId(firstContractOwner))

	secondContractOwner := []byte("12345678901234567890123456789012")
	require.Equal(t, uint32(2), testContextSender.ShardCoordinator.ComputeId(secondContractOwner))

	token := []byte("miiutoken")
	egldBalance := big.NewInt(10000000)
	esdtBalance := big.NewInt(10000000)
	utils.CreateAccountWithESDTBalance(t, testContextSender.Accounts, senderAddr, egldBalance, token, 0, esdtBalance)

	// create accounts for owners
	_, _ = vm.CreateAccount(testContextFirstContract.Accounts, firstContractOwner, 0, egldBalance)
	_, _ = vm.CreateAccount(testContextSecondContract.Accounts, secondContractOwner, 0, egldBalance)

	// deploy contracts on shard 1 and shard 2
	gasPrice := uint64(10)
	gasLimitDeploy := uint64(50000)

	secondContractOwnerAcc, _ := testContextSecondContract.Accounts.LoadAccount(secondContractOwner)
	argsSecond := [][]byte{[]byte(hex.EncodeToString(token))}
	secondSCAddress := utils.DoDeploySecond(t, testContextSecondContract, "../../esdt/testdata/second-contract.wasm", secondContractOwnerAcc, gasPrice, gasLimitDeploy, argsSecond, big.NewInt(0))
	testContextSecondContract.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())

	firstContractOwnerAcc, _ := testContextFirstContract.Accounts.LoadAccount(firstContractOwner)
	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	firstSCAddress := utils.DoDeploySecond(t, testContextFirstContract, "../../esdt/testdata/first-contract.wasm", firstContractOwnerAcc, gasPrice, gasLimitDeploy, args, big.NewInt(0))
	testContextFirstContract.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())

	require.Equal(t, uint32(1), testContextSender.ShardCoordinator.ComputeId(firstSCAddress))
	require.Equal(t, uint32(2), testContextSender.ShardCoordinator.ComputeId(secondSCAddress))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextFirstContract)
	utils.CleanAccumulatedIntermediateTransactions(t, testContextSecondContract)

	gasLimit := uint64(500000)
	tx := utils.CreateESDTTransferTx(0, senderAddr, firstSCAddress, token, big.NewInt(5000), gasPrice, gasLimit)
	tx.Data = []byte(string(tx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractHalf")))

	// execute on the source shard
	retCode, err := testContextSender.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	logs := testContextSender.TxsLogsProcessor.GetAllCurrentLogs()
	require.Len(t, logs[0].GetLogEvents(), 1)

	_, err = testContextSender.Accounts.Commit()
	require.Nil(t, err)

	expectedAccumulatedFees := big.NewInt(950)
	require.Equal(t, expectedAccumulatedFees, testContextSender.TxFeeHandler.GetAccumulatedFees())

	// execute on the destination shard
	retCode, err = testContextFirstContract.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextSender.Accounts.Commit()
	require.Nil(t, err)

	expectedAccumulatedFees = big.NewInt(189890)
	require.Equal(t, expectedAccumulatedFees, testContextFirstContract.TxFeeHandler.GetAccumulatedFees())

	developerFees := big.NewInt(18989)
	require.Equal(t, developerFees, testContextFirstContract.TxFeeHandler.GetDeveloperFees())

	utils.CheckESDTBalance(t, testContextFirstContract, firstSCAddress, token, big.NewInt(2500))

	intermediateTxs := testContextFirstContract.GetIntermediateTransactions(t)

	scrForSecondContract := intermediateTxs[1]
	require.Equal(t, scrForSecondContract.GetSndAddr(), firstSCAddress)
	require.Equal(t, scrForSecondContract.GetRcvAddr(), secondSCAddress)
	utils.ProcessSCRResult(t, testContextSecondContract, scrForSecondContract, vmcommon.Ok, nil)

	utils.CheckESDTBalance(t, testContextSecondContract, secondSCAddress, token, big.NewInt(2500))

	accumulatedFee := big.NewInt(62340)
	require.Equal(t, accumulatedFee, testContextSecondContract.TxFeeHandler.GetAccumulatedFees())

	developerFees = big.NewInt(6234)
	require.Equal(t, developerFees, testContextSecondContract.TxFeeHandler.GetDeveloperFees())

	intermediateTxs = testContextSecondContract.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	utils.ProcessSCRResult(t, testContextFirstContract, intermediateTxs[1], vmcommon.Ok, nil)

	require.Equal(t, big.NewInt(4936720), testContextFirstContract.TxFeeHandler.GetAccumulatedFees())
}

func TestAsyncESDTTransferWithSCCallSecondContractAnotherToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochs := config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	}

	testContextSender, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, enableEpochs)
	require.Nil(t, err)
	defer testContextSender.Close()

	testContextFirstContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, enableEpochs)
	require.Nil(t, err)
	defer testContextFirstContract.Close()

	testContextSecondContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, enableEpochs)
	require.Nil(t, err)
	defer testContextSecondContract.Close()

	senderAddr := []byte("12345678901234567890123456789030")
	require.Equal(t, uint32(0), testContextSender.ShardCoordinator.ComputeId(senderAddr))

	firstContractOwner := []byte("12345678901234567890123456789011")
	require.Equal(t, uint32(1), testContextSender.ShardCoordinator.ComputeId(firstContractOwner))

	secondContractOwner := []byte("12345678901234567890123456789012")
	require.Equal(t, uint32(2), testContextSender.ShardCoordinator.ComputeId(secondContractOwner))

	token := []byte("miiutoken")
	egldBalance := big.NewInt(10000000)
	esdtBalance := big.NewInt(10000000)
	utils.CreateAccountWithESDTBalance(t, testContextSender.Accounts, senderAddr, egldBalance, token, 0, esdtBalance)

	// create accounts for owners
	_, _ = vm.CreateAccount(testContextFirstContract.Accounts, firstContractOwner, 0, egldBalance)
	_, _ = vm.CreateAccount(testContextSecondContract.Accounts, secondContractOwner, 0, egldBalance)

	// deploy contracts on shard 1 and shard 2
	gasPrice := uint64(10)
	gasLimitDeploy := uint64(50000)

	secondContractOwnerAcc, _ := testContextSecondContract.Accounts.LoadAccount(secondContractOwner)
	argsSecond := [][]byte{[]byte(hex.EncodeToString(append(token, []byte("aaa")...)))}
	secondSCAddress := utils.DoDeploySecond(t, testContextSecondContract, "../../esdt/testdata/second-contract.wasm", secondContractOwnerAcc, gasPrice, gasLimitDeploy, argsSecond, big.NewInt(0))
	testContextSecondContract.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())

	firstContractOwnerAcc, _ := testContextFirstContract.Accounts.LoadAccount(firstContractOwner)
	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	firstSCAddress := utils.DoDeploySecond(t, testContextFirstContract, "../../esdt/testdata/first-contract.wasm", firstContractOwnerAcc, gasPrice, gasLimitDeploy, args, big.NewInt(0))
	testContextFirstContract.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())

	require.Equal(t, uint32(1), testContextSender.ShardCoordinator.ComputeId(firstSCAddress))
	require.Equal(t, uint32(2), testContextSender.ShardCoordinator.ComputeId(secondSCAddress))

	utils.CleanAccumulatedIntermediateTransactions(t, testContextFirstContract)
	utils.CleanAccumulatedIntermediateTransactions(t, testContextSecondContract)

	gasLimit := uint64(500000)
	tx := utils.CreateESDTTransferTx(0, senderAddr, firstSCAddress, token, big.NewInt(5000), gasPrice, gasLimit)
	tx.Data = []byte(string(tx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractHalf")))

	// execute on the source shard
	retCode, err := testContextSender.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextSender.Accounts.Commit()
	require.Nil(t, err)

	expectedAccumulatedFees := big.NewInt(950)
	require.Equal(t, expectedAccumulatedFees, testContextSender.TxFeeHandler.GetAccumulatedFees())

	// execute on the destination shard
	retCode, err = testContextFirstContract.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextSender.Accounts.Commit()
	require.Nil(t, err)

	expectedAccumulatedFees = big.NewInt(189890)
	require.Equal(t, expectedAccumulatedFees, testContextFirstContract.TxFeeHandler.GetAccumulatedFees())

	developerFees := big.NewInt(18989)
	require.Equal(t, developerFees, testContextFirstContract.TxFeeHandler.GetDeveloperFees())

	utils.CheckESDTBalance(t, testContextFirstContract, firstSCAddress, token, big.NewInt(2500))

	intermediateTxs := testContextFirstContract.GetIntermediateTransactions(t)
	scrForSecondContract := intermediateTxs[1]
	require.Equal(t, scrForSecondContract.GetSndAddr(), firstSCAddress)
	require.Equal(t, scrForSecondContract.GetRcvAddr(), secondSCAddress)
	utils.ProcessSCRResult(t, testContextSecondContract, scrForSecondContract, vmcommon.UserError, nil)

	utils.CheckESDTBalance(t, testContextSecondContract, secondSCAddress, token, big.NewInt(0))

	accumulatedFee := big.NewInt(3720770)
	require.Equal(t, accumulatedFee, testContextSecondContract.TxFeeHandler.GetAccumulatedFees())

	developerFees = big.NewInt(0)
	require.Equal(t, developerFees, testContextSecondContract.TxFeeHandler.GetDeveloperFees())
	// consumed fee 5 000 000 = 950 + 3 740 770 + 1 258 270 + 10 (built in function call)
	intermediateTxs = testContextSecondContract.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	utils.ProcessSCRResult(t, testContextFirstContract, intermediateTxs[0], vmcommon.Ok, nil)

	require.Equal(t, big.NewInt(1278290), testContextFirstContract.TxFeeHandler.GetAccumulatedFees())
}
