package multiShard

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestMoveBalanceShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789010")
	shardID := testContext.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(0), shardID)

	rcvAddr := []byte("12345678901234567890123456789021")
	shardID = testContext.ShardCoordinator.ComputeId(rcvAddr)
	require.Equal(t, uint32(1), shardID)

	senderNonce := uint64(0)
	gasPrice := uint64(10)
	gasLimit := uint64(100)

	tx := vm.CreateTransaction(senderNonce, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("aaaa"))

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// verify receiver
	expectedBalanceReceiver := big.NewInt(100)
	utils.TestAccount(t, testContext.Accounts, rcvAddr, 0, expectedBalanceReceiver)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(0), accumulatedFees)
}

func TestMoveBalanceContractAddressDataFieldNilShouldConsumeGas(t *testing.T) {
	t.Parallel()

	testContext, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789010")
	shardID := testContext.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(0), shardID)

	scAddress := "00000000000000000500dbb53e4b23392b0d6f36cce32deb2d623e9625ab3132"
	scAddrBytes, _ := hex.DecodeString(scAddress)
	scAddrBytes[31] = 1
	shardID = testContext.ShardCoordinator.ComputeId(scAddrBytes)
	require.Equal(t, uint32(1), shardID)

	senderNonce := uint64(0)
	gasPrice := uint64(10)
	gasLimit := uint64(100)

	tx := vm.CreateTransaction(senderNonce, big.NewInt(100), sndAddr, scAddrBytes, gasPrice, gasLimit, nil)

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)
	require.Equal(t, errors.New("sending value to non payable contract"), testContext.GetCompositeTestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	account, err := testContext.Accounts.GetExistingAccount(scAddrBytes)
	require.NotNil(t, err)
	require.Nil(t, account)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(0), accumulatedFees)
}

func TestMoveBalanceContractAddressDataFieldNotNilShouldConsumeGas(t *testing.T) {
	t.Parallel()

	testContext, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789010")
	shardID := testContext.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(0), shardID)

	scAddress := "00000000000000000500dbb53e4b23392b0d6f36cce32deb2d623e9625ab3132"
	scAddrBytes, _ := hex.DecodeString(scAddress)
	scAddrBytes[31] = 1
	shardID = testContext.ShardCoordinator.ComputeId(scAddrBytes)
	require.Equal(t, uint32(1), shardID)

	senderNonce := uint64(0)
	gasPrice := uint64(10)
	gasLimit := uint64(100)

	tx := vm.CreateTransaction(senderNonce, big.NewInt(100), sndAddr, scAddrBytes, gasPrice, gasLimit, []byte("function"))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)
	require.Equal(t, errors.New("invalid contract code (not found): contract not found"), testContext.GetCompositeTestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// verify receiver
	expectedBalanceReceiver := big.NewInt(0)
	utils.TestAccount(t, testContext.Accounts, scAddrBytes, 0, expectedBalanceReceiver)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(910), accumulatedFees)
}

func TestMoveBalanceExecuteOneSourceAndDestinationShard(t *testing.T) {
	testContextSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSource.Close()

	testContextDst, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextDst.Close()

	sndAddr := []byte("12345678901234567890123456789010")
	shardID := testContextDst.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(0), shardID)

	rcvAddr := []byte("12345678901234567890123456789011")
	shardID = testContextDst.ShardCoordinator.ComputeId(rcvAddr)
	require.Equal(t, uint32(1), shardID)

	senderNonce := uint64(0)
	gasPrice := uint64(10)
	gasLimit := uint64(100)

	_, _ = vm.CreateAccount(testContextSource.Accounts, sndAddr, 0, big.NewInt(100000))

	tx := vm.CreateTransaction(senderNonce, big.NewInt(100), sndAddr, rcvAddr, gasPrice, gasLimit, []byte("function"))

	// execute on source shard
	retCode, err := testContextSource.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	// verify sender
	expectedBalanceSender := big.NewInt(99810)
	utils.TestAccount(t, testContextSource.Accounts, sndAddr, 1, expectedBalanceSender)

	// check accumulated fees
	accumulatedFees := testContextSource.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(90), accumulatedFees)

	// execute on destination shard
	retCode, err = testContextDst.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextDst.Accounts.Commit()
	require.Nil(t, err)

	// verify receiver
	expectedBalanceReceiver := big.NewInt(100)
	utils.TestAccount(t, testContextDst.Accounts, rcvAddr, 0, expectedBalanceReceiver)

	// check accumulated fees
	accumulatedFees = testContextDst.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(0), accumulatedFees)
}
