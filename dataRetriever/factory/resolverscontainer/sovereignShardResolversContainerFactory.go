package resolverscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignShardResolversContainerFactory struct {
	*shardResolversContainerFactory
}

// NewSovereignShardResolversContainerFactory creates a new sovereign shard resolvers container factory
func NewSovereignShardResolversContainerFactory(shardContainer *shardResolversContainerFactory) (*sovereignShardResolversContainerFactory, error) {
	if check.IfNil(shardContainer) {
		return nil, errors.ErrNilShardResolversContainerFactory
	}

	return &sovereignShardResolversContainerFactory{
		shardResolversContainerFactory: shardContainer,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (srcf *sovereignShardResolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
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

	err = srcf.generateHeaderResolvers()
	if err != nil {
		return nil, err
	}

	err = srcf.generateSovereignHeaderResolvers()
	if err != nil {
		return nil, err
	}

	err = srcf.generateMiniBlocksResolvers()
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

func (srcf *sovereignShardResolversContainerFactory) generateSovereignHeaderResolvers() error {
	shardC := srcf.shardCoordinator

	hdrStorer, err := srcf.store.GetStorer(dataRetriever.ExtendedShardHeadersUnit)
	if err != nil {
		return err
	}

	identifierHdr := factory.ExtendedHeaderProofTopic + shardC.CommunicationIdentifier(shardC.SelfId())
	resolverSender, err := srcf.createOneResolverSenderWithSpecifiedNumRequests(identifierHdr, EmptyExcludePeersOnTopic, core.MainChainShardId)
	if err != nil {
		return err
	}

	hdrNonceStorer, err := srcf.store.GetStorer(dataRetriever.ExtendedShardHeadersNonceHashDataUnit)
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
		HeadersNoncesStorage: hdrNonceStorer,
		NonceConverter:       srcf.uint64ByteSliceConverter,
		ShardCoordinator:     srcf.shardCoordinator,
		IsFullHistoryNode:    srcf.isFullHistoryNode,
	}
	resolver, err := resolvers.NewHeaderResolver(arg)
	if err != nil {
		return err
	}

	err = srcf.mainMessenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return err
	}

	err = srcf.fullArchiveMessenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, resolver)
}

func (srcf *sovereignShardResolversContainerFactory) generateTxResolvers(
	topic string,
	unit dataRetriever.UnitType,
	dataPool dataRetriever.ShardedDataCacherNotifier,
) error {

	shardC := srcf.shardCoordinator
	idx := shardC.SelfId()
	identifierTx := topic + shardC.CommunicationIdentifier(idx)
	excludePeersFromTopic := EmptyExcludePeersOnTopic

	resolver, err := srcf.createTxResolver(identifierTx, excludePeersFromTopic, unit, dataPool, idx)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierTx, resolver)
}

func (srcf *sovereignShardResolversContainerFactory) generateHeaderResolvers() error {
	shardC := srcf.shardCoordinator

	// only one shard header topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(core.SovereignChainShardId)

	hdrStorer, err := srcf.store.GetStorer(dataRetriever.BlockHeaderUnit)
	if err != nil {
		return err
	}
	resolverSender, err := srcf.createOneResolverSenderWithSpecifiedNumRequests(identifierHdr, EmptyExcludePeersOnTopic, shardC.SelfId())
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

	err = srcf.mainMessenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return err
	}

	err = srcf.fullArchiveMessenger.RegisterMessageProcessor(resolver.RequestTopic(), common.DefaultResolversIdentifier, resolver)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, resolver)
}

func (srcf *sovereignShardResolversContainerFactory) generateMiniBlocksResolvers() error {
	shardC := srcf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	resolverSlice := make([]dataRetriever.Resolver, noOfShards)

	idx := core.SovereignChainShardId
	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(idx)
	excludePeersFromTopic := EmptyExcludePeersOnTopic

	resolver, err := srcf.createMiniBlocksResolver(identifierMiniBlocks, excludePeersFromTopic, idx)
	if err != nil {
		return err
	}

	resolverSlice[idx] = resolver
	keys[idx] = identifierMiniBlocks

	return srcf.container.AddMultiple(keys, resolverSlice)
}
func (srcf *sovereignShardResolversContainerFactory) generateTrieNodesResolvers() error {
	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
	resolver, err := srcf.createTrieNodesResolver(
		identifierTrieNodes,
		dataRetriever.UserAccountsUnit.String(),
		core.SovereignChainShardId,
	)
	if err != nil {
		return err
	}

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
	resolver, err = srcf.createTrieNodesResolver(
		identifierTrieNodes,
		dataRetriever.PeerAccountsUnit.String(),
		core.SovereignChainShardId,
	)
	if err != nil {
		return err
	}

	resolversSlice = append(resolversSlice, resolver)
	keys = append(keys, identifierTrieNodes)

	return srcf.container.AddMultiple(keys, resolversSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *sovereignShardResolversContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
