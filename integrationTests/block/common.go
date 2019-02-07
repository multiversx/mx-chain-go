package block

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/resolver"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
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

func createMemNode(port int, dPool data.TransientDataHolder) (*node.Node, p2p.Messenger, process.ProcessorFactory) {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	cp, _ := p2p.NewConnectParamsFromPort(port)
	mes, _ := p2p.NewMemMessenger(marshalizer, hasher, cp)

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")

	keyGen := schnorr.NewKeyGenerator()
	blockChain := createTestBlockChain()
	shardCoordinator := &sharding.OneShardCoordinator{}
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	pFactory, _ := factory.NewProcessorsCreator(factory.ProcessorsCreatorConfig{
		InterceptorContainer:     interceptor.NewContainer(),
		ResolverContainer:        resolver.NewContainer(),
		Messenger:                mes,
		Blockchain:               blockChain,
		DataPool:                 dPool,
		ShardCoordinator:         shardCoordinator,
		AddrConverter:            addrConverter,
		Hasher:                   hasher,
		Marshalizer:              marshalizer,
		SingleSignKeyGen:         keyGen,
		Uint64ByteSliceConverter: uint64Converter,
	})

	n, _ := node.NewNode(
		node.WithMessenger(mes),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithContext(context.Background()),
		node.WithDataPool(dPool),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blockChain),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMessenger(mes),
		node.WithProcessorCreator(pFactory),
	)

	_ = pFactory.CreateInterceptors()
	_ = pFactory.CreateResolvers()

	return n, mes, pFactory
}
