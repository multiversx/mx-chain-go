package settlement

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
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
	return startSupernovaSimulatorWithNumShards(t, 1)
}

func startSupernovaSimulatorWithNumShards(t *testing.T, numOfShards uint32) testsChainSimulator.ChainSimulator {
	simulator, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            "../../../cmd/node/config/",
		NumOfShards:                    numOfShards,
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

// TestChainSimulator_ProofedMetaSiblingDelaysShardSettlementUntilReconciliation verifies the
// metachain-to-shard settlement source gate. Shard block S executes transfer Tref locally and is
// referenced by withheld metablock M_A. A later-round proofed sibling M_B references the same S and
// is delivered first. Because M_A now has a known sibling, its usual one-descendant fast path must
// not publish S as settled authority: child M_C is insufficient, while grandchild M_D establishes
// the required depth-two evidence. The test then replays both sibling artifacts to prove idempotence
// and executes Tpost to prove liveness. It uses direct block creators only, without consensus.
func TestChainSimulator_ProofedMetaSiblingDelaysShardSettlementUntilReconciliation(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Step 1: Start in Supernova, establish a clean chain, and fund two accounts in the same shard so
	// Tref and Tpost can be checked through exact sender nonce and account balance changes.
	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)
	require.NoError(t, simulator.GenerateBlocks(3))

	initialBalance := new(big.Int).Mul(big.NewInt(10), testsChainSimulator.OneEGLD)
	sender, err := simulator.GenerateAndMintWalletAddress(shardID, initialBalance)
	require.NoError(t, err)
	receiver, err := simulator.GenerateAndMintWalletAddress(shardID, big.NewInt(0))
	require.NoError(t, err)
	require.NoError(t, simulator.GenerateBlocks(2))

	// Step 2: Snapshot both accounts and submit Tref without producing a block. The snapshot is the
	// baseline used later to prove that settlement and evidence replay never apply Tref twice.
	beforeReference := txAndAccountsSnapshot{
		sender:   getAccountSnapshot(t, simulator, sender),
		receiver: getAccountSnapshot(t, simulator, receiver),
	}
	referenceValue := new(big.Int).Set(testsChainSimulator.OneEGLD)
	referenceTx := testsChainSimulator.GenerateTransaction(
		sender.Bytes,
		beforeReference.sender.nonce,
		receiver.Bytes,
		referenceValue,
		"",
		50_000,
	)
	referenceTxHash := sendTxWithoutGeneratingBlocks(t, simulator, referenceTx)

	// Step 3: Commit S locally while withholding its network artifact. The meta block created in the
	// same simulator round cannot reference S; nevertheless, local account state executes Tref once.
	shardArtifact, err := simulator.GenerateBlockWithoutBroadcast(shardID)
	require.NoError(t, err)
	require.NotNil(t, shardArtifact)
	shardHeaderHash := calculateHeaderHash(t, shardNode, shardArtifact.Header)
	require.Equal(t, shardHeaderHash, shardNode.GetChainHandler().GetCurrentBlockHeaderHash())
	require.Less(t, getLastCrossNotarizedNonce(t, metaNode, shardID), shardArtifact.Header.GetNonce())
	shardAuthorityPublished := observeSelfNotarizedFromMeta(shardNode, shardHeaderHash)

	afterReference := getTxAndAccounts(t, simulator, referenceTxHash, sender, receiver)
	require.Equal(t, transaction.TxStatusPending, afterReference.txResult.Status)
	requireTransferAppliedExactlyOnce(t, beforeReference, afterReference, referenceValue)

	// Step 4: Publish S, then directly create M_A without producing another shard block. M_A references
	// the exact S hash but is withheld so its later sibling can reach the shard first.
	broadcastAll(t, shardNode, shardArtifact)
	metaArtifactA := generateMetaBlockWithoutShardBlock(t, simulator)
	metaHashA := calculateHeaderHash(t, metaNode, metaArtifactA.Header)
	requireMetaReferencesShardHeader(t, metaArtifactA.Header, shardID, shardHeaderHash)

	// Step 5: Create and broadcast later-round sibling M_B, then publish M_A. Verify their shared
	// parent/nonce and common reference to S. With both proofs present, neither source may publish S:
	// M_B is contended and M_A loses its otherwise immediate one-descendant fast path.
	metaArtifactB := broadcastMetaCompetitorWithoutShardBlock(t, simulator)
	metaHashB := calculateHeaderHash(t, metaNode, metaArtifactB.Header)
	require.Equal(t, metaArtifactA.Header.GetNonce(), metaArtifactB.Header.GetNonce())
	require.Equal(t, metaArtifactA.Header.GetPrevHash(), metaArtifactB.Header.GetPrevHash())
	require.Greater(t, metaArtifactB.Header.GetRound(), metaArtifactA.Header.GetRound())
	require.NotEqual(t, metaHashA, metaHashB)
	requireMetaReferencesShardHeader(t, metaArtifactB.Header, shardID, shardHeaderHash)
	broadcastAll(t, metaNode, metaArtifactA)

	metaProofsPool := shardNode.GetDataComponents().Datapool().Proofs()
	require.True(t, metaProofsPool.HasProof(core.MetachainShardId, metaHashA))
	require.True(t, metaProofsPool.HasProof(core.MetachainShardId, metaHashB))
	require.False(t, shardAuthorityPublished.Load())
	require.Less(t, getLastCrossNotarizedNonce(t, shardNode, core.MetachainShardId), metaArtifactA.Header.GetNonce())
	require.Less(t, getSettledNonce(shardNode), shardArtifact.Header.GetNonce())

	// Step 6: Extend the local M_A branch with one proofed child M_C. Because M_A has sibling M_B,
	// one descendant is insufficient: S must remain unpublished and below the settled checkpoint.
	metaChildArtifact := generateMetaBlockWithoutShardBlock(t, simulator)
	broadcastAll(t, metaNode, metaChildArtifact)
	metaChild, metaChildHash := metaChildArtifact.Header, calculateHeaderHash(t, metaNode, metaChildArtifact.Header)
	require.Equal(t, metaArtifactA.Header.GetNonce()+1, metaChild.GetNonce())
	require.Equal(t, metaHashA, metaChild.GetPrevHash())
	require.True(t, metaProofsPool.HasProof(core.MetachainShardId, metaChildHash))
	lastCrossMeta, lastCrossMetaHash, err := shardNode.GetProcessComponents().BlockTracker().GetLastCrossNotarizedHeader(core.MetachainShardId)
	require.NoError(t, err)
	require.LessOrEqual(t, lastCrossMeta.GetNonce(), metaArtifactA.Header.GetNonce())
	if lastCrossMeta.GetNonce() == metaArtifactA.Header.GetNonce() {
		require.Equal(t, metaHashA, lastCrossMetaHash)
	}
	require.False(t, shardAuthorityPublished.Load())
	require.Less(t, getSettledNonce(shardNode), shardArtifact.Header.GetNonce())

	// Step 7: Add M_D as M_C's proofed child, completing depth-two evidence for M_A. The bounded wait
	// allows delivery of the resulting authority callback and requires shard settlement to cover S.
	metaGrandchildArtifact := generateMetaBlockWithoutShardBlock(t, simulator)
	broadcastAll(t, metaNode, metaGrandchildArtifact)
	metaGrandchild, metaGrandchildHash := metaGrandchildArtifact.Header, calculateHeaderHash(t, metaNode, metaGrandchildArtifact.Header)
	require.Equal(t, metaChild.GetNonce()+1, metaGrandchild.GetNonce())
	require.Equal(t, metaChildHash, metaGrandchild.GetPrevHash())
	require.True(t, metaProofsPool.HasProof(core.MetachainShardId, metaGrandchildHash))
	generateBlocksUntil(t, simulator, 4, func() bool {
		return shardAuthorityPublished.Load() && getSettledNonce(shardNode) >= shardArtifact.Header.GetNonce()
	})

	settledNonce, settledHash := shardNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
	require.GreaterOrEqual(t, settledNonce, shardArtifact.Header.GetNonce())
	if settledNonce == shardArtifact.Header.GetNonce() {
		require.Equal(t, shardHeaderHash, settledHash)
	}
	requireSettledNotAheadOfFinal(t, shardNode)
	afterSettlement := getTxAndAccounts(t, simulator, referenceTxHash, sender, receiver)
	require.Equal(t, transaction.TxStatusSuccess, afterSettlement.txResult.Status)
	requireTransferAppliedExactlyOnce(t, beforeReference, afterSettlement, referenceValue)

	// Step 8: Replay the exact M_A and M_B artifacts. Duplicate evidence must neither move the already
	// established settlement checkpoint nor reapply Tref to either account.
	settledBeforeReplayNonce, settledBeforeReplayHash := shardNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
	broadcastAll(t, metaNode, metaArtifactA)
	broadcastAll(t, metaNode, metaArtifactB)
	settledAfterReplayNonce, settledAfterReplayHash := shardNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
	require.Equal(t, settledBeforeReplayNonce, settledAfterReplayNonce)
	require.Equal(t, settledBeforeReplayHash, settledAfterReplayHash)
	afterReplay := getTxAndAccounts(t, simulator, referenceTxHash, sender, receiver)
	require.Equal(t, transaction.TxStatusSuccess, afterReplay.txResult.Status)
	require.Equal(t, afterSettlement.sender, afterReplay.sender)
	require.Equal(t, afterSettlement.receiver, afterReplay.receiver)

	// Step 9: Submit Tpost with the next sender nonce. Its successful exactly-once execution proves
	// that normal transaction processing remains live after metachain reconciliation.
	postValue := new(big.Int).Div(new(big.Int).Set(testsChainSimulator.OneEGLD), big.NewInt(2))
	postTx := testsChainSimulator.GenerateTransaction(
		sender.Bytes,
		afterReplay.sender.nonce,
		receiver.Bytes,
		postValue,
		"",
		50_000,
	)
	postTxHash := sendTxWithoutGeneratingBlocks(t, simulator, postTx)
	generateBlocksUntil(t, simulator, 5, func() bool {
		return isTransactionSuccessful(simulator, postTxHash, receiver)
	})
	afterPost := getTxAndAccounts(t, simulator, postTxHash, sender, receiver)
	require.Equal(t, transaction.TxStatusSuccess, afterPost.txResult.Status)
	beforePost := txAndAccountsSnapshot{sender: afterReplay.sender, receiver: afterReplay.receiver}
	requireTransferAppliedExactlyOnce(t, beforePost, afterPost, postValue)
	requireSettledNotAheadOfFinal(t, shardNode)
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

// TestChainSimulator_EquivocatingLeaderMetaArbitratesCompetitor verifies safe containment when the
// direct simulator's shard and metachain choose different same-nonce blocks. The shard locally
// commits and withholds clean block A, making A its current and final tip. It then broadcasts only a
// later-round sibling B, so metachain eventually cross-notarizes B. Because direct mode has no sync
// loop, the shard is not expected to switch to B. Instead, the invariant is that settlement remains
// strictly below the divergent nonce, never labels A or B as settled, and retains both headers and
// proofs for later diagnosis or reconciliation. No consensus behavior is exercised.
func TestChainSimulator_EquivocatingLeaderMetaArbitratesCompetitor(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Step 1: Start in Supernova and build a clean prefix whose next shard block can be withheld while
	// metachain continues independently.
	simulator := startSupernovaSimulator(t)
	defer simulator.Close()

	shardNode := simulator.GetNodeHandler(shardID)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)

	require.NoError(t, simulator.GenerateBlocks(3))

	// Step 2: Commit clean block A locally without broadcasting it. A becomes the shard's exact final
	// tip, while metachain remains at A's parent because it has received no A artifact.
	withheld, err := simulator.GenerateBlockWithoutBroadcast(shardID)
	require.NoError(t, err)
	require.NotNil(t, withheld)
	withheldNonce := withheld.Header.GetNonce()
	withheldHash := calculateHeaderHash(t, shardNode, withheld.Header)
	require.Equal(t, withheldNonce, getFinalNonce(shardNode))
	require.Equal(t, withheldHash, shardNode.GetProcessComponents().ForkDetector().GetHighestFinalBlockHash())
	require.Equal(t, withheldNonce-1, getLastCrossNotarizedNonce(t, metaNode, shardID))
	settledBefore := getSettledNonce(shardNode)
	require.LessOrEqual(t, settledBefore, withheldNonce-1)

	// Step 3: Create and broadcast higher-round sibling B without committing it locally. Verify that
	// A and B share the fork nonce but have distinct hashes and rounds.
	competitor, err := simulator.BroadcastCompetingBlock(shardID)
	require.NoError(t, err)
	require.NotNil(t, competitor)
	require.Equal(t, withheldNonce, competitor.Header.GetNonce())
	require.Greater(t, competitor.Header.GetRound(), withheld.Header.GetRound())
	competitorHash := calculateHeaderHash(t, shardNode, competitor.Header)
	require.NotEqual(t, withheldHash, competitorHash)

	// Step 4: Freeze shard production and let metachain wait out the discovery window. Since A stays
	// withheld, B is the only network-visible candidate and must be cross-notarized by its exact hash.
	generateBlocksUntilSkipping(t, simulator, metaArbitrationWindowRounds+3, []uint32{shardID}, func() bool {
		return getLastCrossNotarizedNonce(t, metaNode, shardID) >= withheldNonce
	}, diagnosticNode{name: "shard", node: shardNode}, diagnosticNode{name: "meta", node: metaNode})

	notarizedHeader, notarizedHash, err := metaNode.GetProcessComponents().BlockTracker().GetLastCrossNotarizedHeader(shardID)
	require.NoError(t, err)
	require.Equal(t, withheldNonce, notarizedHeader.GetNonce())
	require.Equal(t, competitorHash, notarizedHash)

	// Step 5: Check containment rather than convergence. Direct mode has no sync loop, so the shard
	// keeps A while metachain selects B. Settlement may advance on the common prefix, but it must stay
	// below the fork nonce and its hash must match neither divergent sibling.
	require.Equal(t, withheldHash, shardNode.GetChainHandler().GetCurrentBlockHeaderHash())
	require.Equal(t, withheldNonce, getFinalNonce(shardNode))
	require.Equal(t, withheldHash, shardNode.GetProcessComponents().ForkDetector().GetHighestFinalBlockHash())
	settledNonce, settledHash := shardNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
	require.GreaterOrEqual(t, settledNonce, settledBefore)
	require.Less(t, settledNonce, withheldNonce)
	require.NotEqual(t, withheldHash, settledHash)
	require.NotEqual(t, competitorHash, settledHash)
	requireSettledNotAheadOfFinal(t, shardNode)

	// Step 6: Confirm that both proofs remain available as equivocation evidence and B's exact header
	// remains in the hot pool, while A remains the shard's current local tip.
	proofsPool := shardNode.GetDataComponents().Datapool().Proofs()
	require.True(t, proofsPool.HasProof(shardID, withheldHash))
	require.True(t, proofsPool.HasProof(shardID, competitorHash))
	pooledCompetitor, err := shardNode.GetDataComponents().Datapool().Headers().GetHeaderByHash(competitorHash)
	require.NoError(t, err)
	require.Equal(t, competitorHash, calculateHeaderHash(t, shardNode, pooledCompetitor))
}
