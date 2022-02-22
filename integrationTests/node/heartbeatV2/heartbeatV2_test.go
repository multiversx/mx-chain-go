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
	"github.com/ElrondNetwork/elrond-go/heartbeat"
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

	// Check sent messages
	maxMessageAgeAllowed := time.Second * 7
	checkMessages(t, nodes, dataPools, maxMessageAgeAllowed)

	closeComponents(t, nodes, senders, dataPools, nil)
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
	time.Sleep(time.Second * 3)

	// Check sent messages
	maxMessageAgeAllowed := time.Second * 5
	checkMessages(t, nodes, dataPools, maxMessageAgeAllowed)

	// Add new delayed node which requests messages
	delayedNode, delayedNodeDataPool := createDelayedNode(nodes, sigHandler)
	nodes = append(nodes, delayedNode)
	dataPools = append(dataPools, delayedNodeDataPool)

	pksArray := make([][]byte, 0)
	for _, node := range nodes {
		pksArray = append(pksArray, node.ID().Bytes())
	}

	// Create resolvers and request chunk from delayed node
	paResolvers := createPeerAuthResolvers(pksArray, nodes, dataPools, shardCoordinator)
	newNodeIndex := len(nodes) - 1
	_ = paResolvers[newNodeIndex].RequestDataFromChunk(0, 0)

	// Wait for messages to broadcast
	time.Sleep(time.Second * 3)

	delayedNodeCache := delayedNodeDataPool.PeerAuthentications()
	assert.Equal(t, len(nodes)-1, delayedNodeCache.Len())

	// Only search for messages from initially created nodes.
	// Last one does not send peerAuthentication yet
	for i := 0; i < len(nodes)-1; i++ {
		assert.True(t, delayedNodeCache.Has(nodes[i].ID().Bytes()))
	}

	// Create sender for last node
	nodeName := fmt.Sprintf("%s%d", defaultNodeName, newNodeIndex)
	sk, _ := keyGen.GeneratePair()
	s := createSender(nodeName, delayedNode, sigHandler, sk)
	senders = append(senders, s)

	// Wait to make sure all peers send messages again
	time.Sleep(time.Second * 3)

	// Check sent messages again - now should have from all peers
	maxMessageAgeAllowed = time.Second * 5 // should not have messages from first Send
	checkMessages(t, nodes, dataPools, maxMessageAgeAllowed)

	// Add new delayed node which requests messages by hash array
	delayedNode, delayedNodeDataPool = createDelayedNode(nodes, sigHandler)
	nodes = append(nodes, delayedNode)
	dataPools = append(dataPools, delayedNodeDataPool)
	delayedNodeResolver := createPeerAuthResolver(pksArray, delayedNodeDataPool.PeerAuthentications(), delayedNode, shardCoordinator)
	_ = delayedNodeResolver.RequestDataFromHashArray(pksArray, 0)

	// Wait for messages to broadcast
	time.Sleep(time.Second * 3)

	// Check that the node received peer auths from all of them
	assert.Equal(t, len(nodes)-1, delayedNodeDataPool.PeerAuthentications().Len())
	for _, node := range nodes {
		assert.True(t, delayedNodeDataPool.PeerAuthentications().Has(node.ID().Bytes()))
	}

	closeComponents(t, nodes, senders, dataPools, paResolvers)
}

func TestHeartbeatV2_NetworkShouldSendMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	sigHandler := createMockPeerSignatureHandler(keyGen)

	nodes, _ := integrationTests.CreateFixedNetworkOf8Peers()
	interactingNodes := len(nodes)

	// Create components
	dataPools := make([]dataRetrieverInterface.PoolsHolder, interactingNodes)
	senders := make([]factory.HeartbeatV2Sender, interactingNodes)
	for i := 0; i < interactingNodes; i++ {
		dataPools[i] = dataRetriever.NewPoolsHolderMock()
		createPeerAuthMultiDataInterceptor(nodes[i], dataPools[i].PeerAuthentications(), sigHandler)
		createHeartbeatMultiDataInterceptor(nodes[i], dataPools[i].Heartbeats(), sigHandler)

		nodeName := fmt.Sprintf("%s%d", defaultNodeName, i)
		sk, _ := keyGen.GeneratePair()

		s := createSender(nodeName, nodes[i], sigHandler, sk)
		senders[i] = s
	}

	// Wait for all peers to send peer auth messages twice
	time.Sleep(time.Second * 15)

	checkMessages(t, nodes, dataPools, time.Second*7)

	closeComponents(t, nodes, senders, dataPools, nil)
}

func createDelayedNode(nodes []p2p.Messenger, sigHandler crypto.PeerSignatureHandler) (p2p.Messenger, dataRetrieverInterface.PoolsHolder) {
	node := integrationTests.CreateMessengerWithNoDiscovery()
	connectNodeToPeers(node, nodes)

	// Wait for last peer to join
	time.Sleep(time.Second * 2)

	dataPool := dataRetriever.NewPoolsHolderMock()

	// Create multi data interceptors for the delayed node in order to receive messages
	createPeerAuthMultiDataInterceptor(node, dataPool.PeerAuthentications(), sigHandler)
	createHeartbeatMultiDataInterceptor(node, dataPool.Heartbeats(), sigHandler)

	return node, dataPool
}

func checkMessages(t *testing.T, nodes []p2p.Messenger, dataPools []dataRetrieverInterface.PoolsHolder, maxMessageAgeAllowed time.Duration) {
	numOfNodes := len(nodes)
	for i := 0; i < numOfNodes; i++ {
		paCache := dataPools[i].PeerAuthentications()
		hbCache := dataPools[i].Heartbeats()

		assert.Equal(t, numOfNodes, paCache.Len())
		assert.Equal(t, numOfNodes, hbCache.Len())

		// Check this node received messages from all peers
		for _, node := range nodes {
			assert.True(t, paCache.Has(node.ID().Bytes()))
			assert.True(t, hbCache.Has(node.ID().Bytes()))

			// Also check message age
			value, _ := paCache.Get(node.ID().Bytes())
			msg := value.(heartbeat.PeerAuthentication)

			marshaller := testscommon.MarshalizerMock{}
			payload := &heartbeat.Payload{}
			err := marshaller.Unmarshal(payload, msg.Payload)
			assert.Nil(t, err)

			currentTimestamp := time.Now().Unix()
			messageAge := time.Duration(currentTimestamp - payload.Timestamp)
			assert.True(t, messageAge < maxMessageAgeAllowed)
		}
	}
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
	nodes []p2p.Messenger,
	senders []factory.HeartbeatV2Sender,
	dataPools []dataRetrieverInterface.PoolsHolder,
	resolvers []dataRetrieverInterface.PeerAuthenticationResolver) {
	interactingNodes := len(nodes)
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
