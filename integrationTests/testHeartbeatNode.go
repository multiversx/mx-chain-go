package integrationTests

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/display"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/heartbeat/processor"
	"github.com/ElrondNetwork/elrond-go/heartbeat/sender"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	interceptorsProcessor "github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	processMock "github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/networksharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/update"
)

const (
	defaultNodeName           = "heartbeatNode"
	timeBetweenPeerAuths      = 15 * time.Second
	timeBetweenHeartbeats     = 5 * time.Second
	timeBetweenSendsWhenError = time.Second
	thresholdBetweenSends     = 0.2
	timeBetweenHardforks      = 2 * time.Second

	messagesInChunk         = 10
	minPeersThreshold       = 1.0
	delayBetweenRequests    = time.Second * 5
	maxTimeout              = time.Minute
	maxMissingKeysInRequest = 1
	providedHardforkPubKey  = "provided pub key"
)

// TestMarshaller represents the main marshaller
var TestMarshaller = &testscommon.MarshalizerMock{}

// TestThrottler -
var TestThrottler = &processMock.InterceptorThrottlerStub{
	CanProcessCalled: func() bool {
		return true
	},
}

// TestHeartbeatNode represents a container type of class used in integration tests
// with all its fields exported
type TestHeartbeatNode struct {
	ShardCoordinator           sharding.Coordinator
	NodesCoordinator           nodesCoordinator.NodesCoordinator
	PeerShardMapper            process.NetworkShardingCollector
	Messenger                  p2p.Messenger
	NodeKeys                   TestKeyPair
	DataPool                   dataRetriever.PoolsHolder
	Sender                     update.Closer
	PeerAuthInterceptor        *interceptors.MultiDataInterceptor
	HeartbeatInterceptor       *interceptors.MultiDataInterceptor
	ValidatorInfoInterceptor   *interceptors.SingleDataInterceptor
	PeerSigHandler             crypto.PeerSignatureHandler
	WhiteListHandler           process.WhiteListHandler
	Storage                    dataRetriever.StorageService
	ResolversContainer         dataRetriever.ResolversContainer
	ResolverFinder             dataRetriever.ResolversFinder
	RequestHandler             process.RequestHandler
	RequestedItemsHandler      dataRetriever.RequestedItemsHandler
	RequestsProcessor          update.Closer
	DirectConnectionsProcessor update.Closer
	Interceptor                *CountInterceptor
}

// NewTestHeartbeatNode returns a new TestHeartbeatNode instance with a libp2p messenger
func NewTestHeartbeatNode(
	maxShards uint32,
	nodeShardId uint32,
	minPeersWaiting int,
	p2pConfig config.P2PConfig,
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
			validator, _ := nodesCoordinator.NewValidator(publicKey, defaultChancesSelection, 1)
			return validator, 0, nil
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
	pidPk, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	pkShardId, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	pidShardId, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	startInEpoch := uint32(0)
	arg := networksharding.ArgPeerShardMapper{
		PeerIdPkCache:         pidPk,
		FallbackPkShardCache:  pkShardId,
		FallbackPidShardCache: pidShardId,
		NodesCoordinator:      nodesCoordinatorInstance,
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
		StartEpoch:            startInEpoch,
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
		ShardCoordinator: shardCoordinator,
		NodesCoordinator: nodesCoordinatorInstance,
		Messenger:        messenger,
		PeerSigHandler:   peerSigHandler,
		PeerShardMapper:  peerShardMapper,
	}

	localId := thn.Messenger.ID()
	thn.PeerShardMapper.UpdatePeerIDInfo(localId, []byte(""), shardCoordinator.SelfId())

	thn.NodeKeys = TestKeyPair{
		Sk: sk,
		Pk: pk,
	}

	// start a go routine in order to allow peers to connect first
	go thn.InitTestHeartbeatNode(minPeersWaiting)

	return thn
}

// NewTestHeartbeatNodeWithCoordinator returns a new TestHeartbeatNode instance with a libp2p messenger
// using provided coordinator and keys
func NewTestHeartbeatNodeWithCoordinator(
	maxShards uint32,
	nodeShardId uint32,
	p2pConfig config.P2PConfig,
	coordinator nodesCoordinator.NodesCoordinator,
	keys TestKeyPair,
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
	pidPk, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	pkShardId, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	pidShardId, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	startInEpoch := uint32(0)
	arg := networksharding.ArgPeerShardMapper{
		PeerIdPkCache:         pidPk,
		FallbackPkShardCache:  pkShardId,
		FallbackPidShardCache: pidShardId,
		NodesCoordinator:      coordinator,
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
		StartEpoch:            startInEpoch,
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
		ShardCoordinator: shardCoordinator,
		NodesCoordinator: coordinator,
		Messenger:        messenger,
		PeerSigHandler:   peerSigHandler,
		PeerShardMapper:  peerShardMapper,
		Interceptor:      NewCountInterceptor(),
	}

	localId := thn.Messenger.ID()
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
	p2pConfig config.P2PConfig,
) map[uint32][]*TestHeartbeatNode {

	cp := CreateCryptoParams(nodesPerShard, numMetaNodes, uint32(numShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(numShards))
	validatorsForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)
	nodesMap := make(map[uint32][]*TestHeartbeatNode)
	cacherCfg := storageUnit.CacheConfig{Capacity: 10000, Type: storageUnit.LRUCache, Shards: 1}
	cache, _ := storageUnit.NewCache(cacherCfg)
	for shardId, validatorList := range validatorsMap {
		argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
			ShardConsensusGroupSize:    shardConsensusGroupSize,
			MetaConsensusGroupSize:     metaConsensusGroupSize,
			Marshalizer:                TestMarshalizer,
			Hasher:                     TestHasher,
			ShardIDAsObserver:          shardId,
			NbShards:                   uint32(numShards),
			EligibleNodes:              validatorsForNodesCoordinator,
			SelfPublicKey:              []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:        cache,
			Shuffler:                   &shardingMocks.NodeShufflerMock{},
			BootStorer:                 CreateMemUnit(),
			WaitingNodes:               make(map[uint32][]nodesCoordinator.Validator),
			Epoch:                      0,
			EpochStartNotifier:         notifier.NewEpochStartSubscriptionHandler(),
			ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
			WaitingListFixEnabledEpoch: 0,
			ChanStopNode:               endProcess.GetDummyEndProcessChannel(),
			NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
			IsFullArchive:              false,
		}
		nodesCoordinatorInstance, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
		log.LogIfError(err)

		nodesList := make([]*TestHeartbeatNode, len(validatorList))
		for i := range validatorList {
			kp := cp.Keys[shardId][i]
			nodesList[i] = NewTestHeartbeatNodeWithCoordinator(
				uint32(numShards),
				shardId,
				p2pConfig,
				nodesCoordinatorInstance,
				*kp,
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
				ShardConsensusGroupSize:    shardConsensusGroupSize,
				MetaConsensusGroupSize:     metaConsensusGroupSize,
				Marshalizer:                TestMarshalizer,
				Hasher:                     TestHasher,
				ShardIDAsObserver:          shardId,
				NbShards:                   uint32(numShards),
				EligibleNodes:              validatorsForNodesCoordinator,
				SelfPublicKey:              []byte(strconv.Itoa(int(shardId))),
				ConsensusGroupCache:        cache,
				Shuffler:                   &shardingMocks.NodeShufflerMock{},
				BootStorer:                 CreateMemUnit(),
				WaitingNodes:               make(map[uint32][]nodesCoordinator.Validator),
				Epoch:                      0,
				EpochStartNotifier:         notifier.NewEpochStartSubscriptionHandler(),
				ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
				WaitingListFixEnabledEpoch: 0,
				ChanStopNode:               endProcess.GetDummyEndProcessChannel(),
				NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
				IsFullArchive:              false,
			}
			nodesCoordinatorInstance, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
			log.LogIfError(err)

			n := NewTestHeartbeatNodeWithCoordinator(
				uint32(numShards),
				shardId,
				p2pConfig,
				nodesCoordinatorInstance,
				createCryptoPair(),
			)

			nodesMap[shardId] = append(nodesMap[shardId], n)
		}
	}

	return nodesMap
}

// InitTestHeartbeatNode initializes all the components and starts sender
func (thn *TestHeartbeatNode) InitTestHeartbeatNode(minPeersWaiting int) {
	thn.initStorage()
	thn.initDataPools()
	thn.initRequestedItemsHandler()
	thn.initResolvers()
	thn.initInterceptors()
	thn.initDirectConnectionsProcessor()

	for len(thn.Messenger.Peers()) < minPeersWaiting {
		time.Sleep(time.Second)
	}

	thn.initSender()
	thn.initRequestsProcessor()
}

func (thn *TestHeartbeatNode) initDataPools() {
	thn.DataPool = dataRetrieverMock.CreatePoolsHolder(1, thn.ShardCoordinator.SelfId())

	cacherCfg := storageUnit.CacheConfig{Capacity: 10000, Type: storageUnit.LRUCache, Shards: 1}
	cache, _ := storageUnit.NewCache(cacherCfg)
	thn.WhiteListHandler, _ = interceptors.NewWhiteListDataVerifier(cache)
}

func (thn *TestHeartbeatNode) initStorage() {
	thn.Storage = CreateStore(thn.ShardCoordinator.NumberOfShards())
}

func (thn *TestHeartbeatNode) initSender() {
	identifierHeartbeat := common.HeartbeatV2Topic + thn.ShardCoordinator.CommunicationIdentifier(thn.ShardCoordinator.SelfId())
	argsSender := sender.ArgSender{
		Messenger:               thn.Messenger,
		Marshaller:              TestMarshaller,
		PeerAuthenticationTopic: common.PeerAuthenticationTopic,
		HeartbeatTopic:          identifierHeartbeat,
		VersionNumber:           "v01",
		NodeDisplayName:         defaultNodeName,
		Identity:                defaultNodeName + "_identity",
		PeerSubType:             core.RegularPeer,
		CurrentBlockProvider:    &testscommon.ChainHandlerStub{},
		PeerSignatureHandler:    thn.PeerSigHandler,
		PrivateKey:              thn.NodeKeys.Sk,
		RedundancyHandler:       &mock.RedundancyHandlerStub{},
		NodesCoordinator:        thn.NodesCoordinator,
		HardforkTrigger:         &testscommon.HardforkTriggerStub{},
		HardforkTriggerPubKey:   []byte(providedHardforkPubKey),

		PeerAuthenticationTimeBetweenSends:          timeBetweenPeerAuths,
		PeerAuthenticationTimeBetweenSendsWhenError: timeBetweenSendsWhenError,
		PeerAuthenticationThresholdBetweenSends:     thresholdBetweenSends,
		HeartbeatTimeBetweenSends:                   timeBetweenHeartbeats,
		HeartbeatTimeBetweenSendsWhenError:          timeBetweenSendsWhenError,
		HeartbeatThresholdBetweenSends:              thresholdBetweenSends,
		HardforkTimeBetweenSends:                    timeBetweenHardforks,
	}

	thn.Sender, _ = sender.NewSender(argsSender)
}

func (thn *TestHeartbeatNode) initResolvers() {
	dataPacker, _ := partitioning.NewSimpleDataPacker(TestMarshaller)

	_ = thn.Messenger.CreateTopic(common.ConsensusTopic+thn.ShardCoordinator.CommunicationIdentifier(thn.ShardCoordinator.SelfId()), true)

	resolverContainerFactory := resolverscontainer.FactoryArgs{
		ShardCoordinator:         thn.ShardCoordinator,
		Messenger:                thn.Messenger,
		Store:                    thn.Storage,
		Marshalizer:              TestMarshaller,
		DataPools:                thn.DataPool,
		Uint64ByteSliceConverter: TestUint64Converter,
		DataPacker:               dataPacker,
		TriesContainer: &mock.TriesHolderStub{
			GetCalled: func(bytes []byte) common.Trie {
				return &trieMock.TrieStub{}
			},
		},
		SizeCheckDelta:              100,
		InputAntifloodHandler:       &mock.NilAntifloodHandler{},
		OutputAntifloodHandler:      &mock.NilAntifloodHandler{},
		NumConcurrentResolvingJobs:  10,
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		PreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		ResolverConfig: config.ResolverConfig{
			NumCrossShardPeers:  2,
			NumIntraShardPeers:  1,
			NumFullHistoryPeers: 3,
		},
		NodesCoordinator:                     thn.NodesCoordinator,
		MaxNumOfPeerAuthenticationInResponse: 5,
		PeerShardMapper:                      thn.PeerShardMapper,
		PeersRatingHandler:                   &p2pmocks.PeersRatingHandlerStub{},
	}

	if thn.ShardCoordinator.SelfId() == core.MetachainShardId {
		thn.createMetaResolverContainer(resolverContainerFactory)
	} else {
		thn.createShardResolverContainer(resolverContainerFactory)
	}
}

func (thn *TestHeartbeatNode) createMetaResolverContainer(args resolverscontainer.FactoryArgs) {
	resolversContainerFactory, _ := resolverscontainer.NewMetaResolversContainerFactory(args)

	var err error
	thn.ResolversContainer, err = resolversContainerFactory.Create()
	log.LogIfError(err)

	thn.createRequestHandler()
}

func (thn *TestHeartbeatNode) createShardResolverContainer(args resolverscontainer.FactoryArgs) {
	resolversContainerFactory, _ := resolverscontainer.NewShardResolversContainerFactory(args)

	var err error
	thn.ResolversContainer, err = resolversContainerFactory.Create()
	log.LogIfError(err)

	thn.createRequestHandler()
}

func (thn *TestHeartbeatNode) createRequestHandler() {
	thn.ResolverFinder, _ = containers.NewResolversFinder(thn.ResolversContainer, thn.ShardCoordinator)
	thn.RequestHandler, _ = requestHandlers.NewResolverRequestHandler(
		thn.ResolverFinder,
		thn.RequestedItemsHandler,
		thn.WhiteListHandler,
		100,
		thn.ShardCoordinator.SelfId(),
		time.Second,
	)
}

func (thn *TestHeartbeatNode) initRequestedItemsHandler() {
	thn.RequestedItemsHandler = timecache.NewTimeCache(roundDuration)
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
		HeartbeatExpiryTimespanInSec: 60,
		PeerID:                       thn.Messenger.ID(),
	}

	thn.createPeerAuthInterceptor(argsFactory)
	thn.createHeartbeatInterceptor(argsFactory)
	thn.createValidatorInfoInterceptor(argsFactory)
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
	thn.HeartbeatInterceptor = thn.initMultiDataInterceptor(identifierHeartbeat, hbFactory, hbProcessor)
}

func (thn *TestHeartbeatNode) createValidatorInfoInterceptor(argsFactory interceptorFactory.ArgInterceptedDataFactory) {
	args := interceptorsProcessor.ArgValidatorInfoInterceptorProcessor{
		PeerShardMapper: thn.PeerShardMapper,
	}
	sviProcessor, _ := interceptorsProcessor.NewValidatorInfoInterceptorProcessor(args)
	sviFactory, _ := interceptorFactory.NewInterceptedValidatorInfoFactory(argsFactory)
	thn.ValidatorInfoInterceptor = thn.initSingleDataInterceptor(common.ConnectionTopic, sviFactory, sviProcessor)
}

func (thn *TestHeartbeatNode) initMultiDataInterceptor(topic string, dataFactory process.InterceptedDataFactory, processor process.InterceptorProcessor) *interceptors.MultiDataInterceptor {
	mdInterceptor, _ := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:            topic,
			Marshalizer:      testscommon.MarshalizerMock{},
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
			CurrentPeerId:        thn.Messenger.ID(),
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
			CurrentPeerId:        thn.Messenger.ID(),
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
		MessagesInChunk:         messagesInChunk,
		MinPeersThreshold:       minPeersThreshold,
		DelayBetweenRequests:    delayBetweenRequests,
		MaxTimeout:              maxTimeout,
		MaxMissingKeysInRequest: maxMissingKeysInRequest,
		Randomizer:              &random.ConcurrentSafeIntRandomizer{},
	}
	thn.RequestsProcessor, _ = processor.NewPeerAuthenticationRequestsProcessor(args)
}

func (thn *TestHeartbeatNode) initDirectConnectionsProcessor() {
	args := processor.ArgDirectConnectionsProcessor{
		Messenger:                 thn.Messenger,
		Marshaller:                testscommon.MarshalizerMock{},
		ShardCoordinator:          thn.ShardCoordinator,
		DelayBetweenNotifications: 5 * time.Second,
	}

	thn.DirectConnectionsProcessor, _ = processor.NewDirectConnectionsProcessor(args)
}

// ConnectTo will try to initiate a connection to the provided parameter
func (thn *TestHeartbeatNode) ConnectTo(connectable Connectable) error {
	if check.IfNil(connectable) {
		return fmt.Errorf("trying to connect to a nil Connectable parameter")
	}

	return thn.Messenger.ConnectToPeer(connectable.GetConnectableAddress())
}

// GetConnectableAddress returns a non circuit, non windows default connectable p2p address
func (thn *TestHeartbeatNode) GetConnectableAddress() string {
	if thn == nil {
		return "nil"
	}

	return GetConnectableAddress(thn.Messenger)
}

// MakeDisplayTableForHeartbeatNodes returns a string containing counters for received messages for all provided test nodes
func MakeDisplayTableForHeartbeatNodes(nodes map[uint32][]*TestHeartbeatNode) string {
	header := []string{"pk", "pid", "shard ID", "messages global", "messages intra", "messages cross", "conns Total/IntraVal/CrossVal/IntraObs/CrossObs/FullObs/Unk/Sed"}
	dataLines := make([]*display.LineData, 0)

	for shardId, nodesList := range nodes {
		for _, n := range nodesList {
			buffPk, _ := n.NodeKeys.Pk.ToByteArray()

			peerInfo := n.Messenger.GetConnectedPeersInfo()

			pid := n.Messenger.ID().Pretty()
			lineData := display.NewLineData(
				false,
				[]string{
					core.GetTrimmedPk(hex.EncodeToString(buffPk)),
					pid[len(pid)-6:],
					fmt.Sprintf("%d", shardId),
					fmt.Sprintf("%d", n.CountGlobalMessages()),
					fmt.Sprintf("%d", n.CountIntraShardMessages()),
					fmt.Sprintf("%d", n.CountCrossShardMessages()),
					fmt.Sprintf("%d/%d/%d/%d/%d/%d/%d/%d",
						len(n.Messenger.ConnectedPeers()),
						peerInfo.NumIntraShardValidators,
						peerInfo.NumCrossShardValidators,
						peerInfo.NumIntraShardObservers,
						peerInfo.NumCrossShardObservers,
						peerInfo.NumFullHistoryObservers,
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
	err := thn.Messenger.CreateTopic(topic, true)
	if err != nil {
		fmt.Printf("error while creating topic %s: %s\n", topic, err.Error())
		return
	}

	err = thn.Messenger.RegisterMessageProcessor(topic, "test", processor)
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
	_ = thn.ResolversContainer.Close()
	_ = thn.DirectConnectionsProcessor.Close()
	_ = thn.Messenger.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (thn *TestHeartbeatNode) IsInterfaceNil() bool {
	return thn == nil
}
