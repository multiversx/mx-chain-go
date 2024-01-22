package polynetworkbridge

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBridgeSetupAndBurn(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	enableEpochs := config.EnableEpochs{
		GlobalMintBurnDisableEpoch:          integrationTests.UnreachableEpoch,
		BuiltInFunctionOnMetaEnableEpoch:    integrationTests.UnreachableEpoch,
		SCProcessorV2EnableEpoch:            integrationTests.UnreachableEpoch,
		FixAsyncCallBackArgsListEnableEpoch: integrationTests.UnreachableEpoch,
	}
	arwenVersion := config.WasmVMVersionByEpoch{Version: "v1.4"}
	vmConfig := &config.VirtualMachineConfig{WasmVMVersions: []config.WasmVMVersionByEpoch{arwenVersion}}
	nodes := integrationTests.CreateNodesWithEnableEpochsAndVmConfig(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochs,
		vmConfig,
	)

	ownerNode := nodes[0]
	shard := nodes[0:nodesPerShard]

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000000)
	initialVal.Mul(initialVal, initialVal)
	fmt.Printf("Initial minted sum: %s\n", initialVal.String())
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	tokenManagerPath := "../testdata/polynetworkbridge/esdt_token_manager.wasm"
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)

	blockChainHook := ownerNode.BlockchainHook
	scAddressBytes, _ := blockChainHook.NewAddress(
		ownerNode.OwnAccount.Address,
		ownerNode.OwnAccount.Nonce,
		factory.WasmVirtualMachine,
	)

	scCode, err := os.ReadFile(tokenManagerPath)
	if err != nil {
		panic(fmt.Sprintf("putDeploySCToDataPool(): %s", err))
	}

	scCodeString := hex.EncodeToString(scCode)
	scCodeMetadataString := "0000"

	deploymentData := scCodeString + "@" + hex.EncodeToString(factory.WasmVirtualMachine) + "@" + scCodeMetadataString

	integrationTests.CreateAndSendTransaction(
		ownerNode,
		shard,
		big.NewInt(0),
		make([]byte, 32),
		deploymentData,
		100000,
	)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 1, nonce, round, idxProposers)

	txValue := big.NewInt(1000)
	txData := "performWrappedEgldIssue@05"
	integrationTests.CreateAndSendTransaction(
		ownerNode,
		shard,
		txValue,
		scAddressBytes,
		txData,
		integrationTests.AdditionalGasLimit,
	)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 8, nonce, round, idxProposers)

	scQuery := &process.SCQuery{
		CallerAddr: ownerNode.OwnAccount.Address,
		ScAddress:  scAddressBytes,
		FuncName:   "getWrappedEgldTokenIdentifier",
		Arguments:  [][]byte{},
	}
	vmOutput, _, err := ownerNode.SCQueryService.ExecuteQuery(scQuery)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.NotZero(t, len(vmOutput.ReturnData[0]))

	tokenIdentifier := vmOutput.ReturnData[0]
	require.Equal(t, []byte("WEGLD"), tokenIdentifier[:5])

	valueToBurn := big.NewInt(5)
	txValue = big.NewInt(0)
	txData = "burnEsdtToken@" + hex.EncodeToString(tokenIdentifier) + "@" + hex.EncodeToString(valueToBurn.Bytes())
	integrationTests.CreateAndSendTransaction(
		ownerNode,
		shard,
		txValue,
		scAddressBytes,
		txData,
		integrationTests.AdditionalGasLimit,
	)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 12, nonce, round, idxProposers)

	checkBurnedOnESDTContract(t, nodes, tokenIdentifier, valueToBurn)
}

func checkBurnedOnESDTContract(t *testing.T, nodes []*integrationTests.TestProcessorNode, tokenIdentifier []byte, burntValue *big.Int) {
	esdtSCAcc := getUserAccountWithAddress(t, vm.ESDTSCAddress, nodes)
	retrievedData, _, _ := esdtSCAcc.RetrieveValue(tokenIdentifier)
	tokenInSystemSC := &systemSmartContracts.ESDTDataV2{}
	_ = integrationTests.TestMarshalizer.Unmarshal(tokenInSystemSC, retrievedData)

	assert.Equal(t, tokenInSystemSC.BurntValue.String(), burntValue.String())
}

func getUserAccountWithAddress(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
) state.UserAccountHandler {
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				acc, err := helperNode.AccntState.LoadAccount(address)
				require.Nil(t, err)
				return acc.(state.UserAccountHandler)
			}
		}
	}

	return nil
}
