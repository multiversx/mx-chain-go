package factory

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers/requesters"
	"github.com/multiversx/mx-chain-go/dataRetriever/topicSender"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/update"
)

const (
	numCrossShardPeers  = 2
	numIntraShardPeers  = 2
	numFullHistoryPeers = 3
)

type requestersContainerFactory struct {
	shardCoordinator       sharding.Coordinator
	mainMessenger          p2p.Messenger
	fullArchiveMessenger   p2p.Messenger
	marshaller             marshal.Marshalizer
	intRandomizer          dataRetriever.IntRandomizer
	container              dataRetriever.RequestersContainer
	outputAntifloodHandler dataRetriever.P2PAntifloodHandler
	peersRatingHandler     dataRetriever.PeersRatingHandler
}

// ArgsRequestersContainerFactory defines the arguments for the requestersContainerFactory constructor
type ArgsRequestersContainerFactory struct {
	ShardCoordinator       sharding.Coordinator
	MainMessenger          p2p.Messenger
	FullArchiveMessenger   p2p.Messenger
	Marshaller             marshal.Marshalizer
	ExistingRequesters     dataRetriever.RequestersContainer
	OutputAntifloodHandler dataRetriever.P2PAntifloodHandler
	PeersRatingHandler     dataRetriever.PeersRatingHandler
}

// NewRequestersContainerFactory creates a new container filled with topic requesters
func NewRequestersContainerFactory(args ArgsRequestersContainerFactory) (*requestersContainerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.MainMessenger) {
		return nil, fmt.Errorf("%w on main network", update.ErrNilMessenger)
	}
	if check.IfNil(args.FullArchiveMessenger) {
		return nil, fmt.Errorf("%w on full archive network", update.ErrNilMessenger)
	}
	if check.IfNil(args.Marshaller) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.ExistingRequesters) {
		return nil, update.ErrNilRequestersContainer
	}
	if check.IfNil(args.OutputAntifloodHandler) {
		return nil, update.ErrNilAntiFloodHandler
	}
	if check.IfNil(args.PeersRatingHandler) {
		return nil, update.ErrNilPeersRatingHandler
	}

	return &requestersContainerFactory{
		shardCoordinator:       args.ShardCoordinator,
		mainMessenger:          args.MainMessenger,
		fullArchiveMessenger:   args.FullArchiveMessenger,
		marshaller:             args.Marshaller,
		intRandomizer:          &random.ConcurrentSafeIntRandomizer{},
		container:              args.ExistingRequesters,
		outputAntifloodHandler: args.OutputAntifloodHandler,
		peersRatingHandler:     args.PeersRatingHandler,
	}, nil
}

// Create returns a requesters container that will hold all requesters in the system
func (rcf *requestersContainerFactory) Create() (dataRetriever.RequestersContainer, error) {
	err := rcf.generateTrieNodesRequesters()
	if err != nil {
		return nil, err
	}

	return rcf.container, nil
}

func (rcf *requestersContainerFactory) generateTrieNodesRequesters() error {
	shardC := rcf.shardCoordinator

	keys := make([]string, 0)
	requestersSlice := make([]dataRetriever.Requester, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(i, core.MetachainShardId)
		if rcf.checkIfRequesterExists(identifierTrieNodes) {
			continue
		}

		requester, err := rcf.createTrieNodesRequester(identifierTrieNodes, i)
		if err != nil {
			return err
		}

		requestersSlice = append(requestersSlice, requester)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !rcf.checkIfRequesterExists(identifierTrieNodes) {
		requester, err := rcf.createTrieNodesRequester(identifierTrieNodes, core.MetachainShardId)
		if err != nil {
			return err
		}

		requestersSlice = append(requestersSlice, requester)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !rcf.checkIfRequesterExists(identifierTrieNodes) {
		requester, err := rcf.createTrieNodesRequester(identifierTrieNodes, core.MetachainShardId)
		if err != nil {
			return err
		}

		requestersSlice = append(requestersSlice, requester)
		keys = append(keys, identifierTrieNodes)
	}

	return rcf.container.AddMultiple(keys, requestersSlice)
}

func (rcf *requestersContainerFactory) checkIfRequesterExists(topic string) bool {
	_, err := rcf.container.Get(topic)
	return err == nil
}

func (rcf *requestersContainerFactory) createTrieNodesRequester(baseTopic string, targetShardID uint32) (dataRetriever.Requester, error) {
	// for each requester we create a pseudo-intra shard topic as to make at least of half of the requests target the proper peers
	// this pseudo-intra shard topic is the consensus_targetShardID
	targetShardCoordinator, err := sharding.NewMultiShardCoordinator(rcf.shardCoordinator.NumberOfShards(), targetShardID)
	if err != nil {
		return nil, err
	}

	targetConsensusTopic := common.ConsensusTopic + targetShardCoordinator.CommunicationIdentifier(targetShardID)
	peerListCreator, err := topicsender.NewDiffPeerListCreator(
		rcf.mainMessenger,
		baseTopic,
		targetConsensusTopic,
		resolverscontainer.EmptyExcludePeersOnTopic,
	)
	if err != nil {
		return nil, err
	}

	arg := topicsender.ArgTopicRequestSender{
		ArgBaseTopicSender: topicsender.ArgBaseTopicSender{
			MainMessenger:                   rcf.mainMessenger,
			FullArchiveMessenger:            rcf.fullArchiveMessenger,
			TopicName:                       baseTopic,
			OutputAntiflooder:               rcf.outputAntifloodHandler,
			MainPreferredPeersHolder:        disabled.NewPreferredPeersHolder(),
			FullArchivePreferredPeersHolder: disabled.NewPreferredPeersHolder(),
			TargetShardId:                   defaultTargetShardID,
		},
		Marshaller:                  rcf.marshaller,
		Randomizer:                  rcf.intRandomizer,
		PeerListCreator:             peerListCreator,
		NumIntraShardPeers:          numIntraShardPeers,
		NumCrossShardPeers:          numCrossShardPeers,
		NumFullHistoryPeers:         numFullHistoryPeers,
		CurrentNetworkEpochProvider: disabled.NewCurrentNetworkEpochProviderHandler(),
		SelfShardIdProvider:         rcf.shardCoordinator,
		PeersRatingHandler:          rcf.peersRatingHandler,
	}
	requestSender, err := topicsender.NewTopicRequestSender(arg)
	if err != nil {
		return nil, err
	}

	argTrieRequester := requesters.ArgTrieNodeRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    rcf.marshaller,
		},
	}
	return requesters.NewTrieNodeRequester(argTrieRequester)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcf *requestersContainerFactory) IsInterfaceNil() bool {
	return rcf == nil
}
