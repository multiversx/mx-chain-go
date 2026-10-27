package basicSync

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"
)

// Supernova activates at round 2: the first block is v2, so the test also covers
// the v2 -> v3 transition (a v3 block extending a v2 parent through the synthetic
// execution result). Activation at genesis is not possible: the v3 proposal path
// needs a previous header.
func TestSyncWorksInShard_EmptyBlocksNoForks_Supernova(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// 3 shard nodes and 1 metachain node
	maxShards := uint32(1)
	shardId := uint32(0)
	numNodesPerShard := 3

	enableEpochs := integrationTests.CreateEnableEpochsConfig()
	enableEpochs.AndromedaEnableEpoch = uint32(0)
	enableEpochs.SupernovaEnableEpoch = uint32(0)
	roundsConfig := integrationTests.GetSupernovaRoundConfigActivatedAt(2)

	nodes := make([]*integrationTests.TestProcessorNode, numNodesPerShard+1)
	connectableNodes := make([]integrationTests.Connectable, 0)
	for i := 0; i < numNodesPerShard; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
			MaxShards:            maxShards,
			NodeShardId:          shardId,
			TxSignPrivKeyShardId: shardId,
			WithSync:             true,
			EpochsConfig:         &enableEpochs,
			RoundsConfig:         &roundsConfig,
		})
		connectableNodes = append(connectableNodes, nodes[i])
	}

	metachainNode := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            maxShards,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: shardId,
		WithSync:             true,
		EpochsConfig:         &enableEpochs,
		RoundsConfig:         &roundsConfig,
	})
	idxProposerMeta := numNodesPerShard
	nodes[idxProposerMeta] = metachainNode
	connectableNodes = append(connectableNodes, metachainNode)

	idxProposerShard0 := 0
	leaders := []*integrationTests.TestProcessorNode{nodes[idxProposerShard0], nodes[idxProposerMeta]}

	integrationTests.ConnectNodes(connectableNodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.StartSync()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(nodes, round)
	nonce++

	numRoundsToTest := 8

	for i := 0; i < numRoundsToTest; i++ {
		integrationTests.ProposeBlockWithProof(nodes, leaders, round, nonce)

		time.Sleep(integrationTests.SyncDelay)

		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(nodes, round)
		nonce++
	}

	time.Sleep(integrationTests.SyncDelay)

	require.NotNil(t, nodes[0].BlockChain.GetCurrentBlockHeader())
	expectedNonce := nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce()
	assert.Equal(t, uint64(numRoundsToTest), expectedNonce)
	for i := 1; i < len(nodes); i++ {
		if check.IfNil(nodes[i].BlockChain.GetCurrentBlockHeader()) {
			assert.Fail(t, fmt.Sprintf("Node with idx %d does not have a current block", i))
		} else {
			assert.Equal(t, expectedNonce, nodes[i].BlockChain.GetCurrentBlockHeader().GetNonce())
		}
	}

	// v3 finality must progress on every node (proof-gated)
	for i, n := range nodes {
		assert.Greater(t, n.ForkDetector.GetHighestFinalBlockNonce(), uint64(0),
			fmt.Sprintf("node %d final checkpoint did not advance", i))
	}
}
