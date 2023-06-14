package integrationTests

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/heartbeat/monitor"
	"github.com/multiversx/mx-chain-go/heartbeat/processor"
	"github.com/multiversx/mx-chain-go/heartbeat/sender"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/heartbeat/validator"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	interceptorFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
	interceptorsProcessor "github.com/multiversx/mx-chain-go/process/interceptors/processor"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/networksharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/stretchr/testify/require"
)

const (
	defaultNodeName           = "heartbeatNode"
	timeBetweenPeerAuths      = 10 * time.Second
	timeBetweenHeartbeats     = 5 * time.Second
	timeBetweenSendsWhenError = time.Second
	thresholdBetweenSends     = 0.2
	timeBetweenHardforks      = 2 * time.Second

	minPeersThreshold       = 1.0
	delayBetweenRequests    = time.Second
	maxTimeout              = time.Minute
	maxMissingKeysInRequest = 1
	providedHardforkPubKey  = "provided pub key"
)

// TestMarshaller represents the main marshaller
var TestMarshaller = &marshal.GogoProtoMarshalizer{}

// TestThrottler -
var TestThrottler = &processMock.InterceptorThrottlerStub{
	CanProcessCalled: func() bool {
		return true
	},
}

// TestHeartbeatNode represents a container type of class used in integration tests
// with all its fields exported
type TestHeartbeatNode struct {
	ShardCoordinator                     sharding.Coordinator
	NodesCoordinator                     nodesCoordinator.NodesCoordinator
	PeerShardMapper                      process.NetworkShardingCollector
	MainMessenger                        p2p.Messenger
	FullArchiveMessenger                 p2p.Messenger
	NodeKeys                             *TestNodeKeys
	DataPool                             dataRetriever.PoolsHolder
	Sender                               update.Closer
	PeerAuthInterceptor                  *interceptors.MultiDataInterceptor
	HeartbeatInterceptor                 *interceptors.SingleDataInterceptor
	PeerShardInterceptor                 *interceptors.SingleDataInterceptor
	PeerSigHandler                       crypto.PeerSignatureHandler
	WhiteListHandler                     process.WhiteListHandler
	Storage                              dataRetriever.StorageService
	ResolversContainer                   dataRetriever.ResolversContainer
	RequestersContainer                  dataRetriever.RequestersContainer
	RequestersFinder                     dataRetriever.RequestersFinder
	RequestHandler                       process.RequestHandler
	RequestedItemsHandler                dataRetriever.RequestedItemsHandler
	RequestsProcessor                    update.Closer
	ShardSender                          update.Closer
	MainDirectConnectionProcessor        update.Closer
	FullArchiveDirectConnectionProcessor update.Closer
	Interceptor                          *CountInterceptor
	heartbeatExpiryTimespanInSec         int64
}

// NewTestHeartbeatNode returns a new TestHeartbeatNode instance with a libp2p messenger
func NewTestHeartbeatNode(
	tb testing.TB,
	maxShards uint32,
	nodeShardId uint32,
	minPeersWaiting int,
	p2pConfig p2pConfig.P2PConfig,
	heartbeatExpiryTimespanInSec int64,
) *TestHeartbeatNode {
	keygen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	sk, pk := keygen.GeneratePair()

	pksBytes := make(map[uint32][]byte, maxShards)
	pksBytes[nodeShardId], _ = pk.ToByteArray()

	nodesCoordinatorInstance := &shardingMocks.NodesCoordinatorStub{
		GetAllValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
			keys := make(map[uint32][][]byte)
			for shardID := uint32(0); shardID < maxShards; shardID++ {
				keys[shardID] = append(keys[shardID], pksBytes[shardID])
			}

			shardID := core.MetachainShardId
			keys[shardID] = append(keys[shardID], pksBytes[shardID])

			return keys, nil
		},
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (nodesCoordinator.Validator, uint32, error) {
			validatorInstance, _ := nodesCoordinator.NewValidator(publicKey, defaultChancesSelection, 1)
			return validatorInstance, 0, nil
		},
	}
	singleSigner := singlesig.NewBlsSigner()

	peerSigHandler := &cryptoMocks.PeerSignatureHandlerStub{
		VerifyPeerSignatureCalled: func(pk []byte, pid core.PeerID, signature []byte) error {
			senderPubKey, err := keygen.PublicKeyFromByteArray(pk)
			if err != nil {
				return err
			}
			return singleSigner.Verify(senderPubKey, pid.Bytes(), signature)
		},
		GetPeerSignatureCalled: func(privateKey crypto.PrivateKey, pid []byte) ([]byte, error) {
			return singleSigner.Sign(privateKey, pid)
		},
	}

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerFromConfig(p2pConfig)
	pidPk, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	pkShardId, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	pidShardId, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	arg := networksharding.ArgPeerShardMapper{
		PeerIdPkCache:         pidPk,
		FallbackPkShardCache:  pkShardId,
		FallbackPidShardCache: pidShardId,
		NodesCoordinator:      nodesCoordinatorInstance,
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
	}
	peerShardMapper, err := networksharding.NewPeerShardMapper(arg)
	if err != nil {
		log.Error("error creating NewPeerShardMapper", "error", err)
	}
	err = messenger.SetPeerShardResolver(peerShardMapper)
	if err != nil {
		log.Error("error setting NewPeerShardMapper in p2p messenger", "error", err)
	}

	thn := &TestHeartbeatNode{
		ShardCoordinator:             shardCoordinator,
		NodesCoordinator:             nodesCoordinatorInstance,
		MainMessenger:                messenger,
		FullArchiveMessenger:         &p2pmocks.MessengerStub{}, // TODO[Sorin]: inject a proper messenger when all pieces are done to test this network as well
		PeerSigHandler:               peerSigHandler,
		PeerShardMapper:              peerShardMapper,
		heartbeatExpiryTimespanInSec: heartbeatExpiryTimespanInSec,
	}

	localId := thn.MainMessenger.ID()
	pkBytes, _ := pk.ToByteArray()
	thn.PeerShardMapper.UpdatePeerIDInfo(localId, pkBytes, shardCoordinator.SelfId())

	thn.NodeKeys = &TestNodeKeys{
		MainKey: &TestKeyPair{
			Sk: sk,
			Pk: pk,
		},
	}

	// start a go routine in order to allow peers to connect first
	go thn.InitTestHeartbeatNode(tb, minPeersWaiting)

	return thn
}

// NewTestHeartbeatNodeWithCoordinator returns a new TestHeartbeatNode instance with a libp2p messenger
// using provided coordinator and keys
func NewTestHeartbeatNodeWithCoordinator(
	maxShards uint32,
	nodeShardId uint32,
	p2pConfig p2pConfig.P2PConfig,
	coordinator nodesCoordinator.NodesCoordinator,
	keys *TestNodeKeys,
) *TestHeartbeatNode {
	keygen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	singleSigner := singlesig.NewBlsSigner()

	peerSigHandler := &cryptoMocks.PeerSignatureHandlerStub{
		VerifyPeerSignatureCalled: func(pk []byte, pid core.PeerID, signature []byte) error {
			senderPubKey, err := keygen.PublicKeyFromByteArray(pk)
			if err != nil {
				return err
			}
			return singleSigner.Verify(senderPubKey, pid.Bytes(), signature)
		},
		GetPeerSignatureCalled: func(privateKey crypto.PrivateKey, pid []byte) ([]byte, error) {
			return singleSigner.Sign(privateKey, pid)
		},
	}

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerFromConfig(p2pConfig)
	pidPk, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	pkShardId, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	pidShardId, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	arg := networksharding.ArgPeerShardMapper{
		PeerIdPkCache:         pidPk,
		FallbackPkShardCache:  pkShardId,
		FallbackPidShardCache: pidShardId,
		NodesCoordinator:      coordinator,
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
	}
	peerShardMapper, err := networksharding.NewPeerShardMapper(arg)
	if err != nil {
		log.Error("error creating NewPeerShardMapper", "error", err)
	}
	err = messenger.SetPeerShardResolver(peerShardMapper)
	if err != nil {
		log.Error("error setting NewPeerShardMapper in p2p messenger", "error", err)
	}

	thn := &TestHeartbeatNode{
		ShardCoordinator:             shardCoordinator,
		NodesCoordinator:             coordinator,
		MainMessenger:                messenger,
		FullArchiveMessenger:         &p2pmocks.MessengerStub{},
		PeerSigHandler:               peerSigHandler,
		PeerShardMapper:              peerShardMapper,
		Interceptor:                  NewCountInterceptor(),
		heartbeatExpiryTimespanInSec: 30,
	}

	localId := thn.MainMessenger.ID()
	thn.PeerShardMapper.UpdatePeerIDInfo(localId, []byte(""), shardCoordinator.SelfId())

	thn.NodeKeys = keys

	return thn
}

// CreateNodesWithTestHeartbeatNode returns a map with nodes per shard each using a real nodes coordinator
// and TestHeartbeatNode
func CreateNodesWithTestHeartbeatNode(
	nodesPerShard int,
	numMetaNodes int,
	numShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	numObserversOnShard int,
	p2pConfig p2pConfig.P2PConfig,
) map[uint32][]*TestHeartbeatNode {

	cp := CreateCryptoParams(nodesPerShard, numMetaNodes, uint32(numShards), 1)
	pubKeys := PubKeysMapFromNodesKeysMap(cp.NodesKeys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(numShards))
	validatorsForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)
	nodesMap := make(map[uint32][]*TestHeartbeatNode)
	cacherCfg := storageunit.CacheConfig{Capacity: 10000, Type: storageunit.LRUCache, Shards: 1}
	suCache, _ := storageunit.NewCache(cacherCfg)
	for shardId, validatorList := range validatorsMap {
		argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Marshalizer:             TestMarshalizer,
			Hasher:                  TestHasher,
			ShardIDAsObserver:       shardId,
			NbShards:                uint32(numShards),
			EligibleNodes:           validatorsForNodesCoordinator,
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:     suCache,
			Shuffler:                &shardingMocks.NodeShufflerMock{},
			BootStorer:              CreateMemUnit(),
			WaitingNodes:            make(map[uint32][]nodesCoordinator.Validator),
			Epoch:                   0,
			EpochStartNotifier:      notifier.NewEpochStartSubscriptionHandler(),
			ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
			ChanStopNode:            endProcess.GetDummyEndProcessChannel(),
			NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
			IsFullArchive:           false,
			EnableEpochsHandler:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
		}
		nodesCoordinatorInstance, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
		log.LogIfError(err)

		nodesList := make([]*TestHeartbeatNode, len(validatorList))
		for i := range validatorList {
			kp := cp.NodesKeys[shardId][i]
			nodesList[i] = NewTestHeartbeatNodeWithCoordinator(
				uint32(numShards),
				shardId,
				p2pConfig,
				nodesCoordinatorInstance,
				kp,
			)
		}
		nodesMap[shardId] = nodesList
	}

	for counter := uint32(0); counter < uint32(numShards+1); counter++ {
		for j := 0; j < numObserversOnShard; j++ {
			shardId := counter
			if shardId == uint32(numShards) {
				shardId = core.MetachainShardId
			}

			argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
				ShardConsensusGroupSize: shardConsensusGroupSize,
				MetaConsensusGroupSize:  metaConsensusGroupSize,
				Marshalizer:             TestMarshalizer,
				Hasher:                  TestHasher,
				ShardIDAsObserver:       shardId,
				NbShards:                uint32(numShards),
				EligibleNodes:           validatorsForNodesCoordinator,
				SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
				ConsensusGroupCache:     suCache,
				Shuffler:                &shardingMocks.NodeShufflerMock{},
				BootStorer:              CreateMemUnit(),
				WaitingNodes:            make(map[uint32][]nodesCoordinator.Validator),
				Epoch:                   0,
				EpochStartNotifier:      notifier.NewEpochStartSubscriptionHandler(),
				ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
				ChanStopNode:            endProcess.GetDummyEndProcessChannel(),
				NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
				IsFullArchive:           false,
				EnableEpochsHandler:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
				ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
			}
			nodesCoordinatorInstance, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
			log.LogIfError(err)

			nodeKeysInstance := &TestNodeKeys{
				MainKey: createCryptoPair(),
			}
			n := NewTestHeartbeatNodeWithCoordinator(
				uint32(numShards),
				shardId,
				p2pConfig,
				nodesCoordinatorInstance,
				nodeKeysInstance,
			)

			nodesMap[shardId] = append(nodesMap[shardId], n)
		}
	}

	return nodesMap
}

// InitTestHeartbeatNode initializes all the components and starts sender
func (thn *TestHeartbeatNode) InitTestHeartbeatNode(tb testing.TB, minPeersWaiting int) {
	thn.initStorage()
	thn.initDataPools()
	thn.initRequestedItemsHandler()
	thn.initResolversAndRequesters()
	thn.initInterceptors()
	thn.initShardSender(tb)
	thn.initCrossShardPeerTopicNotifier(tb)
	thn.initDirectConnectionProcessor(tb)

	for len(thn.MainMessenger.Peers()) < minPeersWaiting {
		time.Sleep(time.Second)
	}

	thn.initSender()
	thn.initRequestsProcessor()
}

func (thn *TestHeartbeatNode) initDataPools() {
	thn.DataPool = dataRetrieverMock.CreatePoolsHolder(1, thn.ShardCoordinator.SelfId())

	cacherCfg := storageunit.CacheConfig{Capacity: 10000, Type: storageunit.LRUCache, Shards: 1}
	suCache, _ := storageunit.NewCache(cacherCfg)
	thn.WhiteListHandler, _ = interceptors.NewWhiteListDataVerifier(suCache)
}

func (thn *TestHeartbeatNode) initStorage() {
	thn.Storage = CreateStore(thn.ShardCoordinator.NumberOfShards())
}

func (thn *TestHeartbeatNode) initSender() {
	identifierHeartbeat := common.HeartbeatV2Topic + thn.ShardCoordinator.CommunicationIdentifier(thn.ShardCoordinator.SelfId())
	argsSender := sender.ArgSender{
		MainMessenger:           thn.MainMessenger,
		FullArchiveMessenger:    thn.FullArchiveMessenger,
		Marshaller:              TestMarshaller,
		PeerAuthenticationTopic: common.PeerAuthenticationTopic,
		HeartbeatTopic:          identifierHeartbeat,
		BaseVersionNumber:       "v01-base",
		VersionNumber:           "v01",
		NodeDisplayName:         defaultNodeName,
		Identity:                defaultNodeName + "_identity",
		PeerSubType:             core.RegularPeer,
		CurrentBlockProvider:    &testscommon.ChainHandlerStub{},
		PeerSignatureHandler:    thn.PeerSigHandler,
		PrivateKey:              thn.NodeKeys.MainKey.Sk,
		RedundancyHandler:       &mock.RedundancyHandlerStub{},
		NodesCoordinator:        thn.NodesCoordinator,
		HardforkTrigger:         &testscommon.HardforkTriggerStub{},
		HardforkTriggerPubKey:   []byte(providedHardforkPubKey),
		PeerTypeProvider:        &mock.PeerTypeProviderStub{},
		ManagedPeersHolder:      &testscommon.ManagedPeersHolderStub{},
		ShardCoordinator:        thn.ShardCoordinator,

		PeerAuthenticationTimeBetweenSends:          timeBetweenPeerAuths,
		PeerAuthenticationTimeBetweenSendsWhenError: timeBetweenSendsWhenError,
		PeerAuthenticationTimeThresholdBetweenSends: thresholdBetweenSends,
		HeartbeatTimeBetweenSends:                   timeBetweenHeartbeats,
		HeartbeatTimeBetweenSendsWhenError:          timeBetweenSendsWhenError,
		HeartbeatTimeThresholdBetweenSends:          thresholdBetweenSends,
		HardforkTimeBetweenSends:                    timeBetweenHardforks,
		PeerAuthenticationTimeBetweenChecks:         time.Second * 2,
	}

	thn.Sender, _ = sender.NewSender(argsSender)
}

func (thn *TestHeartbeatNode) initResolversAndRequesters() {
	dataPacker, _ := partitioning.NewSimpleDataPacker(TestMarshaller)

	_ = thn.MainMessenger.CreateTopic(common.ConsensusTopic+thn.ShardCoordinator.CommunicationIdentifier(thn.ShardCoordinator.SelfId()), true)
	_ = thn.FullArchiveMessenger.CreateTopic(common.ConsensusTopic+thn.ShardCoordinator.CommunicationIdentifier(thn.ShardCoordinator.SelfId()), true)

	payloadValidator, _ := validator.NewPeerAuthenticationPayloadValidator(thn.heartbeatExpiryTimespanInSec)
	resolverContainerFactoryArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:         thn.ShardCoordinator,
		Messenger:                thn.MainMessenger,
		Store:                    thn.Storage,
		Marshalizer:              TestMarshaller,
		DataPools:                thn.DataPool,
		Uint64ByteSliceConverter: TestUint64Converter,
		DataPacker:               dataPacker,
		TriesContainer: &trieMock.TriesHolderStub{
			GetCalled: func(bytes []byte) common.Trie {
				return &trieMock.TrieStub{}
			},
		},
		SizeCheckDelta:             100,
		InputAntifloodHandler:      &mock.NilAntifloodHandler{},
		OutputAntifloodHandler:     &mock.NilAntifloodHandler{},
		NumConcurrentResolvingJobs: 10,
		PreferredPeersHolder:       &p2pmocks.PeersHolderStub{},
		PayloadValidator:           payloadValidator,
	}

	requestersContainerFactoryArgs := requesterscontainer.FactoryArgs{
		RequesterConfig: config.RequesterConfig{
			NumCrossShardPeers:  2,
			NumTotalPeers:       3,
			NumFullHistoryPeers: 3},
		ShardCoordinator:            thn.ShardCoordinator,
		Messenger:                   thn.MainMessenger,
		Marshaller:                  TestMarshaller,
		Uint64ByteSliceConverter:    TestUint64Converter,
		OutputAntifloodHandler:      &mock.NilAntifloodHandler{},
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		PreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:          &p2pmocks.PeersRatingHandlerStub{},
		SizeCheckDelta:              0,
	}

	if thn.ShardCoordinator.SelfId() == core.MetachainShardId {
		thn.createMetaResolverContainer(resolverContainerFactoryArgs)
		thn.createMetaRequestersContainer(requestersContainerFactoryArgs)
	} else {
		thn.createShardResolverContainer(resolverContainerFactoryArgs)
		thn.createShardRequestersContainer(requestersContainerFactoryArgs)
	}
	thn.createRequestHandler()
}

func (thn *TestHeartbeatNode) createMetaResolverContainer(args resolverscontainer.FactoryArgs) {
	resolversContainerFactory, _ := resolverscontainer.NewMetaResolversContainerFactory(args)

	var err error
	thn.ResolversContainer, err = resolversContainerFactory.Create()
	log.LogIfError(err)
}

func (thn *TestHeartbeatNode) createShardResolverContainer(args resolverscontainer.FactoryArgs) {
	resolversContainerFactory, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	var err error
	thn.ResolversContainer, err = resolversContainerFactory.Create()
	log.LogIfError(err)
}

func (thn *TestHeartbeatNode) createMetaRequestersContainer(args requesterscontainer.FactoryArgs) {
	requestersContainerFactory, _ := requesterscontainer.NewMetaRequestersContainerFactory(args)

	var err error
	thn.RequestersContainer, err = requestersContainerFactory.Create()
	log.LogIfError(err)
}

func (thn *TestHeartbeatNode) createShardRequestersContainer(args requesterscontainer.FactoryArgs) {
	requestersContainerFactory, _ := requesterscontainer.NewShardRequestersContainerFactory(args)

	var err error
	thn.RequestersContainer, err = requestersContainerFactory.Create()
	log.LogIfError(err)
}

func (thn *TestHeartbeatNode) createRequestHandler() {
	thn.RequestersFinder, _ = containers.NewRequestersFinder(thn.RequestersContainer, thn.ShardCoordinator)
	thn.RequestHandler, _ = requestHandlers.NewResolverRequestHandler(
		thn.RequestersFinder,
		thn.RequestedItemsHandler,
		thn.WhiteListHandler,
		100,
		thn.ShardCoordinator.SelfId(),
		time.Second,
	)
}

func (thn *TestHeartbeatNode) initRequestedItemsHandler() {
	thn.RequestedItemsHandler = cache.NewTimeCache(roundDuration)
}

func (thn *TestHeartbeatNode) initInterceptors() {
	argsFactory := interceptorFactory.ArgInterceptedDataFactory{
		CoreComponents: &processMock.CoreComponentsMock{
			IntMarsh:                   TestMarshaller,
			HardforkTriggerPubKeyField: []byte(providedHardforkPubKey),
		},
		ShardCoordinator:             thn.ShardCoordinator,
		NodesCoordinator:             thn.NodesCoordinator,
		PeerSignatureHandler:         thn.PeerSigHandler,
		SignaturesHandler:            &processMock.SignaturesHandlerStub{},
		HeartbeatExpiryTimespanInSec: thn.heartbeatExpiryTimespanInSec,
		PeerID:                       thn.MainMessenger.ID(),
	}

	thn.createPeerAuthInterceptor(argsFactory)
	thn.createHeartbeatInterceptor(argsFactory)
	thn.createPeerShardInterceptor(argsFactory)
}

func (thn *TestHeartbeatNode) createPeerAuthInterceptor(argsFactory interceptorFactory.ArgInterceptedDataFactory) {
	args := interceptorsProcessor.ArgPeerAuthenticationInterceptorProcessor{
		PeerAuthenticationCacher: thn.DataPool.PeerAuthentications(),
		PeerShardMapper:          thn.PeerShardMapper,
		Marshaller:               TestMarshaller,
		HardforkTrigger:          &testscommon.HardforkTriggerStub{},
	}
	paProcessor, _ := interceptorsProcessor.NewPeerAuthenticationInterceptorProcessor(args)
	paFactory, _ := interceptorFactory.NewInterceptedPeerAuthenticationDataFactory(argsFactory)
	thn.PeerAuthInterceptor = thn.initMultiDataInterceptor(common.PeerAuthenticationTopic, paFactory, paProcessor)
}

func (thn *TestHeartbeatNode) createHeartbeatInterceptor(argsFactory interceptorFactory.ArgInterceptedDataFactory) {
	args := interceptorsProcessor.ArgHeartbeatInterceptorProcessor{
		HeartbeatCacher:  thn.DataPool.Heartbeats(),
		ShardCoordinator: thn.ShardCoordinator,
		PeerShardMapper:  thn.PeerShardMapper,
	}
	hbProcessor, _ := interceptorsProcessor.NewHeartbeatInterceptorProcessor(args)
	hbFactory, _ := interceptorFactory.NewInterceptedHeartbeatDataFactory(argsFactory)
	identifierHeartbeat := common.HeartbeatV2Topic + thn.ShardCoordinator.CommunicationIdentifier(thn.ShardCoordinator.SelfId())
	thn.HeartbeatInterceptor = thn.initSingleDataInterceptor(identifierHeartbeat, hbFactory, hbProcessor)
}

func (thn *TestHeartbeatNode) createPeerShardInterceptor(argsFactory interceptorFactory.ArgInterceptedDataFactory) {
	args := interceptorsProcessor.ArgPeerShardInterceptorProcessor{
		PeerShardMapper: thn.PeerShardMapper,
	}
	dciProcessor, _ := interceptorsProcessor.NewPeerShardInterceptorProcessor(args)
	dciFactory, _ := interceptorFactory.NewInterceptedPeerShardFactory(argsFactory)
	thn.PeerShardInterceptor = thn.initSingleDataInterceptor(common.ConnectionTopic, dciFactory, dciProcessor)
}

func (thn *TestHeartbeatNode) initMultiDataInterceptor(topic string, dataFactory process.InterceptedDataFactory, processor process.InterceptorProcessor) *interceptors.MultiDataInterceptor {
	mdInterceptor, _ := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:            topic,
			Marshalizer:      TestMarshalizer,
			DataFactory:      dataFactory,
			Processor:        processor,
			Throttler:        TestThrottler,
			AntifloodHandler: &mock.NilAntifloodHandler{},
			WhiteListRequest: &testscommon.WhiteListHandlerStub{
				IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
					return true
				},
			},
			PreferredPeersHolder: &p2pmocks.PeersHolderStub{},
			CurrentPeerId:        thn.MainMessenger.ID(),
		},
	)

	thn.registerTopicValidator(topic, mdInterceptor)

	return mdInterceptor
}

func (thn *TestHeartbeatNode) initSingleDataInterceptor(topic string, dataFactory process.InterceptedDataFactory, processor process.InterceptorProcessor) *interceptors.SingleDataInterceptor {
	sdInterceptor, _ := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:            topic,
			DataFactory:      dataFactory,
			Processor:        processor,
			Throttler:        TestThrottler,
			AntifloodHandler: &mock.NilAntifloodHandler{},
			WhiteListRequest: &testscommon.WhiteListHandlerStub{
				IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
					return true
				},
			},
			PreferredPeersHolder: &p2pmocks.PeersHolderStub{},
			CurrentPeerId:        thn.MainMessenger.ID(),
		},
	)

	thn.registerTopicValidator(topic, sdInterceptor)

	return sdInterceptor
}

func (thn *TestHeartbeatNode) initRequestsProcessor() {
	args := processor.ArgPeerAuthenticationRequestsProcessor{
		RequestHandler:          thn.RequestHandler,
		NodesCoordinator:        thn.NodesCoordinator,
		PeerAuthenticationPool:  thn.DataPool.PeerAuthentications(),
		ShardId:                 thn.ShardCoordinator.SelfId(),
		Epoch:                   0,
		MinPeersThreshold:       minPeersThreshold,
		DelayBetweenRequests:    delayBetweenRequests,
		MaxTimeoutForRequests:   maxTimeout,
		MaxMissingKeysInRequest: maxMissingKeysInRequest,
		Randomizer:              &random.ConcurrentSafeIntRandomizer{},
	}
	thn.RequestsProcessor, _ = processor.NewPeerAuthenticationRequestsProcessor(args)
}

func (thn *TestHeartbeatNode) initShardSender(tb testing.TB) {
	args := sender.ArgPeerShardSender{
		MainMessenger:             thn.MainMessenger,
		FullArchiveMessenger:      thn.FullArchiveMessenger,
		Marshaller:                TestMarshaller,
		ShardCoordinator:          thn.ShardCoordinator,
		TimeBetweenSends:          5 * time.Second,
		TimeThresholdBetweenSends: 0.1,
		NodesCoordinator:          thn.NodesCoordinator,
	}

	var err error
	thn.ShardSender, err = sender.NewPeerShardSender(args)
	require.Nil(tb, err)
}

func (thn *TestHeartbeatNode) initDirectConnectionProcessor(tb testing.TB) {
	argsDirectConnectionProcessor := processor.ArgsDirectConnectionProcessor{
		TimeToReadDirectConnections: 5 * time.Second,
		Messenger:                   thn.MainMessenger,
		PeerShardMapper:             thn.PeerShardMapper,
		ShardCoordinator:            thn.ShardCoordinator,
		BaseIntraShardTopic:         ShardTopic,
		BaseCrossShardTopic:         ShardTopic,
	}

	var err error
	thn.MainDirectConnectionProcessor, err = processor.NewDirectConnectionProcessor(argsDirectConnectionProcessor)
	require.Nil(tb, err)

	argsDirectConnectionProcessor = processor.ArgsDirectConnectionProcessor{
		TimeToReadDirectConnections: 5 * time.Second,
		Messenger:                   thn.FullArchiveMessenger,
		PeerShardMapper:             thn.PeerShardMapper, // TODO[Sorin]: replace this with the full archive psm
		ShardCoordinator:            thn.ShardCoordinator,
		BaseIntraShardTopic:         ShardTopic,
		BaseCrossShardTopic:         ShardTopic,
	}

	thn.FullArchiveDirectConnectionProcessor, err = processor.NewDirectConnectionProcessor(argsDirectConnectionProcessor)
	require.Nil(tb, err)
}

func (thn *TestHeartbeatNode) initCrossShardPeerTopicNotifier(tb testing.TB) {
	argsCrossShardPeerTopicNotifier := monitor.ArgsCrossShardPeerTopicNotifier{
		ShardCoordinator: thn.ShardCoordinator,
		PeerShardMapper:  thn.PeerShardMapper,
	}
	crossShardPeerTopicNotifier, err := monitor.NewCrossShardPeerTopicNotifier(argsCrossShardPeerTopicNotifier)
	require.Nil(tb, err)

	err = thn.MainMessenger.AddPeerTopicNotifier(crossShardPeerTopicNotifier)
	require.Nil(tb, err)

	argsCrossShardPeerTopicNotifier = monitor.ArgsCrossShardPeerTopicNotifier{
		ShardCoordinator: thn.ShardCoordinator,
		PeerShardMapper:  thn.PeerShardMapper, // TODO[Sorin]: replace this with the full archive psm
	}
	crossShardPeerTopicNotifier, err = monitor.NewCrossShardPeerTopicNotifier(argsCrossShardPeerTopicNotifier)
	require.Nil(tb, err)

	err = thn.FullArchiveMessenger.AddPeerTopicNotifier(crossShardPeerTopicNotifier)
	require.Nil(tb, err)

}

// ConnectTo will try to initiate a connection to the provided parameter
func (thn *TestHeartbeatNode) ConnectTo(connectable Connectable) error {
	if check.IfNil(connectable) {
		return fmt.Errorf("trying to connect to a nil Connectable parameter")
	}

	return thn.MainMessenger.ConnectToPeer(connectable.GetConnectableAddress())
}

// GetConnectableAddress returns a non circuit, non windows default connectable p2p address
func (thn *TestHeartbeatNode) GetConnectableAddress() string {
	if thn == nil {
		return "nil"
	}

	return GetConnectableAddress(thn.MainMessenger)
}

// MakeDisplayTableForHeartbeatNodes returns a string containing counters for received messages for all provided test nodes
func MakeDisplayTableForHeartbeatNodes(nodes map[uint32][]*TestHeartbeatNode) string {
	header := []string{"pk", "pid", "shard ID", "messages global", "messages intra", "messages cross", "conns Total/IntraVal/CrossVal/IntraObs/CrossObs/FullObs/Unk/Sed"}
	dataLines := make([]*display.LineData, 0)

	for shardId, nodesList := range nodes {
		for _, n := range nodesList {
			buffPk, _ := n.NodeKeys.MainKey.Pk.ToByteArray()

			peerInfo := n.MainMessenger.GetConnectedPeersInfo()

			pid := n.MainMessenger.ID().Pretty()
			lineData := display.NewLineData(
				false,
				[]string{
					core.GetTrimmedPk(hex.EncodeToString(buffPk)),
					pid[len(pid)-6:],
					fmt.Sprintf("%d", shardId),
					fmt.Sprintf("%d", n.CountGlobalMessages()),
					fmt.Sprintf("%d", n.CountIntraShardMessages()),
					fmt.Sprintf("%d", n.CountCrossShardMessages()),
					fmt.Sprintf("%d/%d/%d/%d/%d/%d/%d",
						len(n.MainMessenger.ConnectedPeers()),
						peerInfo.NumIntraShardValidators,
						peerInfo.NumCrossShardValidators,
						peerInfo.NumIntraShardObservers,
						peerInfo.NumCrossShardObservers,
						len(peerInfo.UnknownPeers),
						len(peerInfo.Seeders),
					),
				},
			)

			dataLines = append(dataLines, lineData)
		}
	}
	table, _ := display.CreateTableString(header, dataLines)

	return table
}

// registerTopicValidator registers a message processor instance on the provided topic
func (thn *TestHeartbeatNode) registerTopicValidator(topic string, processor p2p.MessageProcessor) {
	err := thn.MainMessenger.CreateTopic(topic, true)
	if err != nil {
		fmt.Printf("error while creating topic %s: %s\n", topic, err.Error())
		return
	}

	err = thn.MainMessenger.RegisterMessageProcessor(topic, "test", processor)
	if err != nil {
		fmt.Printf("error while registering topic validator %s: %s\n", topic, err.Error())
		return
	}
}

// CreateTestInterceptors creates test interceptors that count the number of received messages
func (thn *TestHeartbeatNode) CreateTestInterceptors() {
	thn.registerTopicValidator(GlobalTopic, thn.Interceptor)

	metaIdentifier := ShardTopic + thn.ShardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	thn.registerTopicValidator(metaIdentifier, thn.Interceptor)

	for i := uint32(0); i < thn.ShardCoordinator.NumberOfShards(); i++ {
		identifier := ShardTopic + thn.ShardCoordinator.CommunicationIdentifier(i)
		thn.registerTopicValidator(identifier, thn.Interceptor)
	}
}

// CountGlobalMessages returns the messages count on the global topic
func (thn *TestHeartbeatNode) CountGlobalMessages() int {
	return thn.Interceptor.MessageCount(GlobalTopic)
}

// CountIntraShardMessages returns the messages count on the intra-shard topic
func (thn *TestHeartbeatNode) CountIntraShardMessages() int {
	identifier := ShardTopic + thn.ShardCoordinator.CommunicationIdentifier(thn.ShardCoordinator.SelfId())
	return thn.Interceptor.MessageCount(identifier)
}

// CountCrossShardMessages returns the messages count on the cross-shard topics
func (thn *TestHeartbeatNode) CountCrossShardMessages() int {
	messages := 0

	if thn.ShardCoordinator.SelfId() != core.MetachainShardId {
		metaIdentifier := ShardTopic + thn.ShardCoordinator.CommunicationIdentifier(core.MetachainShardId)
		messages += thn.Interceptor.MessageCount(metaIdentifier)
	}

	for i := uint32(0); i < thn.ShardCoordinator.NumberOfShards(); i++ {
		if i == thn.ShardCoordinator.SelfId() {
			continue
		}

		metaIdentifier := ShardTopic + thn.ShardCoordinator.CommunicationIdentifier(i)
		messages += thn.Interceptor.MessageCount(metaIdentifier)
	}

	return messages
}

// Close -
func (thn *TestHeartbeatNode) Close() {
	_ = thn.Sender.Close()
	_ = thn.PeerAuthInterceptor.Close()
	_ = thn.RequestsProcessor.Close()
	_ = thn.RequestersContainer.Close()
	_ = thn.ResolversContainer.Close()
	_ = thn.ShardSender.Close()
	_ = thn.MainMessenger.Close()
	_ = thn.FullArchiveMessenger.Close()
	_ = thn.MainDirectConnectionProcessor.Close()
	_ = thn.FullArchiveDirectConnectionProcessor.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (thn *TestHeartbeatNode) IsInterfaceNil() bool {
	return thn == nil
}

func createCryptoPair() *TestKeyPair {
	suite := mcl.NewSuiteBLS12()
	keyGen := signing.NewKeyGenerator(suite)

	kp := &TestKeyPair{}
	kp.Sk, kp.Pk = keyGen.GeneratePair()

	return kp
}
