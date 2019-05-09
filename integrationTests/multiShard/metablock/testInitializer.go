package metablock

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"

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
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/multiShard/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	metaProcess "github.com/ElrondNetwork/elrond-go-sandbox/process/factory/metachain"
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
var testMultiSig = mock.NewMultiSigner()

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type metachainNode struct {
	node         *node.Node
	mesenger     p2p.Messenger
	accntState   state.AccountsAdapter
	blkc         data.ChainHandler
	blkProcessor process.BlockProcessor
	sk           crypto.PrivateKey
	pk           crypto.PublicKey
	dPool        dataRetriever.MetaPoolsHolder
	resFinder    dataRetriever.ResolversFinder
}

func createTestMetaChain() data.ChainHandler {
	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size)
	metaChain, _ := blockchain.NewMetaChain(
		badBlockCache,
	)
	metaChain.GenesisBlock = &dataBlock.MetaBlock{}

	return metaChain
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func createTestStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaPeerDataUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaShardDataUnit, createMemUnit())

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

func createNetNode(
	dPool dataRetriever.MetaPoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	targetShardId uint32,
	initialAddr string,
) (
	*node.Node,
	p2p.Messenger,
	crypto.PrivateKey,
	dataRetriever.ResolversFinder,
	process.BlockProcessor,
	data.ChainHandler) {

	messenger := createMessengerWithKadDht(context.Background(), initialAddr)
	suite := kyber.NewBlakeSHA256Ed25519()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	for {
		pkBytes, _ := pk.ToByteArray()
		addr, _ := testAddressConverter.CreateAddressFromPublicKeyBytes(pkBytes)
		if shardCoordinator.ComputeId(addr) == targetShardId {
			break
		}
		sk, pk = keyGen.GeneratePair()
	}

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s\n", hex.EncodeToString(pkBuff))

	blkc := createTestMetaChain()
	store := createTestStore()
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	interceptorContainerFactory, _ := metaProcess.NewInterceptorsContainerFactory(
		shardCoordinator,
		messenger,
		store,
		testMarshalizer,
		testHasher,
		testMultiSig,
		dPool,
		&mock.ChronologyValidatorMock{},
	)
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		fmt.Println(err.Error())
	}

	resolversContainerFactory, _ := metafactoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		messenger,
		store,
		testMarshalizer,
		dPool,
		uint64Converter,
	)
	resolversContainer, _ := resolversContainerFactory.Create()
	resolversFinder, _ := containers.NewResolversFinder(resolversContainer, shardCoordinator)

	blockProcessor, _ := block.NewMetaProcessor(
		accntAdapter,
		dPool,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, isProcessed bool) error {
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
		func(shardId uint32, hdrHash []byte) {
			//TODO(js) fix this
			//resolver, err := resolversFinder.CrossShardResolver(factory.TransactionTopic, destShardID)
			//if err != nil {
			//	fmt.Println(err.Error())
			//	return
			//}
			//
			//err = resolver.RequestDataFromHash(txHash)
			//if err != nil {
			//	fmt.Println(err.Error())
			//}
		},
	)

	n, err := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithMetaDataPool(dPool),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGenerator(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultisig(testMultiSig),
		node.WithSinglesig(singleSigner),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataStore(store),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return n, messenger, sk, resolversFinder, blockProcessor, blkc
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

func makeDisplayTable(nodes []*metachainNode) string {
	header := []string{"pk", "miniblocks", "headers", "metachain headers", "connections"}
	dataLines := make([]*display.LineData, len(nodes))
	for idx, n := range nodes {
		buffPk, _ := n.pk.ToByteArray()

		dataLines[idx] = display.NewLineData(
			false,
			[]string{
				hex.EncodeToString(buffPk),
				fmt.Sprintf("%d", 0), //atomic.LoadInt32(&n.miniblocksRecv)),
				fmt.Sprintf("%d", 0), //atomic.LoadInt32(&n.headersRecv)),
				fmt.Sprintf("%d", 0), //atomic.LoadInt32(&n.metachainHdrRecv)),
				fmt.Sprintf("%d / %d",
					0,
					len(n.mesenger.ConnectedPeers())), //len(n.mesenger.ConnectedPeersOnTopic(factory.TransactionTopic+"_"+
				//fmt.Sprintf("%d", n.shardId))), ),
			},
		)
	}
	table, _ := display.CreateTableString(header, dataLines)
	return table
}

func displayAndStartNodes(nodes []*metachainNode) {
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
	numOfShards int,
	nodesPerShard int,
	serviceID string,
) []*metachainNode {

	//first node generated will have is pk belonging to firstSkShardId
	nodes := make([]*metachainNode, int(numOfShards)*nodesPerShard)

	idx := 0
	for shardId := 0; shardId < numOfShards; shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			testNode := &metachainNode{
				dPool: createTestMetaDataPool(),
			}

			shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), uint32(shardId))
			accntAdapter := createAccountsDB()
			n, mes, sk, resFinder, blkProcessor, blkc := createNetNode(
				testNode.dPool,
				accntAdapter,
				shardCoordinator,
				sharding.MetachainShardId,
				serviceID,
			)
			_ = n.CreateShardedStores()

			testNode.node = n
			testNode.sk = sk
			testNode.mesenger = mes
			testNode.pk = sk.GeneratePublic()
			testNode.resFinder = resFinder
			testNode.accntState = accntAdapter
			testNode.blkProcessor = blkProcessor
			testNode.blkc = blkc
			//TODO(jls) fix/remove this
			//testNode.dPool.Headers().RegisterHandler(func(key []byte) {
			//	atomic.AddInt32(&testNode.headersRecv, 1)
			//	testNode.mutHeaders.Lock()
			//	testNode.headersHashes = append(testNode.headersHashes, key)
			//	header, _ := testNode.dPool.Headers().Peek(key)
			//	testNode.headers = append(testNode.headers, header.(data.HeaderHandler))
			//	testNode.mutHeaders.Unlock()
			//})
			//testNode.dPool.MiniBlocks().RegisterHandler(func(key []byte) {
			//	atomic.AddInt32(&testNode.miniblocksRecv, 1)
			//	testNode.mutMiniblocks.Lock()
			//	testNode.miniblocksHashes = append(testNode.miniblocksHashes, key)
			//	miniblock, _ := testNode.dPool.MiniBlocks().Peek(key)
			//	testNode.miniblocks = append(testNode.miniblocks, miniblock.(*dataBlock.MiniBlock))
			//	testNode.mutMiniblocks.Unlock()
			//})
			//testNode.dPool.MetaBlocks().RegisterHandler(func(key []byte) {
			//	fmt.Printf("Got metachain header: %v\n", base64.StdEncoding.EncodeToString(key))
			//	atomic.AddInt32(&testNode.metachainHdrRecv, 1)
			//})
			//testNode.dPool.Transactions().RegisterHandler(func(key []byte) {
			//	atomic.AddInt32(&testNode.txsRecv, 1)
			//})

			nodes[idx] = testNode
			idx++
		}
	}

	return nodes
}

func equalSlices(slice1 [][]byte, slice2 [][]byte) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	//check slice1 has all elements in slice2
	for _, buff1 := range slice1 {
		found := false
		for _, buff2 := range slice2 {
			if bytes.Equal(buff1, buff2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	//check slice2 has all elements in slice1
	for _, buff2 := range slice2 {
		found := false
		for _, buff1 := range slice1 {
			if bytes.Equal(buff1, buff2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func uint32InSlice(searched uint32, list []uint32) bool {
	for _, val := range list {
		if val == searched {
			return true
		}
	}
	return false
}
