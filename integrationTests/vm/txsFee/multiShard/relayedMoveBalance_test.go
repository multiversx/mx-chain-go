package multiShard

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	"github.com/stretchr/testify/require"
)

func TestRelayedMoveBalanceRelayerShard0InnerTxSenderAndReceiverShard1ShouldWork(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMsMultiShard(t, 1)
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

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("aaaa"))

	rtxData := utils.PrepareRelayerTxData(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// check balance inner tx sender
	utils.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(0))

	// check balance inner tx receiver
	utils.TestAccount(t, testContext.Accounts, rcvAddr, 0, big.NewInt(100))

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(1000), accumulatedFees)
}

func TestRelayedMoveBalanceRelayerAndInnerTxSenderShard0ReceiverShard1(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMsMultiShard(t, 1)
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

	// gas consumed = 50
	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddrBytes, gasPrice, gasLimit, nil)

	rtxData := utils.PrepareRelayerTxData(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// check inner tx receiver
	account, err := testContext.Accounts.GetExistingAccount(scAddrBytes)
	require.Nil(t, account)
	require.NotNil(t, err)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(1000), accumulatedFees)
}
