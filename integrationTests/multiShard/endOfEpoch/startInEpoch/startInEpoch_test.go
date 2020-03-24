package startInEpoch

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

func TestStartInEpochForAShardNodeInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	totalNodesPerShard := 4
	numNodesPerShardOnline := totalNodesPerShard - 1
	shardCnsSize := 2
	metaCnsSize := 3
	numMetachainNodes := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		numNodesPerShardOnline,
		numMetachainNodes,
		numOfShards,
		shardCnsSize,
		metaCnsSize,
		integrationTests.GetConnectableAddress(advertiser),
	)

	nodes := convertToSlice(nodesMap)

	nodeToJoinLate := nodes[numNodesPerShardOnline] // will return the last node in shard 0 which was not used in consensus
	_ = nodeToJoinLate.Messenger.Close()            // set not offline

	nodes = append(nodes[:numNodesPerShardOnline], nodes[numNodesPerShardOnline+1:]...)
	nodes = append(nodes[:2*numNodesPerShardOnline], nodes[2*numNodesPerShardOnline+1:]...)

	roundsPerEpoch := uint64(10)
	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * numNodesPerShardOnline
	}
	idxProposers[numOfShards] = numOfShards * numNodesPerShardOnline

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	initialVal := big.NewInt(10000000)
	sendValue := big.NewInt(5)
	integrationTests.MintAllNodes(nodes, initialVal)
	receiverAddress := []byte("12345678901234567890123456789012")

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	/////////----- wait for epoch end period
	epoch := uint32(2)
	nrRoundsToPropagateMultiShard := uint64(5)
	for i := uint64(0); i <= (uint64(epoch)*roundsPerEpoch)+nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		for _, node := range nodes {
			integrationTests.CreateAndSendTransaction(node, sendValue, receiverAddress, "")
		}

		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)

	endOfEpoch.VerifyThatNodesHaveCorrectEpoch(t, epoch, nodes)
	endOfEpoch.VerifyIfAddedShardHeadersAreWithNewEpoch(t, nodes)

	epochHandler := &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return epoch
		},
	}
	for _, node := range nodes {
		_ = dataRetriever.SetEpochHandlerToHdrResolver(node.ResolversContainer, epochHandler)
	}

	nodesConfig := sharding.NodesSetup{
		RoundDuration: 4000,
		InitialNodes:  getInitialNodes(nodesMap),
	}
	nodesConfig.SetNumberOfShards(uint32(numOfShards))
}

func convertToSlice(originalMap map[uint32][]*integrationTests.TestProcessorNode) []*integrationTests.TestProcessorNode {
	sliceToRet := make([]*integrationTests.TestProcessorNode, 0)
	for _, nodesPerShard := range originalMap {
		for _, node := range nodesPerShard {
			sliceToRet = append(sliceToRet, node)
		}
	}

	return sliceToRet
}

func getInitialNodes(nodesMap map[uint32][]*integrationTests.TestProcessorNode) []*sharding.InitialNode {
	sliceToRet := make([]*sharding.InitialNode, 0)
	for _, nodesPerShard := range nodesMap {
		for _, node := range nodesPerShard {
			pubKeyBytes, _ := node.NodeKeys.Pk.ToByteArray()
			addressBytes := node.OwnAccount.Address.Bytes()
			entry := &sharding.InitialNode{
				PubKey:   hex.EncodeToString(pubKeyBytes),
				Address:  hex.EncodeToString(addressBytes),
				NodeInfo: sharding.NodeInfo{},
			}
			sliceToRet = append(sliceToRet, entry)
		}
	}

	return sliceToRet
}
