package polynetworkbridge

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/require"
)

func TestBridge_Setup(t *testing.T) {
	t.Skip("contract is not yet finished")

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
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
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
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
	// The ESDT Token Manager contract needs the address of the Cross-Chain
	// Management contract at initialization, so we provide a dummy address here.
	tokenManagerInitParams := "ef2b8bbc7777f5cb92c1c13a2eb307067ce76c3d1bd9fda8d69495ba2ae1e334"

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)

	blockChainHook := ownerNode.BlockchainHook
	scAddressBytes, _ := blockChainHook.NewAddress(
		ownerNode.OwnAccount.Address,
		ownerNode.OwnAccount.Nonce,
		factory.ArwenVirtualMachine,
	)

	scCode, err := ioutil.ReadFile(tokenManagerPath)
	if err != nil {
		panic(fmt.Sprintf("putDeploySCToDataPool(): %s", err))
	}

	scCodeString := hex.EncodeToString(scCode)
	scCodeMetadataString := "0000"

	deploymentData := scCodeString + "@" + hex.EncodeToString(factory.ArwenVirtualMachine) + "@" + scCodeMetadataString + "@" + tokenManagerInitParams

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
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 6, nonce, round, idxProposers)

	scQuery := &process.SCQuery{
		CallerAddr: ownerNode.OwnAccount.Address,
		ScAddress:  scAddressBytes,
		FuncName:   "getWrappedEgldTokenIdentifier",
		Arguments:  [][]byte{},
	}
	vmOutput, _ := ownerNode.SCQueryService.ExecuteQuery(scQuery)
	require.NotNil(t, vmOutput)
	fmt.Println()
	fmt.Println(vmOutput.ReturnData)
	require.NotZero(t, len(vmOutput.ReturnData[0]))
	require.Equal(t, []byte("WEGLD"), vmOutput.ReturnData[0][:4])
}
