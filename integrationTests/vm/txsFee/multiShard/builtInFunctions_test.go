package multiShard

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	"github.com/stretchr/testify/require"
)

// Test scenario
// 1. Do a SC deploy on shard 1
// 2. Do a ChangeOwnerAddress (owner of the deployed contract will be moved in shard 0)
// 3. Do a ClaimDeveloperReward (cross shard call , the transaction will be executed on the source shard and the destination shard)
// 4. Execute SCR from context destination on context source ( the new owner will receive the developer rewards)
func TestBuiltInFunctionExecuteOnSourceAndDestinationShouldWork(t *testing.T) {
	testContextSource := vm.CreatePreparedTxProcessorWithVMsMultiShard(t, 0)
	defer testContextSource.Close()

	testContextDst := vm.CreatePreparedTxProcessorWithVMsMultiShard(t, 1)
	defer testContextDst.Close()

	pathToContract := "../../arwen/testdata/counter/output/counter.wasm"
	scAddr, owner := utils.DoDeploy(t, testContextDst, pathToContract)
	require.Equal(t, uint32(1), testContextDst.ShardCoordinator.ComputeId(scAddr))
	require.Equal(t, uint32(1), testContextDst.ShardCoordinator.ComputeId(owner))
	testContextDst.TxFeeHandler.CreateBlockStarted()

	newOwner := []byte("12345678901234567890123456789110")
	require.Equal(t, uint32(0), testContextDst.ShardCoordinator.ComputeId(newOwner))

	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	tx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddr, gasPrice, gasLimit, txData)
	_, err := testContextDst.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Nil(t, testContextDst.GetLatestError())

	_, err = testContextDst.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckOwnerAddr(t, testContextDst, scAddr, newOwner)

	accumulatedFees := testContextDst.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(850), accumulatedFees)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextDst)

	// do a sc call intra shard
	sndAddr := []byte("12345678901234567890123456789111")
	shardID := testContextDst.ShardCoordinator.ComputeId(sndAddr)
	require.Equal(t, uint32(1), shardID)

	gasLimit = uint64(500)
	_, _ = vm.CreateAccount(testContextDst.Accounts, sndAddr, 0, big.NewInt(10000))
	tx = vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddr, gasPrice, gasLimit, []byte("increment"))

	retCode, err := testContextDst.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
	require.Nil(t, testContextDst.GetLatestError())

	_, err = testContextDst.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(6140)
	vm.TestAccount(t, testContextDst.Accounts, sndAddr, 1, expectedBalance)

	accumulatedFees = testContextDst.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(4710), accumulatedFees)

	developerFees := testContextDst.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(376), developerFees)

	// call get developer rewards
	gasLimit = 500
	_, _ = vm.CreateAccount(testContextSource.Accounts, newOwner, 0, big.NewInt(10000))
	txData = []byte(core.BuiltInFunctionClaimDeveloperRewards)
	tx = vm.CreateTransaction(0, big.NewInt(0), newOwner, scAddr, gasPrice, gasLimit, txData)

	// execute claim on source shard
	retCode, err = testContextSource.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
	require.Nil(t, testContextSource.GetLatestError())

	expectedBalance = big.NewInt(9770)
	utils.TestAccount(t, testContextSource.Accounts, newOwner, 1, expectedBalance)

	accumulatedFees = testContextSource.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(230), accumulatedFees)

	developerFees = testContextSource.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)

	// execute claim on destination shard
	retCode, err = testContextDst.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
	require.Nil(t, testContextSource.GetLatestError())

	txs := utils.GetIntermediateTransactions(t, testContextDst)
	scr := txs[1]
	utils.ProcessSCRResult(t, testContextSource, scr, vmcommon.Ok, nil)

	expectedBalance = big.NewInt(9770 + 376)
	utils.TestAccount(t, testContextSource.Accounts, newOwner, 1, expectedBalance)
}
