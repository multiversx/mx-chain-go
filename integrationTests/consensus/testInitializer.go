package consensus

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
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
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	syncFork "github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	beevikntp "github.com/beevik/ntp"
	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
)

const blsConsensusType = "bls"
const bnConsensusType = "bn"

var r *rand.Rand

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
	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)
	blockChain.GenesisHeader = &dataBlock.Header{}

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10, 1)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func createTestStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaShardDataUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaPeerDataUnit, createMemUnit())
	return store
}

func createTestShardDataPool() dataRetriever.PoolsHolder {
	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})
	cacherCfg := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	hdrPool, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	hdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	txBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	peerChangeBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	metaHdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	metaHdrNonces, _ := dataPool.NewNonceToHashCacher(metaHdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())
	metaBlocks, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

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

func createAccountsDB(marshalizer marshal.Marshalizer) state.AccountsAdapter {
	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (wrapper state.AccountHandler, e error) {
			return state.NewAccount(address, tracker)
		},
	})
	return adb
}

func initialPrivPubKeys(numConsensus int) ([]crypto.PrivateKey, []crypto.PublicKey, crypto.KeyGenerator) {
	privKeys := make([]crypto.PrivateKey, 0)
	pubKeys := make([]crypto.PublicKey, 0)

	testSuite := kyber.NewSuitePairingBn256()
	testKeyGen := signing.NewKeyGenerator(testSuite)

	for i := 0; i < numConsensus; i++ {
		sk, pk := testKeyGen.GeneratePair()

		privKeys = append(privKeys, sk)
		pubKeys = append(pubKeys, pk)
	}

	return privKeys, pubKeys, testKeyGen
}

func createHasher(consensusType string) hashing.Hasher {
	if consensusType == blsConsensusType {
		return blake2b.Blake2b{HashSize: 16}
	}
	return blake2b.Blake2b{}
}

func createConsensusOnlyNode(
	shardCoordinator sharding.Coordinator,
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
		CreateBlockCalled: func(round int32, haveTime func() bool) (handler data.BodyHandler, e error) {
			return &dataBlock.Body{}, nil
		},
		CreateBlockHeaderCalled: func(body data.BodyHandler, round int32, haveTime func() bool) (handler data.HeaderHandler, e error) {
			return &dataBlock.Header{Round: uint32(round)}, nil
		},
		MarshalizedDataToBroadcastCalled: func(header data.HeaderHandler, body data.BodyHandler) (bytes map[uint32][]byte, bytes2 map[uint32][][]byte, e error) {
			mrsData := make(map[uint32][]byte)
			mrsTxs := make(map[uint32][][]byte)
			return mrsData, mrsTxs, nil
		},
	}

	blockProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
		blockProcessor.NrCommitBlockCalled++
		_ = blockChain.SetCurrentBlockHeader(header)
		_ = blockChain.SetCurrentBlockBody(body)
		return nil
	}
	blockProcessor.Marshalizer = testMarshalizer
	blockTracker := &mock.BlocksTrackerMock{
		UnnotarisedBlocksCalled: func() []data.HeaderHandler {
			return make([]data.HeaderHandler, 0)
		},
	}
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

	blockChain.SetGenesisHeader(header)
	hdrMarshalized, _ := testMarshalizer.Marshal(header)
	blockChain.SetGenesisHeaderHash(testHasher.Compute(string(hdrMarshalized)))

	startTime := int64(0)

	singlesigner := &singlesig.SchnorrSigner{}
	singleBlsSigner := &singlesig.BlsSingleSigner{}

	syncer := ntp.NewSyncTime(time.Hour, beevikntp.Query)
	go syncer.StartSync()

	rounder, err := round.NewRound(
		time.Unix(startTime, 0),
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(uint64(roundTime)),
		syncer)

	forkDetector, _ := syncFork.NewBasicForkDetector(rounder)

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

	testMultiSig := mock.NewMultiSigner(uint32(consensusSize))
	_ = testMultiSig.Reset(inPubKeys[shardId], uint16(selfId))

	accntAdapter := createAccountsDB(testMarshalizer)

	n, err := node.NewNode(
		node.WithInitialNodesPubKeys(inPubKeys),
		node.WithRoundDuration(uint64(roundTime)),
		node.WithConsensusGroupSize(int(consensusSize)),
		node.WithSyncer(syncer),
		node.WithGenesisTime(time.Unix(startTime, 0)),
		node.WithRounder(rounder),
		node.WithSingleSigner(singleBlsSigner),
		node.WithPrivKey(privKey),
		node.WithForkDetector(forkDetector),
		node.WithMessenger(messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGen(testKeyGen),
		node.WithShardCoordinator(shardCoordinator),
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
		node.WithBlockTracker(blockTracker),
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
) []*testNode {

	privKeys, pubKeys, testKeyGen := initialPrivPubKeys(nodesPerShard)
	//first node generated will have is pk belonging to firstSkShardId
	nodes := make([]*testNode, nodesPerShard)

	for i := 0; i < nodesPerShard; i++ {
		testNode := &testNode{
			shardId: uint32(0),
		}

		shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(1), uint32(0))
		n, mes, blkProcessor, blkc := createConsensusOnlyNode(
			shardCoordinator,
			testNode.shardId,
			uint32(i),
			serviceID,
			uint32(consensusSize),
			roundTime,
			privKeys[i],
			pubKeys,
			testKeyGen,
			consensusType,
		)

		testNode.node = n
		testNode.node = n
		testNode.sk = privKeys[i]
		testNode.mesenger = mes
		testNode.pk = pubKeys[i]
		testNode.blkProcessor = blkProcessor
		testNode.blkc = blkc

		nodes[i] = testNode
	}

	return nodes
}
