package basicSync

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
)

// mirrors process/block.metaArbitrationWindowRounds, the arbitration discovery window
const metaArbitrationWindowRounds = 3

func injectBlock(target *integrationTests.TestProcessorNode, header data.HeaderHandler, hash []byte, proof data.HeaderProofHandler) {
	target.DataPool.Headers().AddHeader(hash, header)
	_ = target.DataPool.Proofs().AddProof(proof)
}

func mirrorCurrentBlock(t *testing.T, target *integrationTests.TestProcessorNode, src *integrationTests.TestProcessorNode) {
	header, hash, proof := grabCurrentBlock(t, src)
	injectBlock(target, header, hash, proof)
}

func grabCurrentBlock(t *testing.T, src *integrationTests.TestProcessorNode) (data.HeaderHandler, []byte, data.HeaderProofHandler) {
	header := src.BlockChain.GetCurrentBlockHeader()
	hash := src.BlockChain.GetCurrentBlockHeaderHash()
	require.NotNil(t, header)
	proof, err := src.DataPool.Proofs().GetProof(header.GetShardID(), hash)
	require.Nil(t, err)
	return header, hash, proof
}

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
			mirrorCurrentBlock(t, target, pA)
		}
		time.Sleep(integrationTests.SyncDelay)

		integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{metaNode}, round, nonce)
		time.Sleep(integrationTests.SyncDelay)

		for _, target := range island1 {
			mirrorCurrentBlock(t, target, metaNode)
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

	headerA, hashA, _ := grabCurrentBlock(t, pA)
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

	headerB, hashB, proofB := grabCurrentBlock(t, pB)
	require.Equal(t, uint64(5), headerB.GetNonce())
	require.NotEqual(t, string(hashA), string(hashB))
	require.Equal(t, headerA.GetPrevHash(), headerB.GetPrevHash())

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	nonce++

	// round 8: island 2 extends B with proofed child C -- B's branch is the settled one
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{pB}, round, nonce)
	time.Sleep(integrationTests.SyncDelay)

	headerC, hashC, proofC := grabCurrentBlock(t, pB)
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
	notarizingMeta, notarizingMetaHash, notarizingMetaProof := grabCurrentBlock(t, metaNode)
	requireNotarizes(t, notarizingMeta, shardId, hashB, hashC)

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{metaNode}, round, nextMetaNonce+1)
	time.Sleep(integrationTests.SyncDelay)
	settlingMeta, settlingMetaHash, settlingMetaProof := grabCurrentBlock(t, metaNode)
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

// mirrors process/track.maxMetaBlocksScannedForInclusion, the descending inclusion-scan window
const maxMetaBlocksScannedForInclusion = 16

// long-partition variant: the stranded island receives the evidence only
// after meta has advanced far past the notarizing block. The descending inclusion scan alone
// cannot reach it; the backstop must find it through the window anchored at the stranded island's
// last cross-notarized meta header.
func TestSupernovaSync_ReconcileBackstop_LongPartitionConverges(t *testing.T) {
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

	pA := newShardNode()
	obsA := newShardNode()
	pB := newShardNode()
	obsB := newShardNode()
	metaNode := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            maxShards,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: shardId,
		WithSync:             true,
		EpochsConfig:         &enableEpochs,
		RoundsConfig:         &roundsConfig,
	})

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

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	nonce++

	numPrefixBlocks := 4
	for i := 0; i < numPrefixBlocks; i++ {
		integrationTests.ProposeBlockWithProof(island1, []*integrationTests.TestProcessorNode{pA}, round, nonce)
		time.Sleep(integrationTests.SyncDelay)

		for _, target := range island2 {
			mirrorCurrentBlock(t, target, pA)
		}
		time.Sleep(integrationTests.SyncDelay)

		integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{metaNode}, round, nonce)
		time.Sleep(integrationTests.SyncDelay)

		for _, target := range island1 {
			mirrorCurrentBlock(t, target, metaNode)
		}
		time.Sleep(integrationTests.SyncDelay)

		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		nonce++
	}

	nextMetaNonce := nonce

	// island 1 commits and instantly finalizes the clean sibling A
	integrationTests.ProposeBlockWithProof(island1, []*integrationTests.TestProcessorNode{pA}, round, nonce)
	time.Sleep(integrationTests.SyncDelay)

	headerA, hashA, _ := grabCurrentBlock(t, pA)
	require.Equal(t, uint64(5), headerA.GetNonce())

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	time.Sleep(integrationTests.SyncDelay)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	require.Equal(t, uint64(5), pA.ForkDetector.GetHighestFinalBlockNonce())

	// island 2 commits the contended sibling B and extends it with proofed child C
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{pB}, round, nonce)
	time.Sleep(integrationTests.SyncDelay)

	headerB, hashB, proofB := grabCurrentBlock(t, pB)
	require.Equal(t, uint64(5), headerB.GetNonce())
	require.NotEqual(t, string(hashA), string(hashB))

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	nonce++

	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{pB}, round, nonce)
	time.Sleep(integrationTests.SyncDelay)

	headerC, hashC, proofC := grabCurrentBlock(t, pB)
	require.Equal(t, string(hashB), string(headerC.GetPrevHash()))

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	for i := 0; i < metaArbitrationWindowRounds; i++ {
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		time.Sleep(integrationTests.SyncDelay)
	}

	// meta arbitrates B and settles the notarizing block with a child
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{metaNode}, round, nextMetaNonce)
	time.Sleep(integrationTests.SyncDelay)
	notarizingMeta, notarizingMetaHash, notarizingMetaProof := grabCurrentBlock(t, metaNode)
	requireNotarizes(t, notarizingMeta, shardId, hashB, hashC)

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	integrationTests.ProposeBlockWithProof(island2, []*integrationTests.TestProcessorNode{metaNode}, round, nextMetaNonce+1)
	time.Sleep(integrationTests.SyncDelay)
	settlingMeta, settlingMetaHash, settlingMetaProof := grabCurrentBlock(t, metaNode)
	require.Equal(t, string(notarizingMetaHash), string(settlingMeta.GetPrevHash()))

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// the partition persists long enough for the meta head to move far past the notarizing block.
	// MaxProposalNonceGap caps proposals relative to the proposer's own execution results, which
	// advance with the chain in production but stay frozen in this harness, so the healed pool
	// state is emulated the way interceptors create it: broadcast headers land before their proofs
	type pooledMeta struct {
		header data.HeaderHandler
		hash   []byte
	}
	numExtraMetaHeaders := maxMetaBlocksScannedForInclusion + 4
	extraMetas := make([]pooledMeta, 0, numExtraMetaHeaders)
	prevHash := settlingMetaHash
	for i := 2; i < 2+numExtraMetaHeaders; i++ {
		header := &block.MetaBlockV3{
			Nonce:    nextMetaNonce + uint64(i),
			Round:    round + uint64(i),
			PrevHash: prevHash,
			ChainID:  integrationTests.ChainID,
		}
		hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, header)
		require.Nil(t, err)
		extraMetas = append(extraMetas, pooledMeta{header, hash})
		prevHash = hash
	}

	// heal the partition: island 1 receives the CURRENT meta headers first (pool head jumps far
	// ahead of the fork era), then the fork-era evidence arrives through backfill
	deliverEvidence := func() {
		for _, n := range island1 {
			for _, extra := range extraMetas {
				n.DataPool.Headers().AddHeader(extra.hash, extra.header)
			}
			injectBlock(n, notarizingMeta, notarizingMetaHash, notarizingMetaProof)
			injectBlock(n, settlingMeta, settlingMetaHash, settlingMetaProof)
			injectBlock(n, headerB, hashB, proofB)
			injectBlock(n, headerC, hashC, proofC)
		}
	}
	deliverEvidence()

	island1OnC := func() bool {
		for _, n := range island1 {
			if string(n.BlockChain.GetCurrentBlockHeaderHash()) != string(hashC) {
				return false
			}
		}
		return true
	}
	maxReconcileRounds := 8
	for i := 0; i < maxReconcileRounds && !island1OnC(); i++ {
		deliverEvidence()
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		time.Sleep(integrationTests.SyncDelay)
	}

	require.True(t, island1OnC(), "backstop did not converge across the long partition")

	for _, n := range island1 {
		assert.Equal(t, uint64(4), n.ForkDetector.GetHighestFinalBlockNonce())
	}
}

// the shard cross-notarizes a dead meta block only it has seen; the authority refuses to notarize
// the referencing shard block (the ancestor gate) and builds past the dead block, and the shard
// reverts on that evidence and converges forward (the divergence backstop)
func TestSupernovaSync_DivergenceBackstop_DeadMetaReferenceConverges(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := uint32(1)
	shardId := uint32(0)

	enableEpochs := integrationTests.CreateEnableEpochsConfig()
	enableEpochs.AndromedaEnableEpoch = uint32(0)
	enableEpochs.SupernovaEnableEpoch = uint32(0)
	roundsConfig := integrationTests.GetSupernovaRoundConfigActivatedAt(2)

	newNode := func(nodeShardID uint32) *integrationTests.TestProcessorNode {
		return integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
			MaxShards:            maxShards,
			NodeShardId:          nodeShardID,
			TxSignPrivKeyShardId: shardId,
			WithSync:             true,
			EpochsConfig:         &enableEpochs,
			RoundsConfig:         &roundsConfig,
		})
	}

	pA := newNode(shardId)                     // shard proposer, will reference the dead meta block
	obsA := newNode(shardId)                   // shard observer, reverts through sync alone
	metaNode := newNode(core.MetachainShardId) // the authority: never sees the dead block

	allNodes := []*integrationTests.TestProcessorNode{pA, obsA, metaNode}
	shardNodes := []*integrationTests.TestProcessorNode{pA, obsA}

	integrationTests.ConnectNodes([]integrationTests.Connectable{pA, obsA, metaNode})

	defer func() {
		for _, n := range allNodes {
			n.Close()
		}
	}()

	for _, n := range allNodes {
		_ = n.StartSync()
	}
	time.Sleep(integrationTests.P2pBootstrapDelay)

	round := uint64(0)
	shardNonce := uint64(1)
	metaNonce := uint64(1)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// common prefix: shard blocks notarized by the meta node, one chain everywhere
	numPrefixBlocks := 4
	for i := 0; i < numPrefixBlocks; i++ {
		integrationTests.ProposeBlockWithProof(allNodes, []*integrationTests.TestProcessorNode{pA}, round, shardNonce)
		time.Sleep(integrationTests.SyncDelay)

		integrationTests.ProposeBlockWithProof(allNodes, []*integrationTests.TestProcessorNode{metaNode}, round, metaNonce)
		time.Sleep(integrationTests.SyncDelay)

		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		shardNonce++
		metaNonce++
	}

	prefixMeta, prefixMetaHash, _ := grabCurrentBlock(t, metaNode)
	require.Equal(t, metaNonce-1, prefixMeta.GetNonce())

	// the pointer the revert must land back on: the referencing block will consume the not yet
	// referenced prefix tip together with the dead block, so both pop with it
	_, preForkPointerHash, err := pA.BlockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	require.Nil(t, err)

	// the dead meta block exists only on the stranded shard side, injected the way a partitioned
	// broadcast would have landed it: header plus proof, never seen by the authority
	deadMeta := &block.MetaBlockV3{
		Nonce:        metaNonce,
		Round:        round,
		PrevHash:     prefixMetaHash,
		PrevRandSeed: prefixMeta.GetRandSeed(),
		RandSeed:     []byte("deadMetaRandSeed"),
		ChainID:      integrationTests.ChainID,
	}
	deadMetaHash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, deadMeta)
	require.Nil(t, err)
	deadMetaProof := &block.HeaderProof{
		HeaderHash:    deadMetaHash,
		HeaderNonce:   deadMeta.GetNonce(),
		HeaderRound:   deadMeta.GetRound(),
		HeaderShardId: core.MetachainShardId,
	}
	for _, n := range shardNodes {
		injectBlock(n, deadMeta, deadMetaHash, deadMetaProof)
	}
	time.Sleep(integrationTests.SyncDelay)

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// the shard cross-notarizes the dead block, then commits a meta-less block on top
	integrationTests.ProposeBlockWithProof(allNodes, []*integrationTests.TestProcessorNode{pA}, round, shardNonce)
	time.Sleep(integrationTests.SyncDelay)
	referencingNonce := shardNonce
	referencingHeader, referencingHash, _ := grabCurrentBlock(t, pA)
	require.Contains(t, hashesAsStrings(referencingHeader.(data.ShardHeaderHandler).GetMetaBlockHashes()), string(deadMetaHash))
	_, pointerHash, err := pA.BlockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	require.Nil(t, err)
	require.Equal(t, string(deadMetaHash), string(pointerHash))

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)
	shardNonce++

	integrationTests.ProposeBlockWithProof(allNodes, []*integrationTests.TestProcessorNode{pA}, round, shardNonce)
	time.Sleep(integrationTests.SyncDelay)
	metaLessHeader, metaLessHash, _ := grabCurrentBlock(t, pA)
	require.Empty(t, metaLessHeader.(data.ShardHeaderHandler).GetMetaBlockHashes())

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// the authority builds past the dead block: its own sibling plus two extensions, broadcast
	// the ancestor gate keeps the dead-referencing shard blocks out of every one of them
	canonicalHashes := make([][]byte, 0, 3)
	for i := uint64(0); i < 3; i++ {
		integrationTests.ProposeBlockWithProof(allNodes, []*integrationTests.TestProcessorNode{metaNode}, round, metaNonce+i)
		time.Sleep(integrationTests.SyncDelay)
		header, hash, _ := grabCurrentBlock(t, metaNode)
		require.Equal(t, metaNonce+i, header.GetNonce())
		canonicalHashes = append(canonicalHashes, hash)

		shardInfo := process.GetShardHeadersReferencedByMeta(header.(data.MetaHeaderHandler))
		for _, info := range shardInfo {
			require.NotEqual(t, string(referencingHash), string(info.GetHeaderHash()), "the authority notarized a dead-referencing shard block")
			require.NotEqual(t, string(metaLessHash), string(info.GetHeaderHash()), "the authority notarized a dead descendant")
		}

		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
	}
	require.NotEqual(t, string(deadMetaHash), string(canonicalHashes[0]))

	// the divergence backstop fires on sync-loop iterations: tick rounds until the revert lands
	revertedBelowReferencing := func() bool {
		for _, n := range shardNodes {
			currentHeader := n.BlockChain.GetCurrentBlockHeader()
			if currentHeader == nil || currentHeader.GetNonce() != referencingNonce-1 {
				return false
			}
		}
		return true
	}
	maxBackstopRounds := 10
	for i := 0; i < maxBackstopRounds && !revertedBelowReferencing(); i++ {
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		time.Sleep(integrationTests.SyncDelay)
	}
	require.True(t, revertedBelowReferencing(), "divergence backstop did not revert the dead meta reference")

	// depth bound: the regression stops right below the referencing block and the pointer popped
	// back to the shared prefix
	for _, n := range shardNodes {
		require.Equal(t, referencingNonce-1, n.ForkDetector.GetHighestFinalBlockNonce())
		_, revertedPointerHash, errPointer := n.BlockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
		require.Nil(t, errPointer)
		require.Equal(t, string(preForkPointerHash), string(revertedPointerHash))
	}

	// convergence forward: the next shard proposal references the canonical branch; the first
	// attempts may be refused while the tx selection tracker realigns after the revert, so retry
	converged := func() bool {
		currentHeader := pA.BlockChain.GetCurrentBlockHeader()
		return currentHeader != nil && currentHeader.GetNonce() == referencingNonce
	}
	maxConvergenceRounds := 6
	for i := 0; i < maxConvergenceRounds && !converged(); i++ {
		integrationTests.ProposeBlockWithProof(allNodes, []*integrationTests.TestProcessorNode{pA}, round, referencingNonce)
		time.Sleep(integrationTests.SyncDelay)
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
	}
	require.True(t, converged(), "the shard did not rebuild on the canonical branch after the revert")

	convergedHeader, _, _ := grabCurrentBlock(t, pA)
	convergedRefs := hashesAsStrings(convergedHeader.(data.ShardHeaderHandler).GetMetaBlockHashes())
	require.NotContains(t, convergedRefs, string(deadMetaHash))
	_, convergedPointerHash, err := pA.BlockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	require.Nil(t, err)
	require.Equal(t, string(canonicalHashes[len(canonicalHashes)-1]), string(convergedPointerHash))
}

// meta reconcile roll back: the concrete meta processor must restore a committed v3 head,
// re-adding the shard headers it notarized through the proposal shard info
func TestSupernovaSync_ReconcileBackstop_MetaV3HeadRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := uint32(1)
	shardId := uint32(0)

	enableEpochs := integrationTests.CreateEnableEpochsConfig()
	enableEpochs.AndromedaEnableEpoch = uint32(0)
	enableEpochs.SupernovaEnableEpoch = uint32(0)
	roundsConfig := integrationTests.GetSupernovaRoundConfigActivatedAt(2)

	newNode := func(nodeShardID uint32) *integrationTests.TestProcessorNode {
		return integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
			MaxShards:            maxShards,
			NodeShardId:          nodeShardID,
			TxSignPrivKeyShardId: shardId,
			WithSync:             true,
			EpochsConfig:         &enableEpochs,
			RoundsConfig:         &roundsConfig,
		})
	}

	pShard := newNode(shardId)
	metaNode := newNode(core.MetachainShardId)

	allNodes := []*integrationTests.TestProcessorNode{pShard, metaNode}
	integrationTests.ConnectNodes([]integrationTests.Connectable{pShard, metaNode})

	defer func() {
		for _, n := range allNodes {
			n.Close()
		}
	}()

	for _, n := range allNodes {
		_ = n.StartSync()
	}
	time.Sleep(integrationTests.P2pBootstrapDelay)

	round := uint64(0)
	shardNonce := uint64(1)
	metaNonce := uint64(1)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(allNodes, round)

	// build a real v3 meta chain: the shard proposes, the meta node notarizes it into meta blocks
	numMetaBlocks := 4
	for i := 0; i < numMetaBlocks; i++ {
		integrationTests.ProposeBlockWithProof(allNodes, []*integrationTests.TestProcessorNode{pShard}, round, shardNonce)
		time.Sleep(integrationTests.SyncDelay)

		integrationTests.ProposeBlockWithProof(allNodes, []*integrationTests.TestProcessorNode{metaNode}, round, metaNonce)
		time.Sleep(integrationTests.SyncDelay)

		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(allNodes, round)
		shardNonce++
		metaNonce++
	}

	headHeader, _, _ := grabCurrentBlock(t, metaNode)
	require.True(t, headHeader.IsHeaderV3())
	metaHandler, ok := headHeader.(data.MetaHeaderHandler)
	require.True(t, ok)

	referencedShardHashes := make([][]byte, 0)
	for _, shardInfo := range process.GetShardHeadersReferencedByMeta(metaHandler) {
		referencedShardHashes = append(referencedShardHashes, shardInfo.GetHeaderHash())
	}
	require.NotEmpty(t, referencedShardHashes, "the committed meta head must notarize shard headers")

	// clear the referenced shard headers from the pool so the restore has to re-add them
	for _, shardHash := range referencedShardHashes {
		metaNode.DataPool.Headers().RemoveHeaderByHash(shardHash)
	}

	err := metaNode.BlockProcessor.RestoreBlockIntoPools(headHeader, &block.Body{})
	require.Nil(t, err, "the concrete meta processor must restore a v3 head")

	// the notarized shard headers referenced through the v3 proposal shard info are back in the pool
	for _, shardHash := range referencedShardHashes {
		_, err = metaNode.DataPool.Headers().GetHeaderByHash(shardHash)
		require.Nil(t, err, "the restore must re-add the notarized shard header from the v3 proposal shard info")
	}
}

func hashesAsStrings(hashes [][]byte) []string {
	asStrings := make([]string, 0, len(hashes))
	for _, hash := range hashes {
		asStrings = append(asStrings, string(hash))
	}
	return asStrings
}
