package resolverscontainer

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	"github.com/ElrondNetwork/elrond-go-core/core/throttler"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	triesFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
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

	numIntraShardPeers := args.ResolverConfig.NumTotalPeers - args.ResolverConfig.NumCrossShardPeers
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
		preferredPeersHolder:        args.PreferredPeersHolder,
		peersRatingHandler:          args.PeersRatingHandler,
		numCrossShardPeers:          int(args.ResolverConfig.NumCrossShardPeers),
		numIntraShardPeers:          int(numIntraShardPeers),
		numTotalPeers:               int(args.ResolverConfig.NumTotalPeers),
		numFullHistoryPeers:         int(args.ResolverConfig.NumFullHistoryPeers),
		payloadValidator:            args.PayloadValidator,
	}

	err = base.checkParams()
	if err != nil {
		return nil, err
	}

	base.intraShardTopic = common.ConsensusTopic +
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

	err = srcf.generatePeerAuthenticationResolver()
	if err != nil {
		return nil, err
	}

	err = srcf.generateValidatorInfoResolver()
	if err != nil {
		return nil, err
	}

	return srcf.container, nil
}

// ------- Hdr resolver

func (srcf *shardResolversContainerFactory) generateHeaderResolvers() error {
	shardC := srcf.shardCoordinator

	// only one shard header topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	hdrStorer, err := srcf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	if err != nil {
		return err
	}
	resolverSender, err := srcf.createOneResolverSenderWithSpecifiedNumRequests(identifierHdr, EmptyExcludePeersOnTopic, shardC.SelfId(), srcf.numCrossShardPeers, srcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardC.SelfId())
	hdrNonceStore, err := srcf.store.GetStorer(hdrNonceHashDataUnit)
	if err != nil {
		return err
	}

	arg := resolvers.ArgHeaderResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshaller:       srcf.marshalizer,
			AntifloodHandler: srcf.inputAntifloodHandler,
			Throttler:        srcf.throttler,
		},
		Headers:              srcf.dataPools.Headers(),
		HdrStorage:           hdrStorer,
		HeadersNoncesStorage: hdrNonceStore,
		NonceConverter:       srcf.uint64ByteSliceConverter,
		ShardCoordinator:     srcf.shardCoordinator,
		IsFullHistoryNode:    srcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewHeaderResolver(arg)
	if err != nil {
		return err
	}

	err = srcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, resolver)
}

// ------- MetaBlockHeaderResolvers

func (srcf *shardResolversContainerFactory) generateMetablockHeaderResolvers() error {
	// only one metachain header block topic
	// this is: metachainBlocks
	identifierHdr := factory.MetachainBlocksTopic
	hdrStorer, err := srcf.store.GetStorer(dataRetriever.MetaBlockUnit)
	if err != nil {
		return err
	}

	resolverSender, err := srcf.createOneResolverSenderWithSpecifiedNumRequests(identifierHdr, EmptyExcludePeersOnTopic, core.MetachainShardId, srcf.numCrossShardPeers, srcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	hdrNonceStore, err := srcf.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if err != nil {
		return err
	}

	arg := resolvers.ArgHeaderResolver{
		ArgBaseResolver: resolvers.ArgBaseResolver{
			SenderResolver:   resolverSender,
			Marshaller:       srcf.marshalizer,
			AntifloodHandler: srcf.inputAntifloodHandler,
			Throttler:        srcf.throttler,
		},
		Headers:              srcf.dataPools.Headers(),
		HdrStorage:           hdrStorer,
		HeadersNoncesStorage: hdrNonceStore,
		NonceConverter:       srcf.uint64ByteSliceConverter,
		ShardCoordinator:     srcf.shardCoordinator,
		IsFullHistoryNode:    srcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewHeaderResolver(arg)
	if err != nil {
		return err
	}

	err = srcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
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
	resolver, err := srcf.createTrieNodesResolver(
		identifierTrieNodes,
		triesFactory.UserAccountTrie,
		0,
		srcf.numTotalPeers,
		core.MetachainShardId,
	)
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

	resolver, err := srcf.createTxResolver(identifierTx, excludedPeersOnTopic, unit, dataPool, core.MetachainShardId, srcf.numCrossShardPeers, srcf.numIntraShardPeers)
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
