package hardFork

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update/factory"
	"github.com/stretchr/testify/assert"
)

var log = logger.GetOrCreate("integrationTests/hardfork")

func TestEpochStartChangeWithoutTransactionInMultiShardedEnvironment(t *testing.T) {
	//TODO continue writing the tests in the next PR
	//if testing.Short() {
	t.Skip("this is not a short test")
	//}

	numOfShards := 2
	nodesPerShard := 1
	numMetachainNodes := 1

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	roundsPerEpoch := uint64(10)
	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 5
	/////////----- wait for epoch end period
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, int(roundsPerEpoch), nonce, round, idxProposers)

	time.Sleep(time.Second)

	_, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	time.Sleep(time.Second)

	epoch := uint32(1)
	verifyIfNodesHasCorrectEpoch(t, epoch, nodes)
	verifyIfAddedShardHeadersAreWithNewEpoch(t, nodes)

	createHardForkExporter(t, nodes)

	_ = logger.SetLogLevel("*:TRACE")

	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, node := range nodes {
		//go func() {
		log.Warn("***********************************************************************************")
		log.Warn("starting to export for node with shard", "id", node.ShardCoordinator.SelfId())
		err := node.ExportHandler.ExportAll(1)
		assert.Nil(t, err)
		log.Warn("***********************************************************************************")
		wg.Done()
		//}()
	}
	wg.Wait()
}

func TestEpochStartChangeWithContinuousTransactionsInMultiShardedEnvironment(t *testing.T) {
	//TODO continue writing the tests in the next PR
	//if testing.Short() {
	t.Skip("this is not a short test")
	//}

	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	roundsPerEpoch := uint64(10)
	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	initialVal := big.NewInt(10000000)
	sendValue := big.NewInt(5)
	integrationTests.MintAllNodes(nodes, initialVal)
	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	_ = logger.SetLogLevel("*:TRACE")
	/////////----- wait for epoch end period
	epoch := uint32(2)
	nrRoundsToPropagateMultiShard := uint64(5)
	for i := uint64(0); i <= (uint64(epoch)*roundsPerEpoch)+nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		for _, node := range nodes {
			integrationTests.CreateAndSendTransaction(node, sendValue, receiverAddress1, "")
			integrationTests.CreateAndSendTransaction(node, sendValue, receiverAddress2, "")
		}

		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)

	verifyIfNodesHasCorrectEpoch(t, epoch, nodes)
	verifyIfAddedShardHeadersAreWithNewEpoch(t, nodes)

	createHardForkExporter(t, nodes)

	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, node := range nodes {
		var err error
		go func(err error) {
			err = node.ExportHandler.ExportAll(2)
			assert.Nil(t, err)
			wg.Done()
		}(err)
	}
	wg.Wait()
}

func createHardForkExporter(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
) {
	for id, node := range nodes {
		argsExportHandler := factory.ArgsExporter{
			Marshalizer:      integrationTests.TestMarshalizer,
			Hasher:           integrationTests.TestHasher,
			HeaderValidator:  node.HeaderValidator,
			Uint64Converter:  integrationTests.TestUint64Converter,
			DataPool:         node.DataPool,
			StorageService:   node.Storage,
			RequestHandler:   node.RequestHandler,
			ShardCoordinator: node.ShardCoordinator,
			Messenger:        node.Messenger,
			ActiveTries:      node.TrieContainer,
			ExportFolder:     "export" + fmt.Sprintf("%d", id),
			ExportTriesStorageConfig: config.StorageConfig{
				Cache: config.CacheConfig{
					Size: 10000, Type: "LRU", Shards: 1,
				},
				DB: config.DBConfig{
					FilePath:          "ExportTrie" + fmt.Sprintf("%d", id),
					Type:              "LvlDBSerial",
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
			ExportTriesCacheConfig: config.CacheConfig{
				Size: 10000, Type: "LRU", Shards: 1,
			},
			ExportStateStorageConfig: config.StorageConfig{
				Cache: config.CacheConfig{
					Size: 10000, Type: "LRU", Shards: 1,
				},
				DB: config.DBConfig{
					FilePath:          "ExportState" + fmt.Sprintf("%d", id),
					Type:              "LvlDBSerial",
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
			WhiteListHandler:      node.WhiteListHandler,
			InterceptorsContainer: node.InterceptorsContainer,
			ExistingResolvers:     node.ResolversContainer,
		}

		exportHandler, err := factory.NewExportHandlerFactory(argsExportHandler)
		assert.Nil(t, err)
		assert.NotNil(t, exportHandler)

		node.ExportHandler, err = exportHandler.Create()
		assert.Nil(t, err)
		assert.NotNil(t, node.ExportHandler)
	}
}

func verifyIfNodesHasCorrectEpoch(
	t *testing.T,
	epoch uint32,
	nodes []*integrationTests.TestProcessorNode,
) {
	for _, node := range nodes {
		currentShId := node.ShardCoordinator.SelfId()
		currentHeader := node.BlockChain.GetCurrentBlockHeader()
		assert.Equal(t, epoch, currentHeader.GetEpoch())

		for _, testNode := range nodes {
			if testNode.ShardCoordinator.SelfId() == currentShId {
				testHeader := testNode.BlockChain.GetCurrentBlockHeader()
				assert.Equal(t, testHeader.GetNonce(), currentHeader.GetNonce())
			}
		}
	}
}

func verifyIfAddedShardHeadersAreWithNewEpoch(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
) {
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != sharding.MetachainShardId {
			continue
		}

		currentMetaHdr, ok := node.BlockChain.GetCurrentBlockHeader().(*block.MetaBlock)
		if !ok {
			assert.Fail(t, "metablock should have been in current block header")
		}

		shardHDrStorage := node.Storage.GetStorer(dataRetriever.BlockHeaderUnit)
		for _, shardInfo := range currentMetaHdr.ShardInfo {
			value, err := node.DataPool.Headers().GetHeaderByHash(shardInfo.HeaderHash)
			if err == nil {
				header, ok := value.(data.HeaderHandler)
				if !ok {
					assert.Fail(t, "wrong type in shard header pool")
				}

				assert.Equal(t, header.GetEpoch(), currentMetaHdr.GetEpoch())
				continue
			}

			buff, err := shardHDrStorage.Get(shardInfo.HeaderHash)
			assert.Nil(t, err)

			shardHeader := block.Header{}
			err = integrationTests.TestMarshalizer.Unmarshal(&shardHeader, buff)
			assert.Nil(t, err)
			assert.Equal(t, shardHeader.Epoch, currentMetaHdr.Epoch)
		}
	}
}
