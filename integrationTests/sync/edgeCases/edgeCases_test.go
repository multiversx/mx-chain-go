package edgeCases

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSyncMetaNodeIsSyncingReceivedHigherRoundBlockFromShard tests the following scenario:
// 1. Meta and shard 0 are in sync, producing blocks
// 2. At nonce 3, shard 0 makes a rollback and stops producing blocks for 2 rounds, meta keeps producing blocks
// 3. Shard 0 resumes creating blocks starting with nonce 3
// 3. A bootstrapping meta node should be able to pass meta block with nonce 2
func TestSyncMetaNodeIsSyncingReceivedHigherRoundBlockFromShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodesPerShard := 3
	numNodesMeta := 3

	nodes, advertiser, idxProposers := integrationTests.SetupSyncNodesOneShardAndMeta(numNodesPerShard, numNodesMeta)
	idxProposerMeta := idxProposers[1]
	defer integrationTests.CloseProcessorNodes(nodes, advertiser)

	integrationTests.StartP2PBootstrapOnProcessorNodes(nodes)
	integrationTests.StartSyncingBlocks(nodes)

	round := uint64(0)
	idxNonceShard := 0
	idxNonceMeta := 1
	nonces := []*uint64{new(uint64), new(uint64)}

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(nodes, round)
	integrationTests.IncrementNonces(nonces)

	numRoundsBlocksAreProposedCorrectly := 3
	integrationTests.ProposeBlocks(
		nodes,
		&round,
		idxProposers,
		nonces,
		numRoundsBlocksAreProposedCorrectly,
	)

	shardIdToRollbackLastBlock := uint32(0)
	integrationTests.ForkChoiceOneBlock(nodes, shardIdToRollbackLastBlock)
	integrationTests.ResetHighestProbableNonce(nodes, shardIdToRollbackLastBlock, 2)
	integrationTests.EmptyDataPools(nodes, shardIdToRollbackLastBlock)

	//revert also the nonce, so the same block nonce will be used when shard will propose the next block
	atomic.AddUint64(nonces[idxNonceShard], ^uint64(0))

	numRoundsBlocksAreProposedOnlyByMeta := 2
	integrationTests.ProposeBlocks(
		nodes,
		&round,
		[]int{idxProposerMeta},
		[]*uint64{nonces[idxNonceMeta]},
		numRoundsBlocksAreProposedOnlyByMeta,
	)

	secondNumRoundsBlocksAreProposedCorrectly := 2
	integrationTests.ProposeBlocks(
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
		core.MetachainShardId,
		shardId,
		advertiserAddr,
	)
	nodes = append(nodes, syncMetaNode)
	syncMetaNode.RoundHandler.IndexField = int64(round)

	syncNodesSlice := []*integrationTests.TestProcessorNode{syncMetaNode}
	integrationTests.StartP2PBootstrapOnProcessorNodes(syncNodesSlice)

	require.True(t, len(syncMetaNode.Messenger.ConnectedPeers()) > 1, "not enough peers connected to this node."+
		" Check that the peer discovery mechanism works properly.")

	integrationTests.StartSyncingBlocks(syncNodesSlice)

	//after joining the network we must propose a new block on the metachain as to be received by the sync
	//node and to start the bootstrapping process
	integrationTests.ProposeBlocks(
		nodes,
		&round,
		[]int{idxProposerMeta},
		[]*uint64{nonces[idxNonceMeta]},
		1,
	)

	numOfRoundsToWaitToCatchUp := numRoundsBlocksAreProposedCorrectly +
		numRoundsBlocksAreProposedOnlyByMeta +
		secondNumRoundsBlocksAreProposedCorrectly
	time.Sleep(integrationTests.SyncDelay * time.Duration(numOfRoundsToWaitToCatchUp))
	integrationTests.UpdateRound(nodes, round)

	nonceProposerMeta := nodes[idxProposerMeta].BlockChain.GetCurrentBlockHeader().GetNonce()
	nonceSyncNode := syncMetaNode.BlockChain.GetCurrentBlockHeader().GetNonce()
	assert.Equal(t, nonceProposerMeta, nonceSyncNode)
}
