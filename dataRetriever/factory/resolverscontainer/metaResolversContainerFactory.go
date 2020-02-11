package resolverscontainer

import (
	"github.com/ElrondNetwork/elrond-go/core/random"
	"github.com/ElrondNetwork/elrond-go/data/state"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ dataRetriever.ResolversContainerFactory = (*metaResolversContainerFactory)(nil)

type metaResolversContainerFactory struct {
	*baseResolversContainerFactory
}

// NewMetaResolversContainerFactory creates a new container filled with topic resolvers for metachain
func NewMetaResolversContainerFactory(
	shardCoordinator sharding.Coordinator,
	messenger dataRetriever.TopicMessageHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	dataPools dataRetriever.PoolsHolder,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
	dataPacker dataRetriever.DataPacker,
	triesContainer state.TriesHolder,
	sizeCheckDelta uint32,
) (*metaResolversContainerFactory, error) {
	if sizeCheckDelta > 0 {
		marshalizer = marshal.NewSizeCheckUnmarshalizer(marshalizer, sizeCheckDelta)
	}

	container := containers.NewResolversContainer()
	base := &baseResolversContainerFactory{
		container:                container,
		shardCoordinator:         shardCoordinator,
		messenger:                messenger,
		store:                    store,
		marshalizer:              marshalizer,
		dataPools:                dataPools,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		intRandomizer:            &random.ConcurrentSafeIntRandomizer{},
		dataPacker:               dataPacker,
		triesContainer:           triesContainer,
	}

	err := base.checkParams()
	if err != nil {
		return nil, err
	}

	return &metaResolversContainerFactory{
		baseResolversContainerFactory: base,
	}, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (mrcf *metaResolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	err := mrcf.generateShardHeaderResolvers()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateMetaChainHeaderResolvers()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateTxResolvers(
		factory.TransactionTopic,
		dataRetriever.TransactionUnit,
		mrcf.dataPools.Transactions(),
	)
	if err != nil {
		return nil, err
	}

	err = mrcf.generateTxResolvers(
		factory.UnsignedTransactionTopic,
		dataRetriever.UnsignedTransactionUnit,
		mrcf.dataPools.UnsignedTransactions(),
	)
	if err != nil {
		return nil, err
	}

	err = mrcf.generateMiniBlocksResolvers()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateTrieNodesResolvers()
	if err != nil {
		return nil, err
	}

	return mrcf.container, nil
}

//------- Shard header resolvers

func (mrcf *metaResolversContainerFactory) generateShardHeaderResolvers() error {
	shardC := mrcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	resolversSlice := make([]dataRetriever.Resolver, noOfShards)

	//wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierHeader := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := emptyExcludePeersOnTopic

		resolver, err := mrcf.createShardHeaderResolver(identifierHeader, excludePeersFromTopic, idx)
		if err != nil {
			return err
		}

		resolversSlice[idx] = resolver
		keys[idx] = identifierHeader
	}

	return mrcf.container.AddMultiple(keys, resolversSlice)
}

func (mrcf *metaResolversContainerFactory) createShardHeaderResolver(topic string, excludedTopic string, shardID uint32) (dataRetriever.Resolver, error) {
	hdrStorer := mrcf.store.GetStorer(dataRetriever.BlockHeaderUnit)

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(mrcf.messenger, topic, excludedTopic)
	if err != nil {
		return nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		mrcf.messenger,
		topic,
		peerListCreator,
		mrcf.marshalizer,
		mrcf.intRandomizer,
		shardID,
	)
	if err != nil {
		return nil, err
	}

	//TODO change this data unit creation method through a factory or func
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardID)
	hdrNonceStore := mrcf.store.GetStorer(hdrNonceHashDataUnit)
	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		mrcf.dataPools.Headers(),
		hdrStorer,
		hdrNonceStore,
		mrcf.marshalizer,
		mrcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return mrcf.createTopicAndAssignHandler(
		topic+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

//------- Meta header resolvers

func (mrcf *metaResolversContainerFactory) generateMetaChainHeaderResolvers() error {
	identifierHeader := factory.MetachainBlocksTopic
	resolver, err := mrcf.createMetaChainHeaderResolver(identifierHeader, sharding.MetachainShardId)
	if err != nil {
		return err
	}

	return mrcf.container.Add(identifierHeader, resolver)
}

func (mrcf *metaResolversContainerFactory) createMetaChainHeaderResolver(identifier string, shardId uint32) (dataRetriever.Resolver, error) {
	hdrStorer := mrcf.store.GetStorer(dataRetriever.MetaBlockUnit)

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(mrcf.messenger, identifier, emptyExcludePeersOnTopic)
	if err != nil {
		return nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		mrcf.messenger,
		identifier,
		peerListCreator,
		mrcf.marshalizer,
		mrcf.intRandomizer,
		shardId,
	)
	if err != nil {
		return nil, err
	}

	hdrNonceStore := mrcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		mrcf.dataPools.Headers(),
		hdrStorer,
		hdrNonceStore,
		mrcf.marshalizer,
		mrcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, err
	}

	//add on the request topic
	return mrcf.createTopicAndAssignHandler(
		identifier+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
}

func (mrcf *metaResolversContainerFactory) generateTrieNodesResolvers() error {
	shardC := mrcf.shardCoordinator

	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(i)
		resolver, err := mrcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.UserAccountTrie)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)

		identifierTrieNodes = factory.ValidatorTrieNodesTopic + shardC.CommunicationIdentifier(i)
		resolver, err = mrcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.PeerAccountTrie)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	resolver, err := mrcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.UserAccountTrie)
	if err != nil {
		return err
	}

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	resolver, err = mrcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.PeerAccountTrie)
	if err != nil {
		return err
	}

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	return mrcf.container.AddMultiple(keys, resolversSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mrcf *metaResolversContainerFactory) IsInterfaceNil() bool {
	return mrcf == nil
}
