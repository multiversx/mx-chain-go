package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
)

func createTestBlockChain() *blockchain.BlockChain {

	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}

	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size)

	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
		createMemUnit(),
		createMemUnit(),
		createMemUnit(),
		createMemUnit(),
		createMemUnit())

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func createTestDataPool() data.TransientDataHolder {
	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100, Type: storage.LRUCache})
	hdrPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100, Type: storage.LRUCache})

	cacherCfg := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	hdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	txBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	peerChangeBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	stateBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	dPool, _ := dataPool.NewDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		stateBlockBody,
	)

	return dPool
}
