package settlement

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

const (
	shardID                     = uint32(0)
	supernovaEnableEpoch        = uint32(3)
	metaArbitrationWindowRounds = 3
)

func startSupernovaSimulator(t *testing.T) testsChainSimulator.ChainSimulator {
	simulator, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            "../../../cmd/node/config/",
		NumOfShards:                    1,
		RoundDurationInMillis:          uint64(4000),
		SupernovaRoundDurationInMillis: uint64(400),
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    10,
		},
		SupernovaRoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    20,
		},
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         1,
		MetaChainMinNodes:        1,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = supernovaEnableEpoch
		},
	})
	require.NoError(t, err)
	require.NotNil(t, simulator)

	err = simulator.GenerateBlocksUntilEpochIsReached(int32(supernovaEnableEpoch) + 1)
	require.NoError(t, err)

	return simulator
}

func getFinalNonce(node process.NodeHandler) uint64 {
	return node.GetProcessComponents().ForkDetector().GetHighestFinalBlockNonce()
}

func getSettledNonce(node process.NodeHandler) uint64 {
	nonce, _ := node.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
	return nonce
}

func requireSettledNotAheadOfFinal(t *testing.T, node process.NodeHandler) {
	require.LessOrEqual(t, getSettledNonce(node), getFinalNonce(node))
}

func getLastCrossNotarizedNonce(t *testing.T, node process.NodeHandler, ofShard uint32) uint64 {
	header, _, err := node.GetProcessComponents().BlockTracker().GetLastCrossNotarizedHeader(ofShard)
	require.NoError(t, err)

	return header.GetNonce()
}

func generateBlocksUntil(t *testing.T, simulator testsChainSimulator.ChainSimulator, maxBlocks int, condition func() bool) {
	for i := 0; i < maxBlocks && !condition(); i++ {
		require.NoError(t, simulator.GenerateBlocks(1))
	}
	require.True(t, condition())
}

func generateBlocksUntilSkipping(t *testing.T, simulator testsChainSimulator.ChainSimulator, maxBlocks int, skippedShardIDs []uint32, condition func() bool) {
	for i := 0; i < maxBlocks && !condition(); i++ {
		require.NoError(t, simulator.GenerateBlocksSkippingShards(1, skippedShardIDs))
	}
	require.True(t, condition())
}

// a contended shard block (skipped rounds before it) commits without instant finality and without
// being referenced by meta; its own proofed child does not settle it, so meta holds it for the
// discovery window and notarizes it through arbitration, after which finality catches up
func TestChainSimulator_ContendedShardBlockDefersFinalityAndSettles(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)

	// clean path: finality is instant at commit
	require.NoError(t, simulator.GenerateBlocks(3))
	parentHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, parentHeader.GetNonce(), getFinalNonce(shardNode))

	// two skipped shard rounds make the next shard block contended
	require.NoError(t, simulator.GenerateBlocksSkippingShards(2, []uint32{shardID}))
	require.NoError(t, simulator.GenerateBlocks(1))

	contendedHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, parentHeader.GetNonce()+1, contendedHeader.GetNonce())
	require.Greater(t, contendedHeader.GetRound(), parentHeader.GetRound()+1)

	// finality is deferred at commit and meta has not referenced the contended block
	require.Equal(t, contendedHeader.GetNonce()-1, getFinalNonce(shardNode))
	require.Equal(t, contendedHeader.GetNonce()-1, getLastCrossNotarizedNonce(t, metaNode, shardID))

	// the shard extends the contended block with a proofed child; that child does not settle it,
	// so meta keeps holding it for the rest of the discovery window
	require.NoError(t, simulator.GenerateBlocks(metaArbitrationWindowRounds-1))
	require.Equal(t, contendedHeader.GetNonce()-1, getLastCrossNotarizedNonce(t, metaNode, shardID))

	// once the window elapses meta notarizes it through arbitration
	generateBlocksUntil(t, simulator, 5, func() bool {
		return getLastCrossNotarizedNonce(t, metaNode, shardID) >= contendedHeader.GetNonce()
	})

	// the notarization reaches the shard and finality catches up over the clean descendants
	generateBlocksUntil(t, simulator, 5, func() bool {
		return getFinalNonce(shardNode) >= contendedHeader.GetNonce()
	})

	// clean path restored: instant finality resumes
	require.NoError(t, simulator.GenerateBlocks(1))
	currentHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, currentHeader.GetNonce(), getFinalNonce(shardNode))
}

// a contended shard block whose shard then stalls is held by meta for the discovery window and
// afterwards notarized through arbitration, without any settling child
func TestChainSimulator_MetaArbitratesStalledContendedShardNonce(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)

	require.NoError(t, simulator.GenerateBlocks(3))

	// two skipped shard rounds make the next shard block contended, then the shard stalls
	require.NoError(t, simulator.GenerateBlocksSkippingShards(2, []uint32{shardID}))
	require.NoError(t, simulator.GenerateBlocks(1))
	contendedHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, contendedHeader.GetNonce()-1, getFinalNonce(shardNode))
	require.Equal(t, contendedHeader.GetNonce()-1, getLastCrossNotarizedNonce(t, metaNode, shardID))

	// within the discovery window meta holds the contended block, no settling child exists
	require.NoError(t, simulator.GenerateBlocksSkippingShards(metaArbitrationWindowRounds-1, []uint32{shardID}))
	require.Equal(t, contendedHeader.GetNonce()-1, getLastCrossNotarizedNonce(t, metaNode, shardID))

	// once the window elapses, meta notarizes the contended block through arbitration alone; the
	// shard stays stalled throughout, so this cannot be a settle-on-child
	generateBlocksUntilSkipping(t, simulator, 5, []uint32{shardID}, func() bool {
		return getLastCrossNotarizedNonce(t, metaNode, shardID) >= contendedHeader.GetNonce()
	})

	// the shard resumes; the arbitration notarization settles the block and the chain stays live
	generateBlocksUntil(t, simulator, 10, func() bool {
		currentNonce := shardNode.GetChainHandler().GetCurrentBlockHeader().GetNonce()
		return currentNonce > contendedHeader.GetNonce() && getFinalNonce(shardNode) >= contendedHeader.GetNonce()
	})
}

// a contended meta block commits without instant finality and is not referenced by the shard until
// its next meta block settles it (strict referencing gate, no arbitration for the meta direction)
func TestChainSimulator_ContendedMetaBlockNotReferencedUntilSettled(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)

	require.NoError(t, simulator.GenerateBlocks(3))
	parentHeader := metaNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, parentHeader.GetNonce(), getFinalNonce(metaNode))

	// two skipped meta rounds make the next meta block contended
	require.NoError(t, simulator.GenerateBlocksSkippingShards(2, []uint32{core.MetachainShardId}))
	require.NoError(t, simulator.GenerateBlocks(1))

	contendedHeader := metaNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, parentHeader.GetNonce()+1, contendedHeader.GetNonce())
	require.Greater(t, contendedHeader.GetRound(), parentHeader.GetRound()+1)

	// meta finality is deferred at commit; the shard has not referenced the contended meta block
	require.Equal(t, contendedHeader.GetNonce()-1, getFinalNonce(metaNode))
	require.Equal(t, contendedHeader.GetNonce()-1, getLastCrossNotarizedNonce(t, shardNode, core.MetachainShardId))

	// the next meta block settles its parent: meta finality is restored at commit (settle-on-child)
	require.NoError(t, simulator.GenerateBlocks(1))
	require.Equal(t, contendedHeader.GetNonce()+1, getFinalNonce(metaNode))

	// the shard references the settled meta block once its proofed child is visible
	generateBlocksUntil(t, simulator, 5, func() bool {
		return getLastCrossNotarizedNonce(t, shardNode, core.MetachainShardId) >= contendedHeader.GetNonce()
	})

	// clean path restored on meta as well
	require.NoError(t, simulator.GenerateBlocks(1))
	currentHeader := metaNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, currentHeader.GetNonce(), getFinalNonce(metaNode))
}

// on the clean path shard finality is instant at commit while the settled watermark follows meta
// notarization: freezing meta grows the settled lag without touching finality, resuming meta
// catches the watermark up; settled never exceeds final
func TestChainSimulator_SettledWatermarkFollowsMetaNotarization(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)

	require.NoError(t, simulator.GenerateBlocks(3))
	requireSettledNotAheadOfFinal(t, shardNode)

	// freeze meta; the first two frozen rounds let in-flight notarizations land
	require.NoError(t, simulator.GenerateBlocksSkippingShards(2, []uint32{core.MetachainShardId}))
	settledBefore := getSettledNonce(shardNode)

	// with meta frozen the shard stays instantly final but nothing new settles
	require.NoError(t, simulator.GenerateBlocksSkippingShards(4, []uint32{core.MetachainShardId}))
	frozenTip := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, frozenTip.GetNonce(), getFinalNonce(shardNode))
	require.Equal(t, settledBefore, getSettledNonce(shardNode))

	// meta resumes and the settled watermark catches up over the blocks committed during the freeze
	generateBlocksUntil(t, simulator, 8, func() bool {
		requireSettledNotAheadOfFinal(t, shardNode)
		return getSettledNonce(shardNode) >= frozenTip.GetNonce()
	})
}

// contended blocks land on shard and meta in the same round (both chains skipped the same rounds);
// each defers finality at commit and both settle, converging back to the clean path
func TestChainSimulator_SimultaneousShardAndMetaContentionConverges(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)

	require.NoError(t, simulator.GenerateBlocks(3))
	shardParent := shardNode.GetChainHandler().GetCurrentBlockHeader()
	metaParent := metaNode.GetChainHandler().GetCurrentBlockHeader()

	// two fully skipped rounds make the next block on each chain contended
	require.NoError(t, simulator.GenerateBlocksSkippingShards(2, []uint32{shardID, core.MetachainShardId}))
	require.NoError(t, simulator.GenerateBlocks(1))

	shardContended := shardNode.GetChainHandler().GetCurrentBlockHeader()
	metaContended := metaNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, shardParent.GetNonce()+1, shardContended.GetNonce())
	require.Equal(t, metaParent.GetNonce()+1, metaContended.GetNonce())
	require.Greater(t, shardContended.GetRound(), shardParent.GetRound()+1)
	require.Greater(t, metaContended.GetRound(), metaParent.GetRound()+1)

	// finality is deferred on both chains at commit
	require.Equal(t, shardContended.GetNonce()-1, getFinalNonce(shardNode))
	require.Equal(t, metaContended.GetNonce()-1, getFinalNonce(metaNode))

	// both chains settle their contended blocks and finality converges
	generateBlocksUntil(t, simulator, 10, func() bool {
		return getFinalNonce(shardNode) >= shardContended.GetNonce() &&
			getFinalNonce(metaNode) >= metaContended.GetNonce()
	})

	// clean path restored on both chains
	require.NoError(t, simulator.GenerateBlocks(1))
	require.Equal(t, shardNode.GetChainHandler().GetCurrentBlockHeader().GetNonce(), getFinalNonce(shardNode))
	require.Equal(t, metaNode.GetChainHandler().GetCurrentBlockHeader().GetNonce(), getFinalNonce(metaNode))
}

// the shard stalls across a meta epoch change, so its epoch-start block is contended; the epoch
// transition still completes and the contended epoch-start block settles
func TestChainSimulator_EpochBoundaryContendedShardEpochStartSettles(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)

	require.NoError(t, simulator.GenerateBlocks(3))
	shardParent := shardNode.GetChainHandler().GetCurrentBlockHeader()
	metaEpoch := metaNode.GetChainHandler().GetCurrentBlockHeader().GetEpoch()

	// stall the shard until meta crosses into the next epoch
	generateBlocksUntilSkipping(t, simulator, 30, []uint32{shardID}, func() bool {
		return metaNode.GetChainHandler().GetCurrentBlockHeader().GetEpoch() > metaEpoch
	})

	// the shard resumes: its next block bridges the epoch change and is contended
	require.NoError(t, simulator.GenerateBlocks(1))
	contendedHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, shardParent.GetNonce()+1, contendedHeader.GetNonce())
	require.Greater(t, contendedHeader.GetRound(), shardParent.GetRound()+1)
	require.Equal(t, contendedHeader.GetNonce()-1, getFinalNonce(shardNode))

	// the shard enters the new epoch, the contended block settles and the chain stays live
	generateBlocksUntil(t, simulator, 15, func() bool {
		currentHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
		return currentHeader.GetEpoch() > shardParent.GetEpoch() &&
			getFinalNonce(shardNode) >= contendedHeader.GetNonce()
	})

	// clean path restored in the new epoch
	require.NoError(t, simulator.GenerateBlocks(1))
	currentHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, currentHeader.GetNonce(), getFinalNonce(shardNode))
}

// with meta frozen, clean descendants of a contended shard block commit but stay non-final until
// the ancestor settles (transitivity); when meta resumes, finality catches up over the descendants
func TestChainSimulator_DescendantsNotFinalUntilContendedAncestorSettles(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)

	require.NoError(t, simulator.GenerateBlocks(3))
	parentHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, parentHeader.GetNonce(), getFinalNonce(shardNode))

	// freeze meta and skip two shard rounds; the next shard block is contended
	require.NoError(t, simulator.GenerateBlocksSkippingShards(2, []uint32{shardID, core.MetachainShardId}))
	require.NoError(t, simulator.GenerateBlocksSkippingShards(1, []uint32{core.MetachainShardId}))
	contendedHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, parentHeader.GetNonce()+1, contendedHeader.GetNonce())
	require.Greater(t, contendedHeader.GetRound(), parentHeader.GetRound()+1)
	require.Equal(t, contendedHeader.GetNonce()-1, getFinalNonce(shardNode))

	// with meta still frozen the shard builds clean descendants; none becomes final while the
	// contended ancestor is unsettled
	require.NoError(t, simulator.GenerateBlocksSkippingShards(3, []uint32{core.MetachainShardId}))
	tipHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, contendedHeader.GetNonce()+3, tipHeader.GetNonce())
	require.Equal(t, contendedHeader.GetNonce()-1, getFinalNonce(shardNode))

	// meta resumes, notarizes the contended block, and finality catches up over the descendants
	generateBlocksUntil(t, simulator, 8, func() bool {
		return getFinalNonce(shardNode) >= tipHeader.GetNonce()
	})

	// clean path restored
	require.NoError(t, simulator.GenerateBlocks(1))
	currentHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	require.Equal(t, currentHeader.GetNonce(), getFinalNonce(shardNode))
}

// an equivocating shard leader commits a withheld block and broadcasts a competitor at the same
// nonce; meta arbitrates the competitor while the shard holds its own block instantly final; the
// settled watermark never covers the equivocated nonce, so exports stay behind the divergence
func TestChainSimulator_EquivocatingLeaderMetaArbitratesCompetitor(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)

	require.NoError(t, simulator.GenerateBlocks(3))

	// the shard commits the withheld block and holds it instantly final; meta never sees it
	withheld, err := simulator.GenerateBlockWithoutBroadcast(shardID)
	require.NoError(t, err)
	require.NotNil(t, withheld)
	withheldNonce := withheld.Header.GetNonce()
	require.Equal(t, withheldNonce, getFinalNonce(shardNode))
	require.Equal(t, withheldNonce-1, getLastCrossNotarizedNonce(t, metaNode, shardID))
	settledBefore := getSettledNonce(shardNode)
	require.LessOrEqual(t, settledBefore, withheldNonce-1)

	// the competitor lands on meta at the same nonce, one round later, without a local commit
	competitor, err := simulator.BroadcastCompetingBlock(shardID)
	require.NoError(t, err)
	require.NotNil(t, competitor)
	require.Equal(t, withheldNonce, competitor.Header.GetNonce())
	require.Greater(t, competitor.Header.GetRound(), withheld.Header.GetRound())

	// the competitor is contended and childless, so meta notarizes it through arbitration
	generateBlocksUntilSkipping(t, simulator, metaArbitrationWindowRounds+3, []uint32{shardID}, func() bool {
		return getLastCrossNotarizedNonce(t, metaNode, shardID) >= withheldNonce
	})

	_, notarizedHash, err := metaNode.GetProcessComponents().BlockTracker().GetLastCrossNotarizedHeader(shardID)
	require.NoError(t, err)
	coreComponents := shardNode.GetCoreComponents()
	competitorHash, err := core.CalculateHash(coreComponents.InternalMarshalizer(), coreComponents.Hasher(), competitor.Header)
	require.NoError(t, err)
	require.Equal(t, competitorHash, notarizedHash)

	// the shard keeps its own final block, but settlement never covered the equivocated nonce
	require.Equal(t, withheldNonce, getFinalNonce(shardNode))
	require.Equal(t, settledBefore, getSettledNonce(shardNode))
}
