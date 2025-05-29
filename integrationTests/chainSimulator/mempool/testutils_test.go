package mempool

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	testsChainSimulator "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

var (
	oneEGLD                   = big.NewInt(1000000000000000000)
	oneQuarterOfEGLD          = big.NewInt(250000000000000000)
	durationWaitAfterSendMany = 3000 * time.Millisecond
	durationWaitAfterSendSome = 300 * time.Millisecond
)

func startChainSimulator(t *testing.T, alterConfigsFunction func(cfg *config.Configs)) testsChainSimulator.ChainSimulator {
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

type participantsHolder struct {
	sendersByShard  map[int][]dtos.WalletAddress
	relayerByShard  map[int]dtos.WalletAddress
	receiverByShard map[int]dtos.WalletAddress
}

func newParticipantsHolder() *participantsHolder {
	return &participantsHolder{
		sendersByShard:  make(map[int][]dtos.WalletAddress),
		relayerByShard:  make(map[int]dtos.WalletAddress),
		receiverByShard: make(map[int]dtos.WalletAddress),
	}
}

func createParticipants(t *testing.T, simulator testsChainSimulator.ChainSimulator, numSendersPerShard int) *participantsHolder {
	numShards := int(simulator.GetNodeHandler(0).GetShardCoordinator().NumberOfShards())
	participants := newParticipantsHolder()

	for shard := 0; shard < numShards; shard++ {
		senders := make([]dtos.WalletAddress, 0, numSendersPerShard)

		for i := 0; i < numSendersPerShard; i++ {
			sender, err := simulator.GenerateAndMintWalletAddress(uint32(shard), oneEGLD)
			require.NoError(t, err)

			senders = append(senders, sender)
		}

		relayer, err := simulator.GenerateAndMintWalletAddress(uint32(shard), oneEGLD)
		require.NoError(t, err)

		receiver, err := simulator.GenerateAndMintWalletAddress(uint32(shard), big.NewInt(0))
		require.NoError(t, err)

		participants.sendersByShard[shard] = senders
		participants.relayerByShard[shard] = relayer
		participants.receiverByShard[shard] = receiver
	}

	err := simulator.GenerateBlocks(1)
	require.Nil(t, err)

	return participants
}

type noncesTracker struct {
	nonceByAddress map[string]uint64
}

func newNoncesTracker() *noncesTracker {
	return &noncesTracker{
		nonceByAddress: make(map[string]uint64),
	}
}

func (tracker *noncesTracker) getThenIncrementNonce(address dtos.WalletAddress) uint64 {
	nonce, ok := tracker.nonceByAddress[address.Bech32]
	if !ok {
		tracker.nonceByAddress[address.Bech32] = 0
	}

	tracker.nonceByAddress[address.Bech32]++
	return nonce
}

func sendTransactions(t *testing.T, simulator testsChainSimulator.ChainSimulator, transactions []*transaction.Transaction) {
	transactionsBySenderShard := make(map[int][]*transaction.Transaction)
	shardCoordinator := simulator.GetNodeHandler(0).GetShardCoordinator()

	for _, tx := range transactions {
		shard := int(shardCoordinator.ComputeId(tx.SndAddr))
		transactionsBySenderShard[shard] = append(transactionsBySenderShard[shard], tx)
	}

	for shard, transactionsFromShard := range transactionsBySenderShard {
		node := simulator.GetNodeHandler(uint32(shard))

		for _, tx := range transactionsFromShard {
			err := node.GetFacadeHandler().ValidateTransaction(tx)
			require.NoError(t, err)
		}

		numSent, err := node.GetFacadeHandler().SendBulkTransactions(transactionsFromShard)

		require.NoError(t, err)
		require.Equal(t, len(transactionsFromShard), int(numSent))
	}
}

func sendTransaction(t *testing.T, simulator testsChainSimulator.ChainSimulator, tx *transaction.Transaction) {
	sendTransactions(t, simulator, []*transaction.Transaction{tx})
}

func selectTransactions(t *testing.T, simulator testsChainSimulator.ChainSimulator, shard int) ([]*txcache.WrappedTransaction, uint64) {
	shardAsString := strconv.Itoa(shard)
	node := simulator.GetNodeHandler(uint32(shard))
	accountsAdapter := node.GetStateComponents().AccountsAdapter()
	poolsHolder := node.GetDataComponents().Datapool().Transactions()

	selectionSession, err := preprocess.NewSelectionSession(preprocess.ArgsSelectionSession{
		AccountsAdapter:       accountsAdapter,
		TransactionsProcessor: &testscommon.TxProcessorStub{},
	})
	require.NoError(t, err)

	mempool := poolsHolder.ShardDataStore(shardAsString).(*txcache.TxCache)

	selectedTransactions, gas := mempool.SelectTransactions(
		selectionSession,
		process.TxCacheSelectionGasRequested,
		process.TxCacheSelectionMaxNumTxs,
		process.TxCacheSelectionLoopMaximumDuration,
	)

	return selectedTransactions, gas
}

func getNumTransactionsInPool(simulator testsChainSimulator.ChainSimulator, shard int) int {
	node := simulator.GetNodeHandler(uint32(shard))
	poolsHolder := node.GetDataComponents().Datapool().Transactions()
	return int(poolsHolder.GetCounts().GetTotal())
}

func getNumTransactionsInCurrentBlock(simulator testsChainSimulator.ChainSimulator, shard int) int {
	node := simulator.GetNodeHandler(uint32(shard))
	currentBlock := node.GetDataComponents().Blockchain().GetCurrentBlockHeader()
	return int(currentBlock.GetTxCount())
}

func getTransaction(t *testing.T, simulator testsChainSimulator.ChainSimulator, shard int, hash []byte) *transaction.ApiTransactionResult {
	hashAsHex := hex.EncodeToString(hash)
	transaction, err := simulator.GetNodeHandler(uint32(shard)).GetFacadeHandler().GetTransaction(hashAsHex, true)
	require.NoError(t, err)
	return transaction
}
