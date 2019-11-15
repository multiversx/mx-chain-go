package smartContract

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
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
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
)

//TODO refactor this package to use TestNodeProcessor infrastructure

var r *rand.Rand
var testHasher = sha256.Sha256{}
var testMarshalizer = &marshal.JsonMarshalizer{}
var testAddressConverter, _ = addressConverters.NewPlainAddressConverter(32, "0x")
var testMultiSig = mock.NewMultiSigner(1)
var rootHash = []byte("root hash")
var addrConv, _ = addressConverters.NewPlainAddressConverter(32, "0x")

var opGas = int64(1)

const maxTxNonceDeltaAllowed = 8000

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type testNode struct {
	node               *node.Node
	messenger          p2p.Messenger
	shardId            uint32
	accntState         state.AccountsAdapter
	blkc               data.ChainHandler
	store              dataRetriever.StorageService
	blkProcessor       process.BlockProcessor
	txProcessor        process.TransactionProcessor
	txCoordinator      process.TransactionCoordinator
	scrForwarder       process.IntermediateTransactionHandler
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

type keyPair struct {
	sk crypto.PrivateKey
	pk crypto.PublicKey
}

type cryptoParams struct {
	keyGen       crypto.KeyGenerator
	keys         map[uint32][]*keyPair
	singleSigner crypto.SingleSigner
}

func genValidatorsFromPubKeys(pubKeysMap map[uint32][]string) map[uint32][]sharding.Validator {
	validatorsMap := make(map[uint32][]sharding.Validator)

	for shardId, shardNodesPks := range pubKeysMap {
		shardValidators := make([]sharding.Validator, 0)
		for i := 0; i < len(shardNodesPks); i++ {
			v, _ := sharding.NewValidator(big.NewInt(0), 1, []byte(shardNodesPks[i]), []byte(shardNodesPks[i]))
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

func createCryptoParams(nodesPerShard int, nbMetaNodes int, nbShards int) *cryptoParams {
	suite := kyber.NewBlakeSHA256Ed25519()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)

	keysMap := make(map[uint32][]*keyPair)
	keyPairs := make([]*keyPair, nodesPerShard)
	for shardId := 0; shardId < nbShards; shardId++ {
		for n := 0; n < nodesPerShard; n++ {
			kp := &keyPair{}
			kp.sk, kp.pk = keyGen.GeneratePair()
			keyPairs[n] = kp
		}
		keysMap[uint32(shardId)] = keyPairs
	}

	keyPairs = make([]*keyPair, nbMetaNodes)
	for n := 0; n < nbMetaNodes; n++ {
		kp := &keyPair{}
		kp.sk, kp.pk = keyGen.GeneratePair()
		keyPairs[n] = kp
	}
	keysMap[sharding.MetachainShardId] = keyPairs

	params := &cryptoParams{
		keys:         keysMap,
		keyGen:       keyGen,
		singleSigner: singleSigner,
	}

	return params
}

func pubKeysMapFromKeysMap(keyPairMap map[uint32][]*keyPair) map[uint32][]string {
	keysMap := make(map[uint32][]string, 0)

	for shardId, pairList := range keyPairMap {
		shardKeys := make([]string, len(pairList))
		for i, pair := range pairList {
			bytes, _ := pair.pk.ToByteArray()
			shardKeys[i] = string(bytes)
		}
		keysMap[shardId] = shardKeys
	}

	return keysMap
}

func createTestShardChain() *blockchain.BlockChain {
	cfgCache := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	badBlockCache, _ := storageUnit.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)
	blockChain.GenesisHeader = &dataBlock.Header{}
	genesisHeaderM, _ := testMarshalizer.Marshal(blockChain.GenesisHeader)

	blockChain.SetGenesisHeaderHash(testHasher.Compute(string(genesisHeaderM)))

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
	store.AddStorer(dataRetriever.RewardTransactionUnit, createMemUnit())
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
	rewardsTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache})
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

	currTxs, _ := dataPool.NewCurrentBlockPool()

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		uTxPool,
		rewardsTxPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
		currTxs,
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

func createMockTxFeeHandler() process.FeeHandler {
	return &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return tx.GetGasLimit()
		},
		ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			fee := big.NewInt(0).SetUint64(tx.GetGasPrice())
			fee.Mul(fee, big.NewInt(0).SetUint64(tx.GetGasLimit()))

			return fee
		},
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
			return nil
		},
	}
}

func createNetNode(
	dPool dataRetriever.PoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	targetShardId uint32,
	initialAddr string,
	params *cryptoParams,
	keysIndex int,
) (
	*node.Node,
	p2p.Messenger,
	dataRetriever.ResolversFinder,
	process.BlockProcessor,
	process.TransactionProcessor,
	process.TransactionCoordinator,
	process.IntermediateTransactionHandler,
	data.ChainHandler,
	dataRetriever.StorageService) {

	messenger := createMessengerWithKadDht(context.Background(), initialAddr)
	keyPair := params.keys[targetShardId][keysIndex]
	pkBuff, _ := keyPair.pk.ToByteArray()
	fmt.Printf("pk: %s\n", hex.EncodeToString(pkBuff))

	blkc := createTestShardChain()
	store := createTestShardStore(shardCoordinator.NumberOfShards())
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	dataPacker, _ := partitioning.NewSimpleDataPacker(testMarshalizer)

	interceptorContainerFactory, _ := shard.NewInterceptorsContainerFactory(
		accntAdapter,
		shardCoordinator,
		nodesCoordinator,
		messenger,
		store,
		testMarshalizer,
		testHasher,
		params.keyGen,
		params.singleSigner,
		testMultiSig,
		dPool,
		testAddressConverter,
		maxTxNonceDeltaAllowed,
		createMockTxFeeHandler(),
		timecache.NewTimeCache(time.Second),
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
	requestHandler, _ := requestHandlers.NewShardResolverRequestHandler(
		resolversFinder,
		factory.TransactionTopic,
		factory.UnsignedTransactionTopic,
		factory.RewardsTransactionTopic,
		factory.MiniBlocksTopic,
		factory.HeadersTopic,
		factory.MetachainBlocksTopic,
		100,
	)

	economicsData := &economics.EconomicsData{}

	interimProcFactory, _ := shard.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		testMarshalizer,
		testHasher,
		testAddressConverter,
		mock.NewSpecialAddressHandlerMock(
			testAddressConverter,
			shardCoordinator,
			nodesCoordinator,
		),
		store,
		dPool,
		economicsData,
	)
	interimProcContainer, _ := interimProcFactory.Create()
	scForwarder, _ := interimProcContainer.Get(dataBlock.SmartContractResultBlock)
	rewardsInter, _ := interimProcContainer.Get(dataBlock.RewardsBlock)
	rewardsHandler, _ := rewardsInter.(process.TransactionFeeHandler)
	internalTxProducer, _ := rewardsInter.(process.InternalTransactionProducer)
	rewardProcessor, _ := rewardTransaction.NewRewardTxProcessor(
		accntAdapter,
		addrConv,
		shardCoordinator,
		rewardsInter,
	)
	vm, blockChainHook := createVMAndBlockchainHook(accntAdapter, shardCoordinator)
	vmContainer := &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return vm, nil
		}}
	argsParser, _ := smartContract.NewAtArgumentParser()
	scProcessor, _ := smartContract.NewSmartContractProcessor(
		vmContainer,
		argsParser,
		testHasher,
		testMarshalizer,
		accntAdapter,
		blockChainHook,
		addrConv,
		shardCoordinator,
		scForwarder,
		rewardsHandler,
	)

	txTypeHandler, _ := coordinator.NewTxTypeHandler(addrConv, shardCoordinator, accntAdapter)

	txProcessor, _ := transaction.NewTxProcessor(
		accntAdapter,
		testHasher,
		testAddressConverter,
		testMarshalizer,
		shardCoordinator,
		scProcessor,
		rewardsHandler,
		txTypeHandler,
		createMockTxFeeHandler(),
	)

	miniBlocksCompacter, _ := preprocess.NewMiniBlocksCompaction(createMockTxFeeHandler(), shardCoordinator)

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
		scProcessor,
		scProcessor,
		rewardProcessor,
		internalTxProducer,
		createMockTxFeeHandler(),
		miniBlocksCompacter,
	)
	container, _ := fact.Create()

	tc, _ := coordinator.NewTransactionCoordinator(
		shardCoordinator,
		accntAdapter,
		dPool.MiniBlocks(),
		requestHandler,
		container,
		interimProcContainer,
	)

	genesisBlocks := createGenesisBlocks(shardCoordinator)

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      testHasher,
		Marshalizer: testMarshalizer,
	}
	headerValidator, _ := block.NewHeaderValidator(argsHeaderValidator)

	arguments := block.ArgShardProcessor{
		ArgBaseProcessor: block.ArgBaseProcessor{
			Accounts: accntAdapter,
			ForkDetector: &mock.ForkDetectorMock{
				AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
					return nil
				},
				GetHighestFinalBlockNonceCalled: func() uint64 {
					return 0
				},
				ProbableHighestNonceCalled: func() uint64 {
					return 0
				},
			},
			Hasher:           testHasher,
			Marshalizer:      testMarshalizer,
			Store:            store,
			ShardCoordinator: shardCoordinator,
			NodesCoordinator: nodesCoordinator,
			SpecialAddressHandler: mock.NewSpecialAddressHandlerMock(
				testAddressConverter,
				shardCoordinator,
				nodesCoordinator,
			),
			Uint64Converter:              uint64Converter,
			StartHeaders:                 genesisBlocks,
			RequestHandler:               requestHandler,
			Core:                         &mock.ServiceContainerMock{},
			BlockChainHook:               blockChainHook,
			TxCoordinator:                tc,
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorMock{},
			EpochStartTrigger:            &mock.EpochStartTriggerStub{},
			HeaderValidator:              headerValidator,
			Rounder:                      &mock.RounderMock{},
		},
		DataPool:        dPool,
		TxsPoolsCleaner: &mock.TxPoolsCleanerMock{},
	}

	blockProcessor, _ := block.NewShardProcessor(arguments)

	_ = blkc.SetGenesisHeader(genesisBlocks[shardCoordinator.SelfId()])

	n, err := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithDataPool(dPool),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGen(params.keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultiSigner(testMultiSig),
		node.WithSingleSigner(params.singleSigner),
		node.WithTxSignPrivKey(keyPair.sk),
		node.WithTxSignPubKey(keyPair.pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataStore(store),
		node.WithSyncer(&mock.SyncTimerMock{}),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return n, messenger, resolversFinder, blockProcessor, txProcessor, tc, scForwarder, blkc, store
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

func createNodes(
	numOfShards int,
	nodesPerShard int,
	serviceID string,
) map[uint32][]*testNode {
	//first node generated will have is pk belonging to firstSkShardId
	numMetaChainNodes := 1
	nodes := make(map[uint32][]*testNode)
	cp := createCryptoParams(nodesPerShard, numMetaChainNodes, numOfShards)
	keysMap := pubKeysMapFromKeysMap(cp.keys)
	validatorsMap := genValidatorsFromPubKeys(keysMap)

	for shardId := 0; shardId < numOfShards; shardId++ {
		shardNodes := make([]*testNode, nodesPerShard)

		for j := 0; j < nodesPerShard; j++ {
			testNode := &testNode{
				dPool:   createTestShardDataPool(),
				shardId: uint32(shardId),
			}

			shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), uint32(shardId))
			argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
				ShardConsensusGroupSize: 1,
				MetaConsensusGroupSize:  1,
				Hasher:                  testHasher,
				ShardId:                 uint32(shardId),
				NbShards:                uint32(numOfShards),
				Nodes:                   validatorsMap,
				SelfPublicKey:           []byte(strconv.Itoa(j)),
			}
			nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

			accntAdapter := createAccountsDB()
			n, mes, resFinder, blkProcessor, txProcessor, transactionCoordinator, scrForwarder, blkc, store := createNetNode(
				testNode.dPool,
				accntAdapter,
				shardCoordinator,
				nodesCoordinator,
				testNode.shardId,
				serviceID,
				cp,
				j,
			)
			_ = n.CreateShardedStores()

			KeyPair := cp.keys[uint32(shardId)][j]
			testNode.node = n
			testNode.sk = KeyPair.sk
			testNode.messenger = mes
			testNode.pk = KeyPair.pk
			testNode.resFinder = resFinder
			testNode.accntState = accntAdapter
			testNode.blkProcessor = blkProcessor
			testNode.txProcessor = txProcessor
			testNode.scrForwarder = scrForwarder
			testNode.blkc = blkc
			testNode.store = store
			testNode.txCoordinator = transactionCoordinator
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
				KeyPair.sk,
				&singlesig.SchnorrSigner{},
			)

			shardNodes[j] = testNode
		}

		nodes[uint32(shardId)] = shardNodes
	}

	metaNodes := make([]*testNode, numMetaChainNodes)
	for i := 0; i < numMetaChainNodes; i++ {
		shardCoordinatorMeta, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), sharding.MetachainShardId)
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: 1,
			MetaConsensusGroupSize:  1,
			Hasher:                  testHasher,
			ShardId:                 sharding.MetachainShardId,
			NbShards:                uint32(numOfShards),
			Nodes:                   validatorsMap,
			SelfPublicKey:           []byte(strconv.Itoa(i)),
		}
		nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		metaNodes[i] = createMetaNetNode(
			createTestMetaDataPool(),
			createAccountsDB(),
			shardCoordinatorMeta,
			nodesCoordinator,
			serviceID,
			cp,
			i,
		)
	}

	nodes[sharding.MetachainShardId] = metaNodes

	return nodes
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
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	for i := uint32(0); i < coordinator.NumberOfShards(); i++ {
		store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(i), createMemUnit())
	}

	return store
}

func createTestMetaDataPool() dataRetriever.MetaPoolsHolder {
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	metaBlocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 10000, Type: storageUnit.LRUCache}
	miniblocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	shardHeaders, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	headersNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	headersNonces, _ := dataPool.NewNonceSyncMapCacher(headersNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	txPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	uTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})

	currTxs, _ := dataPool.NewCurrentBlockPool()

	dPool, _ := dataPool.NewMetaDataPool(
		metaBlocks,
		miniblocks,
		shardHeaders,
		headersNonces,
		txPool,
		uTxPool,
		currTxs,
	)

	return dPool
}

func createMetaNetNode(
	dPool dataRetriever.MetaPoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	initialAddr string,
	params *cryptoParams,
	keysIndex int,
) *testNode {

	tn := testNode{}

	tn.messenger = createMessengerWithKadDht(context.Background(), initialAddr)
	keyPair := params.keys[sharding.MetachainShardId][keysIndex]
	pkBuff, _ := keyPair.pk.ToByteArray()
	fmt.Printf("Found pk: %s\n", hex.EncodeToString(pkBuff))

	tn.blkc = createTestMetaChain()
	store := createTestMetaStore(shardCoordinator)
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	feeHandler := &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return tx.GetGasLimit()
		},
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
			return nil
		},
		ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			fee := big.NewInt(0).SetUint64(tx.GetGasLimit())
			fee.Mul(fee, big.NewInt(0).SetUint64(tx.GetGasPrice()))

			return fee
		},
	}

	interceptorContainerFactory, _ := metaProcess.NewInterceptorsContainerFactory(
		shardCoordinator,
		nodesCoordinator,
		tn.messenger,
		store,
		testMarshalizer,
		testHasher,
		testMultiSig,
		dPool,
		accntAdapter,
		testAddressConverter,
		params.singleSigner,
		params.keyGen,
		maxTxNonceDeltaAllowed,
		feeHandler,
		timecache.NewTimeCache(time.Second),
	)
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		fmt.Println(err.Error())
	}

	dataPacker, _ := partitioning.NewSimpleDataPacker(testMarshalizer)

	resolversContainerFactory, _ := metafactoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		tn.messenger,
		store,
		testMarshalizer,
		dPool,
		uint64Converter,
		dataPacker,
	)
	resolversContainer, _ := resolversContainerFactory.Create()
	resolvers, _ := containers.NewResolversFinder(resolversContainer, shardCoordinator)

	requestHandler, _ := requestHandlers.NewMetaResolverRequestHandler(
		resolvers,
		factory.ShardHeadersForMetachainTopic,
		factory.MetachainBlocksTopic,
		factory.TransactionTopic,
		factory.UnsignedTransactionTopic,
		factory.MiniBlocksTopic,
		100,
	)

	genesisBlocks := createGenesisBlocks(shardCoordinator)

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      testHasher,
		Marshalizer: testMarshalizer,
	}
	headerValidator, _ := block.NewHeaderValidator(argsHeaderValidator)

	arguments := block.ArgMetaProcessor{
		ArgBaseProcessor: block.ArgBaseProcessor{
			Accounts: accntAdapter,
			ForkDetector: &mock.ForkDetectorMock{
				AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
					return nil
				},
				GetHighestFinalBlockNonceCalled: func() uint64 {
					return 0
				},
				ProbableHighestNonceCalled: func() uint64 {
					return 0
				},
			},
			Hasher:           testHasher,
			Marshalizer:      testMarshalizer,
			Store:            store,
			ShardCoordinator: shardCoordinator,
			NodesCoordinator: nodesCoordinator,
			SpecialAddressHandler: mock.NewSpecialAddressHandlerMock(
				testAddressConverter,
				shardCoordinator,
				nodesCoordinator,
			),
			Uint64Converter:   uint64Converter,
			StartHeaders:      genesisBlocks,
			RequestHandler:    requestHandler,
			Core:              &mock.ServiceContainerMock{},
			BlockChainHook:    &mock.BlockChainHookHandlerMock{},
			TxCoordinator:     &mock.TransactionCoordinatorMock{},
			Rounder:           &mock.RounderMock{},
			EpochStartTrigger: &mock.EpochStartTriggerStub{},
			HeaderValidator:   headerValidator,
		},
		DataPool:           dPool,
		SCDataGetter:       &mock.ScDataGetterMock{},
		SCToProtocol:       &mock.SCToProtocolStub{},
		PeerChangesHandler: &mock.PeerChangesHandler{},
		PendingMiniBlocks:  &mock.PendingMiniBlocksHandlerStub{},
	}
	blkProc, _ := block.NewMetaProcessor(arguments)

	_ = tn.blkc.SetGenesisHeader(genesisBlocks[sharding.MetachainShardId])

	tn.blkProcessor = blkProc

	tn.broadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		testMarshalizer,
		tn.messenger,
		shardCoordinator,
		keyPair.sk,
		params.singleSigner,
	)

	n, err := node.NewNode(
		node.WithMessenger(tn.messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithMetaDataPool(dPool),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGen(params.keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(tn.blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultiSigner(testMultiSig),
		node.WithSingleSigner(params.singleSigner),
		node.WithPrivKey(keyPair.sk),
		node.WithPubKey(keyPair.pk),
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
	tn.sk = keyPair.sk
	tn.pk = keyPair.pk
	tn.accntState = accntAdapter
	tn.shardId = sharding.MetachainShardId

	dPool.MetaBlocks().RegisterHandler(func(key []byte) {
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
	sendersPublicKeys [][]byte,
	value *big.Int,
) {

	for _, n := range nodes {
		//only sender shard nodes will be minted
		if n.shardId != senderShard {
			continue
		}

		for _, pk := range sendersPublicKeys {
			adr, _ := testAddressConverter.CreateAddressFromPublicKeyBytes(pk)
			account, _ := n.accntState.GetAccountWithJournal(adr)
			_ = account.(*state.Account).SetBalanceWithJournal(value)
		}

		_, _ = n.accntState.Commit()
	}
}

func createVMAndBlockchainHook(
	accnts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
) (vmcommon.VMExecutionHandler, *hooks.BlockChainHookImpl) {
	args := hooks.ArgBlockChainHook{
		Accounts:         accnts,
		AddrConv:         addrConv,
		StorageService:   &mock.ChainStorerMock{},
		BlockChain:       &mock.BlockChainMock{},
		ShardCoordinator: shardCoordinator,
		Marshalizer:      testMarshalizer,
		Uint64Converter:  &mock.Uint64ByteSliceConverterMock{},
	}

	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, testHasher)
	vm.GasForOperation = uint64(opGas)

	return vm, blockChainHook
}
