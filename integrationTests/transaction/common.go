package transaction

import (
	"context"
	"math/rand"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
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
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/singlesig"
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
	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})
	hdrPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})

	cacherCfg := storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	hdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	txBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	peerChangeBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
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

func createDummyHexAddress(chars int) string {
	if chars < 1 {
		return ""
	}

	var characters = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

	rdm := rand.New(rand.NewSource(time.Now().UnixNano()))

	buff := make([]byte, chars)
	for i := 0; i < chars; i++ {
		buff[i] = characters[rdm.Int()%16]
	}

	return string(buff)
}

func adbCreateAccountsDB() *state.AccountsDB {
	marsh := &marshal.JsonMarshalizer{}

	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marsh)

	return adb
}

func createMemNode(port int, dPool data.TransientDataHolder, accntAdapter state.AccountsAdapter) (
	*node.Node,
	p2p.Messenger,
	crypto.PrivateKey,
	process.ProcessorFactory) {

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	cp, _ := p2p.NewConnectParamsFromPort(port)
	mes, _ := p2p.NewMemMessenger(marshalizer, hasher, cp)

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")

	suite := kv2.NewBlakeSHA256Ed25519()
	singlesigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()
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
		SingleSigner:             singlesigner,
		KeyGen:                   keyGen,
		Uint64ByteSliceConverter: uint64Converter,
	})

	n, _ := node.NewNode(
		node.WithMessenger(mes),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithContext(context.Background()),
		node.WithDataPool(dPool),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGenerator(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blockChain),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithSinglesig(singlesigner),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
		node.WithProcessorCreator(pFactory),
	)

	_ = pFactory.CreateInterceptors()
	_ = pFactory.CreateResolvers()

	return n, mes, sk, pFactory
}

func createNetNode(port int, dPool data.TransientDataHolder, accntAdapter state.AccountsAdapter) (
	*node.Node,
	p2p.Messenger,
	crypto.PrivateKey) {

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	messenger := createMessenger(context.Background(), marshalizer, hasher, 4, port)

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")

	suite := kv2.NewBlakeSHA256Ed25519()
	singlesigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()
	blkc := createTestBlockChain()
	shardCoordinator := &sharding.OneShardCoordinator{}
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	pFactory, _ := factory.NewProcessorsCreator(factory.ProcessorsCreatorConfig{
		InterceptorContainer:     interceptor.NewContainer(),
		ResolverContainer:        resolver.NewContainer(),
		Messenger:                messenger,
		Blockchain:               blkc,
		DataPool:                 dPool,
		ShardCoordinator:         shardCoordinator,
		AddrConverter:            addrConverter,
		Hasher:                   hasher,
		Marshalizer:              marshalizer,
		SingleSigner:             singlesigner,
		KeyGen:                   keyGen,
		Uint64ByteSliceConverter: uint64Converter,
	})

	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithContext(context.Background()),
		node.WithDataPool(dPool),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGenerator(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithSinglesig(singlesigner),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
		node.WithProcessorCreator(pFactory),
	)

	return n, nil, sk
}

func createMessenger(ctx context.Context, marshalizer marshal.Marshalizer, hasher hashing.Hasher, maxAllowedPeers int, port int) p2p.Messenger {
	cp := &p2p.ConnectParams{}
	cp.Port = port
	cp.GeneratePrivPubKeys(time.Now().UnixNano())
	cp.GenerateIDFromPubKey()

	nm, _ := p2p.NewNetMessenger(ctx, marshalizer, hasher, cp, maxAllowedPeers, p2p.GossipSub)
	return nm
}
