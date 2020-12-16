package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/random"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	factoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

const defaultTargetShardID = uint32(0)
const numCrossShardPeers = 2
const numIntraShardPeers = 2
const numFullHistoryPeers = 3

type resolversContainerFactory struct {
	shardCoordinator       sharding.Coordinator
	messenger              dataRetriever.TopicMessageHandler
	marshalizer            marshal.Marshalizer
	intRandomizer          dataRetriever.IntRandomizer
	dataTrieContainer      state.TriesHolder
	container              dataRetriever.ResolversContainer
	inputAntifloodHandler  dataRetriever.P2PAntifloodHandler
	outputAntifloodHandler dataRetriever.P2PAntifloodHandler
	throttler              dataRetriever.ResolverThrottler
}

// ArgsNewResolversContainerFactory defines the arguments for the resolversContainerFactory constructor
type ArgsNewResolversContainerFactory struct {
	ShardCoordinator           sharding.Coordinator
	Messenger                  dataRetriever.TopicMessageHandler
	Marshalizer                marshal.Marshalizer
	DataTrieContainer          state.TriesHolder
	ExistingResolvers          dataRetriever.ResolversContainer
	InputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	OutputAntifloodHandler     dataRetriever.P2PAntifloodHandler
	NumConcurrentResolvingJobs int32
}

// NewResolversContainerFactory creates a new container filled with topic resolvers
func NewResolversContainerFactory(args ArgsNewResolversContainerFactory) (*resolversContainerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.Messenger) {
		return nil, update.ErrNilMessenger
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.DataTrieContainer) {
		return nil, update.ErrNilTrieDataGetter
	}
	if check.IfNil(args.ExistingResolvers) {
		return nil, update.ErrNilResolverContainer
	}

	thr, err := throttler.NewNumGoRoutinesThrottler(args.NumConcurrentResolvingJobs)
	if err != nil {
		return nil, err
	}
	return &resolversContainerFactory{
		shardCoordinator:       args.ShardCoordinator,
		messenger:              args.Messenger,
		marshalizer:            args.Marshalizer,
		intRandomizer:          &random.ConcurrentSafeIntRandomizer{},
		dataTrieContainer:      args.DataTrieContainer,
		container:              args.ExistingResolvers,
		inputAntifloodHandler:  args.InputAntifloodHandler,
		outputAntifloodHandler: args.OutputAntifloodHandler,
		throttler:              thr,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	err := rcf.generateTrieNodesResolvers()
	if err != nil {
		return nil, err
	}

	return rcf.container, nil
}

func (rcf *resolversContainerFactory) generateTrieNodesResolvers() error {
	shardC := rcf.shardCoordinator

	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(i, core.MetachainShardId)
		if rcf.checkIfResolverExists(identifierTrieNodes) {
			continue
		}

		trieId := genesis.CreateTrieIdentifier(i, genesis.UserAccount)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, trieId, i)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !rcf.checkIfResolverExists(identifierTrieNodes) {
		trieId := genesis.CreateTrieIdentifier(core.MetachainShardId, genesis.UserAccount)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, trieId, core.MetachainShardId)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !rcf.checkIfResolverExists(identifierTrieNodes) {
		trieID := genesis.CreateTrieIdentifier(core.MetachainShardId, genesis.ValidatorAccount)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, trieID, core.MetachainShardId)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	return rcf.container.AddMultiple(keys, resolversSlice)
}

func (rcf *resolversContainerFactory) checkIfResolverExists(topic string) bool {
	_, err := rcf.container.Get(topic)
	return err == nil
}

func (rcf *resolversContainerFactory) createTrieNodesResolver(baseTopic string, trieId string, targetShardID uint32) (dataRetriever.Resolver, error) {
	//for each resolver we create a pseudo-intra shard topic as to make at least of half of the requests target the proper peers
	//this pseudo-intra shard topic is the consensus_targetShardID
	targetShardCoordinator, err := sharding.NewMultiShardCoordinator(rcf.shardCoordinator.NumberOfShards(), targetShardID)
	if err != nil {
		return nil, err
	}

	targetConsensusStopic := core.ConsensusTopic + targetShardCoordinator.CommunicationIdentifier(targetShardID)
	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(
		rcf.messenger,
		baseTopic,
		targetConsensusStopic,
		factoryDataRetriever.EmptyExcludePeersOnTopic,
	)
	if err != nil {
		return nil, err
	}

	arg := topicResolverSender.ArgTopicResolverSender{
		Messenger:                   rcf.messenger,
		TopicName:                   baseTopic,
		PeerListCreator:             peerListCreator,
		Marshalizer:                 rcf.marshalizer,
		Randomizer:                  rcf.intRandomizer,
		TargetShardId:               defaultTargetShardID,
		OutputAntiflooder:           rcf.outputAntifloodHandler,
		NumCrossShardPeers:          numCrossShardPeers,
		NumIntraShardPeers:          numIntraShardPeers,
		NumFullHistoryPeers:         numFullHistoryPeers,
		CurrentNetworkEpochProvider: &disabled.CurrentNetworkEpochProviderHandler{},
	}
	resolverSender, err := topicResolverSender.NewTopicResolverSender(arg)
	if err != nil {
		return nil, err
	}

	trie := rcf.dataTrieContainer.Get([]byte(trieId))
	argTrieResolver := resolvers.ArgTrieNodeResolver{
		SenderResolver:   resolverSender,
		TrieDataGetter:   trie,
		Marshalizer:      rcf.marshalizer,
		AntifloodHandler: rcf.inputAntifloodHandler,
		Throttler:        rcf.throttler,
	}
	resolver, err := resolvers.NewTrieNodeResolver(argTrieResolver)
	if err != nil {
		return nil, err
	}

	err = rcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), core.HardforkResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcf *resolversContainerFactory) IsInterfaceNil() bool {
	return rcf == nil
}
