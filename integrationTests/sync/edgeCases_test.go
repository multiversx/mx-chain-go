package sync

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

// TestSyncMetaNodeIsSyncingReceivedHigherRoundBlockFromShard tests the following scenario:
// 1. Meta and shard 0 are in sync, shard 0 stops producing blocks
// 2. Shard 0 resumes producing block, having a new block with nonce 5 in a higher round than notarized by metachain
// 3. A bootstrapping meta node should be able to pass block with nonce 4
func TestSyncMetaNodeIsSyncingReceivedHigherRoundBlockFromShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodesPerShard := 3
	numNodesMeta := 3

	nodes, advertiser, idxProposers := setupSyncNodesOneShardAndMeta(numNodesPerShard, numNodesMeta)
	idxProposerMeta := idxProposers[1]
	defer integrationTests.CloseProcessorNodes(nodes, advertiser)

	integrationTests.StartP2pBootstrapOnProcessorNodes(nodes)
	startSyncingBlocks(nodes)

	round := uint64(0)
	idxNonceMeta := 1
	nonces := []*uint64{new(uint64), new(uint64)}

	round = integrationTests.IncrementAndPrintRound(round)
	updateRound(nodes, round)
	incrementNonces(nonces)

	numRoundsBlocksAreProposedCorrectly := 2
	proposeAndSyncBlocks(
		nodes,
		&round,
		idxProposers,
		nonces,
		numRoundsBlocksAreProposedCorrectly,
	)

	numRoundsBlocksAreProposedOnlyByMeta := 2
	proposeAndSyncBlocks(
		nodes,
		&round,
		[]int{idxProposerMeta},
		[]*uint64{nonces[idxNonceMeta]},
		numRoundsBlocksAreProposedOnlyByMeta,
	)

	secondNumRoundsBlocksAreProposedCorrectly := 2
	proposeAndSyncBlocks(
		nodes,
		&round,
		idxProposers,
		nonces,
		secondNumRoundsBlocksAreProposedCorrectly,
	)

	maxShards := uint32(1)
	shardId := uint32(0)
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)
	syncMetaNode := integrationTests.NewTestSyncNode(
		maxShards,
		sharding.MetachainShardId,
		shardId,
		advertiserAddr,
	)
	nodes = append(nodes, syncMetaNode)
	syncMetaNode.Rounder.IndexField = int64(round)

	syncNodesSlice := []*integrationTests.TestProcessorNode{syncMetaNode}
	integrationTests.StartP2pBootstrapOnProcessorNodes(syncNodesSlice)
	startSyncingBlocks(syncNodesSlice)

	//after joining the network we must propose a new block on the metachain as to be received by the sync
	//node and to start the bootstrapping process
	proposeAndSyncBlocks(
		nodes,
		&round,
		[]int{idxProposerMeta},
		[]*uint64{nonces[idxNonceMeta]},
		1,
	)

	numOfRoundsToWaitToCatchUp := numRoundsBlocksAreProposedCorrectly +
		numRoundsBlocksAreProposedOnlyByMeta +
		secondNumRoundsBlocksAreProposedCorrectly
	time.Sleep(stepSync * time.Duration(numOfRoundsToWaitToCatchUp))
	updateRound(nodes, round)

	nonceProposerMeta := nodes[idxProposerMeta].BlockChain.GetCurrentBlockHeader().GetNonce()
	nonceSyncNode := syncMetaNode.BlockChain.GetCurrentBlockHeader().GetNonce()
	assert.Equal(t, nonceProposerMeta, nonceSyncNode)
}
