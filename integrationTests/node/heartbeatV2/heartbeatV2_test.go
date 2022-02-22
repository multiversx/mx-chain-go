package heartbeatV2

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/common"
	dataRetrieverInterface "github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
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
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
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

	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	sigHandler := createMockPeerSignatureHandler(keyGen)

	interactingNodes := 3
	nodes, senders, dataPools := createAndStartNodes(interactingNodes, keyGen, sigHandler)
	assert.Equal(t, interactingNodes, len(nodes))
	assert.Equal(t, interactingNodes, len(senders))
	assert.Equal(t, interactingNodes, len(dataPools))

	// Wait for messages to broadcast
	time.Sleep(time.Second * 5)

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

	closeComponents(t, interactingNodes, nodes, senders, dataPools, nil)
}

func TestHeartbeatV2_PeerJoiningLate(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	sigHandler := createMockPeerSignatureHandler(keyGen)
	shardCoordinator := &sharding.OneShardCoordinator{}

	interactingNodes := 3
	nodes, senders, dataPools := createAndStartNodes(interactingNodes, keyGen, sigHandler)
	assert.Equal(t, interactingNodes, len(nodes))
	assert.Equal(t, interactingNodes, len(senders))
	assert.Equal(t, interactingNodes, len(dataPools))

	// Wait for messages to broadcast
	time.Sleep(time.Second * 5)

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

	// Add new delayed node which requests messages
	newNodeIndex := len(nodes)
	nodes = append(nodes, integrationTests.CreateMessengerWithNoDiscovery())
	connectNodeToPeers(nodes[newNodeIndex], nodes[:newNodeIndex])

	dataPools = append(dataPools, dataRetriever.NewPoolsHolderMock())

	pksArray := make([][]byte, 0)
	for _, node := range nodes {
		pksArray = append(pksArray, node.ID().Bytes())
	}

	// Create multi data interceptor for the delayed node in order to process requested messages
	createPeerAuthMultiDataInterceptor(nodes[newNodeIndex], dataPools[newNodeIndex].PeerAuthentications(), sigHandler)

	// Create resolver and request chunk
	paResolvers := createPeerAuthResolvers(pksArray, nodes, dataPools, shardCoordinator)
	_ = paResolvers[newNodeIndex].RequestDataFromChunk(0, 0)

	time.Sleep(time.Second * 5)

	delayedNodeCache := dataPools[newNodeIndex].PeerAuthentications()
	keysInDelayedNodeCache := delayedNodeCache.Keys()
	assert.Equal(t, len(nodes)-1, len(keysInDelayedNodeCache))

	// Only search for messages from initially created nodes.
	// Last one does not send peerAuthentication
	for i := 0; i < len(nodes)-1; i++ {
		assert.True(t, delayedNodeCache.Has(nodes[i].ID().Bytes()))
	}

	closeComponents(t, interactingNodes, nodes, senders, dataPools, paResolvers)
}

func createAndStartNodes(interactingNodes int, keyGen crypto.KeyGenerator, sigHandler crypto.PeerSignatureHandler) (
	[]p2p.Messenger,
	[]factory.HeartbeatV2Sender,
	[]dataRetrieverInterface.PoolsHolder,
) {
	nodes := make([]p2p.Messenger, interactingNodes)
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
		sk, _ := keyGen.GeneratePair()

		s := createSender(nodeName, nodes[i], sigHandler, sk)
		senders[i] = s
	}

	return nodes, senders, dataPools
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

func createPeerAuthResolvers(pks [][]byte, nodes []p2p.Messenger, dataPools []dataRetrieverInterface.PoolsHolder, shardCoordinator sharding.Coordinator) []dataRetrieverInterface.PeerAuthenticationResolver {
	paResolvers := make([]dataRetrieverInterface.PeerAuthenticationResolver, len(nodes))
	for idx, node := range nodes {
		paResolvers[idx] = createPeerAuthResolver(pks, dataPools[idx].PeerAuthentications(), node, shardCoordinator)
	}

	return paResolvers
}

func createPeerAuthResolver(pks [][]byte, peerAuthPool storage.Cacher, messenger p2p.Messenger, shardCoordinator sharding.Coordinator) dataRetrieverInterface.PeerAuthenticationResolver {
	intraShardTopic := common.ConsensusTopic +
		shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())

	peerListCreator, _ := topicResolverSender.NewDiffPeerListCreator(messenger, common.PeerAuthenticationTopic, intraShardTopic, "")

	argsTopicResolverSender := topicResolverSender.ArgTopicResolverSender{
		Messenger:                   messenger,
		TopicName:                   common.PeerAuthenticationTopic,
		PeerListCreator:             peerListCreator,
		Marshalizer:                 &testscommon.MarshalizerMock{},
		Randomizer:                  &random.ConcurrentSafeIntRandomizer{},
		TargetShardId:               shardCoordinator.SelfId(),
		OutputAntiflooder:           &testsMock.NilAntifloodHandler{},
		NumCrossShardPeers:          len(pks),
		NumIntraShardPeers:          1,
		NumFullHistoryPeers:         3,
		CurrentNetworkEpochProvider: &testsMock.CurrentNetworkEpochProviderStub{},
		PreferredPeersHolder:        &p2pmocks.PeersHolderStub{},
		SelfShardIdProvider:         shardCoordinator,
	}
	resolverSender, _ := topicResolverSender.NewTopicResolverSender(argsTopicResolverSender)

	argsPAResolver := resolvers.ArgPeerAuthenticationResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshalizer:      &testscommon.MarshalizerMock{},
			AntifloodHandler: &testsMock.NilAntifloodHandler{},
			Throttler:        createMockThrottler(),
		},
		PeerAuthenticationPool:               peerAuthPool,
		NodesCoordinator:                     createMockNodesCoordinator(pks),
		MaxNumOfPeerAuthenticationInResponse: 10,
	}
	peerAuthResolver, _ := resolvers.NewPeerAuthenticationResolver(argsPAResolver)

	_ = messenger.CreateTopic(peerAuthResolver.RequestTopic(), true)
	_ = messenger.RegisterMessageProcessor(peerAuthResolver.RequestTopic(), common.DefaultResolversIdentifier, peerAuthResolver)

	return peerAuthResolver
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

func createMockNodesCoordinator(pks [][]byte) dataRetrieverInterface.NodesCoordinator {
	return &processMock.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
			pksMap := make(map[uint32][][]byte, 1)
			pksMap[0] = pks
			return pksMap, nil
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

func closeComponents(t *testing.T,
	interactingNodes int,
	nodes []p2p.Messenger,
	senders []factory.HeartbeatV2Sender,
	dataPools []dataRetrieverInterface.PoolsHolder,
	resolvers []dataRetrieverInterface.PeerAuthenticationResolver) {
	for i := 0; i < interactingNodes; i++ {
		var err error
		if senders != nil && len(senders) > i {
			err = senders[i].Close()
			assert.Nil(t, err)
		}

		if dataPools != nil && len(dataPools) > i {
			err = dataPools[i].Close()
			assert.Nil(t, err)
		}

		if resolvers != nil && len(resolvers) > i {
			err = resolvers[i].Close()
			assert.Nil(t, err)
		}

		if nodes != nil && len(nodes) > i {
			err = nodes[i].Close()
			assert.Nil(t, err)
		}
	}
}
