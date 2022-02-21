package heartbeatV2

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	dataRetrieverInterface "github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	heartbeatProcessor "github.com/ElrondNetwork/elrond-go/heartbeat/processor"
	"github.com/ElrondNetwork/elrond-go/heartbeat/sender"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	testsMock "github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	interceptorsProcessor "github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	processMock "github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/stretchr/testify/assert"
)

const (
	defaultNodeName           = "node"
	timeBetweenPeerAuths      = 10 * time.Second
	timeBetweenHeartbeats     = 2 * time.Second
	timeBetweenSendsWhenError = time.Second
	thresholdBetweenSends     = 0.2
)

func TestHeartbeatV2_AllPeersSendMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes, pks, senders, dataPools, processors := createAndStartNodes(interactingNodes)
	assert.Equal(t, interactingNodes, len(nodes))
	assert.Equal(t, interactingNodes, len(pks))
	assert.Equal(t, interactingNodes, len(senders))
	assert.Equal(t, interactingNodes, len(dataPools))
	assert.Equal(t, interactingNodes, len(processors))

	// Wait for messages to broadcast
	time.Sleep(5 * time.Second)

	for i := 0; i < interactingNodes; i++ {
		paCache := dataPools[i].PeerAuthentications()
		hbCache := dataPools[i].Heartbeats()

		assert.Equal(t, interactingNodes, len(paCache.Keys()))
		assert.Equal(t, interactingNodes, len(hbCache.Keys()))

		// Check this node received messages from all peers
		for _, node := range nodes {
			assert.True(t, paCache.Has(node.ID().Bytes()))
			assert.True(t, hbCache.Has(node.ID().Bytes()))
		}
	}
}

func createAndStartNodes(interactingNodes int) ([]p2p.Messenger,
	[]crypto.PublicKey,
	[]factory.HeartbeatV2Sender,
	[]dataRetrieverInterface.PoolsHolder,
	[]factory.PeerAuthenticationRequestsProcessor,
) {
	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	sigHandler := createMockPeerSignatureHandler(keyGen)

	nodes := make([]p2p.Messenger, interactingNodes)
	pks := make([]crypto.PublicKey, interactingNodes)
	senders := make([]factory.HeartbeatV2Sender, interactingNodes)
	dataPools := make([]dataRetrieverInterface.PoolsHolder, interactingNodes)

	// Create and connect messengers
	for i := 0; i < interactingNodes; i++ {
		nodes[i] = integrationTests.CreateMessengerWithNoDiscovery()
		connectNodeToPeers(nodes[i], nodes[:i])
	}

	// Create data interceptors, senders
	// new for loop is needed as peers must be connected before sender creation
	for i := 0; i < interactingNodes; i++ {
		dataPools[i] = dataRetriever.NewPoolsHolderMock()
		createPeerAuthMultiDataInterceptor(nodes[i], dataPools[i].PeerAuthentications(), sigHandler)
		createHeartbeatMultiDataInterceptor(nodes[i], dataPools[i].Heartbeats(), sigHandler)

		nodeName := fmt.Sprintf("%s%d", defaultNodeName, i)
		sk, pk := keyGen.GeneratePair()
		pks[i] = pk

		s := createSender(nodeName, nodes[i], sigHandler, sk)
		senders[i] = s

	}

	/*pksArray := make([][]byte, 0)
	for i := 0; i < interactingNodes; i++ {
		pk, _ := pks[i].ToByteArray()
		pksArray = append(pksArray, pk)
	}
	for i := 0; i < interactingNodes; i++ {
		// processors[i] = createRequestProcessor(pksArray, nodes[i], dataPools[i])
	}*/
	processors := make([]factory.PeerAuthenticationRequestsProcessor, interactingNodes)

	return nodes, pks, senders, dataPools, processors
}

func connectNodeToPeers(node p2p.Messenger, peers []p2p.Messenger) {
	for _, peer := range peers {
		_ = peer.ConnectToPeer(integrationTests.GetConnectableAddress(node))
	}
}

func createSender(nodeName string, messenger p2p.Messenger, peerSigHandler crypto.PeerSignatureHandler, sk crypto.PrivateKey) factory.HeartbeatV2Sender {
	argsSender := sender.ArgSender{
		Messenger:                          messenger,
		Marshaller:                         testscommon.MarshalizerMock{},
		PeerAuthenticationTopic:            common.PeerAuthenticationTopic,
		HeartbeatTopic:                     common.HeartbeatV2Topic,
		PeerAuthenticationTimeBetweenSends: timeBetweenPeerAuths,
		PeerAuthenticationTimeBetweenSendsWhenError: timeBetweenSendsWhenError,
		PeerAuthenticationThresholdBetweenSends:     thresholdBetweenSends,
		HeartbeatTimeBetweenSends:                   timeBetweenHeartbeats,
		HeartbeatTimeBetweenSendsWhenError:          timeBetweenSendsWhenError,
		HeartbeatThresholdBetweenSends:              thresholdBetweenSends,
		VersionNumber:                               "v01",
		NodeDisplayName:                             nodeName,
		Identity:                                    nodeName + "_identity",
		PeerSubType:                                 core.RegularPeer,
		CurrentBlockProvider:                        &testscommon.ChainHandlerStub{},
		PeerSignatureHandler:                        peerSigHandler,
		PrivateKey:                                  sk,
		RedundancyHandler:                           &mock.RedundancyHandlerStub{},
	}

	msgsSender, _ := sender.NewSender(argsSender)
	return msgsSender
}

func createRequestProcessor(pks [][]byte, messenger p2p.Messenger,
	dataPools dataRetrieverInterface.PoolsHolder,
) factory.PeerAuthenticationRequestsProcessor {

	dataPacker, _ := partitioning.NewSimpleDataPacker(&testscommon.MarshalizerMock{})
	shardCoordinator := &sharding.OneShardCoordinator{}
	trieStorageManager, _ := integrationTests.CreateTrieStorageManager(testscommon.CreateMemUnit())
	trieContainer := state.NewDataTriesHolder()

	_, stateTrie := integrationTests.CreateAccountsDB(integrationTests.UserAccount, trieStorageManager)
	trieContainer.Put([]byte(trieFactory.UserAccountTrie), stateTrie)

	_, peerTrie := integrationTests.CreateAccountsDB(integrationTests.ValidatorAccount, trieStorageManager)
	trieContainer.Put([]byte(trieFactory.PeerAccountTrie), peerTrie)

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[trieFactory.UserAccountTrie] = trieStorageManager
	trieStorageManagers[trieFactory.PeerAccountTrie] = trieStorageManager

	resolverContainerFactory := resolverscontainer.FactoryArgs{
		ShardCoordinator:            shardCoordinator,
		Messenger:                   messenger,
		Store:                       integrationTests.CreateStore(2),
		Marshalizer:                 &testscommon.MarshalizerMock{},
		DataPools:                   dataPools,
		Uint64ByteSliceConverter:    integrationTests.TestUint64Converter,
		DataPacker:                  dataPacker,
		TriesContainer:              trieContainer,
		SizeCheckDelta:              100,
		InputAntifloodHandler:       &testsMock.NilAntifloodHandler{},
		OutputAntifloodHandler:      &testsMock.NilAntifloodHandler{},
		NumConcurrentResolvingJobs:  10,
		CurrentNetworkEpochProvider: &testsMock.CurrentNetworkEpochProviderStub{},
		PreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		ResolverConfig: config.ResolverConfig{
			NumCrossShardPeers:  2,
			NumIntraShardPeers:  1,
			NumFullHistoryPeers: 3,
		},
		NodesCoordinator: &processMock.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
				pksMap := make(map[uint32][][]byte, 1)
				pksMap[0] = pks
				return pksMap, nil
			},
		},
		MaxNumOfPeerAuthenticationInResponse: 10,
	}
	resolversContainerFactory, _ := resolverscontainer.NewShardResolversContainerFactory(resolverContainerFactory)

	resolversContainer, _ := resolversContainerFactory.Create()
	resolverFinder, _ := containers.NewResolversFinder(resolversContainer, shardCoordinator)
	whitelistHandler := &testscommon.WhiteListHandlerStub{
		IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
			return true
		},
	}
	requestedItemsHandler := timecache.NewTimeCache(5 * time.Second)
	requestHandler, _ := requestHandlers.NewResolverRequestHandler(
		resolverFinder,
		requestedItemsHandler,
		whitelistHandler,
		100,
		shardCoordinator.SelfId(),
		time.Second,
	)

	argsProcessor := heartbeatProcessor.ArgPeerAuthenticationRequestsProcessor{
		RequestHandler: requestHandler,
		NodesCoordinator: &processMock.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
				pksMap := make(map[uint32][][]byte, 1)
				pksMap[0] = pks
				return pksMap, nil
			},
		},
		PeerAuthenticationPool:  dataPools.PeerAuthentications(),
		ShardId:                 0,
		Epoch:                   0,
		MessagesInChunk:         10,
		MinPeersThreshold:       1.0,
		DelayBetweenRequests:    2 * time.Second,
		MaxTimeout:              10 * time.Second,
		MaxMissingKeysInRequest: 5,
		Randomizer:              &random.ConcurrentSafeIntRandomizer{},
	}

	requestProcessor, _ := heartbeatProcessor.NewPeerAuthenticationRequestsProcessor(argsProcessor)
	return requestProcessor
}

func createPeerAuthMultiDataInterceptor(messenger p2p.Messenger, peerAuthCacher storage.Cacher, sigHandler crypto.PeerSignatureHandler) {
	argProcessor := interceptorsProcessor.ArgPeerAuthenticationInterceptorProcessor{
		PeerAuthenticationCacher: peerAuthCacher,
	}
	paProcessor, _ := interceptorsProcessor.NewPeerAuthenticationInterceptorProcessor(argProcessor)

	args := createMockInterceptedDataFactoryArgs(sigHandler, messenger.ID())
	paFactory, _ := interceptorFactory.NewInterceptedPeerAuthenticationDataFactory(args)

	createMockMultiDataInterceptor(common.PeerAuthenticationTopic, messenger, paFactory, paProcessor)
}

func createHeartbeatMultiDataInterceptor(messenger p2p.Messenger, heartbeatCacher storage.Cacher, sigHandler crypto.PeerSignatureHandler) {
	argProcessor := interceptorsProcessor.ArgHeartbeatInterceptorProcessor{
		HeartbeatCacher: heartbeatCacher,
	}
	hbProcessor, _ := interceptorsProcessor.NewHeartbeatInterceptorProcessor(argProcessor)

	args := createMockInterceptedDataFactoryArgs(sigHandler, messenger.ID())
	hbFactory, _ := interceptorFactory.NewInterceptedHeartbeatDataFactory(args)

	createMockMultiDataInterceptor(common.HeartbeatV2Topic, messenger, hbFactory, hbProcessor)
}

func createMockInterceptedDataFactoryArgs(sigHandler crypto.PeerSignatureHandler, pid core.PeerID) interceptorFactory.ArgInterceptedDataFactory {
	return interceptorFactory.ArgInterceptedDataFactory{
		CoreComponents: &processMock.CoreComponentsMock{
			IntMarsh: &testscommon.MarshalizerMock{},
		},
		NodesCoordinator: &processMock.NodesCoordinatorMock{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator sharding.Validator, shardId uint32, err error) {
				return nil, 0, nil
			},
		},
		PeerSignatureHandler:         sigHandler,
		SignaturesHandler:            &processMock.SignaturesHandlerStub{},
		HeartbeatExpiryTimespanInSec: 10,
		PeerID:                       pid,
	}
}

func createMockMultiDataInterceptor(topic string, messenger p2p.Messenger, dataFactory process.InterceptedDataFactory, processor process.InterceptorProcessor) {
	mdInterceptor, _ := interceptors.NewMultiDataInterceptor(
		interceptors.ArgMultiDataInterceptor{
			Topic:            topic,
			Marshalizer:      testscommon.MarshalizerMock{},
			DataFactory:      dataFactory,
			Processor:        processor,
			Throttler:        createMockThrottler(),
			AntifloodHandler: &testsMock.P2PAntifloodHandlerStub{},
			WhiteListRequest: &testscommon.WhiteListHandlerStub{
				IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
					return true
				},
			},
			PreferredPeersHolder: &p2pmocks.PeersHolderStub{},
			CurrentPeerId:        messenger.ID(),
		},
	)

	_ = messenger.CreateTopic(topic, true)
	_ = messenger.RegisterMessageProcessor(topic, common.DefaultInterceptorsIdentifier, mdInterceptor)
}

func createMockPeerSignatureHandler(keyGen crypto.KeyGenerator) crypto.PeerSignatureHandler {
	singleSigner := singlesig.NewBlsSigner()

	return &mock.PeerSignatureHandlerStub{
		VerifyPeerSignatureCalled: func(pk []byte, pid core.PeerID, signature []byte) error {
			senderPubKey, err := keyGen.PublicKeyFromByteArray(pk)
			if err != nil {
				return err
			}
			return singleSigner.Verify(senderPubKey, pid.Bytes(), signature)
		},
		GetPeerSignatureCalled: func(privateKey crypto.PrivateKey, pid []byte) ([]byte, error) {
			return singleSigner.Sign(privateKey, pid)
		},
	}
}

func createMockThrottler() *processMock.InterceptorThrottlerStub {
	return &processMock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
}
