package multiShard

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestDoChangeOwnerCrossShardFromAContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochs := config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch:  integrationTests.UnreachableEpoch,
		ChangeOwnerAddressCrossShardThroughSCEnableEpoch: 0,
	}

	testContextSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, enableEpochs)
	require.Nil(t, err)
	defer testContextSource.Close()

	// STEP 1
	// deploy first contract (this contract has a function that will call the `ChangeOwnerAddress` built-in function)
	// on shard 0
	pathToContract := "../testdata/changeOwner/contract.wasm"
	firstContract, firstOwner := utils.DoDeployWithCustomParams(t, testContextSource, pathToContract, big.NewInt(100000000000), 15000, nil)

	utils.CleanAccumulatedIntermediateTransactions(t, testContextSource)
	testContextSource.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	testContextSource.TxsLogsProcessor.Clean()

	require.Equal(t, uint32(0), testContextSource.ShardCoordinator.ComputeId(firstContract))
	require.Equal(t, uint32(0), testContextSource.ShardCoordinator.ComputeId(firstOwner))

	testContextSecondContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, enableEpochs)
	require.Nil(t, err)
	defer testContextSecondContract.Close()

	// STEP 2
	// deploy the second contract on shard 1
	pathToContract = "../testdata/first/output/first.wasm"
	secondSCAddress, deployer := utils.DoDeployWithCustomParams(t, testContextSecondContract, pathToContract, big.NewInt(100000000000), 15000, nil)
	require.Equal(t, uint32(1), testContextSource.ShardCoordinator.ComputeId(secondSCAddress))

	// STEP 3
	// change the owner of the second contract -- the new owner will be the first contract
	gasPrice := uint64(10)
	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(firstContract))
	tx := vm.CreateTransaction(1, big.NewInt(0), deployer, secondSCAddress, gasPrice, 1000, txData)
	returnCode, err := testContextSecondContract.TxProcessor.ProcessTransaction(tx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	_, err = testContextSecondContract.Accounts.Commit()
	require.Nil(t, err)
	utils.CheckOwnerAddr(t, testContextSecondContract, secondSCAddress, firstContract)

	// STEP 3
	// call `change_owner` function from the first contract
	gasLimit := uint64(5000000)
	dataField := append([]byte("change_owner"), []byte("@"+hex.EncodeToString(secondSCAddress))...)
	dataField = append(dataField, []byte("@"+hex.EncodeToString(firstOwner))...)
	tx = vm.CreateTransaction(1, big.NewInt(0), firstOwner, firstContract, gasPrice, gasLimit, dataField)

	retCode, err := testContextSource.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	intermediateTxs := testContextSource.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))

	expectedDataField := core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(firstOwner)
	require.True(t, strings.HasPrefix(string(intermediateTxs[0].GetData()), expectedDataField))

	logs := testContextSource.TxsLogsProcessor.GetAllCurrentLogs()
	require.NotNil(t, logs)

	// STEP 4
	// executed results smart contract results of the shard1 where the second contract was deployed
	testContextSecondContract.TxsLogsProcessor.Clean()
	utils.ProcessSCRResult(t, testContextSecondContract, intermediateTxs[0], vmcommon.Ok, nil)
	utils.CheckOwnerAddr(t, testContextSecondContract, secondSCAddress, firstOwner)

	logs = testContextSecondContract.TxsLogsProcessor.GetAllCurrentLogs()
	require.NotNil(t, logs)
	require.Equal(t, core.BuiltInFunctionChangeOwnerAddress, string(logs[0].GetLogEvents()[0].GetIdentifier()))
}
