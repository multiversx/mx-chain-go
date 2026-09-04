package settlement

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
)

// TestChainSimulator_KnownSiblingBlocksUnsafeDescendantFinality verifies the negative and positive
// sides of descendant finality. Starting from final and settled block Q, the shard commits contended
// parent P, then locally commits clean child A while a proofed same-nonce sibling B is also known.
// When metachain settles only P, finality must stop exactly at P instead of cascading over A. After
// metachain deterministically selects A over B, final and settled may advance to A. A later transfer
// confirms that the finality gate did not leave the direct-production chain stalled. No consensus or
// sibling rollback is involved because the local lower-round child A is the eventual winner.
func TestChainSimulator_KnownSiblingBlocksUnsafeDescendantFinality(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Step 1: Start in Supernova, fund same-shard accounts for the final liveness check, and build a
	// clean prefix before taking control of shard and metachain production independently.
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

	// Step 2: Freeze the shard at Q and let metachain catch up until Q's exact hash is both final and
	// settled. This gives every later assertion an unambiguous common-prefix checkpoint.
	commonHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	commonHash := calculateHeaderHash(t, shardNode, commonHeader)
	generateBlocksUntilSkipping(t, simulator, 5, []uint32{shardID}, func() bool {
		settledNonce, settledHash := shardNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
		return settledNonce == commonHeader.GetNonce() && bytes.Equal(settledHash, commonHash)
	})
	require.Equal(t, commonHeader.GetNonce(), getFinalNonce(shardNode))
	require.Equal(t, commonHash, shardNode.GetProcessComponents().ForkDetector().GetHighestFinalBlockHash())

	// Step 3: Skip two more shard rounds, then directly commit and publish P without creating a meta
	// block in the same step. The round gap makes P contended, so neither final nor settled may leave Q.
	require.NoError(t, simulator.GenerateBlocksSkippingShards(2, []uint32{shardID}))
	parentArtifact := generateShardBlockWithoutMetaBlock(t, simulator)
	parentHeader := parentArtifact.Header
	parentHash := calculateHeaderHash(t, shardNode, parentHeader)
	require.Equal(t, commonHeader.GetNonce()+1, parentHeader.GetNonce())
	require.Greater(t, parentHeader.GetRound(), commonHeader.GetRound()+1)
	require.Equal(t, commonHeader.GetNonce(), getFinalNonce(shardNode))
	require.Equal(t, commonHeader.GetNonce(), getSettledNonce(shardNode))
	broadcastAll(t, shardNode, parentArtifact)

	// Step 4: Directly commit clean child A on P. Create later-round proofed sibling B at A's nonce,
	// deliver B first, and then publish A while metachain production remains frozen.
	localArtifact := generateShardBlockWithoutMetaBlock(t, simulator)
	localHeader := localArtifact.Header
	localHash := calculateHeaderHash(t, shardNode, localHeader)
	require.Equal(t, parentHeader.GetNonce()+1, localHeader.GetNonce())
	require.Equal(t, parentHash, localHeader.GetPrevHash())
	require.Equal(t, parentHeader.GetRound()+1, localHeader.GetRound())

	competitorArtifact := broadcastShardCompetitorWithoutMetaBlock(t, simulator)
	competitorHeader := competitorArtifact.Header
	competitorHash := calculateHeaderHash(t, shardNode, competitorHeader)
	require.Equal(t, localHeader.GetNonce(), competitorHeader.GetNonce())
	require.Equal(t, localHeader.GetPrevHash(), competitorHeader.GetPrevHash())
	require.Greater(t, competitorHeader.GetRound(), localHeader.GetRound())
	require.NotEqual(t, localHash, competitorHash)
	broadcastAll(t, shardNode, localArtifact)

	// Step 5: Confirm that A and B are known to both nodes before arbitration. The shard remains on A,
	// while finality remains at Q because the contended parent P has not settled yet.
	shardProofsPool := shardNode.GetDataComponents().Datapool().Proofs()
	metaProofsPool := metaNode.GetDataComponents().Datapool().Proofs()
	require.True(t, shardProofsPool.HasProof(shardID, localHash))
	require.True(t, shardProofsPool.HasProof(shardID, competitorHash))
	require.True(t, metaProofsPool.HasProof(shardID, localHash))
	require.True(t, metaProofsPool.HasProof(shardID, competitorHash))
	require.Equal(t, localHash, shardNode.GetChainHandler().GetCurrentBlockHeaderHash())
	require.Equal(t, commonHeader.GetNonce(), getFinalNonce(shardNode))

	// Step 6: Let metachain arbitrate and settle P while keeping the shard frozen. At this exact
	// checkpoint, the known A/B sibling pair must stop the normal clean-child finality cascade:
	// current remains A, but both final and settled must identify P and no later block.
	generateBlocksUntilSkipping(t, simulator, metaArbitrationWindowRounds+3, []uint32{shardID}, func() bool {
		header, hash, getErr := metaNode.GetProcessComponents().BlockTracker().GetLastCrossNotarizedHeader(shardID)
		return getErr == nil && header.GetNonce() >= parentHeader.GetNonce() && bytes.Equal(hash, parentHash)
	})
	generateBlocksUntilSkipping(t, simulator, 3, []uint32{shardID}, func() bool {
		return getSettledNonce(shardNode) >= parentHeader.GetNonce()
	})
	settledParentNonce, settledParentHash := shardNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
	require.Equal(t, parentHeader.GetNonce(), settledParentNonce)
	require.Equal(t, parentHash, settledParentHash)
	require.Equal(t, parentHeader.GetNonce(), getFinalNonce(shardNode))
	require.Equal(t, parentHash, shardNode.GetProcessComponents().ForkDetector().GetHighestFinalBlockHash())
	require.Equal(t, localHash, shardNode.GetChainHandler().GetCurrentBlockHeaderHash())

	// Step 7: Continue metachain production until it selects lower-round A and makes that source
	// settlement-ready. Only now may both final and settled advance from P to the exact A hash.
	generateBlocksUntilSkipping(t, simulator, metaArbitrationWindowRounds+4, []uint32{shardID}, func() bool {
		header, hash, getErr := metaNode.GetProcessComponents().BlockTracker().GetLastCrossNotarizedHeader(shardID)
		return getErr == nil && header.GetNonce() >= localHeader.GetNonce() && bytes.Equal(hash, localHash)
	})
	generateBlocksUntilSkipping(t, simulator, 4, []uint32{shardID}, func() bool {
		return getSettledNonce(shardNode) >= localHeader.GetNonce() && getFinalNonce(shardNode) >= localHeader.GetNonce()
	})
	settledLocalNonce, settledLocalHash := shardNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
	require.Equal(t, localHeader.GetNonce(), settledLocalNonce)
	require.Equal(t, localHash, settledLocalHash)
	require.Equal(t, localHeader.GetNonce(), getFinalNonce(shardNode))
	require.Equal(t, localHash, shardNode.GetProcessComponents().ForkDetector().GetHighestFinalBlockHash())
	requireSettledNotAheadOfFinal(t, shardNode)

	// Step 8: Submit a transaction after reconciliation. Its exactly-once execution in a descendant
	// of A proves that normal execution and finality resume after the sibling gate is released.
	beforePost := txAndAccountsSnapshot{
		sender:   getAccountSnapshot(t, simulator, sender),
		receiver: getAccountSnapshot(t, simulator, receiver),
	}
	postValue := new(big.Int).Div(new(big.Int).Set(testsChainSimulator.OneEGLD), big.NewInt(2))
	postTx := testsChainSimulator.GenerateTransaction(
		sender.Bytes,
		beforePost.sender.nonce,
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
	requireTransferAppliedExactlyOnce(t, beforePost, afterPost, postValue)
	require.Greater(t, shardNode.GetChainHandler().GetCurrentBlockHeader().GetNonce(), localHeader.GetNonce())
	requireSettledNotAheadOfFinal(t, shardNode)
}
