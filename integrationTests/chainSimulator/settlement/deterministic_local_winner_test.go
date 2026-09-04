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

// TestChainSimulator_DeterministicLocalWinnerWithTransaction verifies the positive equivocation
// path using direct block production only. The shard commits a lower-round contended block A that
// contains transfer Tfork, but withholds A from the network. A higher-round proofed sibling B is
// delivered to metachain first, followed by A. Metachain must deterministically cross-notarize A,
// and the shard must settle that exact hash without applying Tfork more than once. A second transfer
// proves that normal block production remains live after reconciliation. This test deliberately
// does not run consensus or a sync loop: the locally committed branch is also the selected winner.
func TestChainSimulator_DeterministicLocalWinnerWithTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Step 1: Start in the Supernova regime, establish a clean chain, and fund two same-shard
	// accounts. Same-shard transfers make every balance and nonce effect observable on one node.
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
	requireSettledNotAheadOfFinal(t, shardNode)

	// Step 2: Capture the common parent and submit Tfork without producing a block. Two skipped shard
	// rounds make the next shard block contended, forcing the metachain arbitration path.
	parentHeader := shardNode.GetChainHandler().GetCurrentBlockHeader()
	parentHash := shardNode.GetChainHandler().GetCurrentBlockHeaderHash()
	beforeFork := txAndAccountsSnapshot{
		sender:   getAccountSnapshot(t, simulator, sender),
		receiver: getAccountSnapshot(t, simulator, receiver),
	}
	forkValue := new(big.Int).Set(testsChainSimulator.OneEGLD)
	forkTx := testsChainSimulator.GenerateTransaction(
		sender.Bytes,
		beforeFork.sender.nonce,
		receiver.Bytes,
		forkValue,
		"",
		50_000,
	)
	forkTxHash := sendTxWithoutGeneratingBlocks(t, simulator, forkTx)
	require.NoError(t, simulator.GenerateBlocksSkippingShards(2, []uint32{shardID}))

	// Step 3: Commit transaction-bearing A locally without broadcasting it. Tfork is already applied
	// to local accounts, but A remains above both finality and metachain cross-notarization.
	localArtifact, err := simulator.GenerateBlockWithoutBroadcast(shardID)
	require.NoError(t, err)
	require.NotNil(t, localArtifact)
	localHeader := localArtifact.Header
	localHash := calculateHeaderHash(t, shardNode, localHeader)
	require.Equal(t, parentHeader.GetNonce()+1, localHeader.GetNonce())
	require.Equal(t, parentHash, localHeader.GetPrevHash())
	require.Greater(t, localHeader.GetRound(), parentHeader.GetRound()+1)
	require.Equal(t, localHash, shardNode.GetChainHandler().GetCurrentBlockHeaderHash())
	require.Equal(t, parentHeader.GetNonce(), getFinalNonce(shardNode))
	require.Equal(t, parentHash, shardNode.GetProcessComponents().ForkDetector().GetHighestFinalBlockHash())
	require.Less(t, getLastCrossNotarizedNonce(t, metaNode, shardID), localHeader.GetNonce())

	afterLocalExecution := getTxAndAccounts(t, simulator, forkTxHash, sender, receiver)
	require.Equal(t, transaction.TxStatusPending, afterLocalExecution.txResult.Status)
	requireTransferAppliedExactlyOnce(t, beforeFork, afterLocalExecution, forkValue)

	// Step 4: Create and broadcast higher-round sibling B first, then publish the captured A artifact
	// before another metachain block is produced. Both siblings extend the same parent at the same
	// nonce, so their round is the deterministic tie-breaker.
	competitorArtifact, err := simulator.BroadcastCompetingBlock(shardID)
	require.NoError(t, err)
	require.NotNil(t, competitorArtifact)
	competitorHeader := competitorArtifact.Header
	competitorHash := calculateHeaderHash(t, shardNode, competitorHeader)
	require.Equal(t, localHeader.GetNonce(), competitorHeader.GetNonce())
	require.Equal(t, localHeader.GetPrevHash(), competitorHeader.GetPrevHash())
	require.Greater(t, competitorHeader.GetRound(), localHeader.GetRound())
	require.NotEqual(t, localHash, competitorHash)
	require.Less(t, getLastCrossNotarizedNonce(t, metaNode, shardID), localHeader.GetNonce())
	broadcastAll(t, shardNode, localArtifact)

	// Step 5: Prove that both nodes received both equivocation proofs and that evidence delivery did
	// not execute Tfork a second time. Its API status stays pending until its block is finalized.
	shardProofsPool := shardNode.GetDataComponents().Datapool().Proofs()
	metaProofsPool := metaNode.GetDataComponents().Datapool().Proofs()
	require.True(t, shardProofsPool.HasProof(shardID, localHash))
	require.True(t, shardProofsPool.HasProof(shardID, competitorHash))
	require.True(t, metaProofsPool.HasProof(shardID, localHash))
	require.True(t, metaProofsPool.HasProof(shardID, competitorHash))
	afterEvidenceDelivery := getTxAndAccounts(t, simulator, forkTxHash, sender, receiver)
	require.Equal(t, transaction.TxStatusPending, afterEvidenceDelivery.txResult.Status)
	require.Equal(t, afterLocalExecution.sender, afterEvidenceDelivery.sender)
	require.Equal(t, afterLocalExecution.receiver, afterEvidenceDelivery.receiver)

	// Step 6: Freeze shard production at A while metachain waits out the discovery window. The first
	// condition requires the exact lower-round A hash, rather than accepting either same-nonce block;
	// the second waits until settlement independently reaches that same hash.
	generateBlocksUntilSkipping(t, simulator, metaArbitrationWindowRounds+4, []uint32{shardID}, func() bool {
		header, hash, getErr := metaNode.GetProcessComponents().BlockTracker().GetLastCrossNotarizedHeader(shardID)
		return getErr == nil && header.GetNonce() == localHeader.GetNonce() && bytes.Equal(hash, localHash)
	})
	generateBlocksUntilSkipping(t, simulator, 4, []uint32{shardID}, func() bool {
		settledNonce, settledHash := shardNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
		return settledNonce == localHeader.GetNonce() && bytes.Equal(settledHash, localHash)
	})

	// Step 7: At the controlled settlement checkpoint, current, final, and settled must all identify
	// A. Both proofs must remain retained as evidence, and Tfork's account effects must be unchanged.
	settledNonce, settledHash := shardNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
	require.Equal(t, localHeader.GetNonce(), settledNonce)
	require.Equal(t, localHash, settledHash)
	require.Equal(t, localHash, shardNode.GetChainHandler().GetCurrentBlockHeaderHash())
	require.Equal(t, localHeader.GetNonce(), getFinalNonce(shardNode))
	require.Equal(t, localHash, shardNode.GetProcessComponents().ForkDetector().GetHighestFinalBlockHash())
	requireSettledNotAheadOfFinal(t, shardNode)
	require.True(t, shardProofsPool.HasProof(shardID, localHash))
	require.True(t, shardProofsPool.HasProof(shardID, competitorHash))
	require.True(t, metaProofsPool.HasProof(shardID, localHash))
	require.True(t, metaProofsPool.HasProof(shardID, competitorHash))

	afterSettlement := getTxAndAccounts(t, simulator, forkTxHash, sender, receiver)
	requireTransferAppliedExactlyOnce(t, beforeFork, afterSettlement, forkValue)
	require.Equal(t, afterLocalExecution.sender, afterSettlement.sender)
	require.Equal(t, afterLocalExecution.receiver, afterSettlement.receiver)

	// Step 8: Resume shard production. The direct simulator keeps Tfork in the transaction pool while
	// the shard is frozen, even after A settles; producing again removes that transient API view.
	// Tfork must become successful without changing its already-applied account state.
	generateBlocksUntil(t, simulator, 3, func() bool {
		return isTransactionSuccessful(simulator, forkTxHash, receiver)
	})
	afterForkFinalization := getTxAndAccounts(t, simulator, forkTxHash, sender, receiver)
	require.Equal(t, transaction.TxStatusSuccess, afterForkFinalization.txResult.Status)
	requireTransferAppliedExactlyOnce(t, beforeFork, afterForkFinalization, forkValue)
	require.Equal(t, afterSettlement.sender, afterForkFinalization.sender)
	require.Equal(t, afterSettlement.receiver, afterForkFinalization.receiver)

	// Step 9: Submit Tpost with the next sender nonce. Its successful, exactly-once execution and the
	// advancement beyond A prove that the reconciled branch remains live.
	beforePost := txAndAccountsSnapshot{
		sender:   afterForkFinalization.sender,
		receiver: afterForkFinalization.receiver,
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
	require.GreaterOrEqual(t, getFinalNonce(shardNode), localHeader.GetNonce())
	require.GreaterOrEqual(t, getSettledNonce(shardNode), localHeader.GetNonce())
	requireSettledNotAheadOfFinal(t, shardNode)
}
