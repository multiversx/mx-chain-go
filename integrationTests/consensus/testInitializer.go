package consensus

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus/round"
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
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	syncFork "github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
)

const blsConsensusType = "bls"
const bnConsensusType = "bn"

var r *rand.Rand
var consensusChainID = []byte("consensus chain ID")

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type testNode struct {
	node             *node.Node
	mesenger         p2p.Messenger
	shardId          uint32
	accntState       state.AccountsAdapter
	blkc             data.ChainHandler
	blkProcessor     *mock.BlockProcessorMock
	sk               crypto.PrivateKey
	pk               crypto.PublicKey
	dPool            dataRetriever.PoolsHolder
	dMetaPool        dataRetriever.MetaPoolsHolder
	headersRecv      int32
	mutHeaders       sync.Mutex
	headersHashes    [][]byte
	headers          []data.HeaderHandler
	metachainHdrRecv int32
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
			address := fmt.Sprintf("addr_%d_%d", shardId, i)
			v, _ := sharding.NewValidator(big.NewInt(0), 1, []byte(shardNodesPks[i]), []byte(address))
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

func pubKeysMapFromKeysMap(keyPairMap map[uint32][]*keyPair) map[uint32][]string {
	keysMap := make(map[uint32][]string, 0)

	for shardId, pairList := range keyPairMap {
		shardKeys := make([]string, len(pairList))
		for i, pair := range pairList {
			b, _ := pair.pk.ToByteArray()
			shardKeys[i] = string(b)
		}
		keysMap[shardId] = shardKeys
	}

	return keysMap
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

func displayAndStartNodes(nodes []*testNode) {
	for _, n := range nodes {
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

func createTestBlockChain() data.ChainHandler {
	cfgCache := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	badBlockCache, _ := storageUnit.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)
	blockChain.GenesisHeader = &dataBlock.Header{}

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	persist, _ := memorydb.New()

	unit, _ := storageUnit.NewStorageUnit(cache, persist)
	return unit
}

func createTestStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())
	store.AddStorer(dataRetriever.BootstrapUnit, createMemUnit())
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

func createAccountsDB(marshalizer marshal.Marshalizer) state.AccountsAdapter {
	marsh := &marshal.JsonMarshalizer{}
	hasher := sha256.Sha256{}
	store := createMemUnit()

	tr, _ := trie.NewTrie(store, marsh, hasher)
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (wrapper state.AccountHandler, e error) {
			return state.NewAccount(address, tracker)
		},
	})
	return adb
}

func createCryptoParams(nodesPerShard int, nbMetaNodes int, nbShards int) *cryptoParams {
	suite := kyber.NewSuitePairingBn256()
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

func createHasher(consensusType string) hashing.Hasher {
	if consensusType == blsConsensusType {
		return blake2b.Blake2b{HashSize: 16}
	}
	return blake2b.Blake2b{}
}

func createConsensusOnlyNode(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	shardId uint32,
	selfId uint32,
	initialAddr string,
	consensusSize uint32,
	roundTime uint64,
	privKey crypto.PrivateKey,
	pubKeys []crypto.PublicKey,
	testKeyGen crypto.KeyGenerator,
	consensusType string,
) (
	*node.Node,
	p2p.Messenger,
	*mock.BlockProcessorMock,
	data.ChainHandler) {

	testHasher := createHasher(consensusType)
	testMarshalizer := &marshal.JsonMarshalizer{}
	testAddressConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")

	messenger := createMessengerWithKadDht(context.Background(), initialAddr)
	rootHash := []byte("roothash")

	blockProcessor := &mock.BlockProcessorMock{
		ProcessBlockCalled: func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)
			return nil
		},
		RevertAccountStateCalled: func() {
		},
		CreateBlockCalled: func(header data.HeaderHandler, haveTime func() bool) (handler data.BodyHandler, e error) {
			return &dataBlock.Body{}, nil
		},
		ApplyBodyToHeaderCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			return nil
		},
		MarshalizedDataToBroadcastCalled: func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
			mrsData := make(map[uint32][]byte)
			mrsTxs := make(map[string][][]byte)
			return mrsData, mrsTxs, nil
		},
		CreateNewHeaderCalled: func() data.HeaderHandler {
			return &dataBlock.Header{}
		},
	}

	blockProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
		blockProcessor.NrCommitBlockCalled++
		_ = blockChain.SetCurrentBlockHeader(header)
		_ = blockChain.SetCurrentBlockBody(body)
		return nil
	}
	blockProcessor.Marshalizer = testMarshalizer
	blockChain := createTestBlockChain()

	header := &dataBlock.Header{
		Nonce:         0,
		ShardId:       shardId,
		BlockBodyType: dataBlock.StateBlock,
		Signature:     rootHash,
		RootHash:      rootHash,
		PrevRandSeed:  rootHash,
		RandSeed:      rootHash,
	}

	_ = blockChain.SetGenesisHeader(header)
	hdrMarshalized, _ := testMarshalizer.Marshal(header)
	blockChain.SetGenesisHeaderHash(testHasher.Compute(string(hdrMarshalized)))

	startTime := int64(0)

	singlesigner := &singlesig.SchnorrSigner{}
	singleBlsSigner := &singlesig.BlsSingleSigner{}

	syncer := ntp.NewSyncTime(ntp.NewNTPGoogleConfig(), time.Hour, nil)
	go syncer.StartSync()

	rounder, err := round.NewRound(
		time.Unix(startTime, 0),
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(roundTime),
		syncer)

	argsNewMetaEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
		GenesisTime:        time.Unix(startTime, 0),
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		Settings: &config.EpochStartConfig{
			MinRoundsBetweenEpochs: 1,
			RoundsPerEpoch:         3,
		},
		Epoch: 0,
	}
	epochStartTrigger, _ := metachain.NewEpochStartTrigger(argsNewMetaEpochStart)

	forkDetector, _ := syncFork.NewShardForkDetector(rounder, timecache.NewTimeCache(time.Second), 0)

	hdrResolver := &mock.HeaderResolverMock{}
	mbResolver := &mock.MiniBlocksResolverMock{}
	resolverFinder := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if baseTopic == factory.HeadersTopic {
				return hdrResolver, nil
			}
			if baseTopic == factory.MiniBlocksTopic {
				return mbResolver, nil
			}
			return hdrResolver, nil
		},
	}

	inPubKeys := make(map[uint32][]string)
	for _, val := range pubKeys {
		sPubKey, _ := val.ToByteArray()
		inPubKeys[shardId] = append(inPubKeys[shardId], string(sPubKey))
	}

	testMultiSig := mock.NewMultiSigner(consensusSize)
	_ = testMultiSig.Reset(inPubKeys[shardId], uint16(selfId))

	accntAdapter := createAccountsDB(testMarshalizer)

	n, err := node.NewNode(
		node.WithInitialNodesPubKeys(inPubKeys),
		node.WithRoundDuration(roundTime),
		node.WithConsensusGroupSize(int(consensusSize)),
		node.WithSyncer(syncer),
		node.WithGenesisTime(time.Unix(startTime, 0)),
		node.WithRounder(rounder),
		node.WithSingleSigner(singleBlsSigner),
		node.WithPrivKey(privKey),
		node.WithForkDetector(forkDetector),
		node.WithMessenger(messenger),
		node.WithMarshalizer(testMarshalizer, 0),
		node.WithHasher(testHasher),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGen(testKeyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithNodesCoordinator(nodesCoordinator),
		node.WithBlockChain(blockChain),
		node.WithMultiSigner(testMultiSig),
		node.WithTxSingleSigner(singlesigner),
		node.WithTxSignPrivKey(privKey),
		node.WithPubKey(privKey.GeneratePublic()),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataPool(createTestShardDataPool()),
		node.WithDataStore(createTestStore()),
		node.WithResolversFinder(resolverFinder),
		node.WithConsensusType(consensusType),
		node.WithBlackListHandler(&mock.BlackListHandlerStub{}),
		node.WithEpochStartTrigger(epochStartTrigger),
		node.WithBootStorer(&mock.BoostrapStorerMock{}),
		node.WithRequestedItemsHandler(&mock.RequestedItemsHandlerStub{}),
		node.WithHeaderSigVerifier(&mock.HeaderSigVerifierStub{}),
		node.WithChainID(consensusChainID),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return n, messenger, blockProcessor, blockChain
}

func createNodes(
	nodesPerShard int,
	consensusSize int,
	roundTime uint64,
	serviceID string,
	consensusType string,
) map[uint32][]*testNode {

	nodes := make(map[uint32][]*testNode)
	cp := createCryptoParams(nodesPerShard, 1, 1)
	keysMap := pubKeysMapFromKeysMap(cp.keys)
	validatorsMap := genValidatorsFromPubKeys(keysMap)
	nodesList := make([]*testNode, nodesPerShard)

	pubKeys := make([]crypto.PublicKey, len(cp.keys[0]))
	for idx, keyPairShard := range cp.keys[0] {
		pubKeys[idx] = keyPairShard.pk
	}

	for i := 0; i < nodesPerShard; i++ {
		testNode := &testNode{
			shardId: uint32(0),
		}

		kp := cp.keys[0][i]
		shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(1), uint32(0))

		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: consensusSize,
			MetaConsensusGroupSize:  1,
			Hasher:                  createHasher(consensusType),
			NbShards:                1,
			Nodes:                   validatorsMap,
			SelfPublicKey:           []byte(strconv.Itoa(i)),
		}
		nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		n, mes, blkProcessor, blkc := createConsensusOnlyNode(
			shardCoordinator,
			nodesCoordinator,
			testNode.shardId,
			uint32(i),
			serviceID,
			uint32(consensusSize),
			roundTime,
			kp.sk,
			pubKeys,
			cp.keyGen,
			consensusType,
		)

		testNode.node = n
		testNode.node = n
		testNode.sk = kp.sk
		testNode.mesenger = mes
		testNode.pk = kp.pk
		testNode.blkProcessor = blkProcessor
		testNode.blkc = blkc
		nodesList[i] = testNode
	}
	nodes[0] = nodesList

	return nodes
}
