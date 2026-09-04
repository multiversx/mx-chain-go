package settlement

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	chainSimulatorHeartbeat "github.com/multiversx/mx-chain-go/node/chainSimulator/components/heartbeat"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

const (
	proofBroadcastDelay       = 5 * time.Millisecond
	txPropagationTimeout      = time.Second
	txPropagationPollInterval = time.Millisecond
)

type diagnosticNode struct {
	name string
	node process.NodeHandler
}

type forkStateSnapshot struct {
	currentNonce uint64
	currentHash  []byte
	finalNonce   uint64
	finalHash    []byte
	settledNonce uint64
	settledHash  []byte
	probable     uint64
	forkDetected bool
	forkNonce    uint64
	forkHash     []byte
}

type accountSnapshot struct {
	nonce   uint64
	balance *big.Int
}

type txAndAccountsSnapshot struct {
	txResult *transaction.ApiTransactionResult
	sender   accountSnapshot
	receiver accountSnapshot
}

type directBlockCreator interface {
	IncrementRound()
	CreateNewBlock() (*dtos.BroadcastData, error)
	CreateCompetingBlock(original data.HeaderHandler) (*dtos.BroadcastData, error)
}

func observeSelfNotarizedFromMeta(node process.NodeHandler, targetHash []byte) *atomic.Bool {
	published := &atomic.Bool{}
	node.GetProcessComponents().BlockTracker().RegisterSelfNotarizedFromCrossHeadersHandler(
		func(shardID uint32, _ []data.HeaderHandler, hashes [][]byte) {
			if shardID != core.MetachainShardId {
				return
			}
			for _, hash := range hashes {
				if bytes.Equal(hash, targetHash) {
					published.Store(true)
					return
				}
			}
		},
	)

	return published
}

func (snapshot forkStateSnapshot) String() string {
	return fmt.Sprintf(
		"current=%d/%x final=%d/%x settled=%d/%x probable=%d fork=%t/%d/%x",
		snapshot.currentNonce,
		snapshot.currentHash,
		snapshot.finalNonce,
		snapshot.finalHash,
		snapshot.settledNonce,
		snapshot.settledHash,
		snapshot.probable,
		snapshot.forkDetected,
		snapshot.forkNonce,
		snapshot.forkHash,
	)
}

func captureForkState(node process.NodeHandler) forkStateSnapshot {
	header, currentHash := node.GetChainHandler().GetCurrentBlockHeaderAndHash()
	currentNonce := uint64(0)
	if !check.IfNil(header) {
		currentNonce = header.GetNonce()
	}

	forkDetector := node.GetProcessComponents().ForkDetector()
	settledNonce, settledHash := forkDetector.GetHighestSettledBlockInfo()
	forkInfo := forkDetector.CheckFork()
	snapshot := forkStateSnapshot{
		currentNonce: currentNonce,
		currentHash:  currentHash,
		finalNonce:   forkDetector.GetHighestFinalBlockNonce(),
		finalHash:    forkDetector.GetHighestFinalBlockHash(),
		settledNonce: settledNonce,
		settledHash:  settledHash,
		probable:     forkDetector.ProbableHighestNonce(),
	}
	if forkInfo != nil {
		snapshot.forkDetected = forkInfo.IsDetected
		snapshot.forkNonce = forkInfo.Nonce
		snapshot.forkHash = forkInfo.Hash
	}

	return snapshot
}

func formatForkDiagnostics(nodes []diagnosticNode) string {
	diagnostics := ""
	for _, diagnostic := range nodes {
		diagnostics += fmt.Sprintf("\n%s: %s", diagnostic.name, captureForkState(diagnostic.node))
	}

	return diagnostics
}

func defaultDiagnosticNodes(simulator testsChainSimulator.ChainSimulator) []diagnosticNode {
	return []diagnosticNode{
		{name: "shard", node: simulator.GetNodeHandler(shardID)},
		{name: "meta", node: simulator.GetNodeHandler(core.MetachainShardId)},
	}
}

func generateBlocksUntil(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
	maxBlocks int,
	condition func() bool,
	diagnosticNodes ...diagnosticNode,
) {
	t.Helper()

	for i := 0; i < maxBlocks; i++ {
		if condition() {
			return
		}
		require.NoError(t, simulator.GenerateBlocks(1))
	}
	if condition() {
		return
	}
	if len(diagnosticNodes) == 0 {
		diagnosticNodes = defaultDiagnosticNodes(simulator)
	}

	require.FailNow(t, "condition not reached within block budget", formatForkDiagnostics(diagnosticNodes))
}

func generateBlocksUntilSkipping(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
	maxBlocks int,
	skippedShardIDs []uint32,
	condition func() bool,
	diagnosticNodes ...diagnosticNode,
) {
	t.Helper()

	for i := 0; i < maxBlocks; i++ {
		if condition() {
			return
		}
		require.NoError(t, simulator.GenerateBlocksSkippingShards(1, skippedShardIDs))
	}
	if condition() {
		return
	}
	if len(diagnosticNodes) == 0 {
		diagnosticNodes = defaultDiagnosticNodes(simulator)
	}

	require.FailNow(t, "condition not reached within block budget", formatForkDiagnostics(diagnosticNodes))
}

func sendTxWithoutGeneratingBlocks(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
	tx *transaction.Transaction,
) string {
	t.Helper()
	require.NotNil(t, tx)

	shardCoordinator := simulator.GetNodeHandler(shardID).GetShardCoordinator()
	senderShardID := shardCoordinator.ComputeId(tx.SndAddr)
	senderNode := simulator.GetNodeHandler(senderShardID)
	require.NoError(t, senderNode.GetFacadeHandler().ValidateTransaction(tx))

	txHash, err := core.CalculateHash(
		senderNode.GetCoreComponents().InternalMarshalizer(),
		senderNode.GetCoreComponents().Hasher(),
		tx,
	)
	require.NoError(t, err)

	numSent, err := senderNode.GetFacadeHandler().SendBulkTransactions([]*transaction.Transaction{tx})
	require.NoError(t, err)
	require.Equal(t, uint64(1), numSent)

	txHashHex := hex.EncodeToString(txHash)
	require.Eventually(t, func() bool {
		recoveredTx, _ := senderNode.GetFacadeHandler().GetTransaction(txHashHex, false)
		return recoveredTx != nil
	}, txPropagationTimeout, txPropagationPollInterval, "transaction %s was not propagated to the sender shard", txHashHex)

	return txHashHex
}

func getTxAndAccounts(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
	txHash string,
	sender dtos.WalletAddress,
	receiver dtos.WalletAddress,
) txAndAccountsSnapshot {
	t.Helper()

	destinationShardID := simulator.GetNodeHandler(shardID).GetShardCoordinator().ComputeId(receiver.Bytes)
	txResult, err := simulator.GetNodeHandler(destinationShardID).GetFacadeHandler().GetTransaction(txHash, true)
	require.NoError(t, err)
	require.NotNil(t, txResult)

	return txAndAccountsSnapshot{
		txResult: txResult,
		sender:   getAccountSnapshot(t, simulator, sender),
		receiver: getAccountSnapshot(t, simulator, receiver),
	}
}

func getAccountSnapshot(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
	wallet dtos.WalletAddress,
) accountSnapshot {
	t.Helper()

	account, err := simulator.GetAccount(wallet)
	require.NoError(t, err)
	balance, ok := new(big.Int).SetString(account.Balance, 10)
	require.True(t, ok, "invalid account balance %q", account.Balance)

	return accountSnapshot{
		nonce:   account.Nonce,
		balance: balance,
	}
}

func requireTransferAppliedExactlyOnce(
	t *testing.T,
	before txAndAccountsSnapshot,
	after txAndAccountsSnapshot,
	value *big.Int,
) {
	t.Helper()
	require.Equal(t, before.sender.nonce, after.txResult.Nonce)
	require.Equal(t, before.sender.nonce+1, after.sender.nonce)
	require.Equal(t, before.receiver.nonce, after.receiver.nonce)

	fee, ok := new(big.Int).SetString(after.txResult.Fee, 10)
	require.True(t, ok, "invalid transaction fee %q", after.txResult.Fee)
	expectedSenderBalance := new(big.Int).Sub(before.sender.balance, value)
	expectedSenderBalance.Sub(expectedSenderBalance, fee)
	expectedReceiverBalance := new(big.Int).Add(before.receiver.balance, value)
	require.Zero(t, expectedSenderBalance.Cmp(after.sender.balance))
	require.Zero(t, expectedReceiverBalance.Cmp(after.receiver.balance))
}

func isTransactionSuccessful(
	simulator testsChainSimulator.ChainSimulator,
	txHash string,
	receiver dtos.WalletAddress,
) bool {
	destinationShardID := simulator.GetNodeHandler(shardID).GetShardCoordinator().ComputeId(receiver.Bytes)
	txResult, err := simulator.GetNodeHandler(destinationShardID).GetFacadeHandler().GetTransaction(txHash, true)

	return err == nil && txResult != nil && txResult.Status == transaction.TxStatusSuccess
}

func generateMetaBlockWithoutShardBlock(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
) *dtos.BroadcastData {
	t.Helper()

	_, metaCreator := incrementDirectRound(t, simulator)
	artifact, err := metaCreator.CreateNewBlock()
	require.NoError(t, err)
	require.NotNil(t, artifact)

	return artifact
}

func broadcastMetaCompetitorWithoutShardBlock(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
) *dtos.BroadcastData {
	t.Helper()

	_, metaCreator := incrementDirectRound(t, simulator)
	metaNode := simulator.GetNodeHandler(core.MetachainShardId)
	artifact, err := metaCreator.CreateCompetingBlock(metaNode.GetChainHandler().GetCurrentBlockHeader())
	require.NoError(t, err)
	require.NotNil(t, artifact)
	broadcastAll(t, metaNode, artifact)

	return artifact
}

func generateShardBlockWithoutMetaBlock(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
) *dtos.BroadcastData {
	t.Helper()

	shardCreator, _ := incrementDirectRound(t, simulator)
	artifact, err := shardCreator.CreateNewBlock()
	require.NoError(t, err)
	require.NotNil(t, artifact)

	return artifact
}

func broadcastShardCompetitorWithoutMetaBlock(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
) *dtos.BroadcastData {
	t.Helper()

	shardCreator, _ := incrementDirectRound(t, simulator)
	shardNode := simulator.GetNodeHandler(shardID)
	artifact, err := shardCreator.CreateCompetingBlock(shardNode.GetChainHandler().GetCurrentBlockHeader())
	require.NoError(t, err)
	require.NotNil(t, artifact)
	broadcastAll(t, shardNode, artifact)

	return artifact
}

func incrementDirectRound(
	t *testing.T,
	simulator testsChainSimulator.ChainSimulator,
) (directBlockCreator, directBlockCreator) {
	t.Helper()

	shardCreator, err := process.NewBlocksCreator(
		simulator.GetNodeHandler(shardID),
		chainSimulatorHeartbeat.NewHeartbeatMonitor(),
		0,
		true,
	)
	require.NoError(t, err)
	metaCreator, err := process.NewBlocksCreator(
		simulator.GetNodeHandler(core.MetachainShardId),
		chainSimulatorHeartbeat.NewHeartbeatMonitor(),
		0,
		true,
	)
	require.NoError(t, err)

	shardCreator.IncrementRound()
	metaCreator.IncrementRound()

	return shardCreator, metaCreator
}

func broadcastHeader(t *testing.T, node process.NodeHandler, artifact *dtos.BroadcastData) {
	t.Helper()
	require.NotNil(t, artifact)
	require.False(t, check.IfNil(artifact.Header))
	require.NoError(t, node.GetBroadcastMessenger().BroadcastHeader(artifact.Header, artifact.LeaderKey))
}

func broadcastBodyAndTransactions(t *testing.T, node process.NodeHandler, artifact *dtos.BroadcastData) {
	t.Helper()
	require.NotNil(t, artifact)
	require.NoError(t, node.GetBroadcastMessenger().BroadcastMiniBlocks(artifact.MiniBlocksBytes, artifact.LeaderKey))
	require.NoError(t, node.GetBroadcastMessenger().BroadcastTransactions(artifact.TransactionsBytes, artifact.LeaderKey))
}

func broadcastProof(t *testing.T, node process.NodeHandler, artifact *dtos.BroadcastData) {
	t.Helper()
	require.NotNil(t, artifact)
	if check.IfNil(artifact.Proof) {
		return
	}

	// Keep the same small ordering delay used by the simulator so the proof cannot overtake its header.
	time.Sleep(proofBroadcastDelay)
	require.NoError(t, node.GetBroadcastMessenger().BroadcastEquivalentProof(artifact.Proof, artifact.LeaderKey))
}

func broadcastAll(t *testing.T, node process.NodeHandler, artifact *dtos.BroadcastData) {
	t.Helper()
	broadcastHeader(t, node, artifact)
	broadcastBodyAndTransactions(t, node, artifact)
	broadcastProof(t, node, artifact)
}

func calculateHeaderHash(t *testing.T, node process.NodeHandler, header data.HeaderHandler) []byte {
	t.Helper()
	require.False(t, check.IfNil(header))

	hash, err := core.CalculateHash(
		node.GetCoreComponents().InternalMarshalizer(),
		node.GetCoreComponents().Hasher(),
		header,
	)
	require.NoError(t, err)

	return hash
}

func requireMetaReferencesShardHeader(
	t *testing.T,
	metaHeader data.HeaderHandler,
	referencedShardID uint32,
	referencedHash []byte,
) {
	t.Helper()

	typedMetaHeader, ok := metaHeader.(data.MetaHeaderHandler)
	require.True(t, ok)
	for _, shardInfo := range typedMetaHeader.GetShardInfoProposalHandlers() {
		if shardInfo.GetShardID() == referencedShardID && bytes.Equal(shardInfo.GetHeaderHash(), referencedHash) {
			return
		}
	}

	require.FailNow(t, "metablock does not reference expected shard header",
		"meta nonce=%d shard=%d hash=%x", metaHeader.GetNonce(), referencedShardID, referencedHash)
}
