package storageResolversContainers

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/storageResolvers"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

var _ dataRetriever.ResolversContainerFactory = (*metaResolversContainerFactory)(nil)

type metaResolversContainerFactory struct {
	*baseResolversContainerFactory
}

// NewMetaResolversContainerFactory creates a new container filled with topic resolvers for metachain
func NewMetaResolversContainerFactory(
	args FactoryArgs,
) (*metaResolversContainerFactory, error) {
	container := containers.NewResolversContainer()
	base := &baseResolversContainerFactory{
		container:                container,
		shardCoordinator:         args.ShardCoordinator,
		messenger:                args.Messenger,
		store:                    args.Store,
		marshalizer:              args.Marshalizer,
		uint64ByteSliceConverter: args.Uint64ByteSliceConverter,
		dataPacker:               args.DataPacker,
		manualEpochStartNotifier: args.ManualEpochStartNotifier,
		chanGracefullyClose:      args.ChanGracefullyClose,
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
	)
	if err != nil {
		return nil, err
	}

	err = mrcf.generateTxResolvers(
		factory.UnsignedTransactionTopic,
		dataRetriever.UnsignedTransactionUnit,
	)
	if err != nil {
		return nil, err
	}

	err = mrcf.generateRewardsResolvers(
		factory.RewardsTransactionTopic,
		dataRetriever.RewardTransactionUnit,
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

// AddShardTrieNodeResolvers will add trie node resolvers to the existing container, needed for start in epoch
func (mrcf *metaResolversContainerFactory) AddShardTrieNodeResolvers(container dataRetriever.ResolversContainer) error {
	if check.IfNil(container) {
		return dataRetriever.ErrNilResolverContainer
	}

	shardC := mrcf.shardCoordinator

	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(i)
		resolver := storageResolvers.NewTrieNodeResolver()

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	return container.AddMultiple(keys, resolversSlice)
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
		resolver, err := mrcf.createShardHeaderResolver(identifierHeader, idx)
		if err != nil {
			return err
		}

		resolversSlice[idx] = resolver
		keys[idx] = identifierHeader
	}

	return mrcf.container.AddMultiple(keys, resolversSlice)
}

func (mrcf *metaResolversContainerFactory) createShardHeaderResolver(
	responseTopicName string,
	shardID uint32,
) (dataRetriever.Resolver, error) {
	hdrStorer := mrcf.store.GetStorer(dataRetriever.BlockHeaderUnit)

	//TODO change this data unit creation method through a factory or func
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardID)
	hdrNonceStore := mrcf.store.GetStorer(hdrNonceHashDataUnit)
	arg := storageResolvers.ArgHeaderResolver{
		Messenger:                mrcf.messenger,
		ResponseTopicName:        responseTopicName,
		NonceConverter:           mrcf.uint64ByteSliceConverter,
		HdrStorage:               hdrStorer,
		HeadersNoncesStorage:     hdrNonceStore,
		ManualEpochStartNotifier: mrcf.manualEpochStartNotifier,
		ChanGracefullyClose:      mrcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	resolver, err := storageResolvers.NewHeaderResolver(arg)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

//------- Meta header resolvers

func (mrcf *metaResolversContainerFactory) generateMetaChainHeaderResolvers() error {
	identifierHeader := factory.MetachainBlocksTopic
	resolver, err := mrcf.createMetaChainHeaderResolver()
	if err != nil {
		return err
	}

	return mrcf.container.Add(identifierHeader, resolver)
}

func (mrcf *metaResolversContainerFactory) createMetaChainHeaderResolver() (dataRetriever.Resolver, error) {
	hdrStorer := mrcf.store.GetStorer(dataRetriever.MetaBlockUnit)

	hdrNonceStore := mrcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	arg := storageResolvers.ArgHeaderResolver{
		Messenger:                mrcf.messenger,
		ResponseTopicName:        factory.MetachainBlocksTopic,
		NonceConverter:           mrcf.uint64ByteSliceConverter,
		HdrStorage:               hdrStorer,
		HeadersNoncesStorage:     hdrNonceStore,
		ManualEpochStartNotifier: mrcf.manualEpochStartNotifier,
		ChanGracefullyClose:      mrcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	resolver, err := storageResolvers.NewHeaderResolver(arg)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (mrcf *metaResolversContainerFactory) generateTrieNodesResolvers() error {
	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	resolver := storageResolvers.NewTrieNodeResolver()

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	resolver = storageResolvers.NewTrieNodeResolver()

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	return mrcf.container.AddMultiple(keys, resolversSlice)
}

func (mrcf *metaResolversContainerFactory) generateRewardsResolvers(
	topic string,
	unit dataRetriever.UnitType,
) error {

	shardC := mrcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards)

	//wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		resolver, err := mrcf.createTxResolver(identifierTx, unit)
		if err != nil {
			return err
		}

		resolverSlice[idx] = resolver
		keys[idx] = identifierTx
	}

	return mrcf.container.AddMultiple(keys, resolverSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mrcf *metaResolversContainerFactory) IsInterfaceNil() bool {
	return mrcf == nil
}
