package endOfEpoch

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
)

func TestEndOfEpochChangeWithoutTransactionInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	logger.DefaultLogger().SetLevel("DEBUG")

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	roundsPerEpoch := int64(20)
	for _, node := range nodes {
		node.EndOfEpochTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	/////////----- wait for epoch end period
	for i := uint64(0); i < uint64(roundsPerEpoch); i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	time.Sleep(time.Second)

	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	time.Sleep(time.Second)

	epoch := uint32(1)
	/////////------ verify if current blocks are with new epoch
	for _, node := range nodes {
		currentShId := node.ShardCoordinator.SelfId()
		currentHeader := node.BlockChain.GetCurrentBlockHeader()
		assert.Equal(t, epoch, currentHeader.GetEpoch())

		for _, testNode := range nodes {
			if testNode.ShardCoordinator.SelfId() == currentShId {
				testHeader := testNode.BlockChain.GetCurrentBlockHeader()
				assert.Equal(t, testHeader.GetNonce(), currentHeader.GetNonce())
			}
		}
	}

	//////////------ verify last added shardheaders in meta are with new epoch
}
