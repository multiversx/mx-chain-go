package resolverscontainer

import (
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
	_, err := srcf.shardResolversContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	err = srcf.generateSovereignHeaderResolvers()
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
	resolverSender, err := srcf.createOneResolverSenderWithSpecifiedNumRequests(identifierHdr, EmptyExcludePeersOnTopic, shardC.SelfId())
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

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *sovereignShardResolversContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
