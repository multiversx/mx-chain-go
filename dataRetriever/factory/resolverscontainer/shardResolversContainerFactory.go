package resolverscontainer

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/random"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

var _ dataRetriever.ResolversContainerFactory = (*shardResolversContainerFactory)(nil)

type shardResolversContainerFactory struct {
	*baseResolversContainerFactory
}

// NewShardResolversContainerFactory creates a new container filled with topic resolvers for shards
func NewShardResolversContainerFactory(
	args FactoryArgs,
) (*shardResolversContainerFactory, error) {
	if args.SizeCheckDelta > 0 {
		args.Marshalizer = marshal.NewSizeCheckUnmarshalizer(args.Marshalizer, args.SizeCheckDelta)
	}

	thr, err := throttler.NewNumGoRoutinesThrottler(args.NumConcurrentResolvingJobs)
	if err != nil {
		return nil, err
	}

	container := containers.NewResolversContainer()
	base := &baseResolversContainerFactory{
		container:                   container,
		shardCoordinator:            args.ShardCoordinator,
		messenger:                   args.Messenger,
		store:                       args.Store,
		marshalizer:                 args.Marshalizer,
		dataPools:                   args.DataPools,
		uint64ByteSliceConverter:    args.Uint64ByteSliceConverter,
		intRandomizer:               &random.ConcurrentSafeIntRandomizer{},
		dataPacker:                  args.DataPacker,
		triesContainer:              args.TriesContainer,
		inputAntifloodHandler:       args.InputAntifloodHandler,
		outputAntifloodHandler:      args.OutputAntifloodHandler,
		throttler:                   thr,
		isFullHistoryNode:           args.IsFullHistoryNode,
		currentNetworkEpochProvider: args.CurrentNetworkEpochProvider,
		numCrossShardPeers:          int(args.ResolverConfig.NumCrossShardPeers),
		numIntraShardPeers:          int(args.ResolverConfig.NumIntraShardPeers),
		numFullHistoryPeers:         int(args.ResolverConfig.NumFullHistoryPeers),
	}

	err = base.checkParams()
	if err != nil {
		return nil, err
	}

	base.intraShardTopic = core.ConsensusTopic +
		base.shardCoordinator.CommunicationIdentifier(base.shardCoordinator.SelfId())

	return &shardResolversContainerFactory{
		baseResolversContainerFactory: base,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (srcf *shardResolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	err := srcf.generateTxResolvers(
		factory.TransactionTopic,
		dataRetriever.TransactionUnit,
		srcf.dataPools.Transactions(),
	)
	if err != nil {
		return nil, err
	}

	err = srcf.generateTxResolvers(
		factory.UnsignedTransactionTopic,
		dataRetriever.UnsignedTransactionUnit,
		srcf.dataPools.UnsignedTransactions(),
	)
	if err != nil {
		return nil, err
	}

	err = srcf.generateRewardResolver(
		factory.RewardsTransactionTopic,
		dataRetriever.RewardTransactionUnit,
		srcf.dataPools.RewardTransactions(),
	)
	if err != nil {
		return nil, err
	}

	err = srcf.generateHeaderResolvers()
	if err != nil {
		return nil, err
	}

	err = srcf.generateMiniBlocksResolvers()
	if err != nil {
		return nil, err
	}

	err = srcf.generateMetablockHeaderResolvers()
	if err != nil {
		return nil, err
	}

	err = srcf.generateTrieNodesResolvers()
	if err != nil {
		return nil, err
	}

	return srcf.container, nil
}

//------- Hdr resolver

func (srcf *shardResolversContainerFactory) generateHeaderResolvers() error {
	shardC := srcf.shardCoordinator

	//only one shard header topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	hdrStorer := srcf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	resolverSender, err := srcf.createOneResolverSender(identifierHdr, EmptyExcludePeersOnTopic, shardC.SelfId())
	if err != nil {
		return err
	}

	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardC.SelfId())
	hdrNonceStore := srcf.store.GetStorer(hdrNonceHashDataUnit)
	arg := resolvers.ArgHeaderResolver{
		SenderResolver:       resolverSender,
		Headers:              srcf.dataPools.Headers(),
		HdrStorage:           hdrStorer,
		HeadersNoncesStorage: hdrNonceStore,
		Marshalizer:          srcf.marshalizer,
		NonceConverter:       srcf.uint64ByteSliceConverter,
		ShardCoordinator:     srcf.shardCoordinator,
		AntifloodHandler:     srcf.inputAntifloodHandler,
		Throttler:            srcf.throttler,
		IsFullHistoryNode:    srcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewHeaderResolver(arg)
	if err != nil {
		return err
	}

	err = srcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), core.DefaultResolversIdentifier, resolver)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, resolver)
}

//------- MetaBlockHeaderResolvers

func (srcf *shardResolversContainerFactory) generateMetablockHeaderResolvers() error {
	//only one metachain header block topic
	//this is: metachainBlocks
	identifierHdr := factory.MetachainBlocksTopic
	hdrStorer := srcf.store.GetStorer(dataRetriever.MetaBlockUnit)

	resolverSender, err := srcf.createOneResolverSender(identifierHdr, EmptyExcludePeersOnTopic, core.MetachainShardId)
	if err != nil {
		return err
	}

	hdrNonceStore := srcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	arg := resolvers.ArgHeaderResolver{
		SenderResolver:       resolverSender,
		Headers:              srcf.dataPools.Headers(),
		HdrStorage:           hdrStorer,
		HeadersNoncesStorage: hdrNonceStore,
		Marshalizer:          srcf.marshalizer,
		NonceConverter:       srcf.uint64ByteSliceConverter,
		ShardCoordinator:     srcf.shardCoordinator,
		AntifloodHandler:     srcf.inputAntifloodHandler,
		Throttler:            srcf.throttler,
		IsFullHistoryNode:    srcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewHeaderResolver(arg)
	if err != nil {
		return err
	}

	err = srcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), core.DefaultResolversIdentifier, resolver)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, resolver)
}

func (srcf *shardResolversContainerFactory) generateTrieNodesResolvers() error {
	shardC := srcf.shardCoordinator

	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	resolver, err := srcf.createTrieNodesResolver(identifierTrieNodes, triesFactory.UserAccountTrie,
		0, srcf.numIntraShardPeers+srcf.numCrossShardPeers, srcf.numFullHistoryPeers, srcf.currentNetworkEpochProvider)
	if err != nil {
		return err
	}

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	return srcf.container.AddMultiple(keys, resolversSlice)
}

func (srcf *shardResolversContainerFactory) generateRewardResolver(
	topic string,
	unit dataRetriever.UnitType,
	dataPool dataRetriever.ShardedDataCacherNotifier,
) error {

	shardC := srcf.shardCoordinator

	keys := make([]string, 0)
	resolverSlice := make([]dataRetriever.Resolver, 0)

	identifierTx := topic + shardC.CommunicationIdentifier(core.MetachainShardId)
	excludedPeersOnTopic := factory.TransactionTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	resolver, err := srcf.createTxResolver(identifierTx, excludedPeersOnTopic, unit, dataPool)
	if err != nil {
		return err
	}

	resolverSlice = append(resolverSlice, resolver)
	keys = append(keys, identifierTx)

	return srcf.container.AddMultiple(keys, resolverSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *shardResolversContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
