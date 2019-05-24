package metablock

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/splitters"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/containers"
	metafactoryDataRetriever "github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/metachain"
	shard2 "github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	metaProcess "github.com/ElrondNetwork/elrond-go-sandbox/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/shard"
	mock2 "github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
)

var r *rand.Rand
var testHasher = sha256.Sha256{}
var testMarshalizer = &marshal.JsonMarshalizer{}
var testAddressConverter, _ = addressConverters.NewPlainAddressConverter(32, "0x")
var testMultiSig = mock.NewMultiSigner(1)

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type testNode struct {
	node          *node.Node
	messenger     p2p.Messenger
	sk            crypto.PrivateKey
	pk            crypto.PublicKey
	shard         string
	shardDataPool dataRetriever.PoolsHolder
	metaDataPool  dataRetriever.MetaPoolsHolder
	resolvers     dataRetriever.ResolversFinder

	shardHdrRecv     int32
	metachainHdrRecv int32
}

//------- Common

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func createAccountsDB() *state.AccountsDB {
	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, testMarshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (wrapper state.AccountHandler, e error) {
			return state.NewAccount(address, tracker)
		},
	})
	return adb
}

func createMessengerWithKadDht(ctx context.Context, initialAddr string) p2p.Messenger {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto2.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessengerOnFreePort(
		ctx,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewKadDhtPeerDiscoverer(time.Second, "test", []string{initialAddr}),
	)
	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

func getConnectableAddress(mes p2p.Messenger) string {
	for _, addr := range mes.Addresses() {
		if strings.Contains(addr, "circuit") || strings.Contains(addr, "169.254") {
			continue
		}
		return addr
	}
	return ""
}

func makeDisplayTable(nodes []*testNode) string {
	header := []string{"pk", "shard", "headers", "metachain headers", "connections"}
	dataLines := make([]*display.LineData, len(nodes))
	for idx, tn := range nodes {
		buffPk, _ := tn.pk.ToByteArray()

		dataLines[idx] = display.NewLineData(
			false,
			[]string{
				hex.EncodeToString(buffPk),
				fmt.Sprintf("%s", tn.shard),
				fmt.Sprintf("%d", atomic.LoadInt32(&tn.shardHdrRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&tn.metachainHdrRecv)),
				fmt.Sprintf("%d / %d",
					len(tn.messenger.ConnectedPeersOnTopic(factory.MetachainBlocksTopic)),
					len(tn.messenger.ConnectedPeers())),
			},
		)
	}
	table, _ := display.CreateTableString(header, dataLines)
	return table
}

func displayAndStartNodes(nodes []*testNode) {
	for _, n := range nodes {
		skBuff, _ := n.sk.ToByteArray()
		pkBuff, _ := n.pk.ToByteArray()

		fmt.Printf("Metachain node: sk: %s, pk: %s\n",
			hex.EncodeToString(skBuff),
			hex.EncodeToString(pkBuff),
		)
		_ = n.node.Start()
		_ = n.node.P2PBootstrap()
	}
}

func createNodes(
	nodesInMetachain int,
	senderShard uint32,
	initialAddr string,
) []*testNode {

	nodes := make([]*testNode, nodesInMetachain+1)
	//first node is a shard node
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(1, senderShard)
	nodes[0] = createShardNetNode(
		createTestShardDataPool(),
		createAccountsDB(),
		shardCoordinator,
		initialAddr,
	)

	for i := 0; i < nodesInMetachain; i++ {
		shardCoordinator, _ = sharding.NewMultiShardCoordinator(1, sharding.MetachainShardId)
		nodes[i+1] = createMetaNetNode(
			createTestMetaDataPool(),
			createAccountsDB(),
			shardCoordinator,
			initialAddr,
		)
	}

	return nodes
}

//------- Shard

func createTestShardChain() *blockchain.BlockChain {
	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)
	blockChain.GenesisHeader = &dataBlock.Header{}

	return blockChain
}

func createTestShardStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())

	return store
}

func createTestShardDataPool() dataRetriever.PoolsHolder {
	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})
	cacherCfg := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	hdrPool, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	hdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	txBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	peerChangeBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	metaHdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	metaHdrNonces, _ := dataPool.NewNonceToHashCacher(metaHdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())
	metaBlocks, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 10, Type: storage.LRUCache}

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
		metaHdrNonces,
	)

	return dPool
}

func createShardNetNode(
	dPool dataRetriever.PoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	initialAddr string,
) *testNode {

	tn := testNode{}

	tn.messenger = createMessengerWithKadDht(context.Background(), initialAddr)
	suite := kyber.NewBlakeSHA256Ed25519()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s\n", hex.EncodeToString(pkBuff))

	blkc := createTestShardChain()
	store := createTestShardStore()
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, uint64(time.Second*4))
	addConverter, _ := addressConverters.NewPlainAddressConverter(32, "")
	sliceSplitter, _ := splitters.NewSliceSplitter(testMarshalizer)

	interceptorContainerFactory, _ := shard.NewInterceptorsContainerFactory(
		shardCoordinator,
		tn.messenger,
		store,
		testMarshalizer,
		testHasher,
		keyGen,
		singleSigner,
		testMultiSig,
		dPool,
		addConverter,
		&mock.ChronologyValidatorMock{},
		tpsBenchmark,
	)
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		fmt.Println(err.Error())
	}

	resolversContainerFactory, _ := shard2.NewResolversContainerFactory(
		shardCoordinator,
		tn.messenger,
		store,
		testMarshalizer,
		dPool,
		uint64Converter,
		sliceSplitter,
	)
	resolversContainer, _ := resolversContainerFactory.Create()
	tn.resolvers, _ = containers.NewResolversFinder(resolversContainer, shardCoordinator)

	blockProcessor, _ := block.NewShardProcessor(
		dPool,
		store,
		testHasher,
		testMarshalizer,
		&mock2.TxProcessorMock{},
		accntAdapter,
		shardCoordinator,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
				return nil
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		func(shardId uint32, txHash [][]byte) {

		},
		func(shardId uint32, miniblockHash []byte) {

		},
	)

	n, err := node.NewNode(
		node.WithMessenger(tn.messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithDataPool(dPool),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGen(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultiSigner(testMultiSig),
		node.WithSingleSigner(singleSigner),
		node.WithPrivKey(sk),
		node.WithPubKey(pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(tn.resolvers),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataStore(store),
	)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	tn.node = n
	tn.sk = sk
	tn.pk = pk
	tn.shard = fmt.Sprintf("%d", shardCoordinator.SelfId())
	tn.shardDataPool = dPool

	dPool.MetaBlocks().RegisterHandler(func(key []byte) {
		atomic.AddInt32(&tn.metachainHdrRecv, 1)
	})
	dPool.Headers().RegisterHandler(func(key []byte) {
		atomic.AddInt32(&tn.shardHdrRecv, 1)
	})

	return &tn
}

//------- Meta

func createTestMetaChain() data.ChainHandler {
	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size)
	metaChain, _ := blockchain.NewMetaChain(
		badBlockCache,
	)
	metaChain.GenesisBlock = &dataBlock.MetaBlock{}

	return metaChain
}

func createTestMetaStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaPeerDataUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaShardDataUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())

	return store
}

func createTestMetaDataPool() dataRetriever.MetaPoolsHolder {
	cacherCfg := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	metaBlocks, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 10000, Type: storage.LRUCache}
	miniblockHashes, _ := shardedData.NewShardedData(cacherCfg)

	cacherCfg = storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	shardHeaders, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	metaBlockNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	metaBlockNonces, _ := dataPool.NewNonceToHashCacher(metaBlockNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	dPool, _ := dataPool.NewMetaDataPool(
		metaBlocks,
		miniblockHashes,
		shardHeaders,
		metaBlockNonces,
	)

	return dPool
}

func createMetaNetNode(
	dPool dataRetriever.MetaPoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	initialAddr string,
) *testNode {

	tn := testNode{}

	tn.messenger = createMessengerWithKadDht(context.Background(), initialAddr)
	suite := kyber.NewBlakeSHA256Ed25519()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s\n", hex.EncodeToString(pkBuff))

	blkc := createTestMetaChain()
	store := createTestMetaStore()
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	interceptorContainerFactory, _ := metaProcess.NewInterceptorsContainerFactory(
		shardCoordinator,
		tn.messenger,
		store,
		testMarshalizer,
		testHasher,
		testMultiSig,
		dPool,
		&mock.ChronologyValidatorMock{},
		nil,
	)
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		fmt.Println(err.Error())
	}

	resolversContainerFactory, _ := metafactoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		tn.messenger,
		store,
		testMarshalizer,
		dPool,
		uint64Converter,
	)
	resolversContainer, _ := resolversContainerFactory.Create()
	tn.resolvers, _ = containers.NewResolversFinder(resolversContainer, shardCoordinator)

	blockProcessor, _ := block.NewMetaProcessor(
		accntAdapter,
		dPool,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
				return nil
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		shardCoordinator,
		testHasher,
		testMarshalizer,
		store,
		func(shardId uint32, hdrHash []byte) {},
	)

	n, err := node.NewNode(
		node.WithMessenger(tn.messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithMetaDataPool(dPool),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGen(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultiSigner(testMultiSig),
		node.WithSingleSigner(singleSigner),
		node.WithPrivKey(sk),
		node.WithPubKey(pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(tn.resolvers),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataStore(store),
	)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	tn.node = n
	tn.sk = sk
	tn.pk = pk
	tn.shard = "meta"
	tn.metaDataPool = dPool

	dPool.MetaChainBlocks().RegisterHandler(func(key []byte) {
		atomic.AddInt32(&tn.metachainHdrRecv, 1)
	})
	dPool.ShardHeaders().RegisterHandler(func(key []byte) {
		atomic.AddInt32(&tn.shardHdrRecv, 1)
	})

	return &tn
}
