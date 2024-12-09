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

	err = srcf.generateHeaderResolvers(core.SovereignChainShardId)
	if err != nil {
		return nil, err
	}

	err = srcf.generateMiniBlocksResolvers()
	if err != nil {
		return nil, err
	}

	err = srcf.generateAccountAndValidatorTrieNodesResolvers(core.SovereignChainShardId)
	if err != nil {
		return nil, err
	}

	err = srcf.generatePeerAuthenticationResolver()
	if err != nil {
		return nil, err
	}

	validatorInfoTopicID := common.ValidatorInfoTopic + srcf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	err = srcf.generateValidatorInfoResolver(validatorInfoTopicID)
	if err != nil {
		return nil, err
	}

	err = srcf.generateSovereignExtendedHeaderResolvers()
	if err != nil {
		return nil, err
	}

	return srcf.container, nil
}

func (srcf *sovereignShardResolversContainerFactory) generateSovereignExtendedHeaderResolvers() error {
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
	id := shardC.SelfId()
	identifierTx := topic + shardC.CommunicationIdentifier(id)

	resolver, err := srcf.createTxResolver(identifierTx, EmptyExcludePeersOnTopic, unit, dataPool, id)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierTx, resolver)
}

func (srcf *sovereignShardResolversContainerFactory) generateMiniBlocksResolvers() error {
	shardC := srcf.shardCoordinator
	id := shardC.SelfId()
	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(id)

	resolver, err := srcf.createMiniBlocksResolver(identifierMiniBlocks, EmptyExcludePeersOnTopic, id)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierMiniBlocks, resolver)
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *sovereignShardResolversContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
