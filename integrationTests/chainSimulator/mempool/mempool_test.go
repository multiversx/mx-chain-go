package mempool

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func TestMempoolWithChainSimulator_Selection(t *testing.T) {
	logger.SetLogLevel("*:INFO,txcache:DEBUG")

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
	require.Equal(t, 30_000, getNumTransactionsInPool(simulator, shard))

	selectedTransactions, gas := selectTransactions(t, simulator, shard)
	require.Equal(t, 30_000, len(selectedTransactions))
	require.Equal(t, 50_000*30_000, int(gas))

	err := simulator.GenerateBlocks(1)
	require.Nil(t, err)
	require.Equal(t, 27_756, getNumTransactionsInCurrentBlock(simulator, shard))

	selectedTransactions, gas = selectTransactions(t, simulator, shard)
	require.Equal(t, 30_000, len(selectedTransactions))
	require.Equal(t, 50_000*30_000, int(gas))
}

func TestMempoolWithChainSimulator_Eviction(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startChainSimulator(t, func(cfg *config.Configs) {})
	node := simulator.GetNodeHandler(0)
	mempool := node.GetDataComponents().Datapool().Transactions()

	defer simulator.Close()

	numSenders := 10000
	numTransactionsPerSender := 30

	senders := make([]dtos.WalletAddress, numSenders)
	sendersNonces := make([]uint64, numSenders)

	for i := 0; i < numSenders; i++ {
		sender, err := simulator.GenerateAndMintWalletAddress(0, oneEGLD)
		require.NoError(t, err)

		senders[i] = sender
	}

	receiver, err := simulator.GenerateAndMintWalletAddress(0, big.NewInt(0))
	require.NoError(t, err)

	err = simulator.GenerateBlocks(1)
	require.Nil(t, err)

	transactions := make([]*transaction.Transaction, 0, numSenders*numTransactionsPerSender)

	for i := 0; i < numSenders; i++ {
		for j := 0; j < numTransactionsPerSender; j++ {
			tx := &transaction.Transaction{
				Nonce:     sendersNonces[i],
				Value:     oneEGLD,
				SndAddr:   senders[i].Bytes,
				RcvAddr:   receiver.Bytes,
				Data:      []byte{},
				GasLimit:  50000,
				GasPrice:  1_000_000_000,
				ChainID:   []byte(configs.ChainID),
				Version:   2,
				Signature: []byte("signature"),
			}

			sendersNonces[i]++
			transactions = append(transactions, tx)
		}
	}

	numSent, err := node.GetFacadeHandler().SendBulkTransactions(transactions)
	require.NoError(t, err)
	require.Equal(t, 300000, int(numSent))

	time.Sleep(1 * time.Second)
	require.Equal(t, 300000, int(mempool.GetCounts().GetTotal()))

	// Send one more transaction (fill up the mempool)
	_, err = node.GetFacadeHandler().SendBulkTransactions([]*transaction.Transaction{
		{
			Nonce:     42,
			Value:     oneEGLD,
			SndAddr:   senders[7].Bytes,
			RcvAddr:   receiver.Bytes,
			Data:      []byte{},
			GasLimit:  50000,
			GasPrice:  1_000_000_000,
			ChainID:   []byte(configs.ChainID),
			Version:   2,
			Signature: []byte("signature"),
		},
	})
	require.NoError(t, err)

	time.Sleep(42 * time.Millisecond)
	require.Equal(t, 300001, int(mempool.GetCounts().GetTotal()))

	// Send one more transaction to trigger eviction
	_, err = node.GetFacadeHandler().SendBulkTransactions([]*transaction.Transaction{
		{
			Nonce:     42,
			Value:     oneEGLD,
			SndAddr:   senders[7].Bytes,
			RcvAddr:   receiver.Bytes,
			Data:      []byte{},
			GasLimit:  50000,
			GasPrice:  1_000_000_000,
			ChainID:   []byte(configs.ChainID),
			Version:   2,
			Signature: []byte("signature"),
		},
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	require.Equal(t, 300000+1+1-int(storage.TxPoolSourceMeNumItemsToPreemptivelyEvict), int(mempool.GetCounts().GetTotal()))
}
