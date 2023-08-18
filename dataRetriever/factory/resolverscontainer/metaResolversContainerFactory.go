package resolverscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"

	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process/factory"
)

var _ dataRetriever.ResolversContainerFactory = (*metaResolversContainerFactory)(nil)

type metaResolversContainerFactory struct {
	*baseResolversContainerFactory
}

// NewMetaResolversContainerFactory creates a new container filled with topic resolvers for metachain
func NewMetaResolversContainerFactory(
	args FactoryArgs,
) (*metaResolversContainerFactory, error) {
	if args.SizeCheckDelta > 0 {
		args.Marshalizer = marshal.NewSizeCheckUnmarshalizer(args.Marshalizer, args.SizeCheckDelta)
	}

	thr, err := throttler.NewNumGoRoutinesThrottler(args.NumConcurrentResolvingJobs)
	if err != nil {
		return nil, err
	}

	container := containers.NewResolversContainer()
	base := &baseResolversContainerFactory{
		container:                       container,
		shardCoordinator:                args.ShardCoordinator,
		mainMessenger:                   args.MainMessenger,
		fullArchiveMessenger:            args.FullArchiveMessenger,
		store:                           args.Store,
		marshalizer:                     args.Marshalizer,
		dataPools:                       args.DataPools,
		uint64ByteSliceConverter:        args.Uint64ByteSliceConverter,
		dataPacker:                      args.DataPacker,
		triesContainer:                  args.TriesContainer,
		inputAntifloodHandler:           args.InputAntifloodHandler,
		outputAntifloodHandler:          args.OutputAntifloodHandler,
		throttler:                       thr,
		isFullHistoryNode:               args.IsFullHistoryNode,
		mainPreferredPeersHolder:        args.MainPreferredPeersHolder,
		fullArchivePreferredPeersHolder: args.FullArchivePreferredPeersHolder,
		payloadValidator:                args.PayloadValidator,
	}

	err = base.checkParams()
	if err != nil {
		return nil, err
	}

	base.intraShardTopic = common.ConsensusTopic +
		base.shardCoordinator.CommunicationIdentifier(base.shardCoordinator.SelfId())

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

	err = mrcf.generateRewardsResolvers(
		factory.RewardsTransactionTopic,
		dataRetriever.RewardTransactionUnit,
		mrcf.dataPools.RewardTransactions(),
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

	err = mrcf.generatePeerAuthenticationResolver()
	if err != nil {
		return nil, err
	}

	err = mrcf.generateValidatorInfoResolver()
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

	for idx := uint32(0); idx < shardC.NumberOfShards(); idx++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(idx)
		resolver, err := mrcf.createTrieNodesResolver(
			identifierTrieNodes,
			dataRetriever.UserAccountsUnit.String(),
			idx,
		)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	return container.AddMultiple(keys, resolversSlice)
}

// ------- Shard header resolvers

func (mrcf *metaResolversContainerFactory) generateShardHeaderResolvers() error {
	shardC := mrcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	resolversSlice := make([]dataRetriever.Resolver, noOfShards)

	// wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierHeader := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := EmptyExcludePeersOnTopic

		resolver, err := mrcf.createShardHeaderResolver(identifierHeader, excludePeersFromTopic, idx)
		if err != nil {
			return err
		}

		resolversSlice[idx] = resolver
		keys[idx] = identifierHeader
	}

	return mrcf.container.AddMultiple(keys, resolversSlice)
}

func (mrcf *metaResolversContainerFactory) createShardHeaderResolver(
	topic string,
	excludedTopic string,
	shardID uint32,
) (dataRetriever.Resolver, error) {
	hdrStorer, err := mrcf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	resolverSender, err := mrcf.createOneResolverSenderWithSpecifiedNumRequests(topic, excludedTopic, shardID)
	if err != nil {
		return nil, err
	}

	// TODO change this data unit creation method through a factory or func
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardID)
	hdrNonceStore, err := mrcf.store.GetStorer(hdrNonceHashDataUnit)
	if err != nil {
		return nil, err
	}

	arg := resolvers.ArgHeaderResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshaller:       mrcf.marshalizer,
			AntifloodHandler: mrcf.inputAntifloodHandler,
			Throttler:        mrcf.throttler,
		},
		Headers:              mrcf.dataPools.Headers(),
		HdrStorage:           hdrStorer,
		HeadersNoncesStorage: hdrNonceStore,
		NonceConverter:       mrcf.uint64ByteSliceConverter,
		ShardCoordinator:     mrcf.shardCoordinator,
		IsFullHistoryNode:    mrcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewHeaderResolver(arg)
	if err != nil {
		return nil, err
	}

	err = mrcf.mainMessenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	err = mrcf.fullArchiveMessenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

// ------- Meta header resolvers

func (mrcf *metaResolversContainerFactory) generateMetaChainHeaderResolvers() error {
	identifierHeader := factory.MetachainBlocksTopic
	resolver, err := mrcf.createMetaChainHeaderResolver(identifierHeader, core.MetachainShardId)
	if err != nil {
		return err
	}

	return mrcf.container.Add(identifierHeader, resolver)
}

func (mrcf *metaResolversContainerFactory) createMetaChainHeaderResolver(
	identifier string,
	shardId uint32,
) (dataRetriever.Resolver, error) {
	hdrStorer, err := mrcf.store.GetStorer(dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	resolverSender, err := mrcf.createOneResolverSenderWithSpecifiedNumRequests(identifier, EmptyExcludePeersOnTopic, shardId)
	if err != nil {
		return nil, err
	}

	hdrNonceStore, err := mrcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if err != nil {
		return nil, err
	}

	arg := resolvers.ArgHeaderResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshaller:       mrcf.marshalizer,
			AntifloodHandler: mrcf.inputAntifloodHandler,
			Throttler:        mrcf.throttler,
		},
		Headers:              mrcf.dataPools.Headers(),
		HdrStorage:           hdrStorer,
		HeadersNoncesStorage: hdrNonceStore,
		NonceConverter:       mrcf.uint64ByteSliceConverter,
		ShardCoordinator:     mrcf.shardCoordinator,
		IsFullHistoryNode:    mrcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewHeaderResolver(arg)
	if err != nil {
		return nil, err
	}

	err = mrcf.mainMessenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	err = mrcf.fullArchiveMessenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (mrcf *metaResolversContainerFactory) generateTrieNodesResolvers() error {
	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	resolver, err := mrcf.createTrieNodesResolver(
		identifierTrieNodes,
		dataRetriever.UserAccountsUnit.String(),
		core.MetachainShardId,
	)
	if err != nil {
		return err
	}

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	resolver, err = mrcf.createTrieNodesResolver(
		identifierTrieNodes,
		dataRetriever.PeerAccountsUnit.String(),
		core.MetachainShardId,
	)
	if err != nil {
		return err
	}

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	return mrcf.container.AddMultiple(keys, resolversSlice)
}

func (mrcf *metaResolversContainerFactory) generateRewardsResolvers(
	topic string,
	unit dataRetriever.UnitType,
	dataPool dataRetriever.ShardedDataCacherNotifier,
) error {

	shardC := mrcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards)

	// wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := EmptyExcludePeersOnTopic

		resolver, err := mrcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool, idx)
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
