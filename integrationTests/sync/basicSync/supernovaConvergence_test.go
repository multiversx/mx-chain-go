package basicSync

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"
)

// convergence scenario: two disconnected islands of sync nodes are split on
// same-nonce siblings -- clean A (round = parent round + 1) vs contended B (later round,
// same parent). Once the lower-round sibling and its proof reach the island holding B,
// fork detection plus the V3 switch converge everyone on A without crossing any final
// checkpoint. Delivery between islands is direct pool injection: a deterministic
// stand-in for gossip and the proof pull (which are unit-tested separately).
func TestSupernovaSync_NodesSplitOnSiblings_ConvergeOnLowerRound(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := uint32(1)
	shardId := uint32(0)

	enableEpochs := integrationTests.CreateEnableEpochsConfig()
	enableEpochs.AndromedaEnableEpoch = uint32(0)
	enableEpochs.SupernovaEnableEpoch = uint32(0)
	roundsConfig := integrationTests.GetSupernovaRoundConfigActivatedAt(2)

	newShardNode := func() *integrationTests.TestProcessorNode {
		return integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
			MaxShards:            maxShards,
			NodeShardId:          shardId,
			TxSignPrivKeyShardId: shardId,
			WithSync:             true,
			EpochsConfig:         &enableEpochs,
			RoundsConfig:         &roundsConfig,
		})
	}

	pA := newShardNode()   // island 1 proposer
	obsA := newShardNode() // island 1 observer
	pB := newShardNode()   // island 2 proposer
	obsB := newShardNode() // island 2 observer
	metaNode := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            maxShards,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: shardId,
		WithSync:             true,
		EpochsConfig:         &enableEpochs,
		RoundsConfig:         &roundsConfig,
	})

	island1 := []*integrationTests.TestProcessorNode{pA, obsA, metaNode}
	island2 := []*integrationTests.TestProcessorNode{pB, obsB}
	allNodes := []*integrationTests.TestProcessorNode{pA, obsA, metaNode, pB, obsB}
	shardNodes := []*integrationTests.TestProcessorNode{pA, obsA, pB, obsB}

	integrationTests.ConnectNodes([]integrationTests.Connectable{pA, obsA, metaNode})
	integrationTests.ConnectNodes([]integrationTests.Connectable{pB, obsB})

	defer func() {
		for _, n := range allNodes {
			n.Close()
		}
	}()

	for _, n := range allNodes {
		_ = n.StartSync()
	}
	time.Sleep(integrationTests.P2pBootstrapDelay)

	// injectBlock simulates targeted delivery of a committed block and its proof
	injectBlock := func(target *integrationTests.TestProcessorNode, header data.HeaderHandler, hash []byte, proof data.HeaderProofHandler) {
		target.DataPool.Headers().AddHeader(hash, header)
		_ = target.DataPool.Proofs().AddProof(proof)
	}
	// mirrorCurrentBlock delivers src's current committed block (with proof) to target
	mirrorCurrentBlock := func(target *integrationTests.TestProcessorNode, src *integrationTests.TestProcessorNode) {
		header := src.BlockChain.GetCurrentBlockHeader()
		hash := src.BlockChain.GetCurrentBlockHeaderHash()
		require.NotNil(t, header)
		proof, err := src.DataPool.Proofs().GetProof(header.GetShardID(), hash)
		require.Nil(t, err)
		injectBlock(target, header, hash, proof)
	}

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	nonce++

	// common prefix, rounds 1..4: island 1 produces, island 2 receives by injection
	numPrefixBlocks := 4
	for i := 0; i < numPrefixBlocks; i++ {
		integrationTests.ProposeBlockWithProof(island1, []*integrationTests.TestProcessorNode{pA, metaNode}, round, nonce)
		time.Sleep(integrationTests.SyncDelay)

		for _, target := range island2 {
			mirrorCurrentBlock(target, pA)
			mirrorCurrentBlock(target, metaNode)
		}
		time.Sleep(integrationTests.SyncDelay)

		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		nonce++
	}

	for _, n := range shardNodes {
		require.NotNil(t, n.BlockChain.GetCurrentBlockHeader())
		require.Equal(t, uint64(numPrefixBlocks), n.BlockChain.GetCurrentBlockHeader().GetNonce())
	}

	// round 5: island 1 commits the clean sibling A (nonce 5, round = parent round + 1)
	integrationTests.ProposeBlockWithProof(island1, []*integrationTests.TestProcessorNode{pA}, round, nonce)
	time.Sleep(integrationTests.SyncDelay)

	headerA := pA.BlockChain.GetCurrentBlockHeader()
	hashA := pA.BlockChain.GetCurrentBlockHeaderHash()
	require.Equal(t, uint64(5), headerA.GetNonce())
	proofA, err := pA.DataPool.Proofs().GetProof(shardId, hashA)
	require.Nil(t, err)

	// two silent rounds so the island 2 sibling lands with a round gap (contended)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// round 7: island 2, still on nonce 4, commits the contended sibling B
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{pB}, round, nonce)
	time.Sleep(integrationTests.SyncDelay)

	headerB := pB.BlockChain.GetCurrentBlockHeader()
	hashB := pB.BlockChain.GetCurrentBlockHeaderHash()
	require.Equal(t, uint64(5), headerB.GetNonce())
	require.NotEqual(t, string(hashA), string(hashB))
	require.Equal(t, headerA.GetPrevHash(), headerB.GetPrevHash())
	proofB, err := pB.DataPool.Proofs().GetProof(shardId, hashB)
	require.Nil(t, err)

	// the clean sibling finalized instantly on its committers, the contended one did not
	assert.Equal(t, uint64(5), pA.ForkDetector.GetHighestFinalBlockNonce())
	assert.Equal(t, uint64(5), obsA.ForkDetector.GetHighestFinalBlockNonce())
	assert.Equal(t, uint64(4), pB.ForkDetector.GetHighestFinalBlockNonce())
	assert.Equal(t, uint64(4), obsB.ForkDetector.GetHighestFinalBlockNonce())

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// convergence: island 2 learns the lower-round sibling;
	// island 1 learns B and must NOT switch away from its finalized lower-round block
	for _, n := range island2 {
		injectBlock(n, headerA, hashA, proofA)
	}
	injectBlock(pA, headerB, hashB, proofB)
	injectBlock(obsA, headerB, hashB, proofB)

	// the switch executes on sync-loop iterations, which advance with rounds: tick
	// empty rounds until everyone sits on A
	allOnBlockA := func() bool {
		for _, n := range shardNodes {
			currentHeader := n.BlockChain.GetCurrentBlockHeader()
			if currentHeader == nil || currentHeader.GetNonce() != 5 {
				return false
			}
			if string(n.BlockChain.GetCurrentBlockHeaderHash()) != string(hashA) {
				return false
			}
		}
		return true
	}
	maxConvergenceRounds := 6
	for i := 0; i < maxConvergenceRounds && !allOnBlockA(); i++ {
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		time.Sleep(integrationTests.SyncDelay)
	}

	require.True(t, allOnBlockA(), "nodes did not converge on the lower-round sibling within the bounded number of rounds")
	for _, n := range shardNodes {
		assert.Equal(t, uint64(5), n.ForkDetector.GetHighestFinalBlockNonce())
	}

	// the chain stays usable after the switch: island 1 extends A, island 2 follows
	nonce++
	numExtensionBlocks := 2
	for i := 0; i < numExtensionBlocks; i++ {
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)

		integrationTests.ProposeBlockWithProof(island1, []*integrationTests.TestProcessorNode{pA}, round, nonce)
		time.Sleep(integrationTests.SyncDelay)

		for _, target := range island2 {
			mirrorCurrentBlock(target, pA)
		}
		time.Sleep(integrationTests.SyncDelay)
		nonce++
	}

	expectedNonce := uint64(5 + numExtensionBlocks)
	expectedHash := pA.BlockChain.GetCurrentBlockHeaderHash()
	allExtended := func() bool {
		for _, n := range shardNodes {
			currentHeader := n.BlockChain.GetCurrentBlockHeader()
			if currentHeader == nil || currentHeader.GetNonce() != expectedNonce {
				return false
			}
			if string(n.BlockChain.GetCurrentBlockHeaderHash()) != string(expectedHash) {
				return false
			}
		}
		return true
	}
	maxCatchUpRounds := 4
	for i := 0; i < maxCatchUpRounds && !allExtended(); i++ {
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		time.Sleep(integrationTests.SyncDelay)
	}

	require.True(t, allExtended(), "nodes did not extend the adopted branch together")
	// the first extension block carries a round gap (rounds ticked during convergence),
	// so it is contended: finality holds at the fork nonce and never regresses; it would
	// catch up only through meta notarization (covered by the chain-simulator suite)
	for _, n := range shardNodes {
		assert.Equal(t, uint64(5), n.ForkDetector.GetHighestFinalBlockNonce())
	}
}
