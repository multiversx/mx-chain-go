package metachain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core/random"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type resolversContainerFactory struct {
	shardCoordinator         sharding.Coordinator
	messenger                dataRetriever.TopicMessageHandler
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	dataPools                dataRetriever.MetaPoolsHolder
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// NewResolversContainerFactory creates a new container filled with topic resolvers
func NewResolversContainerFactory(
	shardCoordinator sharding.Coordinator,
	messenger dataRetriever.TopicMessageHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	dataPools dataRetriever.MetaPoolsHolder,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (*resolversContainerFactory, error) {

	if shardCoordinator == nil {
		return nil, dataRetriever.ErrNilShardCoordinator
	}
	if messenger == nil {
		return nil, dataRetriever.ErrNilMessenger
	}
	if store == nil {
		return nil, dataRetriever.ErrNilStore
	}
	if marshalizer == nil {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if dataPools == nil {
		return nil, dataRetriever.ErrNilDataPoolHolder
	}
	if uint64ByteSliceConverter == nil {
		return nil, dataRetriever.ErrNilUint64ByteSliceConverter
	}

	return &resolversContainerFactory{
		shardCoordinator:         shardCoordinator,
		messenger:                messenger,
		store:                    store,
		marshalizer:              marshalizer,
		dataPools:                dataPools,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	container := containers.NewResolversContainer()

	keys, interceptorSlice, err := rcf.generateShardHeaderResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	metaKeys, metaInterceptorSlice, err := rcf.generateMetaChainHeaderResolvers()
	err = container.AddMultiple(metaKeys, metaInterceptorSlice)
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

//------- Shard header resolvers

func (rcf *resolversContainerFactory) generateShardHeaderResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := rcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards)

	//wire up to topics: shardHeadersForMetachain_0_META, shardHeadersForMetachain_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierHeader := factory.ShardHeadersForMetachainTopic + shardC.CommunicationIdentifier(idx)
		resolver, err := rcf.createOneShardHeaderResolver(identifierHeader)
		if err != nil {
			return nil, nil, err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierHeader
	}

	return keys, resolverSlice, nil
}

func (rcf *resolversContainerFactory) createOneShardHeaderResolver(identifier string) (dataRetriever.Resolver, error) {
	hdrStorer := rcf.store.GetStorer(dataRetriever.BlockHeaderUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer,
		&random.IntRandomizerConcurrentSafe{},
	)
	if err != nil {
		return nil, err
	}

	resolver, err := resolvers.NewShardHeaderResolver(
		resolverSender,
		rcf.dataPools.ShardHeaders(),
		hdrStorer,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

//------- Meta header resolvers

func (rcf *resolversContainerFactory) generateMetaChainHeaderResolvers() ([]string, []dataRetriever.Resolver, error) {
	identifierHeader := factory.MetachainBlocksTopic
	resolver, err := rcf.createMetaChainHeaderResolver(identifierHeader)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHeader}, []dataRetriever.Resolver{resolver}, nil
}

func (rcf *resolversContainerFactory) createMetaChainHeaderResolver(identifier string) (dataRetriever.Resolver, error) {
	hdrStorer := rcf.store.GetStorer(dataRetriever.MetaBlockUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer,
		&random.IntRandomizerConcurrentSafe{},
	)
	if err != nil {
		return nil, err
	}
	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		rcf.dataPools.MetaChainBlocks(),
		rcf.dataPools.MetaBlockNonces(),
		hdrStorer,
		rcf.marshalizer,
		rcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return rcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}
