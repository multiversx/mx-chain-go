package resolverscontainer

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/random"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer/disabled"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"

	"github.com/ElrondNetwork/elrond-go/marshal"
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
	if args.SizeCheckDelta > 0 {
		args.Marshalizer = marshal.NewSizeCheckUnmarshalizer(args.Marshalizer, args.SizeCheckDelta)
	}

	thr, err := throttler.NewNumGoRoutinesThrottler(args.NumConcurrentResolvingJobs)
	if err != nil {
		return nil, err
	}

	container := containers.NewResolversContainer()
	base := &baseResolversContainerFactory{
		container:                container,
		shardCoordinator:         args.ShardCoordinator,
		messenger:                args.Messenger,
		store:                    args.Store,
		marshalizer:              args.Marshalizer,
		dataPools:                args.DataPools,
		uint64ByteSliceConverter: args.Uint64ByteSliceConverter,
		intRandomizer:            &random.ConcurrentSafeIntRandomizer{},
		dataPacker:               args.DataPacker,
		triesContainer:           args.TriesContainer,
		inputAntifloodHandler:    args.InputAntifloodHandler,
		outputAntifloodHandler:   args.OutputAntifloodHandler,
		throttler:                thr,
		isFullHistoryNode:        args.IsFullHistoryNode,
	}

	err = base.checkParams()
	if err != nil {
		return nil, err
	}

	base.intraShardTopic = core.ConsensusTopic +
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
		resolver, err := mrcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.UserAccountTrie,
			numCrossShardPeers, numIntraShardPeers, numFullHistoryPeers, &disabled.NilCurrentNetworkEpochProviderHandler{})
		if err != nil {
			return err
		}

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
	hdrStorer := mrcf.store.GetStorer(dataRetriever.BlockHeaderUnit)

	resolverSender, err := mrcf.createOneResolverSender(topic, excludedTopic, shardID)
	if err != nil {
		return nil, err
	}

	//TODO change this data unit creation method through a factory or func
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardID)
	hdrNonceStore := mrcf.store.GetStorer(hdrNonceHashDataUnit)
	arg := resolvers.ArgHeaderResolver{
		SenderResolver:       resolverSender,
		Headers:              mrcf.dataPools.Headers(),
		HdrStorage:           hdrStorer,
		HeadersNoncesStorage: hdrNonceStore,
		Marshalizer:          mrcf.marshalizer,
		NonceConverter:       mrcf.uint64ByteSliceConverter,
		ShardCoordinator:     mrcf.shardCoordinator,
		AntifloodHandler:     mrcf.inputAntifloodHandler,
		Throttler:            mrcf.throttler,
		IsFullHistoryNode:    mrcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewHeaderResolver(arg)
	if err != nil {
		return nil, err
	}

	err = mrcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), core.DefaultResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

//------- Meta header resolvers

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
	hdrStorer := mrcf.store.GetStorer(dataRetriever.MetaBlockUnit)

	resolverSender, err := mrcf.createOneResolverSender(identifier, EmptyExcludePeersOnTopic, shardId)
	if err != nil {
		return nil, err
	}

	hdrNonceStore := mrcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	arg := resolvers.ArgHeaderResolver{
		SenderResolver:       resolverSender,
		Headers:              mrcf.dataPools.Headers(),
		HdrStorage:           hdrStorer,
		HeadersNoncesStorage: hdrNonceStore,
		Marshalizer:          mrcf.marshalizer,
		NonceConverter:       mrcf.uint64ByteSliceConverter,
		ShardCoordinator:     mrcf.shardCoordinator,
		AntifloodHandler:     mrcf.inputAntifloodHandler,
		Throttler:            mrcf.throttler,
		IsFullHistoryNode:    mrcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewHeaderResolver(arg)
	if err != nil {
		return nil, err
	}

	err = mrcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), core.DefaultResolversIdentifier, resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (mrcf *metaResolversContainerFactory) generateTrieNodesResolvers() error {
	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	resolver, err := mrcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.UserAccountTrie,
		0, numIntraShardPeers+numCrossShardPeers, numFullHistoryPeers, &disabled.NilCurrentNetworkEpochProviderHandler{})
	if err != nil {
		return err
	}

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	resolver, err = mrcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.PeerAccountTrie,
		0, numIntraShardPeers+numCrossShardPeers, numFullHistoryPeers, &disabled.NilCurrentNetworkEpochProviderHandler{})
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

	//wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierTx := topic + shardC.CommunicationIdentifier(idx)
		excludePeersFromTopic := EmptyExcludePeersOnTopic

		resolver, err := mrcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool)
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
