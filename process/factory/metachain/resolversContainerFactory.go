package metachain

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/metablock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type resolversContainerFactory struct {
	shardCoordinator         sharding.Coordinator
	messenger                process.TopicMessageHandler
	blockchain               data.ChainHandler
	marshalizer              marshal.Marshalizer
	dataPools                data.MetaPoolsHolder
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// NewResolversContainerFactory creates a new container filled with topic resolvers
func NewResolversContainerFactory(
	shardCoordinator sharding.Coordinator,
	messenger process.TopicMessageHandler,
	blockchain data.ChainHandler,
	marshalizer marshal.Marshalizer,
	dataPools data.MetaPoolsHolder,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (*resolversContainerFactory, error) {

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if messenger == nil {
		return nil, process.ErrNilMessenger
	}
	if blockchain == nil {
		return nil, process.ErrNilBlockChain
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if dataPools == nil {
		return nil, process.ErrNilDataPoolHolder
	}
	if uint64ByteSliceConverter == nil {
		return nil, process.ErrNilUint64ByteSliceConverter
	}

	return &resolversContainerFactory{
		shardCoordinator:         shardCoordinator,
		messenger:                messenger,
		blockchain:               blockchain,
		marshalizer:              marshalizer,
		dataPools:                dataPools,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (rcf *resolversContainerFactory) Create() (process.ResolversContainer, error) {
	container := containers.NewResolversContainer()

	keys, interceptorSlice, err := rcf.generateShardHeaderResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, interceptorSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (rcf *resolversContainerFactory) createTopicAndAssignHandler(
	topicName string,
	resolver process.Resolver,
	createChannel bool,
) (process.Resolver, error) {

	err := rcf.messenger.CreateTopic(topicName, createChannel)
	if err != nil {
		return nil, err
	}

	return resolver, rcf.messenger.RegisterMessageProcessor(topicName, resolver)
}

//------- Shard header resolvers

func (rcf *resolversContainerFactory) generateShardHeaderResolvers() ([]string, []process.Resolver, error) {
	shardC := rcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	resolverSlice := make([]process.Resolver, noOfShards)

	//wire up to topics: shardHeadersForMetachain_0, shardHeadersForMetachain_1 ...
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

func (rcf *resolversContainerFactory) createOneShardHeaderResolver(identifier string) (process.Resolver, error) {
	hdrStorer := rcf.blockchain.GetStorer(data.BlockHeaderUnit)

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		rcf.messenger,
		identifier,
		rcf.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	resolver, err := metablock.NewShardHeaderResolver(
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
