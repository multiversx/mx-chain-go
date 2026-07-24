package basicSync

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
)

// mirrors process/block.metaArbitrationWindowRounds, the R-RESOLVE discovery window
const metaArbitrationWindowRounds = 3

func requireNotarizes(t *testing.T, metaHeader data.HeaderHandler, shardID uint32, hashes ...[]byte) {
	metaHandler, ok := metaHeader.(data.MetaHeaderHandler)
	require.True(t, ok)

	for _, shardInfo := range process.GetShardHeadersReferencedByMeta(metaHandler) {
		if shardInfo.GetShardID() != shardID {
			continue
		}
		for _, hash := range hashes {
			if string(shardInfo.GetHeaderHash()) == string(hash) {
				return
			}
		}
	}

	require.Fail(t, "meta block does not notarize the winning branch")
}

// backstop scenario: island 1 instantly finalizes clean sibling A; island 2,
// blind to A, commits contended sibling B AND extends it with a proofed child C.
// When B and C reach island 1, its nodes hold a FINALIZED block that objectively
// lost (childless, competitor settled via proofed child): the reconcile backstop
// must fire -- final checkpoint lowered below the fork nonce (the only sanctioned
// finality regression), the loser blacklisted, and the nodes converge on C.
func TestSupernovaSync_ReconcileBackstop_FinalizedMinorityConverges(t *testing.T) {
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

	pA := newShardNode()   // island 1 proposer: commits and finalizes the losing block
	obsA := newShardNode() // island 1 observer
	pB := newShardNode()   // island 2 proposer: builds the winning branch
	obsB := newShardNode() // island 2 observer
	metaNode := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            maxShards,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: shardId,
		WithSync:             true,
		EpochsConfig:         &enableEpochs,
		RoundsConfig:         &roundsConfig,
	})

	// meta sits on the winning side: it never learns A, so it arbitrates B and its notarization is
	// the authority verdict delivered to the stranded island
	island1 := []*integrationTests.TestProcessorNode{pA, obsA}
	island2 := []*integrationTests.TestProcessorNode{pB, obsB, metaNode}
	allNodes := []*integrationTests.TestProcessorNode{pA, obsA, metaNode, pB, obsB}

	integrationTests.ConnectNodes([]integrationTests.Connectable{pA, obsA})
	integrationTests.ConnectNodes([]integrationTests.Connectable{pB, obsB, metaNode})

	defer func() {
		for _, n := range allNodes {
			n.Close()
		}
	}()

	for _, n := range allNodes {
		_ = n.StartSync()
	}
	time.Sleep(integrationTests.P2pBootstrapDelay)

	injectBlock := func(target *integrationTests.TestProcessorNode, header data.HeaderHandler, hash []byte, proof data.HeaderProofHandler) {
		target.DataPool.Headers().AddHeader(hash, header)
		_ = target.DataPool.Proofs().AddProof(proof)
	}
	mirrorCurrentBlock := func(target *integrationTests.TestProcessorNode, src *integrationTests.TestProcessorNode) {
		header := src.BlockChain.GetCurrentBlockHeader()
		hash := src.BlockChain.GetCurrentBlockHeaderHash()
		require.NotNil(t, header)
		proof, err := src.DataPool.Proofs().GetProof(header.GetShardID(), hash)
		require.Nil(t, err)
		injectBlock(target, header, hash, proof)
	}
	grabCurrentBlock := func(src *integrationTests.TestProcessorNode) (data.HeaderHandler, []byte, data.HeaderProofHandler) {
		header := src.BlockChain.GetCurrentBlockHeader()
		hash := src.BlockChain.GetCurrentBlockHeaderHash()
		require.NotNil(t, header)
		proof, err := src.DataPool.Proofs().GetProof(header.GetShardID(), hash)
		require.Nil(t, err)
		return header, hash, proof
	}

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	nonce++

	// common prefix, rounds 1..4: the shard proposes on island 1, meta notarizes it on island 2,
	// and each side is fed the other's block
	numPrefixBlocks := 4
	for i := 0; i < numPrefixBlocks; i++ {
		integrationTests.ProposeBlockWithProof(island1, []*integrationTests.TestProcessorNode{pA}, round, nonce)
		time.Sleep(integrationTests.SyncDelay)

		for _, target := range island2 {
			mirrorCurrentBlock(target, pA)
		}
		time.Sleep(integrationTests.SyncDelay)

		integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{metaNode}, round, nonce)
		time.Sleep(integrationTests.SyncDelay)

		for _, target := range island1 {
			mirrorCurrentBlock(target, metaNode)
		}
		time.Sleep(integrationTests.SyncDelay)

		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		nonce++
	}

	// meta built its own chain over the prefix and stops here; the shard nonce moves on without it
	nextMetaNonce := nonce

	// round 5: island 1 commits and instantly finalizes the clean sibling A
	integrationTests.ProposeBlockWithProof(island1, []*integrationTests.TestProcessorNode{pA}, round, nonce)
	time.Sleep(integrationTests.SyncDelay)

	headerA, hashA, _ := grabCurrentBlock(pA)
	require.Equal(t, uint64(5), headerA.GetNonce())

	// two silent rounds so the island 2 sibling lands contended
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	time.Sleep(integrationTests.SyncDelay)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// island 1 holds A instantly finalized (clean, proofed)
	require.Equal(t, uint64(5), pA.ForkDetector.GetHighestFinalBlockNonce())
	require.Equal(t, uint64(5), obsA.ForkDetector.GetHighestFinalBlockNonce())

	// round 7: island 2 commits the contended sibling B
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{pB}, round, nonce)
	time.Sleep(integrationTests.SyncDelay)

	headerB, hashB, proofB := grabCurrentBlock(pB)
	require.Equal(t, uint64(5), headerB.GetNonce())
	require.NotEqual(t, string(hashA), string(hashB))
	require.Equal(t, headerA.GetPrevHash(), headerB.GetPrevHash())

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	nonce++

	// round 8: island 2 extends B with proofed child C -- B's branch is the settled one
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{pB}, round, nonce)
	time.Sleep(integrationTests.SyncDelay)

	headerC, hashC, proofC := grabCurrentBlock(pB)
	require.Equal(t, uint64(6), headerC.GetNonce())
	require.Equal(t, string(hashB), string(headerC.GetPrevHash()))

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// B is contended, so meta may only arbitrate it once the discovery window has elapsed
	for i := 0; i < metaArbitrationWindowRounds; i++ {
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		time.Sleep(integrationTests.SyncDelay)
	}

	// meta arbitrates B, then extends itself so the notarizing block is settled and thus held final
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{metaNode}, round, nextMetaNonce)
	time.Sleep(integrationTests.SyncDelay)
	notarizingMeta, notarizingMetaHash, notarizingMetaProof := grabCurrentBlock(metaNode)
	requireNotarizes(t, notarizingMeta, shardId, hashB, hashC)

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{metaNode}, round, nextMetaNonce+1)
	time.Sleep(integrationTests.SyncDelay)
	settlingMeta, settlingMetaHash, settlingMetaProof := grabCurrentBlock(metaNode)
	require.Equal(t, string(notarizingMetaHash), string(settlingMeta.GetPrevHash()))

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// deliver the winning branch AND the authority verdict on it to island 1
	for _, n := range []*integrationTests.TestProcessorNode{pA, obsA} {
		injectBlock(n, notarizingMeta, notarizingMetaHash, notarizingMetaProof)
		injectBlock(n, settlingMeta, settlingMetaHash, settlingMetaProof)
		injectBlock(n, headerB, hashB, proofB)
		injectBlock(n, headerC, hashC, proofC)
	}

	// the backstop fires on sync-loop iterations: tick rounds until island 1 converges on C
	island1OnC := func() bool {
		for _, n := range []*integrationTests.TestProcessorNode{pA, obsA} {
			currentHeader := n.BlockChain.GetCurrentBlockHeader()
			if currentHeader == nil || currentHeader.GetNonce() != 6 {
				return false
			}
			if string(n.BlockChain.GetCurrentBlockHeaderHash()) != string(hashC) {
				return false
			}
		}
		return true
	}
	maxReconcileRounds := 8
	for i := 0; i < maxReconcileRounds && !island1OnC(); i++ {
		// the forced rollback clears pools above the rollback nonce; in production the
		// majority network answers the re-requests -- modeled by re-injecting each round
		for _, n := range []*integrationTests.TestProcessorNode{pA, obsA} {
			injectBlock(n, notarizingMeta, notarizingMetaHash, notarizingMetaProof)
			injectBlock(n, settlingMeta, settlingMetaHash, settlingMetaProof)
			injectBlock(n, headerB, hashB, proofB)
			injectBlock(n, headerC, hashC, proofC)
		}
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		time.Sleep(integrationTests.SyncDelay)
	}

	require.True(t, island1OnC(), "backstop did not converge the finalized minority onto the settled branch")

	// the sanctioned finality regression: final dropped below the fork nonce and B
	// (contended) plus C (descendant of unsettled B) stay non-final pending settlement
	for _, n := range []*integrationTests.TestProcessorNode{pA, obsA} {
		assert.Equal(t, uint64(4), n.ForkDetector.GetHighestFinalBlockNonce())
	}

	// island 2 was never on the losing block and is untouched (the meta node runs its own chain,
	// so only the shard nodes are compared here)
	for _, n := range []*integrationTests.TestProcessorNode{pB, obsB} {
		assert.Equal(t, string(hashC), string(n.BlockChain.GetCurrentBlockHeaderHash()))
		assert.Equal(t, uint64(4), n.ForkDetector.GetHighestFinalBlockNonce())
	}
}
