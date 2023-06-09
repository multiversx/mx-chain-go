package dataRetriever

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool/headersCache"
	"github.com/multiversx/mx-chain-go/dataRetriever/shardedData"
	"github.com/multiversx/mx-chain-go/dataRetriever/txpool"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/multiversx/mx-chain-go/trie/factory"
)

var peerAuthDuration = 10 * time.Second

func panicIfError(message string, err error) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", message, err))
	}
}

// CreateTxPool -
func CreateTxPool(numShards uint32, selfShard uint32) (dataRetriever.ShardedDataCacherNotifier, error) {
	return txpool.NewShardedTxPool(
		txpool.ArgShardedTxPool{
			Config: storageunit.CacheConfig{
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

	unsignedTxPool, err := shardedData.NewShardedData("unsignedTxPool", storageunit.CacheConfig{
		Capacity:    100000,
		SizeInBytes: 1000000000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolder", err)

	rewardsTxPool, err := shardedData.NewShardedData("rewardsTxPool", storageunit.CacheConfig{
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

	cacherConfig := storageunit.CacheConfig{Capacity: 100000, Type: storageunit.LRUCache, Shards: 1}
	txBlockBody, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	cacherConfig = storageunit.CacheConfig{Capacity: 100000, Type: storageunit.LRUCache, Shards: 1}
	peerChangeBlockBody, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	cacherConfig = storageunit.CacheConfig{Capacity: 50000, Type: storageunit.LRUCache}
	cacher, err := cache.NewCapacityLRU(10, 10000)
	panicIfError("Create trieSync cacher", err)

	tempDir, _ := ioutil.TempDir("", "integrationTests")
	cfg := storageunit.ArgDB{
		Path:              tempDir,
		DBType:            storageunit.LvlDBSerial,
		BatchDelaySeconds: 4,
		MaxBatchSize:      10000,
		MaxOpenFiles:      10,
	}
	persister, err := storageunit.NewDB(cfg)
	panicIfError("Create trieSync DB", err)
	tnf := factory.NewTrieNodeFactory()

	adaptedTrieNodesStorage, err := storageunit.NewStorageCacherAdapter(
		cacher,
		persister,
		tnf,
		&marshal.GogoProtoMarshalizer{},
	)
	panicIfError("Create AdaptedTrieNodesStorage", err)

	trieNodesChunks, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	cacherConfig = storageunit.CacheConfig{Capacity: 50000, Type: storageunit.LRUCache}
	smartContracts, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	peerAuthPool, err := cache.NewTimeCacher(cache.ArgTimeCacher{
		DefaultSpan: 60 * time.Second,
		CacheExpiry: 60 * time.Second,
	})
	panicIfError("CreatePoolsHolder", err)

	cacherConfig = storageunit.CacheConfig{Capacity: 50000, Type: storageunit.LRUCache}
	heartbeatPool, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	validatorsInfo, err := shardedData.NewShardedData("validatorsInfoPool", storageunit.CacheConfig{
		Capacity:    300,
		SizeInBytes: 300000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolder", err)

	currentBlockTransactions := dataPool.NewCurrentBlockTransactionsPool()
	currentEpochValidatorInfo := dataPool.NewCurrentEpochValidatorInfoPool()
	dataPoolArgs := dataPool.DataPoolArgs{
		Transactions:              txPool,
		UnsignedTransactions:      unsignedTxPool,
		RewardTransactions:        rewardsTxPool,
		Headers:                   headersPool,
		MiniBlocks:                txBlockBody,
		PeerChangesBlocks:         peerChangeBlockBody,
		TrieNodes:                 adaptedTrieNodesStorage,
		TrieNodesChunks:           trieNodesChunks,
		CurrentBlockTransactions:  currentBlockTransactions,
		CurrentEpochValidatorInfo: currentEpochValidatorInfo,
		SmartContracts:            smartContracts,
		PeerAuthentications:       peerAuthPool,
		Heartbeats:                heartbeatPool,
		ValidatorsInfo:            validatorsInfo,
	}
	holder, err := dataPool.NewDataPool(dataPoolArgs)
	panicIfError("CreatePoolsHolder", err)

	return holder
}

// CreatePoolsHolderWithTxPool -
func CreatePoolsHolderWithTxPool(txPool dataRetriever.ShardedDataCacherNotifier) dataRetriever.PoolsHolder {
	var err error

	unsignedTxPool, err := shardedData.NewShardedData("unsignedTxPool", storageunit.CacheConfig{
		Capacity:    100000,
		SizeInBytes: 1000000000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolderWithTxPool", err)

	rewardsTxPool, err := shardedData.NewShardedData("rewardsTxPool", storageunit.CacheConfig{
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

	cacherConfig := storageunit.CacheConfig{Capacity: 100000, Type: storageunit.LRUCache, Shards: 1}
	txBlockBody, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig = storageunit.CacheConfig{Capacity: 100000, Type: storageunit.LRUCache, Shards: 1}
	peerChangeBlockBody, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig = storageunit.CacheConfig{Capacity: 50000, Type: storageunit.LRUCache}
	trieNodes, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	trieNodesChunks, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig = storageunit.CacheConfig{Capacity: 50000, Type: storageunit.LRUCache}
	smartContracts, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	validatorsInfo, err := shardedData.NewShardedData("validatorsInfoPool", storageunit.CacheConfig{
		Capacity:    300,
		SizeInBytes: 300000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolderWithTxPool", err)

	peerAuthPool, err := cache.NewTimeCacher(cache.ArgTimeCacher{
		DefaultSpan: peerAuthDuration,
		CacheExpiry: peerAuthDuration,
	})
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig = storageunit.CacheConfig{Capacity: 50000, Type: storageunit.LRUCache}
	heartbeatPool, err := storageunit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	currentBlockTransactions := dataPool.NewCurrentBlockTransactionsPool()
	currentEpochValidatorInfo := dataPool.NewCurrentEpochValidatorInfoPool()
	dataPoolArgs := dataPool.DataPoolArgs{
		Transactions:              txPool,
		UnsignedTransactions:      unsignedTxPool,
		RewardTransactions:        rewardsTxPool,
		Headers:                   headersPool,
		MiniBlocks:                txBlockBody,
		PeerChangesBlocks:         peerChangeBlockBody,
		TrieNodes:                 trieNodes,
		TrieNodesChunks:           trieNodesChunks,
		CurrentBlockTransactions:  currentBlockTransactions,
		CurrentEpochValidatorInfo: currentEpochValidatorInfo,
		SmartContracts:            smartContracts,
		PeerAuthentications:       peerAuthPool,
		Heartbeats:                heartbeatPool,
		ValidatorsInfo:            validatorsInfo,
	}
	holder, err := dataPool.NewDataPool(dataPoolArgs)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	return holder
}
