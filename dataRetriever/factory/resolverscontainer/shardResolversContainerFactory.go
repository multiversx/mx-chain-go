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

var _ dataRetriever.ResolversContainerFactory = (*shardResolversContainerFactory)(nil)

type shardResolversContainerFactory struct {
	*baseResolversContainerFactory
}

// NewShardResolversContainerFactory creates a new container filled with topic resolvers for shards
func NewShardResolversContainerFactory(
	shardCoordinator sharding.Coordinator,
	messenger dataRetriever.TopicMessageHandler,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	dataPools dataRetriever.PoolsHolder,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
	dataPacker dataRetriever.DataPacker,
	trieContainer state.TriesHolder,
	sizeCheckDelta uint32,
) (*shardResolversContainerFactory, error) {
	if sizeCheckDelta > 0 {
		marshalizer = marshal.NewSizeCheckUnmarshalizer(marshalizer, sizeCheckDelta)
	}

	base := &baseResolversContainerFactory{
		shardCoordinator:         shardCoordinator,
		messenger:                messenger,
		store:                    store,
		marshalizer:              marshalizer,
		dataPools:                dataPools,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		intRandomizer:            &random.ConcurrentSafeIntRandomizer{},
		dataPacker:               dataPacker,
		triesContainer:           trieContainer,
	}

	err := base.checkParams()
	if err != nil {
		return nil, err
	}

	return &shardResolversContainerFactory{
		baseResolversContainerFactory: base,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (srcf *shardResolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	container := containers.NewResolversContainer()

	keys, resolverSlice, err := srcf.generateTxResolvers(
		factory.TransactionTopic,
		dataRetriever.TransactionUnit,
		srcf.dataPools.Transactions(),
	)
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = srcf.generateTxResolvers(
		factory.UnsignedTransactionTopic,
		dataRetriever.UnsignedTransactionUnit,
		srcf.dataPools.UnsignedTransactions(),
	)
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = srcf.generateTxResolvers(
		factory.RewardsTransactionTopic,
		dataRetriever.RewardTransactionUnit,
		srcf.dataPools.RewardTransactions(),
	)
	if err != nil {
		return nil, err
	}

	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = srcf.generateHeaderResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = srcf.generateMiniBlocksResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = srcf.generatePeerChBlockBodyResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = srcf.generateMetablockHeaderResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	keys, resolverSlice, err = srcf.generateTrieNodesResolvers()
	if err != nil {
		return nil, err
	}
	err = container.AddMultiple(keys, resolverSlice)
	if err != nil {
		return nil, err
	}

	return container, nil
}

//------- Hdr resolver

func (srcf *shardResolversContainerFactory) generateHeaderResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := srcf.shardCoordinator

	//only one shard header topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(srcf.messenger, identifierHdr, emptyExcludePeersOnTopic)
	if err != nil {
		return nil, nil, err
	}

	hdrStorer := srcf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		srcf.messenger,
		identifierHdr,
		peerListCreator,
		srcf.marshalizer,
		srcf.intRandomizer,
		shardC.SelfId(),
	)
	if err != nil {
		return nil, nil, err
	}

	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardC.SelfId())
	hdrNonceStore := srcf.store.GetStorer(hdrNonceHashDataUnit)
	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		srcf.dataPools.Headers(),
		hdrStorer,
		hdrNonceStore,
		srcf.marshalizer,
		srcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, nil, err
	}
	//add on the request topic
	_, err = srcf.createTopicAndAssignHandler(
		identifierHdr+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []dataRetriever.Resolver{resolver}, nil
}

//------- PeerChBlocks resolvers

func (srcf *shardResolversContainerFactory) generatePeerChBlockBodyResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := srcf.shardCoordinator

	//only one intrashard peer change blocks topic
	identifierPeerCh := factory.PeerChBodyTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	peerBlockBodyStorer := srcf.store.GetStorer(dataRetriever.PeerChangesUnit)

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(srcf.messenger, identifierPeerCh, emptyExcludePeersOnTopic)
	if err != nil {
		return nil, nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		srcf.messenger,
		identifierPeerCh,
		peerListCreator,
		srcf.marshalizer,
		srcf.intRandomizer,
		shardC.SelfId(),
	)
	if err != nil {
		return nil, nil, err
	}

	resolver, err := resolvers.NewGenericBlockBodyResolver(
		resolverSender,
		srcf.dataPools.MiniBlocks(),
		peerBlockBodyStorer,
		srcf.marshalizer,
	)
	if err != nil {
		return nil, nil, err
	}
	//add on the request topic
	_, err = srcf.createTopicAndAssignHandler(
		identifierPeerCh+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierPeerCh}, []dataRetriever.Resolver{resolver}, nil
}

//------- MetaBlockHeaderResolvers

func (srcf *shardResolversContainerFactory) generateMetablockHeaderResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := srcf.shardCoordinator

	//only one metachain header block topic
	//this is: metachainBlocks
	identifierHdr := factory.MetachainBlocksTopic
	hdrStorer := srcf.store.GetStorer(dataRetriever.MetaBlockUnit)

	metaAndCrtShardTopic := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	excludedPeersOnTopic := factory.TransactionTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(srcf.messenger, metaAndCrtShardTopic, excludedPeersOnTopic)
	if err != nil {
		return nil, nil, err
	}

	resolverSender, err := topicResolverSender.NewTopicResolverSender(
		srcf.messenger,
		identifierHdr,
		peerListCreator,
		srcf.marshalizer,
		srcf.intRandomizer,
		sharding.MetachainShardId,
	)
	if err != nil {
		return nil, nil, err
	}

	hdrNonceStore := srcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	resolver, err := resolvers.NewHeaderResolver(
		resolverSender,
		srcf.dataPools.Headers(),
		hdrStorer,
		hdrNonceStore,
		srcf.marshalizer,
		srcf.uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, nil, err
	}

	//add on the request topic
	_, err = srcf.createTopicAndAssignHandler(
		identifierHdr+resolverSender.TopicRequestSuffix(),
		resolver,
		false)
	if err != nil {
		return nil, nil, err
	}

	return []string{identifierHdr}, []dataRetriever.Resolver{resolver}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *shardResolversContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}

func (srcf *shardResolversContainerFactory) generateTrieNodesResolvers() ([]string, []dataRetriever.Resolver, error) {
	shardC := srcf.shardCoordinator

	keys := make([]string, 0)
	resolverSlice := make([]dataRetriever.Resolver, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	resolver, err := srcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.UserAccountTrie)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + shardC.CommunicationIdentifier(sharding.MetachainShardId)
	resolver, err = srcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.PeerAccountTrie)
	if err != nil {
		return nil, nil, err
	}

	resolverSlice = append(resolverSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	return keys, resolverSlice, nil
}
