package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/random"
	factory2 "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

const emptyExcludePeersOnTopic = ""

type resolversContainerFactory struct {
	shardCoordinator         sharding.Coordinator
	messenger                dataRetriever.TopicMessageHandler
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	dataPools                dataRetriever.PoolsHolder
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	intRandomizer            dataRetriever.IntRandomizer
	dataPacker               dataRetriever.DataPacker
	trieDataGetter           dataRetriever.TrieDataGetter
}

// NewResolversContainerFactory creates a new container filled with topic resolvers
func NewResolversContainerFactory(
	shardCoordinator sharding.Coordinator,
	messenger dataRetriever.TopicMessageHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	dataPools dataRetriever.PoolsHolder,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
	dataPacker dataRetriever.DataPacker,
	trieDataGetter dataRetriever.TrieDataGetter,
	sizeCheckDelta uint32,
) (*resolversContainerFactory, error) {

	if check.IfNil(shardCoordinator) {
		return nil, dataRetriever.ErrNilShardCoordinator
	}
	if check.IfNil(messenger) {
		return nil, dataRetriever.ErrNilMessenger
	}
	if check.IfNil(store) {
		return nil, dataRetriever.ErrNilTxStorage
	}
	if check.IfNil(marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if sizeCheckDelta > 0 {
		marshalizer = marshal.NewSizeCheckUnmarshalizer(marshalizer, sizeCheckDelta)
	}
	if check.IfNil(dataPools) {
		return nil, dataRetriever.ErrNilDataPoolHolder
	}
	if check.IfNil(uint64ByteSliceConverter) {
		return nil, dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(dataPacker) {
		return nil, dataRetriever.ErrNilDataPacker
	}
	if trieDataGetter == nil || trieDataGetter.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilTrieDataGetter
	}

	return &resolversContainerFactory{
		shardCoordinator:         shardCoordinator,
		messenger:                messenger,
		store:                    store,
		marshalizer:              marshalizer,
		dataPools:                dataPools,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		intRandomizer:            &random.ConcurrentSafeIntRandomizer{},
		dataPacker:               dataPacker,
		trieDataGetter:           trieDataGetter,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	container := containers.NewResolversContainer()

	keys, resolverSlice, err := rcf.generateTrieNodesResolver()
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

func (rcf *resolversContainerFactory) generateTrieNodesResolver() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator

	keys := make([]string, 0)
	resolverSlice := make([]dataRetriever.Resolver, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.TrieNodesTopic + shardC.CommunicationIdentifier(shardC.SelfId())
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, i)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice = append(resolverSlice, resolver)

		actualId := update.CreateTrieIdentifier(i, factory2.UserAccount)
		keys = append(keys, actualId)
	}

	identifierTrieNodes := factory.TrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, sharding.MetachainShardId)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)

	actualId := update.CreateTrieIdentifier(sharding.MetachainShardId, factory2.ValidatorAccount)
	keys = append(keys, actualId)

	identifierTrieNodes = factory.TrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	resolver, err = rcf.createTrieNodesResolver(identifierTrieNodes, sharding.MetachainShardId)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)

	actualId = update.CreateTrieIdentifier(sharding.MetachainShardId, factory2.ValidatorAccount)
	keys = append(keys, actualId)

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createTrieNodesResolver(topic string, shId uint32) (dataRetriever.Resolver, error) {
	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(rcf.messenger, topic, emptyExcludePeersOnTopic)
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
		rcf.trieDataGetter,
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
