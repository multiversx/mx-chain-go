package integrationTests

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	ed25519SingleSig "github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	mclsinglesig "github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus/round"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/factory/peerSignatureHandler"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	syncFork "github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/sharding"
	elrondShardingMocks "github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	vic "github.com/ElrondNetwork/elrond-go/testscommon/validatorInfoCacher"
)

const (
	blsConsensusType = "bls"
	signatureSize    = 48
	publicKeySize    = 96
	maxShards        = 1
	nodeShardId      = 0
)

var testPubkeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(32)

// TestConsensusNode represents a structure used in integration tests used for consensus tests
type TestConsensusNode struct {
	Node             *node.Node
	Messenger        p2p.Messenger
	NodesCoordinator nodesCoordinator.NodesCoordinator
	ShardCoordinator sharding.Coordinator
	ChainHandler     data.ChainHandler
	BlockProcessor   *mock.BlockProcessorMock
	ResolverFinder   dataRetriever.ResolversFinder
	AccountsDB       *state.AccountsDB
	NodeKeys         TestKeyPair
}

// NewTestConsensusNode returns a new TestConsensusNode
func NewTestConsensusNode(
	consensusSize int,
	roundTime uint64,
	consensusType string,
	nodeKeys TestKeyPair,
	eligibleMap map[uint32][]nodesCoordinator.Validator,
	waitingMap map[uint32][]nodesCoordinator.Validator,
	keyGen crypto.KeyGenerator,
) *TestConsensusNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	tcn := &TestConsensusNode{
		NodeKeys:         nodeKeys,
		ShardCoordinator: shardCoordinator,
	}
	tcn.initNode(consensusSize, roundTime, consensusType, eligibleMap, waitingMap, keyGen)

	return tcn
}

// CreateNodesWithTestConsensusNode returns a map with nodes per shard each using TestConsensusNode
func CreateNodesWithTestConsensusNode(
	numMetaNodes int,
	nodesPerShard int,
	consensusSize int,
	roundTime uint64,
	consensusType string,
) map[uint32][]*TestConsensusNode {

	nodes := make(map[uint32][]*TestConsensusNode, nodesPerShard)
	cp := CreateCryptoParams(nodesPerShard, numMetaNodes, maxShards)
	keysMap := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(keysMap, maxShards)
	eligibleMap, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)
	waitingMap := make(map[uint32][]nodesCoordinator.Validator)
	connectableNodes := make([]Connectable, 0)

	for _, keysPair := range cp.Keys[0] {
		tcn := NewTestConsensusNode(
			consensusSize,
			roundTime,
			consensusType,
			*keysPair,
			eligibleMap,
			waitingMap,
			cp.KeyGen)
		nodes[nodeShardId] = append(nodes[nodeShardId], tcn)
		connectableNodes = append(connectableNodes, tcn)
	}

	ConnectNodes(connectableNodes)

	return nodes
}

func (tcn *TestConsensusNode) initNode(
	consensusSize int,
	roundTime uint64,
	consensusType string,
	eligibleMap map[uint32][]nodesCoordinator.Validator,
	waitingMap map[uint32][]nodesCoordinator.Validator,
	keyGen crypto.KeyGenerator,
) {

	testHasher := createHasher(consensusType)
	epochStartRegistrationHandler := notifier.NewEpochStartSubscriptionHandler()
	consensusCache, _ := lrucache.NewCache(10000)
	pkBytes, _ := tcn.NodeKeys.Pk.ToByteArray()

	tcn.initNodesCoordinator(consensusSize, testHasher, epochStartRegistrationHandler, eligibleMap, waitingMap, pkBytes, consensusCache)
	tcn.Messenger = CreateMessengerWithNoDiscovery()
	tcn.initBlockChain(testHasher)
	tcn.initBlockProcessor()

	startTime := time.Now().Unix()

	singlesigner := &ed25519SingleSig.Ed25519Signer{}
	singleBlsSigner := &mclsinglesig.BlsSingleSigner{}

	syncer := ntp.NewSyncTime(ntp.NewNTPGoogleConfig(), nil)
	syncer.StartSyncingTime()

	roundHandler, _ := round.NewRound(
		time.Unix(startTime, 0),
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(roundTime),
		syncer,
		0)

	dataPool := dataRetrieverMock.CreatePoolsHolder(1, 0)

	argsNewMetaEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
		GenesisTime:        time.Unix(startTime, 0),
		EpochStartNotifier: notifier.NewEpochStartSubscriptionHandler(),
		Settings: &config.EpochStartConfig{
			MinRoundsBetweenEpochs: 1,
			RoundsPerEpoch:         3,
		},
		Epoch:            0,
		Storage:          createTestStore(),
		Marshalizer:      TestMarshalizer,
		Hasher:           testHasher,
		AppStatusHandler: &statusHandlerMock.AppStatusHandlerStub{},
		DataPool:         dataPool,
	}
	epochStartTrigger, _ := metachain.NewEpochStartTrigger(argsNewMetaEpochStart)

	forkDetector, _ := syncFork.NewShardForkDetector(
		roundHandler,
		timecache.NewTimeCache(time.Second),
		&mock.BlockTrackerStub{},
		0,
	)

	tcn.initResolverFinder()

	testMultiSig := cryptoMocks.NewMultiSigner(uint32(consensusSize))

	peerSigCache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(peerSigCache, singleBlsSigner, keyGen)

	tcn.initAccountsDB()

	coreComponents := GetDefaultCoreComponents()
	coreComponents.SyncTimerField = syncer
	coreComponents.RoundHandlerField = roundHandler
	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.HasherField = testHasher
	coreComponents.AddressPubKeyConverterField = testPubkeyConverter
	coreComponents.ChainIdCalled = func() string {
		return string(ChainID)
	}
	coreComponents.GenesisTimeField = time.Unix(startTime, 0)
	coreComponents.GenesisNodesSetupField = &testscommon.NodesSetupStub{
		GetShardConsensusGroupSizeCalled: func() uint32 {
			return uint32(consensusSize)
		},
		GetMetaConsensusGroupSizeCalled: func() uint32 {
			return uint32(consensusSize)
		},
	}

	cryptoComponents := GetDefaultCryptoComponents()
	cryptoComponents.PrivKey = tcn.NodeKeys.Sk
	cryptoComponents.PubKey = tcn.NodeKeys.Sk.GeneratePublic()
	cryptoComponents.BlockSig = singleBlsSigner
	cryptoComponents.TxSig = singlesigner
	cryptoComponents.MultiSig = testMultiSig
	cryptoComponents.BlKeyGen = keyGen
	cryptoComponents.PeerSignHandler = peerSigHandler

	processComponents := GetDefaultProcessComponents()
	processComponents.ForkDetect = forkDetector
	processComponents.ShardCoord = tcn.ShardCoordinator
	processComponents.NodesCoord = tcn.NodesCoordinator
	processComponents.BlockProcess = tcn.BlockProcessor
	processComponents.ResFinder = tcn.ResolverFinder
	processComponents.EpochTrigger = epochStartTrigger
	processComponents.EpochNotifier = epochStartRegistrationHandler
	processComponents.BlackListHdl = &testscommon.TimeCacheStub{}
	processComponents.BootSore = &mock.BoostrapStorerMock{}
	processComponents.HeaderSigVerif = &mock.HeaderSigVerifierStub{}
	processComponents.HeaderIntegrVerif = &mock.HeaderIntegrityVerifierStub{}
	processComponents.ReqHandler = &testscommon.RequestHandlerStub{}
	processComponents.PeerMapper = mock.NewNetworkShardingCollectorMock()
	processComponents.RoundHandlerField = roundHandler
	processComponents.ScheduledTxsExecutionHandlerInternal = &testscommon.ScheduledTxsExecutionStub{}
	processComponents.ProcessedMiniBlocksTrackerInternal = &testscommon.ProcessedMiniBlocksTrackerStub{}

	dataComponents := GetDefaultDataComponents()
	dataComponents.BlockChain = tcn.ChainHandler
	dataComponents.DataPool = dataPool
	dataComponents.Store = createTestStore()

	stateComponents := GetDefaultStateComponents()
	stateComponents.Accounts = tcn.AccountsDB
	stateComponents.AccountsAPI = tcn.AccountsDB

	networkComponents := GetDefaultNetworkComponents()
	networkComponents.Messenger = tcn.Messenger
	networkComponents.InputAntiFlood = &mock.NilAntifloodHandler{}
	networkComponents.PeerHonesty = &mock.PeerHonestyHandlerStub{}

	var err error
	tcn.Node, err = node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithProcessComponents(processComponents),
		node.WithDataComponents(dataComponents),
		node.WithStateComponents(stateComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithRoundDuration(roundTime),
		node.WithConsensusGroupSize(consensusSize),
		node.WithConsensusType(consensusType),
		node.WithGenesisTime(time.Unix(startTime, 0)),
		node.WithValidatorSignatureSize(signatureSize),
		node.WithPublicKeySize(publicKeySize),
	)

	if err != nil {
		fmt.Println(err.Error())
	}
}

func (tcn *TestConsensusNode) initNodesCoordinator(
	consensusSize int,
	hasher hashing.Hasher,
	epochStartRegistrationHandler notifier.EpochStartNotifier,
	eligibleMap map[uint32][]nodesCoordinator.Validator,
	waitingMap map[uint32][]nodesCoordinator.Validator,
	pkBytes []byte,
	cache storage.Cacher,
) {
	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize: consensusSize,
		MetaConsensusGroupSize:  1,
		Marshalizer:             TestMarshalizer,
		Hasher:                  hasher,
		Shuffler:                &shardingMocks.NodeShufflerMock{},
		EpochStartNotifier:      epochStartRegistrationHandler,
		BootStorer:              CreateMemUnit(),
		NbShards:                maxShards,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           pkBytes,
		ConsensusGroupCache:     cache,
		ShuffledOutHandler:      &elrondShardingMocks.ShuffledOutHandlerStub{},
		ChanStopNode:            endProcess.GetDummyEndProcessChannel(),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{
			IsWaitingListFixFlagEnabledField: true,
		},
		ValidatorInfoCacher: &vic.ValidatorInfoCacherStub{},
	}

	tcn.NodesCoordinator, _ = nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
}

func (tcn *TestConsensusNode) initBlockChain(hasher hashing.Hasher) {
	if tcn.ShardCoordinator.SelfId() == core.MetachainShardId {
		tcn.ChainHandler = CreateMetaChain()
	} else {
		tcn.ChainHandler = CreateShardChain()
	}

	rootHash := []byte("roothash")
	header := &dataBlock.Header{
		Nonce:         0,
		ShardID:       tcn.ShardCoordinator.SelfId(),
		BlockBodyType: dataBlock.StateBlock,
		Signature:     rootHash,
		RootHash:      rootHash,
		PrevRandSeed:  rootHash,
		RandSeed:      rootHash,
	}

	_ = tcn.ChainHandler.SetGenesisHeader(header)
	hdrMarshalized, _ := TestMarshalizer.Marshal(header)
	tcn.ChainHandler.SetGenesisHeaderHash(hasher.Compute(string(hdrMarshalized)))
}

func (tcn *TestConsensusNode) initBlockProcessor() {
	tcn.BlockProcessor = &mock.BlockProcessorMock{
		CommitBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			_ = tcn.ChainHandler.SetCurrentBlockHeaderAndRootHash(header, header.GetRootHash())
			return nil
		},
		CreateBlockCalled: func(header data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
			return header, &dataBlock.Body{}, nil
		},
		MarshalizedDataToBroadcastCalled: func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
			mrsData := make(map[uint32][]byte)
			mrsTxs := make(map[string][][]byte)
			return mrsData, mrsTxs, nil
		},
		CreateNewHeaderCalled: func(round uint64, nonce uint64) (data.HeaderHandler, error) {
			return &dataBlock.Header{
				Round:           round,
				Nonce:           nonce,
				SoftwareVersion: []byte("version"),
			}, nil
		},
	}

	tcn.BlockProcessor.CommitBlockCalled = func(header data.HeaderHandler, body data.BodyHandler) error {
		tcn.BlockProcessor.NrCommitBlockCalled++
		_ = tcn.ChainHandler.SetCurrentBlockHeaderAndRootHash(header, header.GetRootHash())
		return nil
	}
	tcn.BlockProcessor.Marshalizer = TestMarshalizer
}

func (tcn *TestConsensusNode) initResolverFinder() {
	hdrResolver := &mock.HeaderResolverStub{}
	mbResolver := &mock.MiniBlocksResolverStub{}
	tcn.ResolverFinder = &mock.ResolversFinderStub{
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
}

func (tcn *TestConsensusNode) initAccountsDB() {
	trieStorage, _ := CreateTrieStorageManager(CreateMemUnit())
	tcn.AccountsDB, _ = CreateAccountsDB(UserAccount, trieStorage)
}

func createHasher(consensusType string) hashing.Hasher {
	if consensusType == blsConsensusType {
		hasher, _ := blake2b.NewBlake2bWithSize(32)
		return hasher
	}
	return blake2b.NewBlake2b()
}

func createTestStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.BootstrapUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ScheduledSCRsUnit, CreateMemUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, CreateMemUnit())

	return store
}

// ConnectTo will try to initiate a connection to the provided parameter
func (tcn *TestConsensusNode) ConnectTo(connectable Connectable) error {
	if check.IfNil(connectable) {
		return fmt.Errorf("trying to connect to a nil Connectable parameter")
	}

	return tcn.Messenger.ConnectToPeer(connectable.GetConnectableAddress())
}

// GetConnectableAddress returns a non circuit, non windows default connectable p2p address
func (tcn *TestConsensusNode) GetConnectableAddress() string {
	if tcn == nil {
		return "nil"
	}

	return GetConnectableAddress(tcn.Messenger)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tcn *TestConsensusNode) IsInterfaceNil() bool {
	return tcn == nil
}
