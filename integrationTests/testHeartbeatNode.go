package integrationTests

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
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
	"github.com/ElrondNetwork/elrond-go/factory"
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
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
)

const (
	defaultNodeName           = "heartbeatNode"
	timeBetweenPeerAuths      = 10 * time.Second
	timeBetweenHeartbeats     = 2 * time.Second
	timeBetweenSendsWhenError = time.Second
	thresholdBetweenSends     = 0.2

	messagesInChunk         = 10
	minPeersThreshold       = 1.0
	delayBetweenRequests    = time.Second
	maxTimeout              = time.Minute
	maxMissingKeysInRequest = 1
)

// TestMarshaller represents the main marshaller
var TestMarshaller = &testscommon.MarshalizerMock{}

var TestThrottler = &processMock.InterceptorThrottlerStub{
	CanProcessCalled: func() bool {
		return true
	},
}

// TestHeartbeatNode represents a container type of class used in integration tests
// with all its fields exported
type TestHeartbeatNode struct {
	ShardCoordinator      sharding.Coordinator
	NodesCoordinator      sharding.NodesCoordinator
	PeerShardMapper       process.PeerShardMapper
	Messenger             p2p.Messenger
	NodeKeys              TestKeyPair
	DataPool              dataRetriever.PoolsHolder
	Sender                factory.HeartbeatV2Sender
	PeerAuthInterceptor   *interceptors.MultiDataInterceptor
	HeartbeatInterceptor  *interceptors.MultiDataInterceptor
	PeerAuthResolver      dataRetriever.PeerAuthenticationResolver
	PeerSigHandler        crypto.PeerSignatureHandler
	WhiteListHandler      process.WhiteListHandler
	Storage               dataRetriever.StorageService
	ResolversContainer    dataRetriever.ResolversContainer
	ResolverFinder        dataRetriever.ResolversFinder
	RequestHandler        process.RequestHandler
	RequestedItemsHandler dataRetriever.RequestedItemsHandler
	RequestsProcessor     factory.PeerAuthenticationRequestsProcessor
}

// NewTestHeartbeatNode returns a new TestHeartbeatNode instance with a libp2p messenger
func NewTestHeartbeatNode(
	maxShards uint32,
	nodeShardId uint32,
	minPeersWaiting int,
) *TestHeartbeatNode {
	keygen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	sk, pk := keygen.GeneratePair()

	pksBytes := make(map[uint32][]byte, maxShards)
	pksBytes[nodeShardId], _ = pk.ToByteArray()

	nodesCoordinator := &mock.NodesCoordinatorMock{
		GetAllValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
			keys := make(map[uint32][][]byte)
			for shardID := uint32(0); shardID < maxShards; shardID++ {
				keys[shardID] = append(keys[shardID], pksBytes[shardID])
			}

			shardID := core.MetachainShardId
			keys[shardID] = append(keys[shardID], pksBytes[shardID])

			return keys, nil
		},
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (sharding.Validator, uint32, error) {
			validator, _ := sharding.NewValidator(publicKey, defaultChancesSelection, 1)
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

	messenger := CreateMessengerWithNoDiscovery()
	peerShardMapper := mock.NewNetworkShardingCollectorMock()

	thn := &TestHeartbeatNode{
		ShardCoordinator: shardCoordinator,
		NodesCoordinator: nodesCoordinator,
		Messenger:        messenger,
		PeerSigHandler:   peerSigHandler,
		PeerShardMapper:  peerShardMapper,
	}

	thn.NodeKeys = TestKeyPair{
		Sk: sk,
		Pk: pk,
	}

	// start a go routine in order to allow peers to connect first
	go thn.initTestHeartbeatNode(minPeersWaiting)

	return thn
}

func (thn *TestHeartbeatNode) initTestHeartbeatNode(minPeersWaiting int) {
	thn.initStorage()
	thn.initDataPools()
	thn.initRequestedItemsHandler()
	thn.initResolvers()
	thn.initInterceptors()

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
		Messenger:                          thn.Messenger,
		Marshaller:                         TestMarshaller,
		PeerAuthenticationTopic:            common.PeerAuthenticationTopic,
		HeartbeatTopic:                     identifierHeartbeat,
		PeerAuthenticationTimeBetweenSends: timeBetweenPeerAuths,
		PeerAuthenticationTimeBetweenSendsWhenError: timeBetweenSendsWhenError,
		PeerAuthenticationThresholdBetweenSends:     thresholdBetweenSends,
		HeartbeatTimeBetweenSends:                   timeBetweenHeartbeats,
		HeartbeatTimeBetweenSendsWhenError:          timeBetweenSendsWhenError,
		HeartbeatThresholdBetweenSends:              thresholdBetweenSends,
		VersionNumber:                               "v01",
		NodeDisplayName:                             defaultNodeName,
		Identity:                                    defaultNodeName + "_identity",
		PeerSubType:                                 core.RegularPeer,
		CurrentBlockProvider:                        &testscommon.ChainHandlerStub{},
		PeerSignatureHandler:                        thn.PeerSigHandler,
		PrivateKey:                                  thn.NodeKeys.Sk,
		RedundancyHandler:                           &mock.RedundancyHandlerStub{},
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
	}

	var err error
	if thn.ShardCoordinator.SelfId() == core.MetachainShardId {
		resolversContainerFactory, _ := resolverscontainer.NewMetaResolversContainerFactory(resolverContainerFactory)

		thn.ResolversContainer, err = resolversContainerFactory.Create()
		log.LogIfError(err)

		thn.ResolverFinder, _ = containers.NewResolversFinder(thn.ResolversContainer, thn.ShardCoordinator)
		thn.RequestHandler, _ = requestHandlers.NewResolverRequestHandler(
			thn.ResolverFinder,
			thn.RequestedItemsHandler,
			thn.WhiteListHandler,
			100,
			thn.ShardCoordinator.SelfId(),
			time.Second,
		)
	} else {
		resolversContainerFactory, _ := resolverscontainer.NewShardResolversContainerFactory(resolverContainerFactory)

		thn.ResolversContainer, err = resolversContainerFactory.Create()
		log.LogIfError(err)

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
}

func (thn *TestHeartbeatNode) initRequestedItemsHandler() {
	thn.RequestedItemsHandler = timecache.NewTimeCache(roundDuration)
}

func (thn *TestHeartbeatNode) initInterceptors() {
	argsFactory := interceptorFactory.ArgInterceptedDataFactory{
		CoreComponents: &processMock.CoreComponentsMock{
			IntMarsh: TestMarshaller,
		},
		NodesCoordinator:             thn.NodesCoordinator,
		PeerSignatureHandler:         thn.PeerSigHandler,
		SignaturesHandler:            &processMock.SignaturesHandlerStub{},
		HeartbeatExpiryTimespanInSec: 10,
		PeerID:                       thn.Messenger.ID(),
	}

	// PeerAuthentication interceptor
	argPAProcessor := interceptorsProcessor.ArgPeerAuthenticationInterceptorProcessor{
		PeerAuthenticationCacher: thn.DataPool.PeerAuthentications(),
		PeerShardMapper:          thn.PeerShardMapper,
	}
	paProcessor, _ := interceptorsProcessor.NewPeerAuthenticationInterceptorProcessor(argPAProcessor)
	paFactory, _ := interceptorFactory.NewInterceptedPeerAuthenticationDataFactory(argsFactory)
	thn.PeerAuthInterceptor = thn.initMultiDataInterceptor(common.PeerAuthenticationTopic, paFactory, paProcessor)

	// Heartbeat interceptor
	argHBProcessor := interceptorsProcessor.ArgHeartbeatInterceptorProcessor{
		HeartbeatCacher: thn.DataPool.Heartbeats(),
	}
	hbProcessor, _ := interceptorsProcessor.NewHeartbeatInterceptorProcessor(argHBProcessor)
	hbFactory, _ := interceptorFactory.NewInterceptedHeartbeatDataFactory(argsFactory)
	identifierHeartbeat := common.HeartbeatV2Topic + thn.ShardCoordinator.CommunicationIdentifier(thn.ShardCoordinator.SelfId())
	thn.HeartbeatInterceptor = thn.initMultiDataInterceptor(identifierHeartbeat, hbFactory, hbProcessor)
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

	_ = thn.Messenger.CreateTopic(topic, true)
	_ = thn.Messenger.RegisterMessageProcessor(topic, common.DefaultInterceptorsIdentifier, mdInterceptor)

	return mdInterceptor
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

// Close -
func (thn *TestHeartbeatNode) Close() {
	_ = thn.Sender.Close()
	_ = thn.PeerAuthInterceptor.Close()
	_ = thn.RequestsProcessor.Close()
	_ = thn.ResolversContainer.Close()
	_ = thn.Messenger.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (thn *TestHeartbeatNode) IsInterfaceNil() bool {
	return thn == nil
}
