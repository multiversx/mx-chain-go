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

func TestRelayedMoveBalanceRelayerShard0InnerTxSenderAndReceiverShard1ShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789030")
	shardID := testContext.ShardCoordinator.ComputeId(relayerAddr)
	require.Equal(t, uint32(0), shardID)

	sndAddr := []byte("12345678901234567890123456789011")
	shardID = testContext.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(1), shardID)

	rcvAddr := []byte("12345678901234567890123456789021")
	shardID = testContext.ShardCoordinator.ComputeId(rcvAddr)
	require.Equal(t, uint32(1), shardID)

	gasPrice := uint64(10)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(100))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(3000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("aaaa"))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := gasLimit + 1 + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// check balance inner tx sender
	utils.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

	// check balance inner tx receiver
	utils.TestAccount(t, testContext.Accounts, rcvAddr, 0, big.NewInt(100))

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(50), accumulatedFees)
}

func TestRelayedMoveBalanceRelayerAndInnerTxSenderShard0ReceiverShard1(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	relayerAddr := []byte("12345678901234567890123456789030")
	shardID := testContext.ShardCoordinator.ComputeId(relayerAddr)
	require.Equal(t, uint32(0), shardID)

	sndAddr := []byte("12345678901234567890123456789011")
	shardID = testContext.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(1), shardID)

	scAddress := "00000000000000000000dbb53e4b23392b0d6f36cce32deb2d623e9625ab3132"
	scAddrBytes, _ := hex.DecodeString(scAddress)
	scAddrBytes[31] = 1
	shardID = testContext.ShardCoordinator.ComputeId(scAddrBytes)
	require.Equal(t, uint32(1), shardID)

	gasPrice := uint64(10)
	gasLimit := uint64(100)

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddrBytes, gasPrice, gasLimit, nil)

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := gasLimit + 1 + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// check inner tx receiver
	account, err := testContext.Accounts.GetExistingAccount(scAddrBytes)
	require.Nil(t, account)
	require.NotNil(t, err)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10), accumulatedFees)
}

func TestRelayedMoveBalanceExecuteOnSourceAndDestination(t *testing.T) {
	testContextSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSource.Close()

	testContextDst, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextDst.Close()

	relayerAddr := []byte("12345678901234567890123456789030")
	shardID := testContextSource.ShardCoordinator.ComputeId(relayerAddr)
	require.Equal(t, uint32(0), shardID)

	sndAddr := []byte("12345678901234567890123456789011")
	shardID = testContextSource.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(1), shardID)

	scAddress := "00000000000000000000dbb53e4b23392b0d6f36cce32deb2d623e9625ab3132"
	scAddrBytes, _ := hex.DecodeString(scAddress)
	scAddrBytes[31] = 1
	shardID = testContextSource.ShardCoordinator.ComputeId(scAddrBytes)
	require.Equal(t, uint32(1), shardID)

	gasPrice := uint64(10)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContextSource.Accounts, relayerAddr, 0, big.NewInt(100000))
	_, _ = vm.CreateAccount(testContextSource.Accounts, sndAddr, 0, big.NewInt(100))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddrBytes, gasPrice, gasLimit, nil)

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	// execute on source shard
	retCode, err := testContextSource.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	// check relayed balance
	// 100000 - rTxFee(164)*gasPrice(10) - gasLimitForMoveInner(1)*gasPrice(10) = 98360
	utils.TestAccount(t, testContextSource.Accounts, relayerAddr, 1, big.NewInt(98360))

	// check accumulated fees
	accumulatedFees := testContextSource.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(1630), accumulatedFees)

	// execute on destination shard
	retCode, err = testContextDst.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContextDst.Accounts.Commit()
	require.Nil(t, err)

	// check inner tx receiver
	account, err := testContextDst.Accounts.GetExistingAccount(scAddrBytes)
	require.Nil(t, account)
	require.NotNil(t, err)

	// check accumulated fees
	accumulatedFees = testContextDst.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10), accumulatedFees)
}

func TestRelayedMoveBalanceExecuteOnSourceAndDestinationRelayerAndInnerTxSenderShard0InnerTxReceiverShard1ShouldWork(t *testing.T) {
	testContextSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSource.Close()

	testContextDst, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextDst.Close()

	relayerAddr := []byte("12345678901234567890123456789030")
	shardID := testContextSource.ShardCoordinator.ComputeId(relayerAddr)
	require.Equal(t, uint32(0), shardID)

	sndAddr := []byte("12345678901234567890123456789010")
	shardID = testContextSource.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(0), shardID)

	rcvAddr := []byte("12345678901234567890123456789011")
	shardID = testContextSource.ShardCoordinator.ComputeId(rcvAddr)
	require.Equal(t, uint32(1), shardID)

	gasPrice := uint64(10)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContextSource.Accounts, relayerAddr, 0, big.NewInt(100000))
	_, _ = vm.CreateAccount(testContextSource.Accounts, sndAddr, 0, big.NewInt(100))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, nil)

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := gasLimit + 1 + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	// execute on source shard
	retCode, err := testContextSource.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	// check relayed balance
	// 100000 - rTxFee(164)*gasPrice(10) - gasLimitForMoveInner(1)*gasPrice(10) = 98360
	utils.TestAccount(t, testContextSource.Accounts, relayerAddr, 1, big.NewInt(98360))
	// check inner tx sender
	utils.TestAccount(t, testContextSource.Accounts, sndAddr, 1, big.NewInt(0))

	// check accumulated fees
	accumulatedFees := testContextSource.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(1640), accumulatedFees)

	// get scr for destination shard
	txs := testContextSource.GetIntermediateTransactions(t)
	scr := txs[0]

	utils.ProcessSCRResult(t, testContextDst, scr, vmcommon.Ok, nil)

	// check balance receiver
	utils.TestAccount(t, testContextDst.Accounts, rcvAddr, 0, big.NewInt(100))

	// check accumulated fess
	accumulatedFees = testContextDst.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(0), accumulatedFees)
}

func TestRelayedMoveBalanceRelayerAndInnerTxReceiverShard0SenderShard1(t *testing.T) {
	testContextSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSource.Close()

	testContextDst, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextDst.Close()

	relayerAddr := []byte("12345678901234567890123456789030")
	shardID := testContextSource.ShardCoordinator.ComputeId(relayerAddr)
	require.Equal(t, uint32(0), shardID)

	sndAddr := []byte("12345678901234567890123456789011")
	shardID = testContextSource.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(1), shardID)

	rcvAddr := []byte("12345678901234567890123456789010")
	shardID = testContextSource.ShardCoordinator.ComputeId(rcvAddr)
	require.Equal(t, uint32(0), shardID)

	gasPrice := uint64(10)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContextSource.Accounts, relayerAddr, 0, big.NewInt(100000))
	_, _ = vm.CreateAccount(testContextDst.Accounts, sndAddr, 0, big.NewInt(100))

	innerTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, nil)

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	// execute on relayer shard
	retCode, err := testContextSource.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	// check relayed balance
	// 100000 - rTxFee(164)*gasPrice(10) - gasLimitForMoveInner(1)*gasPrice(10) = 98360
	utils.TestAccount(t, testContextSource.Accounts, relayerAddr, 1, big.NewInt(98360))

	// check inner Tx receiver
	innerTxSenderAccount, err := testContextSource.Accounts.GetExistingAccount(sndAddr)
	require.Nil(t, innerTxSenderAccount)
	require.NotNil(t, err)

	// check accumulated fees
	accumulatedFees := testContextSource.TxFeeHandler.GetAccumulatedFees()
	expectedAccFees := big.NewInt(1630)
	require.Equal(t, expectedAccFees, accumulatedFees)

	// execute on destination shard
	retCode, err = testContextDst.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	utils.TestAccount(t, testContextDst.Accounts, sndAddr, 1, big.NewInt(0))

	// check accumulated fees
	accumulatedFees = testContextDst.TxFeeHandler.GetAccumulatedFees()
	expectedAccFees = big.NewInt(10)
	require.Equal(t, expectedAccFees, accumulatedFees)

	txs := testContextDst.GetIntermediateTransactions(t)
	scr := txs[0]

	// execute generated SCR from shard1 on shard 0
	utils.ProcessSCRResult(t, testContextSource, scr, vmcommon.Ok, nil)

	// check receiver balance
	utils.TestAccount(t, testContextSource.Accounts, rcvAddr, 0, big.NewInt(100))
}

func TestMoveBalanceRelayerShard0InnerTxSenderShard1InnerTxReceiverShard2ShouldWork(t *testing.T) {
	testContextRelayer, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextRelayer.Close()

	testContextInnerSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextInnerSource.Close()

	testContextDst, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextDst.Close()

	relayerAddr := []byte("12345678901234567890123456789030")
	shardID := testContextRelayer.ShardCoordinator.ComputeId(relayerAddr)
	require.Equal(t, uint32(0), shardID)

	sndAddr := []byte("12345678901234567890123456789011")
	shardID = testContextRelayer.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(1), shardID)

	rcvAddr := []byte("12345678901234567890123456789012")
	shardID = testContextRelayer.ShardCoordinator.ComputeId(rcvAddr)
	require.Equal(t, uint32(2), shardID)

	gasPrice := uint64(10)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContextRelayer.Accounts, relayerAddr, 0, big.NewInt(100000))
	_, _ = vm.CreateAccount(testContextInnerSource.Accounts, sndAddr, 0, big.NewInt(100))

	innerTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, nil)

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, big.NewInt(0), relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	// execute on relayer shard
	retCode, err := testContextRelayer.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	// check relayed balance
	// 100000 - rTxFee(164)*gasPrice(10) - gasLimitForMoveInner(1)*gasPrice(10) = 98360
	utils.TestAccount(t, testContextRelayer.Accounts, relayerAddr, 1, big.NewInt(98360))

	// check inner Tx receiver
	innerTxSenderAccount, err := testContextRelayer.Accounts.GetExistingAccount(sndAddr)
	require.Nil(t, innerTxSenderAccount)
	require.NotNil(t, err)

	// check accumulated fees
	accumulatedFees := testContextRelayer.TxFeeHandler.GetAccumulatedFees()
	expectedAccFees := big.NewInt(1630)
	require.Equal(t, expectedAccFees, accumulatedFees)

	// execute on inner tx sender shard
	retCode, err = testContextInnerSource.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	utils.TestAccount(t, testContextInnerSource.Accounts, sndAddr, 1, big.NewInt(0))

	// check accumulated fees
	accumulatedFees = testContextInnerSource.TxFeeHandler.GetAccumulatedFees()
	expectedAccFees = big.NewInt(10)
	require.Equal(t, expectedAccFees, accumulatedFees)

	// execute on inner tx receiver shard
	txs := testContextInnerSource.GetIntermediateTransactions(t)
	scr := txs[0]

	utils.ProcessSCRResult(t, testContextDst, scr, vmcommon.Ok, nil)

	// check receiver balance
	utils.TestAccount(t, testContextDst.Accounts, rcvAddr, 0, big.NewInt(100))
}
