package relayedTx

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/require"
)

var (
	oneEGLD = big.NewInt(1000000000000000000)
)

func TestMempoolWithChainSimulator_Selection(t *testing.T) {
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

	time.Sleep(500 * time.Millisecond)
	require.Equal(t, 300000, int(mempool.GetCounts().GetTotal()))

	err = simulator.GenerateBlocks(1)
	require.Nil(t, err)

	currentBlock := node.GetDataComponents().Blockchain().GetCurrentBlockHeader()
	require.Equal(t, 27755, int(currentBlock.GetTxCount()))

	miniblockHeader := currentBlock.GetMiniBlockHeaderHandlers()[0]
	miniblockHash := miniblockHeader.GetHash()

	miniblocks, _ := node.GetDataComponents().MiniBlocksProvider().GetMiniBlocks([][]byte{miniblockHash})
	require.Equal(t, 1, len(miniblocks))
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

func startChainSimulator(t *testing.T, alterConfigsFunction func(cfg *config.Configs),
) testsChainSimulator.ChainSimulator {
	simulator, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    "../../../cmd/node/config/",
		NumOfShards:            1,
		GenesisTimestamp:       time.Now().Unix(),
		RoundDurationInMillis:  uint64(4000),
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    10,
		},
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         1,
		MetaChainMinNodes:        1,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction:     alterConfigsFunction,
	})
	require.NoError(t, err)
	require.NotNil(t, simulator)

	err = simulator.GenerateBlocksUntilEpochIsReached(1)
	require.NoError(t, err)

	return simulator
}
