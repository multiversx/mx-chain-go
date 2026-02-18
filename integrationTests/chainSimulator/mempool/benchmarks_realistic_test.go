package mempool

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/txcache"
)

// TestBenchmark_RealisticSelectionWithRealTrie measures tx selection performance
// using the chain simulator's real AccountsDB and trie, NOT mocked sessions.
//
// This exercises the full path:
//
//	SelectTransactions -> SelectionSession -> AccountsEphemeralProvider -> AccountsDB -> patriciaMerkleTrie.Get()
//
// The goal is to show that selection time scales with unique senders (trie lookups),
// not total transaction count.
func TestBenchmark_RealisticSelectionWithRealTrie(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	scenarios := []struct {
		name         string
		numSenders   int
		txsPerSender int
	}{
		{"100_senders_x_100_txs", 100, 100},
		{"1000_senders_x_10_txs", 1000, 10},
		{"5000_senders_x_2_txs", 5000, 2},
		{"10000_senders_x_1_tx", 10000, 1},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			benchmarkRealisticSelection(t, sc.numSenders, sc.txsPerSender)
		})
	}
}

// TestBenchmark_RealisticSecondSelectionWithRealTrie measures the second selection
// (after OnProposedBlock), which exercises the virtual session creation path:
// handleGlobalAccountBreadcrumbs -> GetAccountNonceAndBalance for every tracked sender.
func TestBenchmark_RealisticSecondSelectionWithRealTrie(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// Use small txsPerSender to avoid nonce gap validation errors
	// (each sender needs 2x txsPerSender nonces, and high nonces are rejected)
	scenarios := []struct {
		name         string
		numSenders   int
		txsPerSender int
	}{
		{"100_senders_x_5_txs", 100, 5},
		{"1000_senders_x_5_txs", 1000, 5},
		{"10000_senders_x_1_tx", 10000, 1},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			benchmarkRealisticSecondSelection(t, sc.numSenders, sc.txsPerSender)
		})
	}
}

func benchmarkRealisticSelection(t *testing.T, numSenders int, txsPerSender int) {
	simulator := startChainSimulator(t, nil)
	defer simulator.Close()

	shard := 0
	totalTxs := numSenders * txsPerSender

	senders := benchMintSenders(t, simulator, uint32(shard), numSenders)

	receiver, err := simulator.GenerateAndMintWalletAddress(uint32(shard), big.NewInt(0))
	require.NoError(t, err)

	// Commit minted accounts to trie
	err = simulator.GenerateBlocks(1)
	require.NoError(t, err)

	// Send transactions to real mempool
	txs := benchBuildTransactions(senders, receiver, txsPerSender, 0)
	sendTransactions(t, simulator, txs)
	time.Sleep(durationWaitAfterSendSome)

	numInPool := getNumTransactionsInPool(simulator, shard)
	fmt.Printf("  [setup] pool: %d txs, senders: %d, txs/sender: %d\n", numInPool, numSenders, txsPerSender)
	require.Equal(t, totalTxs, numInPool)

	// Create real trie-backed selection session
	node := simulator.GetNodeHandler(uint32(shard))
	accountsAdapter := node.GetStateComponents().AccountsAdapter()

	selectionSession, err := preprocess.NewSelectionSession(preprocess.ArgsSelectionSession{
		AccountsAdapter:         accountsAdapter,
		TransactionsProcessor:   &testscommon.TxProcessorStub{},
		TxVersionCheckerHandler: node.GetCoreComponents().TxVersionChecker(),
	})
	require.NoError(t, err)

	options, _ := holders.NewTxSelectionOptions(
		10_000_000_000,
		totalTxs,
		10,
		haveTimeTrue,
	)

	shardAsString := strconv.Itoa(shard)
	poolsHolder := node.GetDataComponents().Datapool().Transactions()
	mempool := poolsHolder.ShardDataStore(shardAsString).(*txcache.TxCache)

	// Measure first selection (each unique sender = 1 trie lookup via getRecord)
	start := time.Now()
	selectedTxs, _, err := mempool.SelectTransactions(selectionSession, options, 0)
	selectionDuration := time.Since(start)

	require.NoError(t, err)

	fmt.Printf("  [result] selected %d/%d txs in %v (%d unique senders)\n",
		len(selectedTxs), totalTxs, selectionDuration, numSenders)
}

func benchmarkRealisticSecondSelection(t *testing.T, numSenders int, txsPerSender int) {
	simulator := startChainSimulator(t, nil)
	defer simulator.Close()

	shard := 0
	// Send 2x txs per sender so there are enough for two selection rounds
	doubleTxsPerSender := txsPerSender * 2
	txsPerBlock := numSenders * txsPerSender

	// Mint with enough balance for all txs
	twoEGLD := new(big.Int).Mul(oneEGLD, big.NewInt(2))
	senders := make([]dtos.WalletAddress, 0, numSenders)
	for i := 0; i < numSenders; i++ {
		sender, err := simulator.GenerateAndMintWalletAddress(uint32(shard), twoEGLD)
		require.NoError(t, err)
		senders = append(senders, sender)
	}

	receiver, err := simulator.GenerateAndMintWalletAddress(uint32(shard), big.NewInt(0))
	require.NoError(t, err)

	err = simulator.GenerateBlocks(1)
	require.NoError(t, err)

	// Send all txs at once (nonces 0 to doubleTxsPerSender-1 per sender)
	allTxs := benchBuildTransactions(senders, receiver, doubleTxsPerSender, 0)
	sendTransactions(t, simulator, allTxs)
	time.Sleep(durationWaitAfterSendSome)

	node := simulator.GetNodeHandler(uint32(shard))
	accountsAdapter := node.GetStateComponents().AccountsAdapter()
	poolsHolder := node.GetDataComponents().Datapool().Transactions()
	shardAsString := strconv.Itoa(shard)
	mempool := poolsHolder.ShardDataStore(shardAsString).(*txcache.TxCache)

	// --- First selection + propose (creates breadcrumbs) ---
	session1, err := preprocess.NewSelectionSession(preprocess.ArgsSelectionSession{
		AccountsAdapter:         accountsAdapter,
		TransactionsProcessor:   &testscommon.TxProcessorStub{},
		TxVersionCheckerHandler: node.GetCoreComponents().TxVersionChecker(),
	})
	require.NoError(t, err)

	options, _ := holders.NewTxSelectionOptions(
		10_000_000_000,
		txsPerBlock,
		10,
		haveTimeTrue,
	)

	selectedTxs1, _, err := mempool.SelectTransactions(session1, options, 1)
	require.NoError(t, err)
	require.Equal(t, txsPerBlock, len(selectedTxs1))

	// Propose block (creates breadcrumbs for virtual session)
	proposedBlock := createProposedBlock(selectedTxs1)
	blockchain := node.GetDataComponents().Blockchain()
	currentBlockHash := blockchain.GetCurrentBlockHeaderHash()
	currentNonce := blockchain.GetCurrentBlockHeader().GetNonce()

	proposedHeader := &block.Header{
		Nonce:    currentNonce + 1,
		PrevHash: currentBlockHash,
	}

	err = mempool.OnProposedBlock([]byte("benchBlockHash1"), proposedBlock, proposedHeader, session1, currentBlockHash)
	require.NoError(t, err)

	fmt.Printf("  [setup] proposed %d txs, breadcrumbs for %d senders\n", len(selectedTxs1), numSenders)

	// --- Second selection: exercises handleGlobalAccountBreadcrumbs + virtual session ---
	session2, err := preprocess.NewSelectionSession(preprocess.ArgsSelectionSession{
		AccountsAdapter:         accountsAdapter,
		TransactionsProcessor:   &testscommon.TxProcessorStub{},
		TxVersionCheckerHandler: node.GetCoreComponents().TxVersionChecker(),
	})
	require.NoError(t, err)

	// This selection exercises: deriveVirtualSelectionSession -> handleGlobalAccountBreadcrumbs
	// (N trie lookups for N tracked senders) + the selection loop itself
	start := time.Now()
	selectedTxs2, _, err := mempool.SelectTransactions(session2, options, 2)
	selectionDuration := time.Since(start)
	require.NoError(t, err)

	fmt.Printf("  [result] second selection: %d txs in %v (%d senders with breadcrumbs)\n",
		len(selectedTxs2), selectionDuration, numSenders)
}

// benchMintSenders creates numSenders funded wallets in the specified shard.
func benchMintSenders(t *testing.T, simulator testsChainSimulator.ChainSimulator, shardID uint32, numSenders int) []dtos.WalletAddress {
	senders := make([]dtos.WalletAddress, 0, numSenders)
	for i := 0; i < numSenders; i++ {
		sender, err := simulator.GenerateAndMintWalletAddress(shardID, oneEGLD)
		require.NoError(t, err)
		senders = append(senders, sender)
	}
	return senders
}

// TestBenchmark_RealisticSecondSelectionGuarded measures second selection performance
// with ~40% of senders having guarded accounts. This exercises the optimization path:
// PrefetchAccounts -> light batch -> upgradeGuardedAccounts -> CreateFullAccountFromLight.
//
// Guarded accounts have codeMetadata with the guarded bit set and a non-empty data trie,
// so upgradeFromLight must load each guarded account's data trie (the unavoidable cost).
// Without the optimization, each guarded account would also re-read the main trie.
func TestBenchmark_RealisticSecondSelectionGuarded(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	scenarios := []struct {
		name           string
		numSenders     int
		txsPerSender   int
		guardedPercent float64
	}{
		{"1000_senders_40pct_guarded", 1000, 5, 0.40},
		{"5000_senders_40pct_guarded", 5000, 2, 0.40},
		{"10000_senders_40pct_guarded", 10000, 1, 0.40},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			benchmarkRealisticSecondSelectionGuarded(t, sc.numSenders, sc.txsPerSender, sc.guardedPercent)
		})
	}
}

func benchmarkRealisticSecondSelectionGuarded(t *testing.T, numSenders int, txsPerSender int, guardedPercent float64) {
	simulator := startChainSimulator(t, nil)
	defer simulator.Close()

	shard := 0
	doubleTxsPerSender := txsPerSender * 2
	txsPerBlock := numSenders * txsPerSender

	twoEGLD := new(big.Int).Mul(oneEGLD, big.NewInt(2))
	senders := make([]dtos.WalletAddress, 0, numSenders)
	for i := 0; i < numSenders; i++ {
		sender, err := simulator.GenerateAndMintWalletAddress(uint32(shard), twoEGLD)
		require.NoError(t, err)
		senders = append(senders, sender)
	}

	receiver, err := simulator.GenerateAndMintWalletAddress(uint32(shard), big.NewInt(0))
	require.NoError(t, err)

	// Commit minted accounts to trie
	err = simulator.GenerateBlocks(1)
	require.NoError(t, err)

	// Make guardedPercent of senders guarded via direct state manipulation.
	// This sets codeMetadata with guarded bit + writes data trie entry (non-nil rootHash).
	numGuarded := setupGuardedAccountsDirect(t, simulator, uint32(shard), senders, guardedPercent)

	// Generate a block to sync the selection tracker's latestRootHash with
	// the trie state after our direct account manipulation.
	err = simulator.GenerateBlocks(1)
	require.NoError(t, err)

	fmt.Printf("  [setup] %d/%d senders set as guarded (%.0f%%)\n", numGuarded, numSenders, guardedPercent*100)

	// Send benchmark transactions (nonces start at 0 — we only modified codeMetadata/dataTrie, not nonces)
	allTxs := benchBuildTransactions(senders, receiver, doubleTxsPerSender, 0)
	sendTransactions(t, simulator, allTxs)
	time.Sleep(durationWaitAfterSendSome)

	node := simulator.GetNodeHandler(uint32(shard))
	accountsAdapter := node.GetStateComponents().AccountsAdapter()
	poolsHolder := node.GetDataComponents().Datapool().Transactions()
	shardAsString := strconv.Itoa(shard)
	mempool := poolsHolder.ShardDataStore(shardAsString).(*txcache.TxCache)

	// --- First selection + propose (creates breadcrumbs) ---
	session1, err := preprocess.NewSelectionSession(preprocess.ArgsSelectionSession{
		AccountsAdapter:         accountsAdapter,
		TransactionsProcessor:   &testscommon.TxProcessorStub{},
		TxVersionCheckerHandler: node.GetCoreComponents().TxVersionChecker(),
	})
	require.NoError(t, err)

	options, _ := holders.NewTxSelectionOptions(
		10_000_000_000,
		txsPerBlock,
		10,
		haveTimeTrue,
	)

	selectedTxs1, _, err := mempool.SelectTransactions(session1, options, 1)
	require.NoError(t, err)

	proposedBlock := createProposedBlock(selectedTxs1)
	blockchain := node.GetDataComponents().Blockchain()
	currentBlockHash := blockchain.GetCurrentBlockHeaderHash()
	currentNonce := blockchain.GetCurrentBlockHeader().GetNonce()

	proposedHeader := &block.Header{
		Nonce:    currentNonce + 1,
		PrevHash: currentBlockHash,
	}

	err = mempool.OnProposedBlock([]byte("benchBlockHash1"), proposedBlock, proposedHeader, session1, currentBlockHash)
	require.NoError(t, err)

	fmt.Printf("  [setup] proposed %d txs, breadcrumbs for %d senders\n", len(selectedTxs1), numSenders)

	// --- Second selection: exercises prefetch + upgradeGuardedAccounts + virtual session ---
	session2, err := preprocess.NewSelectionSession(preprocess.ArgsSelectionSession{
		AccountsAdapter:         accountsAdapter,
		TransactionsProcessor:   &testscommon.TxProcessorStub{},
		TxVersionCheckerHandler: node.GetCoreComponents().TxVersionChecker(),
	})
	require.NoError(t, err)

	start := time.Now()
	selectedTxs2, _, err := mempool.SelectTransactions(session2, options, 2)
	selectionDuration := time.Since(start)
	require.NoError(t, err)

	fmt.Printf("  [result] second selection: %d txs in %v (%d senders, %d guarded)\n",
		len(selectedTxs2), selectionDuration, numSenders, numGuarded)
}

// setupGuardedAccountsDirect marks a percentage of senders as guarded via direct state
// manipulation (no chain simulator transactions). Sets codeMetadata guarded bit and writes
// a data trie entry so each guarded account has a non-nil rootHash.
func setupGuardedAccountsDirect(t *testing.T, simulator testsChainSimulator.ChainSimulator, shardID uint32, senders []dtos.WalletAddress, guardedPercent float64) int {
	node := simulator.GetNodeHandler(shardID)
	accountsAdapter := node.GetStateComponents().AccountsAdapter()

	numGuarded := int(float64(len(senders)) * guardedPercent)

	for i := 0; i < numGuarded; i++ {
		account, err := accountsAdapter.LoadAccount(senders[i].Bytes)
		require.NoError(t, err)

		userAccount, ok := account.(state.UserAccountHandler)
		require.True(t, ok)

		// Set guarded bit (0x08) in codeMetadata first byte.
		// This makes lightAccountInfo.IsGuarded() return true, triggering upgradeGuardedAccounts.
		userAccount.SetCodeMetadata([]byte{0x08, 0x00})

		// Write a data trie entry so the account has a non-nil rootHash.
		// This ensures CreateFullAccountFromLight exercises real data trie loading.
		err = userAccount.SaveKeyValue([]byte("guardian_bench_data"), []byte("placeholder"))
		require.NoError(t, err)

		err = accountsAdapter.SaveAccount(userAccount)
		require.NoError(t, err)
	}

	_, err := accountsAdapter.Commit()
	require.NoError(t, err)

	return numGuarded
}

// benchBuildTransactions creates txsPerSender transactions per sender.
func benchBuildTransactions(senders []dtos.WalletAddress, receiver dtos.WalletAddress, txsPerSender int, startNonce uint64) []*transaction.Transaction {
	txs := make([]*transaction.Transaction, 0, len(senders)*txsPerSender)

	for nIdx := 0; nIdx < txsPerSender; nIdx++ {
		for _, sender := range senders {
			tx := &transaction.Transaction{
				Nonce:     startNonce + uint64(nIdx),
				Value:     big.NewInt(1),
				SndAddr:   sender.Bytes,
				RcvAddr:   receiver.Bytes,
				Data:      []byte{},
				GasLimit:  uint64(gasLimit),
				GasPrice:  uint64(gasPrice),
				ChainID:   []byte(configs.ChainID),
				Version:   2,
				Signature: []byte("signature"),
			}
			txs = append(txs, tx)
		}
	}

	return txs
}
