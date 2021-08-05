package dataRetriever

import (
	"fmt"
	"io/ioutil"

	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool/headersCache"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache/capacity"
	"github.com/ElrondNetwork/elrond-go/storage/storageCacherAdapter"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon/txcachemocks"
	"github.com/ElrondNetwork/elrond-go/trie/factory"
)

func panicIfError(message string, err error) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", message, err))
	}
}

// CreateTxPool -
func CreateTxPool(numShards uint32, selfShard uint32) (dataRetriever.ShardedDataCacherNotifier, error) {
	return txpool.NewShardedTxPool(
		txpool.ArgShardedTxPool{
			Config: storageUnit.CacheConfig{
				Capacity:             100_000,
				SizePerSender:        1_000_000_000,
				SizeInBytes:          1_000_000_000,
				SizeInBytesPerSender: 33_554_432,
				Shards:               16,
			},
			NumberOfShards: numShards,
			SelfShardID:    selfShard,
			TxGasHandler: &txcachemocks.TxGasHandlerMock{
				MinimumGasMove:       50000,
				MinimumGasPrice:      200000000000,
				GasProcessingDivisor: 100,
			},
		},
	)
}

// CreatePoolsHolder -
func CreatePoolsHolder(numShards uint32, selfShard uint32) dataRetriever.PoolsHolder {
	var err error

	txPool, err := CreateTxPool(numShards, selfShard)
	panicIfError("CreatePoolsHolder", err)

	unsignedTxPool, err := shardedData.NewShardedData("unsignedTxPool", storageUnit.CacheConfig{
		Capacity:    100000,
		SizeInBytes: 1000000000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolder", err)

	rewardsTxPool, err := shardedData.NewShardedData("rewardsTxPool", storageUnit.CacheConfig{
		Capacity:    300,
		SizeInBytes: 300000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolder", err)

	headersPool, err := headersCache.NewHeadersPool(config.HeadersPoolConfig{
		MaxHeadersPerShard:            1000,
		NumElementsToRemoveOnEviction: 100,
	})
	panicIfError("CreatePoolsHolder", err)

	cacherConfig := storageUnit.CacheConfig{Capacity: 100000, Type: storageUnit.LRUCache, Shards: 1}
	txBlockBody, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 100000, Type: storageUnit.LRUCache, Shards: 1}
	peerChangeBlockBody, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 50000, Type: storageUnit.LRUCache}
	cacher, err := capacity.NewCapacityLRU(10, 10000)
	panicIfError("Create trieSync cacher", err)

	tempDir, _ := ioutil.TempDir("", "integrationTests")
	cfg := storageUnit.ArgDB{
		Path:              tempDir,
		DBType:            storageUnit.LvlDBSerial,
		BatchDelaySeconds: 4,
		MaxBatchSize:      10000,
		MaxOpenFiles:      10,
	}
	persister, err := storageUnit.NewDB(cfg)
	panicIfError("Create trieSync DB", err)
	tnf := factory.NewTrieNodeFactory()

	adaptedTrieNodesStorage, err := storageCacherAdapter.NewStorageCacherAdapter(
		cacher,
		persister,
		tnf,
		&marshal.GogoProtoMarshalizer{},
	)
	panicIfError("Create AdaptedTrieNodesStorage", err)

	trieNodesChunks, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 50000, Type: storageUnit.LRUCache}
	smartContracts, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	currentTx, err := dataPool.NewCurrentBlockPool()
	panicIfError("CreatePoolsHolder", err)

	dataPoolArgs := dataPool.DataPoolArgs{
		Transactions:             txPool,
		UnsignedTransactions:     unsignedTxPool,
		RewardTransactions:       rewardsTxPool,
		Headers:                  headersPool,
		MiniBlocks:               txBlockBody,
		PeerChangesBlocks:        peerChangeBlockBody,
		TrieNodes:                adaptedTrieNodesStorage,
		TrieNodesChunks:          trieNodesChunks,
		CurrentBlockTransactions: currentTx,
		SmartContracts:           smartContracts,
	}
	holder, err := dataPool.NewDataPool(dataPoolArgs)
	panicIfError("CreatePoolsHolder", err)

	return holder
}

// CreatePoolsHolderWithTxPool -
func CreatePoolsHolderWithTxPool(txPool dataRetriever.ShardedDataCacherNotifier) dataRetriever.PoolsHolder {
	var err error

	unsignedTxPool, err := shardedData.NewShardedData("unsignedTxPool", storageUnit.CacheConfig{
		Capacity:    100000,
		SizeInBytes: 1000000000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolderWithTxPool", err)

	rewardsTxPool, err := shardedData.NewShardedData("rewardsTxPool", storageUnit.CacheConfig{
		Capacity:    300,
		SizeInBytes: 300000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolderWithTxPool", err)

	headersPool, err := headersCache.NewHeadersPool(config.HeadersPoolConfig{
		MaxHeadersPerShard:            1000,
		NumElementsToRemoveOnEviction: 100,
	})
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig := storageUnit.CacheConfig{Capacity: 100000, Type: storageUnit.LRUCache, Shards: 1}
	txBlockBody, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 100000, Type: storageUnit.LRUCache, Shards: 1}
	peerChangeBlockBody, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 50000, Type: storageUnit.LRUCache}
	trieNodes, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	trieNodesChunks, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 50000, Type: storageUnit.LRUCache}
	smartContracts, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	currentTx, err := dataPool.NewCurrentBlockPool()
	panicIfError("CreatePoolsHolderWithTxPool", err)

	dataPoolArgs := dataPool.DataPoolArgs{
		Transactions:             txPool,
		UnsignedTransactions:     unsignedTxPool,
		RewardTransactions:       rewardsTxPool,
		Headers:                  headersPool,
		MiniBlocks:               txBlockBody,
		PeerChangesBlocks:        peerChangeBlockBody,
		TrieNodes:                trieNodes,
		TrieNodesChunks:          trieNodesChunks,
		CurrentBlockTransactions: currentTx,
		SmartContracts:           smartContracts,
	}
	holder, err := dataPool.NewDataPool(dataPoolArgs)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	return holder
}
