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
// being referenced by meta; the next shard block settles it, meta references it, finality catches
// up through the notarization and instant finality resumes on the clean path
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

	// this round's shard child settles the contended block, but meta proposes before seeing it
	require.NoError(t, simulator.GenerateBlocks(1))
	require.Equal(t, contendedHeader.GetNonce()-1, getLastCrossNotarizedNonce(t, metaNode, shardID))

	// next round meta references the settled block
	require.NoError(t, simulator.GenerateBlocks(1))
	require.GreaterOrEqual(t, getLastCrossNotarizedNonce(t, metaNode, shardID), contendedHeader.GetNonce())

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
// its next meta block settles it (strict R-CROSS, no arbitration for the meta direction)
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
