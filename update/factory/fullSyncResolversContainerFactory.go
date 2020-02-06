package factory

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/random"
	"github.com/ElrondNetwork/elrond-go/data/state"
	accountFactory "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

type resolversContainerFactory struct {
	shardCoordinator  sharding.Coordinator
	messenger         dataRetriever.TopicMessageHandler
	marshalizer       marshal.Marshalizer
	intRandomizer     dataRetriever.IntRandomizer
	dataTrieContainer state.TriesHolder
	existingResolvers dataRetriever.ResolversContainer
}

// ArgsNewResolversContainerFactory defines the arguments for the resolversContainerFactory constructor
type ArgsNewResolversContainerFactory struct {
	ShardCoordinator  sharding.Coordinator
	Messenger         dataRetriever.TopicMessageHandler
	Marshalizer       marshal.Marshalizer
	DataTrieContainer state.TriesHolder
	ExistingResolvers dataRetriever.ResolversContainer
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

	return &resolversContainerFactory{
		shardCoordinator:  args.ShardCoordinator,
		messenger:         args.Messenger,
		marshalizer:       args.Marshalizer,
		intRandomizer:     &random.ConcurrentSafeIntRandomizer{},
		dataTrieContainer: args.DataTrieContainer,
		existingResolvers: args.ExistingResolvers,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	container := rcf.existingResolvers

	keys, resolverSlice, err := rcf.generateTrieNodesResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (rcf *resolversContainerFactory) createTopicAndAssignHandler(
	topicName string,
	resolver dataRetriever.Resolver,
	createChannel bool,
) (dataRetriever.Resolver, error) {

	err := rcf.messenger.CreateTopic(topicName, createChannel)
	if err != nil {
		return nil, err
	}

	return resolver, rcf.messenger.RegisterMessageProcessor(topicName, resolver)
}

func (rcf *resolversContainerFactory) generateTrieNodesResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator

	keys := make([]string, 0)
	resolverSlice := make([]dataRetriever.Resolver, 0)
	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		trieId := genesis.CreateTrieIdentifier(i, accountFactory.UserAccount)
		resolver, err := rcf.createTrieNodesResolver(factory.AccountTrieNodesTopic, i, trieId)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice = append(resolverSlice, resolver)
		keys = append(keys, trieId)
	}

	trieId := genesis.CreateTrieIdentifier(sharding.MetachainShardId, accountFactory.UserAccount)
	resolver, err := rcf.createTrieNodesResolver(factory.AccountTrieNodesTopic, sharding.MetachainShardId, trieId)
	if err != nil && !errors.Is(err, p2p.ErrTopicAlreadyExists) {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)
	keys = append(keys, trieId)

	trieId = genesis.CreateTrieIdentifier(sharding.MetachainShardId, accountFactory.ValidatorAccount)
	resolver, err = rcf.createTrieNodesResolver(factory.ValidatorTrieNodesTopic, sharding.MetachainShardId, trieId)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)
	keys = append(keys, trieId)

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createTrieNodesResolver(baseTopic string, shId uint32, trieId string) (dataRetriever.Resolver, error) {
	topic := baseTopic + rcf.shardCoordinator.CommunicationIdentifier(shId)
	existingResolver, _ := rcf.existingResolvers.Get(topic)
	if !check.IfNil(existingResolver) {
		return existingResolver, nil
	}

	excludePeersFromTopic := topic + rcf.shardCoordinator.CommunicationIdentifier(rcf.shardCoordinator.SelfId())

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(rcf.messenger, topic, excludePeersFromTopic)
	if err != nil {
		return nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		topic,
		peerListCreator,
		rcf.marshalizer,
		rcf.intRandomizer,
		shId,
	)
	if err != nil {
		return nil, err
	}

	resolver, err := resolvers.NewTrieNodeResolver(
		resolverSender,
		rcf.dataTrieContainer.Get([]byte(trieId)),
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcf *resolversContainerFactory) IsInterfaceNil() bool {
	return rcf == nil
}
