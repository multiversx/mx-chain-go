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

// TestChainSimulator_SelectedSiblingExecutesCrossShardTransferExactlyOnce verifies cross-shard
// miniblock processing around a source-shard equivocation. Shard 0 locally commits and withholds
// lower-round contended block A containing transfer Tcross to shard 1. Higher-round sibling B is
// delivered first, followed by A's complete header, miniblock, transaction, and proof artifact.
// Metachain must select A, shard 1 must credit the receiver exactly once, and replaying both sibling
// artifacts must not debit, charge, or credit either account again. Continued production on both
// shards proves liveness. B is a re-signed clone of A, so this tests duplicate delivery of one
// payload around equivocation; it intentionally does not model rollback of branch-specific state or
// run consensus.
func TestChainSimulator_SelectedSiblingExecutesCrossShardTransferExactlyOnce(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	const destinationShardID = uint32(1)

	// Step 1: Start a two-shard Supernova network, establish a clean prefix, and fund a sender in
	// shard 0 and receiver in shard 1. Their initial snapshots are the exactly-once baseline.
	simulator := startSupernovaSimulatorWithNumShards(t, 2)
	defer simulator.Close()

	sourceNode := simulator.GetNodeHandler(shardID)
	destinationNode := simulator.GetNodeHandler(destinationShardID)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)
	require.NoError(t, simulator.GenerateBlocks(3))

	initialBalance := new(big.Int).Mul(big.NewInt(10), testsChainSimulator.OneEGLD)
	sender, err := simulator.GenerateAndMintWalletAddress(shardID, initialBalance)
	require.NoError(t, err)
	receiver, err := simulator.GenerateAndMintWalletAddress(destinationShardID, big.NewInt(0))
	require.NoError(t, err)
	require.NoError(t, simulator.GenerateBlocks(2))

	beforeTransfer := txAndAccountsSnapshot{
		sender:   getAccountSnapshot(t, simulator, sender),
		receiver: getAccountSnapshot(t, simulator, receiver),
	}
	requireSettledNotAheadOfFinal(t, sourceNode)
	requireSettledNotAheadOfFinal(t, destinationNode)

	// Step 2: Submit Tcross without producing a block, then skip two source-shard rounds. The receiver
	// stays untouched, and the round gap forces A through metachain's contention arbitration path.
	transferValue := new(big.Int).Set(testsChainSimulator.OneEGLD)
	crossTx := testsChainSimulator.GenerateTransaction(
		sender.Bytes,
		beforeTransfer.sender.nonce,
		receiver.Bytes,
		transferValue,
		"",
		50_000,
	)
	crossTxHash := sendTxWithoutGeneratingBlocks(t, simulator, crossTx)
	parentHeader := sourceNode.GetChainHandler().GetCurrentBlockHeader()
	parentHash := sourceNode.GetChainHandler().GetCurrentBlockHeaderHash()
	require.NoError(t, simulator.GenerateBlocksSkippingShards(2, []uint32{shardID}))

	// Step 3: Commit A locally without broadcasting it. Source execution consumes the sender nonce
	// and funds the outgoing miniblock, while shard 1 cannot execute it and metachain cannot reference
	// A because neither has received the withheld artifact.
	localArtifact, err := simulator.GenerateBlockWithoutBroadcast(shardID)
	require.NoError(t, err)
	require.NotNil(t, localArtifact)
	localHeader := localArtifact.Header
	localHash := calculateHeaderHash(t, sourceNode, localHeader)
	require.Equal(t, parentHeader.GetNonce()+1, localHeader.GetNonce())
	require.Equal(t, parentHash, localHeader.GetPrevHash())
	require.Greater(t, localHeader.GetRound(), parentHeader.GetRound()+1)
	require.Equal(t, localHash, sourceNode.GetChainHandler().GetCurrentBlockHeaderHash())
	require.Equal(t, parentHeader.GetNonce(), getFinalNonce(sourceNode))
	require.Less(t, getLastCrossNotarizedNonce(t, metaNode, shardID), localHeader.GetNonce())

	afterSourceExecutionSender := getAccountSnapshot(t, simulator, sender)
	afterSourceExecutionReceiver := getAccountSnapshot(t, simulator, receiver)
	require.Equal(t, beforeTransfer.sender.nonce+1, afterSourceExecutionSender.nonce)
	require.True(t, afterSourceExecutionSender.balance.Cmp(beforeTransfer.sender.balance) < 0)
	require.Equal(t, beforeTransfer.receiver, afterSourceExecutionReceiver)

	// Step 4: Broadcast higher-round sibling B first and then publish A's complete artifact before the
	// next metachain block. Equal nonce and parent prove a real sibling pair; distinct rounds and hashes
	// make A the deterministic winner once the discovery window expires.
	competitorArtifact, err := simulator.BroadcastCompetingBlock(shardID)
	require.NoError(t, err)
	require.NotNil(t, competitorArtifact)
	competitorHeader := competitorArtifact.Header
	competitorHash := calculateHeaderHash(t, sourceNode, competitorHeader)
	require.Equal(t, localHeader.GetNonce(), competitorHeader.GetNonce())
	require.Equal(t, localHeader.GetPrevHash(), competitorHeader.GetPrevHash())
	require.Greater(t, competitorHeader.GetRound(), localHeader.GetRound())
	require.NotEqual(t, localHash, competitorHash)
	require.Less(t, getLastCrossNotarizedNonce(t, metaNode, shardID), localHeader.GetNonce())
	broadcastAll(t, sourceNode, localArtifact)

	// Step 5: Confirm that the source and metachain retain both proofs, then freeze source production
	// while shard 1 and metachain continue. Metachain must cross-notarize the exact A hash and the
	// source's final and settled checkpoints must subsequently reach that same hash.
	sourceProofsPool := sourceNode.GetDataComponents().Datapool().Proofs()
	metaProofsPool := metaNode.GetDataComponents().Datapool().Proofs()
	require.True(t, sourceProofsPool.HasProof(shardID, localHash))
	require.True(t, sourceProofsPool.HasProof(shardID, competitorHash))
	require.True(t, metaProofsPool.HasProof(shardID, localHash))
	require.True(t, metaProofsPool.HasProof(shardID, competitorHash))

	diagnostics := []diagnosticNode{
		{name: "source shard", node: sourceNode},
		{name: "destination shard", node: destinationNode},
		{name: "meta", node: metaNode},
	}
	generateBlocksUntilSkipping(t, simulator, metaArbitrationWindowRounds+4, []uint32{shardID}, func() bool {
		header, hash, getErr := metaNode.GetProcessComponents().BlockTracker().GetLastCrossNotarizedHeader(shardID)
		return getErr == nil && header.GetNonce() == localHeader.GetNonce() && bytes.Equal(hash, localHash)
	}, diagnostics...)
	generateBlocksUntilSkipping(t, simulator, 4, []uint32{shardID}, func() bool {
		settledNonce, settledHash := sourceNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
		return settledNonce == localHeader.GetNonce() && bytes.Equal(settledHash, localHash) &&
			getFinalNonce(sourceNode) == localHeader.GetNonce()
	}, diagnostics...)
	settledNonce, settledHash := sourceNode.GetProcessComponents().ForkDetector().GetHighestSettledBlockInfo()
	require.Equal(t, localHeader.GetNonce(), settledNonce)
	require.Equal(t, localHash, settledHash)
	require.Equal(t, localHeader.GetNonce(), getFinalNonce(sourceNode))
	require.Equal(t, localHash, sourceNode.GetProcessComponents().ForkDetector().GetHighestFinalBlockHash())
	require.Equal(t, localHash, sourceNode.GetChainHandler().GetCurrentBlockHeaderHash())

	// Step 6: Resume normal production on both shards and metachain until the receiver balance proves
	// that the incoming miniblock executed, then wait for metachain to notarize that destination
	// execution so the API reports success. The final result supplies the actual fee, allowing exact
	// source debit, nonce, and destination credit checks against the initial snapshots.
	expectedReceiverBalance := new(big.Int).Add(beforeTransfer.receiver.balance, transferValue)
	generateBlocksUntil(t, simulator, 8, func() bool {
		receiverSnapshot := getAccountSnapshot(t, simulator, receiver)
		return receiverSnapshot.balance.Cmp(expectedReceiverBalance) == 0
	}, diagnostics...)
	generateBlocksUntil(t, simulator, 5, func() bool {
		return isTransactionSuccessful(simulator, crossTxHash, receiver)
	}, diagnostics...)
	afterDestinationExecution := getTxAndAccounts(t, simulator, crossTxHash, sender, receiver)
	require.Equal(t, transaction.TxStatusSuccess, afterDestinationExecution.txResult.Status)
	requireTransferAppliedExactlyOnce(t, beforeTransfer, afterDestinationExecution, transferValue)
	require.Equal(t, afterSourceExecutionSender, afterDestinationExecution.sender)

	require.GreaterOrEqual(t, getSettledNonce(sourceNode), localHeader.GetNonce())
	require.GreaterOrEqual(t, getFinalNonce(sourceNode), localHeader.GetNonce())
	requireSettledNotAheadOfFinal(t, sourceNode)
	requireSettledNotAheadOfFinal(t, destinationNode)

	// Step 7: Replay A's full artifact and B's header/proof, then produce several ordinary blocks.
	// Account snapshots must remain byte-for-byte equivalent to the post-execution state, and both
	// shards must advance without final or settled checkpoint regression.
	sourceBeforeReplay := captureForkState(sourceNode)
	destinationBeforeReplay := captureForkState(destinationNode)
	broadcastAll(t, sourceNode, localArtifact)
	broadcastAll(t, sourceNode, competitorArtifact)
	require.True(t, sourceProofsPool.HasProof(shardID, localHash))
	require.True(t, sourceProofsPool.HasProof(shardID, competitorHash))
	require.NoError(t, simulator.GenerateBlocks(3))

	afterReplay := getTxAndAccounts(t, simulator, crossTxHash, sender, receiver)
	require.Equal(t, transaction.TxStatusSuccess, afterReplay.txResult.Status)
	require.Equal(t, afterDestinationExecution.sender, afterReplay.sender)
	require.Equal(t, afterDestinationExecution.receiver, afterReplay.receiver)

	sourceAfterReplay := captureForkState(sourceNode)
	destinationAfterReplay := captureForkState(destinationNode)
	require.Greater(t, sourceAfterReplay.currentNonce, sourceBeforeReplay.currentNonce)
	require.Greater(t, destinationAfterReplay.currentNonce, destinationBeforeReplay.currentNonce)
	require.GreaterOrEqual(t, sourceAfterReplay.finalNonce, sourceBeforeReplay.finalNonce)
	require.GreaterOrEqual(t, sourceAfterReplay.settledNonce, sourceBeforeReplay.settledNonce)
	require.GreaterOrEqual(t, destinationAfterReplay.finalNonce, destinationBeforeReplay.finalNonce)
	require.GreaterOrEqual(t, destinationAfterReplay.settledNonce, destinationBeforeReplay.settledNonce)
	requireSettledNotAheadOfFinal(t, sourceNode)
	requireSettledNotAheadOfFinal(t, destinationNode)
}
