package mempool

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
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

	// Allow the eviction to complete (even if it's quite fast).
	time.Sleep(3 * time.Second)

	expectedNumTransactionsInPool := 300_000 + 1 + 1 - int(storage.TxPoolSourceMeNumItemsToPreemptivelyEvict)
	require.Equal(t, expectedNumTransactionsInPool, getNumTransactionsInPool(simulator, shard))
}
