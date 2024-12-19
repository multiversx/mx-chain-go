package mempool

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/require"
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

	time.Sleep(1 * time.Second)

	expectedNumTransactionsInPool := 300_000 + 1 + 1 - int(storage.TxPoolSourceMeNumItemsToPreemptivelyEvict)
	require.Equal(t, expectedNumTransactionsInPool, getNumTransactionsInPool(simulator, shard))
}
