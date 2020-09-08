package consensus

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus/round"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/peerSignatureHandler"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	ed25519SingleSig "github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	mclsinglesig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/trie/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	syncFork "github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
)

const blsConsensusType = "bls"
const signatureSize = 48
const publicKeySize = 96

var p2pBootstrapDelay = time.Second * 5
var testPubkeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(32)

type testNode struct {
	node         *node.Node
	mesenger     p2p.Messenger
	blkc         data.ChainHandler
	blkProcessor *mock.BlockProcessorMock
	sk           crypto.PrivateKey
	pk           crypto.PublicKey
	shardId      uint32
}

type keyPair struct {
	sk crypto.PrivateKey
	pk crypto.PublicKey
}

type cryptoParams struct {
	keyGen         crypto.KeyGenerator
	keys           map[uint32][]*keyPair
	txSingleSigner crypto.SingleSigner
	singleSigner   crypto.SingleSigner
}

func genValidatorsFromPubKeys(pubKeysMap map[uint32][]string) map[uint32][]sharding.Validator {
	validatorsMap := make(map[uint32][]sharding.Validator)

	for shardId, shardNodesPks := range pubKeysMap {
		shardValidators := make([]sharding.Validator, 0)
		for i := 0; i < len(shardNodesPks); i++ {
			v, _ := sharding.NewValidator([]byte(shardNodesPks[i]), 1, uint32(i))
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

func pubKeysMapFromKeysMap(keyPairMap map[uint32][]*keyPair) map[uint32][]string {
	keysMap := make(map[uint32][]string)

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
			testPubkeyConverter.Encode(pkBuff),
		)
		_ = n.mesenger.Bootstrap()
	}
}

func createTestBlockChain() data.ChainHandler {
	blockChain := blockchain.NewBlockChain()
	_ = blockChain.SetGenesisHeader(&dataBlock.Header{})

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 10, Shards: 1, SizeInBytes: 0})

	unit, _ := storageUnit.NewStorageUnit(cache, memorydb.New())
	return unit
}

func createTestStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())
	store.AddStorer(dataRetriever.BootstrapUnit, createMemUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, createMemUnit())
	return store
}

func createAccountsDB(marshalizer marshal.Marshalizer) state.AccountsAdapter {
	marsh := &marshal.GogoProtoMarshalizer{}
	hasher := sha256.Sha256{}
	store := createMemUnit()
	evictionWaitListSize := uint(100)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(evictionWaitListSize, memorydb.New(), marsh)

	// TODO change this implementation with a factory
	tempDir, _ := ioutil.TempDir("", "integrationTests")
	cfg := config.DBConfig{
		FilePath:          tempDir,
		Type:              string(storageUnit.LvlDBSerial),
		BatchDelaySeconds: 4,
		MaxBatchSize:      10000,
		MaxOpenFiles:      10,
	}
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	trieStorage, _ := trie.NewTrieStorageManager(store, marshalizer, hasher, cfg, ewl, generalCfg)

	maxTrieLevelInMemory := uint(5)
	tr, _ := trie.NewTrie(trieStorage, marsh, hasher, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marshalizer, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address []byte) (wrapper state.AccountHandler, e error) {
			return state.NewUserAccount(address)
		},
	})
	return adb
}

func createCryptoParams(nodesPerShard int, nbMetaNodes int, nbShards int) *cryptoParams {
	suite := mcl.NewSuiteBLS12()
	txSingleSigner := &ed25519SingleSig.Ed25519Signer{}
	singleSigner := &mclsinglesig.BlsSingleSigner{}
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
	keysMap[core.MetachainShardId] = keyPairs

	params := &cryptoParams{
		keys:           keysMap,
		keyGen:         keyGen,
		txSingleSigner: txSingleSigner,
		singleSigner:   singleSigner,
	}

	return params
}

func createHasher(consensusType string) hashing.Hasher {
	if consensusType == blsConsensusType {
		return &blake2b.Blake2b{HashSize: 32}
	}
	return &blake2b.Blake2b{}
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
	epochStartRegistrationHandler epochStart.RegistrationHandler,
) (
	*node.Node,
	p2p.Messenger,
	*mock.BlockProcessorMock,
	data.ChainHandler) {

	testHasher := createHasher(consensusType)
	testMarshalizer := &marshal.GogoProtoMarshalizer{}

	messenger := integrationTests.CreateMessengerWithKadDht(initialAddr)
	rootHash := []byte("roothash")

	blockChain := createTestBlockChain()
	blockProcessor := &mock.BlockProcessorMock{
		ProcessBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
			_ = blockChain.SetCurrentBlockHeader(header)
			return nil
		},
		RevertAccountStateCalled: func(header data.HeaderHandler) {
		},
		CreateBlockCalled: func(header data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
			return header, &dataBlock.Body{}, nil
		},
		MarshalizedDataToBroadcastCalled: func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
			mrsData := make(map[uint32][]byte)
			mrsTxs := make(map[string][][]byte)
			return mrsData, mrsTxs, nil
		},
		CreateNewHeaderCalled: func(round uint64, nonce uint64) data.HeaderHandler {
			return &dataBlock.Header{
				Round:           round,
				Nonce:           nonce,
				SoftwareVersion: []byte("version"),
			}
		},
	}

	blockProcessor.CommitBlockCalled = func(header data.HeaderHandler, body data.BodyHandler) error {
		blockProcessor.NrCommitBlockCalled++
		_ = blockChain.SetCurrentBlockHeader(header)
		return nil
	}
	blockProcessor.Marshalizer = testMarshalizer

	header := &dataBlock.Header{
		Nonce:         0,
		ShardID:       shardId,
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

	singlesigner := &ed25519SingleSig.Ed25519Signer{}
	singleBlsSigner := &mclsinglesig.BlsSingleSigner{}

	syncer := ntp.NewSyncTime(ntp.NewNTPGoogleConfig(), nil)
	syncer.StartSyncingTime()

	rounder, _ := round.NewRound(
		time.Unix(startTime, 0),
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(roundTime),
		syncer,
		0)

	argsNewMetaEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
		GenesisTime:        time.Unix(startTime, 0),
		EpochStartNotifier: notifier.NewEpochStartSubscriptionHandler(),
		Settings: &config.EpochStartConfig{
			MinRoundsBetweenEpochs: 1,
			RoundsPerEpoch:         3,
		},
		Epoch:       0,
		Storage:     createTestStore(),
		Marshalizer: testMarshalizer,
		Hasher:      testHasher,
	}
	epochStartTrigger, _ := metachain.NewEpochStartTrigger(argsNewMetaEpochStart)

	forkDetector, _ := syncFork.NewShardForkDetector(
		rounder,
		timecache.NewTimeCache(time.Second),
		&mock.BlockTrackerStub{},
		0,
	)

	hdrResolver := &mock.HeaderResolverMock{}
	mbResolver := &mock.MiniBlocksResolverMock{}
	resolverFinder := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if baseTopic == factory.MiniBlocksTopic {
				return mbResolver, nil
			}
			return nil, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			if baseTopic == factory.ShardBlocksTopic {
				return hdrResolver, nil
			}
			return nil, nil
		},
	}

	inPubKeys := make(map[uint32][]string)
	for _, val := range pubKeys {
		sPubKey, _ := val.ToByteArray()
		inPubKeys[shardId] = append(inPubKeys[shardId], string(sPubKey))
	}

	testMultiSig := mock.NewMultiSigner(consensusSize)
	_ = testMultiSig.Reset(inPubKeys[shardId], uint16(selfId))

	peerSigCache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(peerSigCache, singleBlsSigner, testKeyGen)

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
		node.WithInternalMarshalizer(testMarshalizer, 0),
		node.WithVmMarshalizer(&marshal.JsonMarshalizer{}),
		node.WithTxSignMarshalizer(&marshal.JsonMarshalizer{}),
		node.WithHasher(testHasher),
		node.WithAddressPubkeyConverter(testPubkeyConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGen(testKeyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithNodesCoordinator(nodesCoordinator),
		node.WithBlockChain(blockChain),
		node.WithMultiSigner(testMultiSig),
		node.WithTxSingleSigner(singlesigner),
		node.WithPubKey(privKey.GeneratePublic()),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataPool(testscommon.CreatePoolsHolder(1, 0)),
		node.WithDataStore(createTestStore()),
		node.WithResolversFinder(resolverFinder),
		node.WithConsensusType(consensusType),
		node.WithBlockBlackListHandler(&mock.TimeCacheStub{}),
		node.WithPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{}),
		node.WithEpochStartTrigger(epochStartTrigger),
		node.WithEpochStartEventNotifier(epochStartRegistrationHandler),
		node.WithNetworkShardingCollector(mock.NewNetworkShardingCollectorMock()),
		node.WithBootStorer(&mock.BoostrapStorerMock{}),
		node.WithRequestedItemsHandler(&mock.RequestedItemsHandlerStub{}),
		node.WithHeaderSigVerifier(&mock.HeaderSigVerifierStub{}),
		node.WithHeaderIntegrityVerifier(&mock.HeaderIntegrityVerifierStub{}),
		node.WithChainID(integrationTests.ChainID),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithUint64ByteSliceConverter(&mock.Uint64ByteSliceConverterMock{}),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithInputAntifloodHandler(&mock.NilAntifloodHandler{}),
		node.WithSignatureSize(signatureSize),
		node.WithPublicKeySize(publicKeySize),
		node.WithPeerHonestyHandler(&mock.PeerHonestyHandlerStub{}),
		node.WithInterceptorsContainer(&mock.InterceptorsContainerStub{}),
		node.WithHardforkTrigger(&mock.HardforkTriggerStub{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
		node.WithPeerSignatureHandler(peerSigHandler),
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
	eligibleMap := genValidatorsFromPubKeys(keysMap)
	waitingMap := make(map[uint32][]sharding.Validator)
	nodesList := make([]*testNode, nodesPerShard)

	nodeShuffler := &mock.NodeShufflerMock{}

	pubKeys := make([]crypto.PublicKey, len(cp.keys[0]))
	for idx, keyPairShard := range cp.keys[0] {
		pubKeys[idx] = keyPairShard.pk
	}

	for i := 0; i < nodesPerShard; i++ {
		testNodeObject := &testNode{
			shardId: uint32(0),
		}

		kp := cp.keys[0][i]
		shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(1), uint32(0))
		epochStartRegistrationHandler := notifier.NewEpochStartSubscriptionHandler()
		bootStorer := integrationTests.CreateMemUnit()
		consensusCache, _ := lrucache.NewCache(10000)

		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: consensusSize,
			MetaConsensusGroupSize:  1,
			Marshalizer:             integrationTests.TestMarshalizer,
			Hasher:                  createHasher(consensusType),
			Shuffler:                nodeShuffler,
			EpochStartNotifier:      epochStartRegistrationHandler,
			BootStorer:              bootStorer,
			NbShards:                1,
			EligibleNodes:           eligibleMap,
			WaitingNodes:            waitingMap,
			SelfPublicKey:           []byte(strconv.Itoa(i)),
			ConsensusGroupCache:     consensusCache,
			ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		}
		nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		n, mes, blkProcessor, blkc := createConsensusOnlyNode(
			shardCoordinator,
			nodesCoordinator,
			testNodeObject.shardId,
			uint32(i),
			serviceID,
			uint32(consensusSize),
			roundTime,
			kp.sk,
			pubKeys,
			cp.keyGen,
			consensusType,
			epochStartRegistrationHandler,
		)

		testNodeObject.node = n
		testNodeObject.sk = kp.sk
		testNodeObject.mesenger = mes
		testNodeObject.pk = kp.pk
		testNodeObject.blkProcessor = blkProcessor
		testNodeObject.blkc = blkc
		nodesList[i] = testNodeObject
	}
	nodes[0] = nodesList

	return nodes
}
