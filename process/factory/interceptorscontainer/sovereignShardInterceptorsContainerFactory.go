package interceptorscontainer

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	interceptorFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
)

// TODO: Implement this in MX-14517

// sovereignShardInterceptorsContainerFactory will handle the creation of sovereign interceptors container
type sovereignShardInterceptorsContainerFactory struct {
	*shardInterceptorsContainerFactory
	IncomingHeaderSubscriber IncomingHeaderSubscriber
}

// NewSovereignShardInterceptorsContainerFactory creates a new sovereign interceptors factory
func NewSovereignShardInterceptorsContainerFactory(
	args CommonInterceptorsContainerFactoryArgs,
) (*sovereignShardInterceptorsContainerFactory, error) {
	shardInterceptorContainer, err := NewShardInterceptorsContainerFactory(args)
	if err != nil {
		return nil, err
	}

	return &sovereignShardInterceptorsContainerFactory{
		shardInterceptorsContainerFactory: shardInterceptorContainer,
		IncomingHeaderSubscriber:          args.IncomingHeaderSubscriber,
	}, nil
}

// Create returns an interceptor container that will hold all sovereign interceptors
func (sicf *sovereignShardInterceptorsContainerFactory) Create() (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	_, _, err := sicf.shardInterceptorsContainerFactory.Create()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateSovereignHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	return sicf.mainContainer, sicf.fullArchiveContainer, nil
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateSovereignHeaderInterceptors() error {
	shardC := sicf.shardCoordinator

	argsHdrFactory := interceptorFactory.ArgsSovereignInterceptedExtendedHeaderFactory{
		Marshaller: sicf.argInterceptorFactory.CoreComponents.InternalMarshalizer(),
		Hasher:     sicf.argInterceptorFactory.CoreComponents.Hasher(),
	}
	hdrFactory, err := interceptorFactory.NewSovereignInterceptedShardHeaderDataFactory(argsHdrFactory)
	if err != nil {
		return err
	}

	argProcessor := &processor.ArgsSovereignHeaderInterceptorProcessor{
		BlockBlackList:           sicf.blockBlackList,
		Hasher:                   sicf.argInterceptorFactory.CoreComponents.Hasher(),
		Marshaller:               sicf.argInterceptorFactory.CoreComponents.InternalMarshalizer(),
		IncomingHeaderSubscriber: sicf.IncomingHeaderSubscriber,
	}
	hdrProcessor, err := processor.NewSovereignHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	identifierHdr := factory.ExtendedHeaderProofTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	// only one intra shard header topic
	interceptor, err := interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                identifierHdr,
			DataFactory:          hdrFactory,
			Processor:            hdrProcessor,
			Throttler:            sicf.globalThrottler,
			AntifloodHandler:     sicf.antifloodHandler,
			WhiteListRequest:     sicf.whiteListHandler,
			CurrentPeerId:        sicf.mainMessenger.ID(),
			PreferredPeersHolder: sicf.preferredPeersHolder,
		},
	)
	if err != nil {
		return err
	}

	_, err = sicf.createTopicAndAssignHandler(identifierHdr, interceptor, true)
	if err != nil {
		return err
	}

	return sicf.addInterceptorsToContainers([]string{identifierHdr}, []process.Interceptor{interceptor})
}

// IsInterfaceNil returns true if there is no value under the interface
func (sicf *sovereignShardInterceptorsContainerFactory) IsInterfaceNil() bool {
	return sicf == nil
}
