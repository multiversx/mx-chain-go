package consensus

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
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
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/consensus/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	syncFork "github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	beevikntp "github.com/beevik/ntp"
)

var r *rand.Rand
var testHasher = sha256.Sha256{}
var testMarshalizer = &marshal.JsonMarshalizer{}
var testAddressConverter, _ = addressConverters.NewPlainAddressConverter(32, "0x")
var testMultiSig *mock.BelNevMock

var testSuite = kyber.NewSuitePairingBn256()
var testKeyGen = signing.NewKeyGenerator(testSuite)

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

func createTestBlockChain() data.ChainHandler {
	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)
	blockChain.GenesisHeader = &dataBlock.Header{}

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
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

func createTestMetaDataPool() dataRetriever.MetaPoolsHolder {
	cacherCfg := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	metaHdrPool, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	metaHdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	shardHdrPool, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 10, Type: storage.LRUCache}
	mbPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})
	dPool, _ := dataPool.NewMetaDataPool(
		metaHdrPool,
		mbPool,
		shardHdrPool,
		metaHdrNonces,
	)

	return dPool
}

func createAccountsDB() state.AccountsAdapter {
	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, testMarshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (wrapper state.AccountHandler, e error) {
			return state.NewAccount(address, tracker)
		},
	})
	return adb
}

func initialPubKeysAndBalances(nrConsens int) map[string]*big.Int {
	pubKeyBalance := make(map[string]*big.Int)

	for i := 0; i < nrConsens; i++ {
		pubKeyBalance[string(i)] = big.NewInt(1000)
	}

	return pubKeyBalance
}

func initialPrivPubKeys(nrConsens int) ([]crypto.PrivateKey, []crypto.PublicKey) {
	testMultiSig = mock.NewMultiSigner(uint32(nrConsens))
	privKeys := make([]crypto.PrivateKey, 0)
	pubKeys := make([]crypto.PublicKey, 0)

	for i := 0; i < nrConsens; i++ {
		sk, pk := testKeyGen.GeneratePair()

		privKeys = append(privKeys, sk)
		pubKeys = append(pubKeys, pk)
	}

	return privKeys, pubKeys
}

// err := ef.node.Start()
//err = ef.node.StartConsensus()

func createConsensusOnlyNode(
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	shardId uint32,
	selfId uint32,
	initialAddr string,
	consensusSize uint32,
	privKey crypto.PrivateKey,
	pubKeys []crypto.PublicKey,
) (
	*node.Node,
	p2p.Messenger,
	*mock.BlockProcessorMock,
	data.ChainHandler) {

	messenger := createMessengerWithKadDht(context.Background(), initialAddr)
	rootHash := []byte("roothash")

	blockProcessor := &mock.BlockProcessorMock{
		CreateGenesisBlockCalled: func(balances map[string]*big.Int) (handler data.HeaderHandler, e error) {
			header := &dataBlock.Header{
				Nonce:         0,
				ShardId:       shardId,
				BlockBodyType: dataBlock.StateBlock,
				Signature:     rootHash,
				RootHash:      rootHash,
				PrevRandSeed:  rootHash,
				RandSeed:      rootHash,
			}

			return header, nil
		},
		ProcessBlockCalled: func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
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
		blockChain.SetCurrentBlockHeader(header)
		blockChain.SetCurrentBlockBody(body)
		return nil
	}

	blockChain := createTestBlockChain()

	roundDuration := uint64(2000)
	startTime := int64(0)

	singlesigner := &singlesig.SchnorrSigner{}
	singleBlsSigner := &singlesig.BlsSingleSigner{}

	syncer := ntp.NewSyncTime(time.Hour, beevikntp.Query)
	go syncer.StartSync()

	//ntpTime := syncer.CurrentTime()
	//startTime = (ntpTime.Unix()/60 + 1) * 58

	rounder, err := round.NewRound(
		time.Unix(startTime, 0),
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(roundDuration),
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

	testMultiSig.Reset(inPubKeys[shardId], uint16(selfId))

	n, err := node.NewNode(
		node.WithInitialNodesPubKeys(inPubKeys),
		node.WithRoundDuration(roundDuration),
		node.WithConsensusGroupSize(int(consensusSize)),
		node.WithSyncer(syncer),
		node.WithGenesisTime(time.Unix(startTime, 0)),
		node.WithRounder(rounder),
		node.WithBlsSinglesig(singleBlsSigner),
		node.WithBlsPrivateKey(privKey),
		node.WithForkDetector(forkDetector),
		node.WithMessenger(messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGenerator(testKeyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blockChain),
		node.WithMultisig(testMultiSig),
		node.WithSinglesig(singlesigner),
		node.WithPrivateKey(privKey),
		node.WithPublicKey(privKey.GeneratePublic()),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataPool(createTestShardDataPool()),
		node.WithDataStore(createTestStore()),
		node.WithResolversFinder(resolverFinder),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return n, messenger, blockProcessor, blockChain
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
	header := []string{"pk", "shard ID", "txs", "miniblocks", "headers", "metachain headers", "connections"}
	dataLines := make([]*display.LineData, len(nodes))
	for idx, n := range nodes {
		buffPk, _ := n.pk.ToByteArray()

		dataLines[idx] = display.NewLineData(
			false,
			[]string{
				hex.EncodeToString(buffPk),
				fmt.Sprintf("%d", n.shardId),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.headersRecv)),
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

		fmt.Printf("Shard ID: %v, sk: %s, pk: %s\n",
			n.shardId,
			hex.EncodeToString(skBuff),
			hex.EncodeToString(pkBuff),
		)
		_ = n.node.Start()
		_ = n.node.P2PBootstrap()
	}
}

func createNodes(
	nodesPerShard int,
	serviceID string,
) []*testNode {

	privKeys, pubKeys := initialPrivPubKeys(nodesPerShard)
	//first node generated will have is pk belonging to firstSkShardId
	nodes := make([]*testNode, nodesPerShard)

	idx := 0
	for i := 0; i < nodesPerShard; i++ {
		testNode := &testNode{
			shardId: uint32(0),
		}

		shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(1), uint32(0))
		accntAdapter := createAccountsDB()
		n, mes, blkProcessor, blkc := createConsensusOnlyNode(
			accntAdapter,
			shardCoordinator,
			testNode.shardId,
			uint32(i),
			serviceID,
			uint32(nodesPerShard),
			privKeys[i],
			pubKeys,
		)

		testNode.node = n
		testNode.accntState = accntAdapter
		testNode.node = n
		testNode.sk = privKeys[i]
		testNode.mesenger = mes
		testNode.pk = pubKeys[i]
		testNode.blkProcessor = blkProcessor
		testNode.blkc = blkc

		nodes[idx] = testNode
		idx++
	}

	return nodes
}
