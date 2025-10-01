package mempool

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common/holders"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/multiversx/mx-chain-go/txcache"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/storage"
)

func TestMempoolWithChainSimulator_Selection(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numSenders := 10000
	numTransactionsPerSender := 3
	shard := 0

	simulator := startChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	participants := createParticipants(t, simulator, numSenders)
	noncesTracker := newNoncesTracker()

	transactions := make([]*transaction.Transaction, 0, numSenders*numTransactionsPerSender)

	for i := 0; i < numSenders; i++ {
		sender := participants.sendersByShard[shard][i]
		receiver := participants.receiverByShard[shard]

		for j := 0; j < numTransactionsPerSender; j++ {
			tx := &transaction.Transaction{
				Nonce:     noncesTracker.getThenIncrementNonce(sender),
				Value:     oneQuarterOfEGLD,
				SndAddr:   sender.Bytes,
				RcvAddr:   receiver.Bytes,
				Data:      []byte{},
				GasLimit:  50_000,
				GasPrice:  1_000_000_000,
				ChainID:   []byte(configs.ChainID),
				Version:   2,
				Signature: []byte("signature"),
			}

			transactions = append(transactions, tx)
		}
	}

	sendTransactions(t, simulator, transactions)
	time.Sleep(durationWaitAfterSendMany)
	require.Equal(t, 30_000, getNumTransactionsInPool(simulator, shard))

	selectedTransactions, gas := selectTransactions(t, simulator, shard)
	require.Equal(t, 30_000, len(selectedTransactions))
	require.Equal(t, 50_000*30_000, int(gas))

	err := simulator.GenerateBlocks(1)
	require.Nil(t, err)
	require.Equal(t, 27_756, getNumTransactionsInCurrentBlock(simulator, shard))

	selectedTransactions, gas = selectTransactions(t, simulator, shard)
	require.Equal(t, 30_000-27_756, len(selectedTransactions))
	require.Equal(t, 50_000*(30_000-27_756), int(gas))
}

func TestMempoolWithChainSimulator_Selection_WhenUsersHaveZeroBalance_WithRelayedV3(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shard := 0

	simulator := startChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	err := simulator.GenerateBlocksUntilEpochIsReached(2)
	require.NoError(t, err)

	relayer, err := simulator.GenerateAndMintWalletAddress(uint32(shard), oneEGLD)
	require.NoError(t, err)

	receiver, err := simulator.GenerateAndMintWalletAddress(uint32(shard), big.NewInt(0))
	require.NoError(t, err)

	alice, err := simulator.GenerateAndMintWalletAddress(uint32(shard), big.NewInt(0))
	require.NoError(t, err)

	bob, err := simulator.GenerateAndMintWalletAddress(uint32(shard), big.NewInt(0))
	require.NoError(t, err)

	err = simulator.GenerateBlocks(1)
	require.Nil(t, err)

	noncesTracker := newNoncesTracker()
	transactions := make([]*transaction.Transaction, 0)

	// Transfer (executable, invalid) from Alice (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            noncesTracker.getThenIncrementNonce(alice),
		Value:            oneQuarterOfEGLD,
		SndAddr:          alice.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_002,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Contract call from Bob (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            noncesTracker.getThenIncrementNonce(bob),
		Value:            big.NewInt(0),
		SndAddr:          bob.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte("hello"),
		GasLimit:         100_000 + 5*1500,
		GasPrice:         1_000_000_001,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	sendTransactions(t, simulator, transactions)
	time.Sleep(durationWaitAfterSendSome)
	require.Equal(t, 2, getNumTransactionsInPool(simulator, shard))

	selectedTransactions, _ := selectTransactions(t, simulator, shard)
	require.Equal(t, 2, len(selectedTransactions))
	require.Equal(t, alice.Bytes, selectedTransactions[0].Tx.GetSndAddr())
	require.Equal(t, bob.Bytes, selectedTransactions[1].Tx.GetSndAddr())

	err = simulator.GenerateBlocks(1)
	require.Nil(t, err)
	require.Equal(t, 2, getNumTransactionsInCurrentBlock(simulator, shard))

	require.Equal(t, "invalid", getTransaction(t, simulator, shard, selectedTransactions[0].TxHash).Status.String())
	require.Equal(t, "success", getTransaction(t, simulator, shard, selectedTransactions[1].TxHash).Status.String())
}

func TestMempoolWithChainSimulator_Selection_WhenInsufficientBalanceForFee_WithRelayedV3(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numSenders := 3
	shard := 0

	simulator := startChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	err := simulator.GenerateBlocksUntilEpochIsReached(2)
	require.NoError(t, err)

	participants := createParticipants(t, simulator, numSenders)
	noncesTracker := newNoncesTracker()

	alice := participants.sendersByShard[shard][0]
	bob := participants.sendersByShard[shard][1]
	carol := participants.sendersByShard[shard][2]
	relayer := participants.relayerByShard[shard]
	receiver := participants.receiverByShard[shard]

	transactions := make([]*transaction.Transaction, 0)

	// Consume most of relayer's balance. Keep an amount that is enough for the fee of two simple transfer transactions.
	currentRelayerBalance := int64(1000000000000000000)
	feeForTransfer := int64(50_000 * 1_000_000_004)
	feeForRelayingTransactionsOfAliceAndBob := int64(100_000*1_000_000_003 + 100_000*1_000_000_002)

	transactions = append(transactions, &transaction.Transaction{
		Nonce:     noncesTracker.getThenIncrementNonce(relayer),
		Value:     big.NewInt(currentRelayerBalance - feeForTransfer - feeForRelayingTransactionsOfAliceAndBob),
		SndAddr:   relayer.Bytes,
		RcvAddr:   receiver.Bytes,
		Data:      []byte{},
		GasLimit:  50_000,
		GasPrice:  1_000_000_004,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	// Transfer from Alice (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            noncesTracker.getThenIncrementNonce(alice),
		Value:            oneQuarterOfEGLD,
		SndAddr:          alice.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_003,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Bob (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            noncesTracker.getThenIncrementNonce(bob),
		Value:            oneQuarterOfEGLD,
		SndAddr:          bob.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_002,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Carol (relayed) - this one should not be selected due to insufficient balance (of the relayer)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            noncesTracker.getThenIncrementNonce(carol),
		Value:            oneQuarterOfEGLD,
		SndAddr:          carol.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_001,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	sendTransactions(t, simulator, transactions)
	time.Sleep(durationWaitAfterSendSome)
	require.Equal(t, 4, getNumTransactionsInPool(simulator, shard))

	selectedTransactions, _ := selectTransactions(t, simulator, shard)
	require.Equal(t, 3, len(selectedTransactions))
	require.Equal(t, relayer.Bytes, selectedTransactions[0].Tx.GetSndAddr())
	require.Equal(t, alice.Bytes, selectedTransactions[1].Tx.GetSndAddr())
	require.Equal(t, bob.Bytes, selectedTransactions[2].Tx.GetSndAddr())
}

func TestMempoolWithChainSimulator_Selection_WhenInitialGap(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numSenders := 2
	shard := 0

	simulator := startChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	err := simulator.GenerateBlocksUntilEpochIsReached(2)
	require.NoError(t, err)

	participants := createParticipants(t, simulator, numSenders)
	noncesTracker := newNoncesTracker()

	alice := participants.sendersByShard[shard][0]
	bob := participants.sendersByShard[shard][1]
	relayer := participants.relayerByShard[shard]
	receiver := participants.receiverByShard[shard]

	transactions := make([]*transaction.Transaction, 0)

	// Transfer from Alice (relayed) with wrong nonce
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            1,
		Value:            oneQuarterOfEGLD,
		SndAddr:          alice.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_003,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Bob (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            noncesTracker.getThenIncrementNonce(bob),
		Value:            oneQuarterOfEGLD,
		SndAddr:          bob.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_002,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	sendTransactions(t, simulator, transactions)
	time.Sleep(durationWaitAfterSendSome)
	require.Equal(t, 2, getNumTransactionsInPool(simulator, shard))

	selectedTransactions, _ := selectTransactions(t, simulator, shard)
	require.Equal(t, 1, len(selectedTransactions))
	require.Equal(t, bob.Bytes, selectedTransactions[0].Tx.GetSndAddr())
}

func TestMempoolWithChainSimulator_Selection_WhenMiddleGap(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numSenders := 2
	shard := 0

	simulator := startChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	err := simulator.GenerateBlocksUntilEpochIsReached(2)
	require.NoError(t, err)

	participants := createParticipants(t, simulator, numSenders)
	noncesTracker := newNoncesTracker()

	alice := participants.sendersByShard[shard][0]
	bob := participants.sendersByShard[shard][1]
	relayer := participants.relayerByShard[shard]
	receiver := participants.receiverByShard[shard]

	transactions := make([]*transaction.Transaction, 0)

	// Transfer from Alice (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            noncesTracker.getThenIncrementNonce(alice),
		Value:            oneQuarterOfEGLD,
		SndAddr:          alice.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_003,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Another transfer from Alice (relayed) which is not following the last nonce
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            100,
		Value:            oneQuarterOfEGLD,
		SndAddr:          alice.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_003,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Bob (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            noncesTracker.getThenIncrementNonce(bob),
		Value:            oneQuarterOfEGLD,
		SndAddr:          bob.Bytes,
		RcvAddr:          receiver.Bytes,
		RelayerAddr:      relayer.Bytes,
		Data:             []byte{},
		GasLimit:         100_000,
		GasPrice:         1_000_000_002,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	sendTransactions(t, simulator, transactions)
	time.Sleep(durationWaitAfterSendSome)
	require.Equal(t, 3, getNumTransactionsInPool(simulator, shard))

	selectedTransactions, _ := selectTransactions(t, simulator, shard)
	require.Equal(t, 2, len(selectedTransactions))
	require.Equal(t, alice.Bytes, selectedTransactions[0].Tx.GetSndAddr())
	require.Equal(t, bob.Bytes, selectedTransactions[1].Tx.GetSndAddr())
}

func TestMempoolWithChainSimulator_Eviction(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numSenders := 10000
	numTransactionsPerSender := 30
	shard := 0

	simulator := startChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	participants := createParticipants(t, simulator, numSenders)
	noncesTracker := newNoncesTracker()

	transactions := make([]*transaction.Transaction, 0, numSenders*numTransactionsPerSender)

	for i := 0; i < numSenders; i++ {
		sender := participants.sendersByShard[shard][i]
		receiver := participants.receiverByShard[shard]

		for j := 0; j < numTransactionsPerSender; j++ {
			tx := &transaction.Transaction{
				Nonce:     noncesTracker.getThenIncrementNonce(sender),
				Value:     oneQuarterOfEGLD,
				SndAddr:   sender.Bytes,
				RcvAddr:   receiver.Bytes,
				Data:      []byte{},
				GasLimit:  50_000,
				GasPrice:  1_000_000_000,
				ChainID:   []byte(configs.ChainID),
				Version:   2,
				Signature: []byte("signature"),
			}

			transactions = append(transactions, tx)
		}
	}

	sendTransactions(t, simulator, transactions)
	time.Sleep(durationWaitAfterSendMany)
	require.Equal(t, 300_000, getNumTransactionsInPool(simulator, shard))

	// Send one more transaction (fill up the mempool)
	sendTransaction(t, simulator, &transaction.Transaction{
		Nonce:     42,
		Value:     oneEGLD,
		SndAddr:   participants.sendersByShard[shard][7].Bytes,
		RcvAddr:   participants.receiverByShard[shard].Bytes,
		Data:      []byte{},
		GasLimit:  50000,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	time.Sleep(durationWaitAfterSendSome)
	require.Equal(t, 300_001, getNumTransactionsInPool(simulator, shard))

	// Send one more transaction to trigger eviction
	sendTransaction(t, simulator, &transaction.Transaction{
		Nonce:     43,
		Value:     oneEGLD,
		SndAddr:   participants.sendersByShard[shard][7].Bytes,
		RcvAddr:   participants.receiverByShard[shard].Bytes,
		Data:      []byte{},
		GasLimit:  50000,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	// Allow the eviction to complete (even if it's quite fast).
	time.Sleep(5 * time.Second)

	expectedNumTransactionsInPool := 300_000 + 1 + 1 - int(storage.TxPoolSourceMeNumItemsToPreemptivelyEvict)
	require.Equal(t, expectedNumTransactionsInPool, getNumTransactionsInPool(simulator, shard))
}

func Test_Selection_ShouldNotSelectSameTransactionsWithSameSender(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	// create the non-virtual selection session, assure we have enough balance
	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: oneEGLD,
			Nonce:   0,
		},
		"receiver": {
			Balance: big.NewInt(0),
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMockWithAccounts(accounts)

	maxNumTxs := 2
	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		maxNumTxs,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	numPoolTxs := maxNumTxs * 2
	txHashes := make([][]byte, 0, numPoolTxs)

	nonceTracker := newNoncesTracker()
	for i := 0; i < numPoolTxs; i++ {
		tx := &transaction.Transaction{
			Nonce:     nonceTracker.getThenIncrementNonceByStringAddress("alice"),
			Value:     oneQuarterOfEGLD,
			SndAddr:   []byte("alice"),
			RcvAddr:   []byte("receiver"),
			GasLimit:  50_000,
			GasPrice:  1_000_000_000,
			ChainID:   []byte(configs.ChainID),
			Version:   2,
			Signature: []byte("signature"),
		}

		txHash := []byte(fmt.Sprintf("txHash%d", i))
		txpool.AddTx(&txcache.WrappedTransaction{
			Tx:               tx,
			TxHash:           txHash,
			Fee:              big.NewInt(int64(tx.GasLimit * tx.GasPrice)),
			TransferredValue: tx.Value,
			FeePayer:         tx.SndAddr,
		})

		txHashes = append(txHashes, txHash)
	}

	blockBody := block.Body{MiniBlocks: []*block.MiniBlock{
		{
			TxHashes: txHashes[:len(txHashes)/2],
		},
	}}

	require.Equal(t, txpool.CountTx(), uint64(4))

	// do the first selection, first two txs should be returned
	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 2, len(selectedTransactions))
	require.Equal(t, "txHash0", string(selectedTransactions[0].TxHash))
	require.Equal(t, "txHash1", string(selectedTransactions[1].TxHash))

	// propose the block
	err = txpool.OnProposedBlock([]byte("blockHash1"), &blockBody, &block.Header{
		Nonce:    0,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		accountsProvider,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection. should not return same txs
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 2, len(selectedTransactions))
	require.Equal(t, "txHash2", string(selectedTransactions[0].TxHash))
	require.Equal(t, "txHash3", string(selectedTransactions[1].TxHash))
}

func Test_Selection_ShouldNotSelectSameTransactionsWithDifferentSenders(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	// assure we have enough balance for each account
	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: oneEGLD,
			Nonce:   0,
		},
		"bob": {
			Balance: oneEGLD,
			Nonce:   0,
		},
		"receiver": {
			Balance: big.NewInt(0),
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMockWithAccounts(accounts)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		2,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	nonceTracker := newNoncesTracker()
	transactions := make([]*transaction.Transaction, 0)

	transactions = append(transactions, &transaction.Transaction{
		Nonce:     nonceTracker.getThenIncrementNonceByStringAddress("alice"),
		Value:     oneQuarterOfEGLD,
		SndAddr:   []byte("alice"),
		RcvAddr:   []byte("receiver"),
		GasLimit:  50_000,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	transactions = append(transactions, &transaction.Transaction{
		Nonce:     nonceTracker.getThenIncrementNonceByStringAddress("bob"),
		Value:     oneQuarterOfEGLD,
		SndAddr:   []byte("bob"),
		RcvAddr:   []byte("receiver"),
		GasLimit:  50_000,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	transactions = append(transactions, &transaction.Transaction{
		Nonce:     nonceTracker.getThenIncrementNonceByStringAddress("alice"),
		Value:     oneQuarterOfEGLD,
		SndAddr:   []byte("alice"),
		RcvAddr:   []byte("receiver"),
		GasLimit:  50_000,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	transactions = append(transactions, &transaction.Transaction{
		Nonce:     nonceTracker.getThenIncrementNonceByStringAddress("bob"),
		Value:     oneQuarterOfEGLD,
		SndAddr:   []byte("bob"),
		RcvAddr:   []byte("receiver"),
		GasLimit:  50_000,
		GasPrice:  1_000_000_000,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	txHashes := make([][]byte, 0)
	for i, tx := range transactions {
		txHash := []byte(fmt.Sprintf("txHash%d", i))
		txHashes = append(txHashes, txHash)
		txpool.AddTx(&txcache.WrappedTransaction{
			Tx:               tx,
			TxHash:           txHash,
			Fee:              big.NewInt(int64(tx.GasLimit * tx.GasPrice)),
			TransferredValue: tx.Value,
			FeePayer:         tx.SndAddr,
		})
	}

	blockBody := block.Body{MiniBlocks: []*block.MiniBlock{
		{
			TxHashes: txHashes[:len(txHashes)/2],
		},
	}}

	require.Equal(t, txpool.CountTx(), uint64(4))

	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 2, len(selectedTransactions))
	require.Equal(t, "txHash0", string(selectedTransactions[0].TxHash))
	require.Equal(t, "txHash1", string(selectedTransactions[1].TxHash))

	// propose the selected transactions
	err = txpool.OnProposedBlock([]byte("blockHash1"), &blockBody,
		&block.Header{
			Nonce:    1,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		accountsProvider,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection. should not return same txs
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 2, len(selectedTransactions))
	require.Equal(t, "txHash2", string(selectedTransactions[0].TxHash))
	require.Equal(t, "txHash3", string(selectedTransactions[1].TxHash))
}

func Test_Selection_ShouldNotSelectSameTransactionsWithManyTransactions(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	numTxsPerSender := 30_000
	initialAmount := big.NewInt(int64(numTxsPerSender) * 50_000 * 1_000_000_000)

	senders := []string{"alice", "bob"}
	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: initialAmount,
			Nonce:   0,
		},
		"bob": {
			Balance: initialAmount,
			Nonce:   0,
		},
		"receiver": {
			Balance: big.NewInt(0),
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMockWithAccounts(accounts)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		numTxsPerSender,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	numTxs := numTxsPerSender * len(senders)
	nonceTracker := newNoncesTracker()

	// create numTxs transactions and save them to txpool
	addTransactionsToTxPool(txpool, nonceTracker, numTxsPerSender, senders)
	require.Equal(t, txpool.CountTx(), uint64(numTxs))

	// do the first selections
	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, numTxsPerSender, len(selectedTransactions))

	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	proposedBlock1 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
		Nonce:    1,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		accountsProvider,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection (the rest of the transactions should be selected)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, numTxsPerSender, len(selectedTransactions))

	// propose the second block
	proposedBlock2 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock2,
		&block.Header{
			Nonce:    2,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		accountsProvider,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the last selection (no tx should be returned)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 0, len(selectedTransactions))
}

func Test_Selection_ProposeEmptyBlocks(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	numTxsPerSender := 30_000
	initialAmount := big.NewInt(int64(numTxsPerSender) * 50_000 * 1_000_000_000)

	senders := []string{"alice", "bob"}
	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: initialAmount,
			Nonce:   0,
		},
		"bob": {
			Balance: initialAmount,
			Nonce:   0,
		},
		"receiver": {
			Balance: big.NewInt(0),
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		numTxsPerSender,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	numTxs := numTxsPerSender * len(senders)
	nonceTracker := newNoncesTracker()

	// create numTxs transactions and save them to txpool
	addTransactionsToTxPool(txpool, nonceTracker, numTxsPerSender, senders)
	require.Equal(t, txpool.CountTx(), uint64(numTxs))

	// do the first selections
	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, numTxsPerSender, len(selectedTransactions))

	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	proposedBlock1 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
		Nonce:    1,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// propose some empty blocks
	err = txpool.OnProposedBlock([]byte("blockHash2"), &block.Body{},
		&block.Header{
			Nonce:    2,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	err = txpool.OnProposedBlock([]byte("blockHash3"), &block.Body{},
		&block.Header{
			Nonce:    3,
			PrevHash: []byte("blockHash2"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection (the rest of the transactions should be selected)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, numTxsPerSender, len(selectedTransactions))

	// propose the second block
	proposedBlock2 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash4"), proposedBlock2,
		&block.Header{
			Nonce:    4,
			PrevHash: []byte("blockHash3"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the last selection (no tx should be returned)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 0, len(selectedTransactions))
}

func Test_Selection_ProposeBlocksWithSameNonce(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	numTxsPerSender := 30_000
	initialAmount := big.NewInt(int64(numTxsPerSender) * 50_000 * 1_000_000_000)

	senders := []string{"alice", "bob"}
	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: initialAmount,
			Nonce:   0,
		},
		"bob": {
			Balance: initialAmount,
			Nonce:   0,
		},
		"receiver": {
			Balance: big.NewInt(0),
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		numTxsPerSender,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	numTxs := numTxsPerSender * len(senders)
	nonceTracker := newNoncesTracker()

	// create numTxs transactions and save them to txpool
	addTransactionsToTxPool(txpool, nonceTracker, numTxsPerSender, senders)
	require.Equal(t, txpool.CountTx(), uint64(numTxs))

	// do the first selection
	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, numTxsPerSender, len(selectedTransactions))

	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	proposedBlock1 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
		Nonce:    1,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// propose an empty block with same nonce as the previous one
	err = txpool.OnProposedBlock([]byte("blockHash1"), &block.Body{},
		&block.Header{
			Nonce:    1,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// because the first one was replaced, the same transactions should be selected again
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, numTxsPerSender, len(selectedTransactions))

	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	proposedBlock2 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock2, &block.Header{
		Nonce:    2,
		PrevHash: []byte("blockHash1"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection (the rest of the transactions should be selected)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, numTxsPerSender, len(selectedTransactions))

	// propose the second block
	proposedBlock3 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash3"), proposedBlock3,
		&block.Header{
			Nonce:    3,
			PrevHash: []byte("blockHash2"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the last selection (no tx should be returned)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 0, len(selectedTransactions))
}

func Test_Selection_ShouldNotSelectSameTransactionsWithManyTransactionsAndExecutedBlockNotification(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	// set the number of transactions that we want for each sender
	numTxsPerSender := 60_000
	// assure that we have enough balance for fees
	initialAmount := big.NewInt(int64(numTxsPerSender) * 50_000 * 1_000_000_000)

	senders := []string{"alice"}
	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: initialAmount,
			Nonce:   0,
		},
		"receiver": {
			Balance: big.NewInt(0),
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMockWithAccounts(accounts)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		30_000,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	nonceTracker := newNoncesTracker()

	// create txs and add them to txpool
	numTxs := numTxsPerSender * len(senders)
	addTransactionsToTxPool(txpool, nonceTracker, numTxsPerSender, senders)
	require.Equal(t, txpool.CountTx(), uint64(numTxs))

	// do the first selection
	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 30_000, len(selectedTransactions))

	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	proposedBlock1 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1,
		&block.Header{
			Nonce:    1,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		accountsProvider,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection (the rest of the transactions should be selected)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 30_000, len(selectedTransactions))

	// execute the first proposed block
	err = txpool.OnExecutedBlock(&block.Header{
		Nonce:    1,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	})
	require.Nil(t, err)

	// remove the executed txs from the pool
	for _, tx := range proposedBlock1.MiniBlocks[0].TxHashes {
		require.True(t, txpool.RemoveTxByHash(tx))
	}

	// update the state of the account on the blockchain
	selectionSession.SetNonce([]byte("alice"), 30_000)

	// propose the second block
	proposedBlock2 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock2,
		&block.Header{
			Nonce:    2,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		accountsProvider,
		holders.NewBlockchainInfo([]byte("blockHash1"), nil, 2),
	)
	require.Nil(t, err)

	// no transactions should be returned
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 0, len(selectedTransactions))

	for _, tx := range proposedBlock2.MiniBlocks[0].TxHashes {
		require.True(t, txpool.RemoveTxByHash(tx))
	}
}

func Test_Selection_ProposeEmptyBlocksAndExecutedBlockNotification(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	// set the number of transactions that we want for each sender
	numTxsPerSender := 60_000
	// assure that we have enough balance for fees
	initialAmount := big.NewInt(int64(numTxsPerSender) * 50_000 * 1_000_000_000)

	// mock the non-virtual selection session
	senders := []string{"alice"}
	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: initialAmount,
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		30_000,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	nonceTracker := newNoncesTracker()
	numTxs := numTxsPerSender * len(senders)

	// create txs and add them to txpool
	addTransactionsToTxPool(txpool, nonceTracker, numTxsPerSender, senders)
	require.Equal(t, txpool.CountTx(), uint64(numTxs))

	// do the first selection
	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 30_000, len(selectedTransactions))

	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	proposedBlock1 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1,
		&block.Header{
			Nonce:    1,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// propose empty blocks
	err = txpool.OnProposedBlock([]byte("blockHash2"), &block.Body{},
		&block.Header{
			Nonce:    2,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	err = txpool.OnProposedBlock([]byte("blockHash3"), &block.Body{},
		&block.Header{
			Nonce:    3,
			PrevHash: []byte("blockHash2"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection (the rest of the transactions should be selected)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 30_000, len(selectedTransactions))

	// execute the first proposed block
	err = txpool.OnExecutedBlock(&block.Header{
		Nonce:    1,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	})
	require.Nil(t, err)

	// remove the executed txs from the pool
	for _, tx := range proposedBlock1.MiniBlocks[0].TxHashes {
		require.True(t, txpool.RemoveTxByHash(tx))
	}

	// update the state of the account on the blockchain
	selectionSession.SetNonce([]byte("alice"), 30_000)

	// propose the second block
	proposedBlock2 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash4"), proposedBlock2,
		&block.Header{
			Nonce:    4,
			PrevHash: []byte("blockHash3"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		selectionSession,
		holders.NewBlockchainInfo([]byte("blockHash1"), nil, 2),
	)
	require.Nil(t, err)

	// execute the empty proposed blocks
	err = txpool.OnExecutedBlock(&block.Header{
		Nonce:    2,
		PrevHash: []byte("blockHash1"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
	})
	require.Nil(t, err)

	// execute the empty proposed blocks
	err = txpool.OnExecutedBlock(&block.Header{
		Nonce:    3,
		PrevHash: []byte("blockHash2"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
	})
	require.Nil(t, err)

	// no transactions should be returned
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.NoError(t, err)
	require.Equal(t, 0, len(selectedTransactions))

	for _, tx := range proposedBlock2.MiniBlocks[0].TxHashes {
		require.True(t, txpool.RemoveTxByHash(tx))
	}
}

func Test_Selection_MaxTrackedBlocksReached(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(txcache.ConfigSourceMe{
		Name:                        "test",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBoundTest,
		CountThreshold:              math.MaxUint32,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
		TxCacheBoundsConfig: config.TxCacheBoundsConfig{
			MaxNumBytesPerSenderUpperBound: maxNumBytesPerSenderUpperBoundTest,
			MaxTrackedBlocks:               3,
		},
	}, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	numTxsPerSender := 30_000
	initialAmount := big.NewInt(int64(numTxsPerSender) * 50_000 * 1_000_000_000)

	senders := []string{"alice", "bob"}
	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: initialAmount,
			Nonce:   0,
		},
		"bob": {
			Balance: initialAmount,
			Nonce:   0,
		},
		"receiver": {
			Balance: big.NewInt(0),
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		numTxsPerSender,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	numTxs := numTxsPerSender * len(senders)
	nonceTracker := newNoncesTracker()

	addTransactionsToTxPool(txpool, nonceTracker, numTxsPerSender, senders)
	require.Equal(t, txpool.CountTx(), uint64(numTxs))

	// do the first selections
	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, numTxsPerSender, len(selectedTransactions))

	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	proposedBlock1 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
		Nonce:    1,
		PrevHash: []byte("blockHash0"),
		RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
	},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection (the rest of the transactions should be selected)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, numTxsPerSender, len(selectedTransactions))

	// propose the second block
	proposedBlock2 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock2,
		&block.Header{
			Nonce:    2,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the last selection (no tx should be returned)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 0, len(selectedTransactions))

	// propose one more block (an empty one) just to trigger the MaxTrackedBlocks
	err = txpool.OnProposedBlock([]byte("blockHash3"), &block.Body{},
		&block.Header{
			Nonce:    3,
			PrevHash: []byte("blockHash2"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// proposing an empty block when MaxTrackedBlocks is reached should work
	err = txpool.OnProposedBlock([]byte("blockHash4"), &block.Body{},
		&block.Header{
			Nonce:    4,
			PrevHash: []byte("blockHash3"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// proposing a block with transactions when MaxTrackedBlocks is reached should not work
	err = txpool.OnProposedBlock([]byte("blockHash4"), &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{},
		},
	},
		&block.Header{
			Nonce:    4,
			PrevHash: []byte("blockHash3"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 1)),
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.ErrorContains(t, err, "bad block received while max tracked blocks is reached")

	// proposing a block with transactions and with new execution results when MaxTrackedBlocks is reached should work
	err = txpool.OnProposedBlock([]byte("blockHash5"), &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{},
		},
	},
		&block.HeaderV3{
			Nonce:    5,
			PrevHash: []byte("blockHash4"),
			ExecutionResults: []*block.ExecutionResult{
				{},
			},
		},
		selectionSession,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)
}

func Test_SelectionWhenFeeExceedsBalanceWithMax3TxsSelected(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: oneQuarterOfEGLD,
			Nonce:   0,
		},
		"bob": {
			Balance: oneQuarterOfEGLD,
			Nonce:   0,
		},
		"carol": {
			Balance: oneQuarterOfEGLD,
			Nonce:   0,
		},
		"receiver": {
			Balance: big.NewInt(0),
			Nonce:   0,
		},
		"relayer": {
			Balance: oneEGLD,
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMockWithAccounts(accounts)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		3,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	// Consume most of relayer's balance. Keep an amount that is enough for the fee of two simple transfer transactions.
	currentRelayerBalance := int64(1000000000000000000)
	feeForTransfer := int64(50_000 * 1_000_000_004)
	feeForRelayingTransactionsOfAliceAndBob := int64(100_000*1_000_000_003 + 100_000*1_000_000_002)

	transactions := make([]*transaction.Transaction, 0)

	transactions = append(transactions, &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(currentRelayerBalance - feeForTransfer - feeForRelayingTransactionsOfAliceAndBob),
		SndAddr:   []byte("relayer"),
		RcvAddr:   []byte("receiver"),
		GasLimit:  50_000,
		GasPrice:  1_000_000_004,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	// Transfer from Alice (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            0,
		Value:            oneQuarterOfEGLD,
		SndAddr:          []byte("alice"),
		RcvAddr:          []byte("receiver"),
		RelayerAddr:      []byte("relayer"),
		GasLimit:         100_000,
		GasPrice:         1_000_000_003,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Bob (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            0,
		Value:            oneQuarterOfEGLD,
		SndAddr:          []byte("bob"),
		RcvAddr:          []byte("receiver"),
		RelayerAddr:      []byte("relayer"),
		GasLimit:         100_000,
		GasPrice:         1_000_000_002,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Carol (relayed) - this one should not be selected due to insufficient balance (of the relayer)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            0,
		Value:            oneQuarterOfEGLD,
		SndAddr:          []byte("carol"),
		RcvAddr:          []byte("receiver"),
		RelayerAddr:      []byte("relayer"),
		GasLimit:         100_000,
		GasPrice:         1_000_000_001,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	for i, tx := range transactions {
		txHash := []byte(fmt.Sprintf("txHash%d", i))
		txpool.AddTx(&txcache.WrappedTransaction{
			Tx:               tx,
			TxHash:           txHash,
			Fee:              big.NewInt(int64(tx.GasLimit * tx.GasPrice)),
			TransferredValue: tx.Value,
			FeePayer:         tx.RelayerAddr,
		})
	}

	require.Equal(t, txpool.CountTx(), uint64(4))

	// do the first selection: first 3 transactions should be returned
	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 3, len(selectedTransactions))
	require.Equal(t, "relayer", string(selectedTransactions[0].Tx.GetSndAddr()))
	require.Equal(t, "alice", string(selectedTransactions[1].Tx.GetSndAddr()))
	require.Equal(t, "bob", string(selectedTransactions[2].Tx.GetSndAddr()))

	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	proposedBlock1 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1,
		&block.Header{
			Nonce:    1,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		accountsProvider,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection, last tx should not be returned (relayer has insufficient balance)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 0, len(selectedTransactions))
}

func Test_SelectionWhenFeeExceedsBalanceWithMax2TxsSelected(t *testing.T) {
	t.Parallel()

	host := txcachemocks.NewMempoolHostMock()
	txpool, err := txcache.NewTxCache(configSourceMe, host)

	require.Nil(t, err)
	require.NotNil(t, txpool)

	accounts := map[string]*stateMock.UserAccountStub{
		"alice": {
			Balance: oneQuarterOfEGLD,
			Nonce:   0,
		},
		"bob": {
			Balance: oneQuarterOfEGLD,
			Nonce:   0,
		},
		"carol": {
			Balance: oneQuarterOfEGLD,
			Nonce:   0,
		},
		"receiver": {
			Balance: big.NewInt(0),
			Nonce:   0,
		},
		"relayer": {
			Balance: oneEGLD,
			Nonce:   0,
		},
	}

	selectionSession := txcachemocks.NewSelectionSessionMockWithAccounts(accounts)
	accountsProvider := txcachemocks.NewAccountNonceAndBalanceProviderMockWithAccounts(accounts)

	options := holders.NewTxSelectionOptions(
		10_000_000_000,
		2,
		int(selectionLoopMaximumDuration.Milliseconds()),
		10,
	)

	// Consume most of relayer's balance. Keep an amount that is enough for the fee of two simple transfer transactions.
	currentRelayerBalance := int64(1000000000000000000)
	feeForTransfer := int64(50_000 * 1_000_000_004)
	feeForRelayingTransactionsOfAliceAndBob := int64(100_000*1_000_000_003 + 100_000*1_000_000_002)

	transactions := make([]*transaction.Transaction, 0)

	transactions = append(transactions, &transaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(currentRelayerBalance - feeForTransfer - feeForRelayingTransactionsOfAliceAndBob),
		SndAddr:   []byte("relayer"),
		RcvAddr:   []byte("receiver"),
		GasLimit:  50_000,
		GasPrice:  1_000_000_004,
		ChainID:   []byte(configs.ChainID),
		Version:   2,
		Signature: []byte("signature"),
	})

	// Transfer from Alice (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            0,
		Value:            oneQuarterOfEGLD,
		SndAddr:          []byte("alice"),
		RcvAddr:          []byte("receiver"),
		RelayerAddr:      []byte("relayer"),
		GasLimit:         100_000,
		GasPrice:         1_000_000_003,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Bob (relayed)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            0,
		Value:            oneQuarterOfEGLD,
		SndAddr:          []byte("bob"),
		RcvAddr:          []byte("receiver"),
		RelayerAddr:      []byte("relayer"),
		GasLimit:         100_000,
		GasPrice:         1_000_000_002,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	// Transfer from Carol (relayed) - this one should not be selected due to insufficient balance (of the relayer)
	transactions = append(transactions, &transaction.Transaction{
		Nonce:            0,
		Value:            oneQuarterOfEGLD,
		SndAddr:          []byte("carol"),
		RcvAddr:          []byte("receiver"),
		RelayerAddr:      []byte("relayer"),
		GasLimit:         100_000,
		GasPrice:         1_000_000_001,
		ChainID:          []byte(configs.ChainID),
		Version:          2,
		Signature:        []byte("signature"),
		RelayerSignature: []byte("signature"),
	})

	for i, tx := range transactions {
		txHash := []byte(fmt.Sprintf("txHash%d", i))
		txpool.AddTx(&txcache.WrappedTransaction{
			Tx:               tx,
			TxHash:           txHash,
			Fee:              big.NewInt(int64(tx.GasLimit * tx.GasPrice)),
			TransferredValue: tx.Value,
			FeePayer:         tx.RelayerAddr,
		})
	}

	require.Equal(t, txpool.CountTx(), uint64(4))

	// do the first selection: first 3 transactions should be returned
	selectedTransactions, _, err := txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 2, len(selectedTransactions))
	require.Equal(t, "relayer", string(selectedTransactions[0].Tx.GetSndAddr()))
	require.Equal(t, "alice", string(selectedTransactions[1].Tx.GetSndAddr()))

	// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
	proposedBlock1 := createProposedBlock(selectedTransactions)
	err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1,
		&block.Header{
			Nonce:    1,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		},
		accountsProvider,
		defaultBlockchainInfo,
	)
	require.Nil(t, err)

	// do the second selection, last tx should not be returned (relayer has insufficient balance)
	selectedTransactions, _, err = txpool.SelectTransactions(selectionSession, options)
	require.Nil(t, err)
	require.Equal(t, 1, len(selectedTransactions))
	require.Equal(t, "bob", string(selectedTransactions[0].Tx.GetSndAddr()))
}
