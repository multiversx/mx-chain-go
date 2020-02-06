package hardFork

import (
	"context"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update/factory"
	"github.com/stretchr/testify/assert"
)

func TestEpochStartChangeWithoutTransactionInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 5
	/////////----- wait for epoch end period
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, int(roundsPerEpoch), nonce, round, idxProposers)

	time.Sleep(time.Second)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	time.Sleep(time.Second)

	epoch := uint32(1)
	verifyIfNodesHasCorrectEpoch(t, epoch, nodes)
	verifyIfAddedShardHeadersAreWithNewEpoch(t, nodes)

	createHardForkExporter(t, nodes)
}

func createHardForkExporter(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
) {
	for _, node := range nodes {
		argsExportHandler := factory.ArgsExporter{
			Marshalizer:              integrationTests.TestMarshalizer,
			Hasher:                   integrationTests.TestHasher,
			HeaderValidator:          node.HeaderValidator,
			Uint64Converter:          integrationTests.TestUint64Converter,
			DataPool:                 node.DataPool,
			StorageService:           node.Storage,
			RequestHandler:           node.RequestHandler,
			ShardCoordinator:         node.ShardCoordinator,
			Messenger:                node.Messenger,
			ActiveTries:              node.TrieContainer,
			ExportFolder:             "export",
			ExportTriesStorageConfig: config.StorageConfig{
				Cache: config.CacheConfig{},
				DB:    config.DBConfig{},
				Bloom: config.BloomFilterConfig{},
			},
			ExportTriesCacheConfig: config.CacheConfig{
				Size: 10000, Type: storageUnit.LRUCache, Shards: 1
			},
			ExportStateStorageConfig: config.StorageConfig{
				Cache: config.CacheConfig{},
				DB:    config.DBConfig{},
				Bloom: config.BloomFilterConfig{},
			},
			WhiteListHandler:         node.WhiteListHandler,
			InterceptorsContainer:    node.InterceptorsContainer,
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
