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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoChangeOwnerCrossShardFromAContract(t *testing.T) {
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	enableEpochs := config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	}

	testContextSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, enableEpochs)
	require.Nil(t, err)
	defer testContextSource.Close()

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

	pathToContract = "../testdata/first/output/first.wasm"
	secondSCAddress, _ := utils.DoDeployWithCustomParams(t, testContextSecondContract, pathToContract, big.NewInt(100000000000), 15000, nil)
	require.Equal(t, uint32(1), testContextSource.ShardCoordinator.ComputeId(secondSCAddress))

	gasPrice := uint64(10)
	gasLimit := uint64(5000000)
	dataField := append([]byte("change_owner"), []byte("@"+hex.EncodeToString(secondSCAddress))...)
	dataField = append(dataField, []byte("@"+hex.EncodeToString(firstOwner))...)
	tx := vm.CreateTransaction(1, big.NewInt(0), firstOwner, firstContract, gasPrice, gasLimit, dataField)

	// execute on the sender shard
	retCode, err := testContextSource.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	// TODO here should be generated a smart contract results cross shard to call the ChangeOwnerAddress built in function
	intermediateTxs := testContextSource.GetIntermediateTransactions(t)
	assert.Equal(t, 1, len(intermediateTxs))

	logs := testContextSource.TxsLogsProcessor.GetAllCurrentLogs()
	require.NotNil(t, logs)
}
