//go:build !race
// +build !race

package multiShard

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestDeployContractAndTransferValueSCProcessorV1(t *testing.T) {
	testDeployContractAndTransferValue(t, false)
}

func TestDeployContractAndTransferValueSCProcessorV2(t *testing.T) {
	testDeployContractAndTransferValue(t, true)
}

func testDeployContractAndTransferValue(t *testing.T, scProcessorV2Enabled bool) {
	testContextSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextSource.Close()

	configEnabledEpochs := config.EnableEpochs{}
	if !scProcessorV2Enabled {
		configEnabledEpochs.SCDeployEnableEpoch = 1000
	}

	testContextDst, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, configEnabledEpochs)
	require.Nil(t, err)
	defer testContextDst.Close()

	pathToContract := "../testdata/transferValueContract/transferValue.wasm"

	// deploy contract on the source context
	scAddrShardSource, ownerAddress := utils.DoDeployWithCustomParams(t, testContextSource, pathToContract, big.NewInt(100000000000), 20000, nil)
	utils.CleanAccumulatedIntermediateTransactions(t, testContextSource)

	// deploy the same contract on the destination context (in this test we need a payable contract)
	scAddressShardDestination, _ := utils.DoDeployWithCustomParams(t, testContextDst, pathToContract, big.NewInt(100000000000), 20000, nil)
	utils.CleanAccumulatedIntermediateTransactions(t, testContextDst)
	testContextDst.TxsLogsProcessor.Clean()

	receiverHex := hex.EncodeToString(scAddressShardDestination)
	tx := vm.CreateTransaction(1, big.NewInt(1000000), ownerAddress, scAddrShardSource, 10, 20000, []byte("send_egld@"+receiverHex))
	retCode, err := testContextSource.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	intermediateTxs := testContextSource.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	utils.ProcessSCRResult(t, testContextDst, intermediateTxs[0], vmcommon.Ok, nil)
	logs := testContextDst.TxsLogsProcessor.GetAllCurrentLogs()
	require.Len(t, logs, 1)

	generatedLogIdentifier := string(logs[0].LogHandler.GetLogEvents()[0].GetIdentifier())
	require.Equal(t, generatedLogIdentifier, core.CompletedTxEventIdentifier)
}
