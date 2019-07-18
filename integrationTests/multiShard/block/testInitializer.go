package block

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	metafactoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/metachain"
	factoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
)

var r *rand.Rand
var testHasher = sha256.Sha256{}
var testMarshalizer = &marshal.JsonMarshalizer{}
var testAddressConverter, _ = addressConverters.NewPlainAddressConverter(32, "0x")
var testMultiSig = mock.NewMultiSigner(1)
var rootHash = []byte("root hash")

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type testNode struct {
	node               *node.Node
	messenger          p2p.Messenger
	shardId            uint32
	accntState         state.AccountsAdapter
	blkc               data.ChainHandler
	blkProcessor       process.BlockProcessor
	txProcessor        process.TransactionProcessor
	broadcastMessenger consensus.BroadcastMessenger
	sk                 crypto.PrivateKey
	pk                 crypto.PublicKey
	dPool              dataRetriever.PoolsHolder
	resFinder          dataRetriever.ResolversFinder
	headersRecv        int32
	miniblocksRecv     int32
	mutHeaders         sync.Mutex
	headersHashes      [][]byte
	headers            []data.HeaderHandler
	mutMiniblocks      sync.Mutex
	miniblocksHashes   [][]byte
	miniblocks         []*dataBlock.MiniBlock
	metachainHdrRecv   int32
	txsRecv            int32
}

func createTestShardChain() *blockchain.BlockChain {
	cfgCache := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	badBlockCache, _ := storageUnit.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)
	blockChain.GenesisHeader = &dataBlock.Header{}
	genisisHeaderM, _ := testMarshalizer.Marshal(blockChain.GenesisHeader)

	blockChain.SetGenesisHeaderHash(testHasher.Compute(string(genisisHeaderM)))

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	persist, _ := memorydb.New()

	unit, _ := storageUnit.NewStorageUnit(cache, persist)
	return unit
}

func createTestShardStore(numOfShards uint32) dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, createMemUnit())

	for i := uint32(0); i < numOfShards; i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, createMemUnit())
	}

	return store
}

func createTestShardDataPool() dataRetriever.PoolsHolder {
	txPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	uTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	hdrPool, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	hdrNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	hdrNonces, _ := dataPool.NewNonceSyncMapCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	txBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	peerChangeBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	metaBlocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		uTxPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
	)

	return dPool
}

func createAccountsDB() *state.AccountsDB {
	hasher := sha256.Sha256{}
	store := createMemUnit()

	tr, _ := trie.NewTrie(store, testMarshalizer, hasher)
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, testMarshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (wrapper state.AccountHandler, e error) {
			return state.NewAccount(address, tracker)
		},
	})
	return adb
}

func createNetNode(
	dPool dataRetriever.PoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	targetShardId uint32,
	initialAddr string,
) (
	*node.Node,
	p2p.Messenger,
	crypto.PrivateKey,
	crypto.PublicKey,
	dataRetriever.ResolversFinder,
	process.BlockProcessor,
	process.TransactionProcessor,
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

	blkc := createTestShardChain()
	store := createTestShardStore(shardCoordinator.NumberOfShards())
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	dataPacker, _ := partitioning.NewSizeDataPacker(testMarshalizer)

	interceptorContainerFactory, _ := shard.NewInterceptorsContainerFactory(
		shardCoordinator,
		nodesCoordinator,
		messenger,
		store,
		testMarshalizer,
		testHasher,
		keyGen,
		singleSigner,
		testMultiSig,
		dPool,
		testAddressConverter,
	)
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		fmt.Println(err.Error())
	}

	resolversContainerFactory, _ := factoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		messenger,
		store,
		testMarshalizer,
		dPool,
		uint64Converter,
		dataPacker,
	)
	resolversContainer, _ := resolversContainerFactory.Create()
	resolversFinder, _ := containers.NewResolversFinder(resolversContainer, shardCoordinator)
	requestHandler, _ := requestHandlers.NewShardResolverRequestHandler(resolversFinder, factory.TransactionTopic, factory.UnsignedTransactionTopic, factory.MiniBlocksTopic, factory.MetachainBlocksTopic, 100)

	txProcessor, _ := transaction.NewTxProcessor(
		accntAdapter,
		testHasher,
		testAddressConverter,
		testMarshalizer,
		shardCoordinator,
		&mock.SCProcessorMock{},
	)

	fact, _ := shard.NewPreProcessorsContainerFactory(
		shardCoordinator,
		store,
		testMarshalizer,
		testHasher,
		dPool,
		testAddressConverter,
		accntAdapter,
		requestHandler,
		txProcessor,
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
	)
	container, _ := fact.Create()

	tc, _ := coordinator.NewTransactionCoordinator(
		shardCoordinator,
		accntAdapter,
		dPool,
		requestHandler,
		container,
		&mock.InterimProcessorContainerMock{},
	)

	genesisBlocks := createGenesisBlocks(shardCoordinator)
	blockProcessor, _ := block.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dPool,
		store,
		testHasher,
		testMarshalizer,
		accntAdapter,
		shardCoordinator,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
				return nil
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlocksTrackerMock{
			AddBlockCalled: func(headerHandler data.HeaderHandler) {
			},
			RemoveNotarisedBlocksCalled: func(headerHandler data.HeaderHandler) error {
				return nil
			},
			UnnotarisedBlocksCalled: func() []data.HeaderHandler {
				return make([]data.HeaderHandler, 0)
			},
		},
		genesisBlocks,
		true,
		requestHandler,
		tc,
		uint64Converter,
	)

	_ = blkc.SetGenesisHeader(genesisBlocks[shardCoordinator.SelfId()])

	n, err := node.NewNode(
		node.WithMessenger(messenger),
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
		node.WithTxSignPrivKey(sk),
		node.WithTxSignPubKey(pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataStore(store),
		node.WithSyncer(&mock.SyncTimerMock{}),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return n, messenger, sk, pk, resolversFinder, blockProcessor, txProcessor, blkc
}

func createMessengerWithKadDht(ctx context.Context, initialAddr string) p2p.Messenger {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)

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

func makeDisplayTable(nodes map[uint32][]*testNode) string {
	header := []string{"pk", "shard ID", "txs", "miniblocks", "headers", "metachain headers", "connections"}
	dataLines := make([]*display.LineData, 0)

	for _, nList := range nodes {
		for _, n := range nList {
			buffPk, _ := n.pk.ToByteArray()

			dataLine := display.NewLineData(
				false,
				[]string{
					hex.EncodeToString(buffPk),
					fmt.Sprintf("%d", n.shardId),
					fmt.Sprintf("%d", atomic.LoadInt32(&n.txsRecv)),
					fmt.Sprintf("%d", atomic.LoadInt32(&n.miniblocksRecv)),
					fmt.Sprintf("%d", atomic.LoadInt32(&n.headersRecv)),
					fmt.Sprintf("%d", atomic.LoadInt32(&n.metachainHdrRecv)),
					fmt.Sprintf("%d / %d", len(n.messenger.ConnectedPeersOnTopic(factory.TransactionTopic+"_"+
						fmt.Sprintf("%d", n.shardId))), len(n.messenger.ConnectedPeers())),
				},
			)

			dataLines = append(dataLines, dataLine)
		}
	}
	table, _ := display.CreateTableString(header, dataLines)
	return table
}

func displayAndStartNodes(nodes map[uint32][]*testNode) {
	for _, nodeList := range nodes {

		for _, n := range nodeList {
			skBuff, _ := n.sk.ToByteArray()
			pkBuff, _ := n.pk.ToByteArray()

			fmt.Printf("Shard ID: %v, sk: %s, pk: %s\n",
				n.shardId,
				hex.EncodeToString(skBuff),
				hex.EncodeToString(pkBuff),
			)
			_ = n.node.Start()
			_ = n.node.P2PBootstrap()
		}
	}
}

func genValidatorsFromPubKeys(pubKeysMap map[uint32][]string) map[uint32][]sharding.Validator {
	validatorsMap := make(map[uint32][]sharding.Validator)

	for shardId, shardNodesPks := range pubKeysMap {
		shardValidators := make([]sharding.Validator, 0)
		for i := 0; i < len(shardNodesPks); i++ {
			v, _ := consensus.NewValidator(big.NewInt(0), 1, []byte(shardNodesPks[i]))
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

func createNodes(
	numOfShards int,
	nodesPerShard int,
	serviceID string,
) map[uint32][]*testNode {

	//first node generated will have is pk belonging to firstSkShardId
	numMetaChainNodes := 1
	nodes := make(map[uint32][]*testNode)
	nodesCoordinators := make(map[uint32][]sharding.NodesCoordinator)
	nodesPublicKeys := make(map[uint32][]string)

	for shardId := 0; shardId < numOfShards; shardId++ {
		shardNodesCoordinators := make([]sharding.NodesCoordinator, 0)
		shardPubKeys := make([]string, 0)
		shardNodes := make([]*testNode, nodesPerShard)

		for j := 0; j < nodesPerShard; j++ {
			testNode := &testNode{
				dPool:   createTestShardDataPool(),
				shardId: uint32(shardId),
			}

			shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), uint32(shardId))
			nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(
				1,
				testHasher,
				uint32(shardId),
				uint32(numOfShards),
			)
			shardNodesCoordinators = append(shardNodesCoordinators, nodesCoordinator)

			accntAdapter := createAccountsDB()
			n, mes, sk, pk, resFinder, blkProcessor, txProcessor, blkc := createNetNode(
				testNode.dPool,
				accntAdapter,
				shardCoordinator,
				nodesCoordinator,
				testNode.shardId,
				serviceID,
			)
			pubKeyBytes, _ := pk.ToByteArray()
			shardPubKeys = append(shardPubKeys, string(pubKeyBytes))
			_ = n.CreateShardedStores()

			testNode.node = n
			testNode.sk = sk
			testNode.messenger = mes
			testNode.pk = sk.GeneratePublic()
			testNode.resFinder = resFinder
			testNode.accntState = accntAdapter
			testNode.blkProcessor = blkProcessor
			testNode.txProcessor = txProcessor
			testNode.blkc = blkc
			testNode.dPool.Headers().RegisterHandler(func(key []byte) {
				atomic.AddInt32(&testNode.headersRecv, 1)
				testNode.mutHeaders.Lock()
				testNode.headersHashes = append(testNode.headersHashes, key)
				header, _ := testNode.dPool.Headers().Peek(key)
				testNode.headers = append(testNode.headers, header.(data.HeaderHandler))
				testNode.mutHeaders.Unlock()
			})
			testNode.dPool.MiniBlocks().RegisterHandler(func(key []byte) {
				atomic.AddInt32(&testNode.miniblocksRecv, 1)
				testNode.mutMiniblocks.Lock()
				testNode.miniblocksHashes = append(testNode.miniblocksHashes, key)
				miniblock, _ := testNode.dPool.MiniBlocks().Peek(key)
				testNode.miniblocks = append(testNode.miniblocks, miniblock.(*dataBlock.MiniBlock))
				testNode.mutMiniblocks.Unlock()
			})
			testNode.dPool.MetaBlocks().RegisterHandler(func(key []byte) {
				fmt.Printf("Got metachain header: %v\n", base64.StdEncoding.EncodeToString(key))
				atomic.AddInt32(&testNode.metachainHdrRecv, 1)
			})
			testNode.dPool.Transactions().RegisterHandler(func(key []byte) {
				atomic.AddInt32(&testNode.txsRecv, 1)
			})
			testNode.broadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
				testMarshalizer,
				mes,
				shardCoordinator,
				sk,
				&singlesig.SchnorrSigner{},
			)

			shardNodes[j] = testNode
		}

		nodes[uint32(shardId)] = shardNodes
		nodesCoordinators[uint32(shardId)] = shardNodesCoordinators
		nodesPublicKeys[uint32(shardId)] = shardPubKeys
	}

	metaNodesCoordinators := make([]sharding.NodesCoordinator, 0)
	metaNodesPubKeys := make([]string, 0)

	metaNodes := make([]*testNode, numMetaChainNodes)
	for i := 0; i < numMetaChainNodes; i++ {
		shardCoordinatorMeta, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), sharding.MetachainShardId)
		nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(
			1,
			testHasher,
			sharding.MetachainShardId,
			uint32(numOfShards),
		)

		metaNodes[i] = createMetaNetNode(
			createTestMetaDataPool(),
			createAccountsDB(),
			shardCoordinatorMeta,
			nodesCoordinator,
			serviceID,
		)

		metaNodesCoordinators = append(metaNodesCoordinators, nodesCoordinator)
		pkBytes, _ := metaNodes[i].pk.ToByteArray()
		metaNodesPubKeys = append(metaNodesPubKeys, string(pkBytes))
	}

	nodes[sharding.MetachainShardId] = metaNodes
	nodesCoordinators[sharding.MetachainShardId] = metaNodesCoordinators
	nodesPublicKeys[sharding.MetachainShardId] = metaNodesPubKeys
	mapValidators := genValidatorsFromPubKeys(nodesPublicKeys)

	for _, shardCoord := range nodesCoordinators {
		for j := 0; j < len(shardCoord); j++ {
			shardCoord[j].LoadNodesPerShards(mapValidators)
		}
	}

	return nodes
}

func getMiniBlocksHashesFromShardIds(body dataBlock.Body, shardIds ...uint32) [][]byte {
	hashes := make([][]byte, 0)

	for _, miniblock := range body {
		for _, shardId := range shardIds {
			if miniblock.ReceiverShardID == shardId {
				buff, _ := testMarshalizer.Marshal(miniblock)
				hashes = append(hashes, testHasher.Compute(string(buff)))
			}
		}
	}

	return hashes
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

func generatePrivateKeyInShardId(
	coordinator sharding.Coordinator,
	shardId uint32,
) crypto.PrivateKey {

	suite := kyber.NewBlakeSHA256Ed25519()
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	for {
		buff, _ := pk.ToByteArray()
		addr, _ := testAddressConverter.CreateAddressFromPublicKeyBytes(buff)

		if coordinator.ComputeId(addr) == shardId {
			return sk
		}

		sk, pk = keyGen.GeneratePair()
	}
}

func createTestMetaChain() data.ChainHandler {
	cfgCache := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	badBlockCache, _ := storageUnit.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)
	metaChain, _ := blockchain.NewMetaChain(
		badBlockCache,
	)
	metaChain.GenesisBlock = &dataBlock.MetaBlock{}

	return metaChain
}

func createTestMetaStore(coordinator sharding.Coordinator) dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())
	for i := uint32(0); i < coordinator.NumberOfShards(); i++ {
		store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(i), createMemUnit())
	}

	return store
}

func createTestMetaDataPool() dataRetriever.MetaPoolsHolder {
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	metaBlocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache}
	miniblockHashes, _ := shardedData.NewShardedData(cacherCfg)

	cacherCfg = storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	shardHeaders, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	shardHeadersNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	shardHeadersNonces, _ := dataPool.NewNonceSyncMapCacher(shardHeadersNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	dPool, _ := dataPool.NewMetaDataPool(
		metaBlocks,
		miniblockHashes,
		shardHeaders,
		shardHeadersNonces,
	)

	return dPool
}

func createMetaNetNode(
	dPool dataRetriever.MetaPoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
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

	tn.blkc = createTestMetaChain()
	store := createTestMetaStore(shardCoordinator)
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	interceptorContainerFactory, _ := metaProcess.NewInterceptorsContainerFactory(
		shardCoordinator,
		nodesCoordinator,
		tn.messenger,
		store,
		testMarshalizer,
		testHasher,
		testMultiSig,
		dPool,
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
	resolvers, _ := containers.NewResolversFinder(resolversContainer, shardCoordinator)

	requestHandler, _ := requestHandlers.NewMetaResolverRequestHandler(resolvers, factory.ShardHeadersForMetachainTopic)

	genesisBlocks := createGenesisBlocks(shardCoordinator)
	blkProc, _ := block.NewMetaProcessor(
		&mock.ServiceContainerMock{},
		accntAdapter,
		dPool,
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
				return nil
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
		},
		shardCoordinator,
		testHasher,
		testMarshalizer,
		store,
		genesisBlocks,
		requestHandler,
	)

	_ = tn.blkc.SetGenesisHeader(genesisBlocks[sharding.MetachainShardId])

	tn.blkProcessor = blkProc

	tn.broadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		testMarshalizer,
		tn.messenger,
		shardCoordinator,
		sk,
		singleSigner,
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
		node.WithBlockChain(tn.blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultiSigner(testMultiSig),
		node.WithSingleSigner(singleSigner),
		node.WithPrivKey(sk),
		node.WithPubKey(pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolvers),
		node.WithBlockProcessor(tn.blkProcessor),
		node.WithDataStore(store),
		node.WithSyncer(&mock.SyncTimerMock{}),
	)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	tn.node = n
	tn.sk = sk
	tn.pk = pk
	tn.accntState = accntAdapter
	tn.shardId = sharding.MetachainShardId

	dPool.MetaChainBlocks().RegisterHandler(func(key []byte) {
		atomic.AddInt32(&tn.metachainHdrRecv, 1)
	})
	dPool.ShardHeaders().RegisterHandler(func(key []byte) {
		atomic.AddInt32(&tn.headersRecv, 1)
		tn.mutHeaders.Lock()
		metaHeader, _ := dPool.ShardHeaders().Peek(key)
		tn.headers = append(tn.headers, metaHeader.(data.HeaderHandler))
		tn.mutHeaders.Unlock()
	})

	return &tn
}

func createGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		genesisBlocks[shardId] = createGenesisBlock(shardId)
	}

	genesisBlocks[sharding.MetachainShardId] = createGenesisMetaBlock()

	return genesisBlocks
}

func createGenesisBlock(shardId uint32) *dataBlock.Header {
	return &dataBlock.Header{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		ShardId:       shardId,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}

func createGenesisMetaBlock() *dataBlock.MetaBlock {
	return &dataBlock.MetaBlock{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}

func createMintingForSenders(
	nodes []*testNode,
	senderShard uint32,
	sendersPrivateKeys []crypto.PrivateKey,
	value *big.Int,
) {

	for _, n := range nodes {
		//only sender shard nodes will be minted
		if n.shardId != senderShard {
			continue
		}

		for _, sk := range sendersPrivateKeys {
			pkBuff, _ := sk.GeneratePublic().ToByteArray()
			adr, _ := testAddressConverter.CreateAddressFromPublicKeyBytes(pkBuff)
			account, _ := n.accntState.GetAccountWithJournal(adr)
			_ = account.(*state.Account).SetBalanceWithJournal(value)
		}

		_, _ = n.accntState.Commit()
	}
}

func skToPk(sk crypto.PrivateKey) []byte {
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	return pkBuff
}
