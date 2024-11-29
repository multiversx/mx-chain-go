package interceptorscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	interceptorFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
)

// ArgsSovereignShardInterceptorsContainerFactory is a struct placeholder for args needed to create a sovereign
// shard interceptors container factory
type ArgsSovereignShardInterceptorsContainerFactory struct {
	ShardContainer           *shardInterceptorsContainerFactory
	IncomingHeaderSubscriber process.IncomingHeaderSubscriber
}

type sovereignShardInterceptorsContainerFactory struct {
	*shardInterceptorsContainerFactory
	incomingHeaderSubscriber process.IncomingHeaderSubscriber
}

// NewSovereignShardInterceptorsContainerFactory creates a new sovereign interceptors factory
func NewSovereignShardInterceptorsContainerFactory(
	args ArgsSovereignShardInterceptorsContainerFactory,
) (*sovereignShardInterceptorsContainerFactory, error) {
	if check.IfNil(args.ShardContainer) {
		return nil, errors.ErrNilShardInterceptorsContainerFactory
	}
	if check.IfNil(args.IncomingHeaderSubscriber) {
		return nil, errors.ErrNilIncomingHeaderSubscriber
	}

	return &sovereignShardInterceptorsContainerFactory{
		shardInterceptorsContainerFactory: args.ShardContainer,
		incomingHeaderSubscriber:          args.IncomingHeaderSubscriber,
	}, nil
}

// Create returns an interceptor container that will hold all sovereign interceptors
func (sicf *sovereignShardInterceptorsContainerFactory) Create() (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	err := sicf.generateTxInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateSovereignHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateValidatorAndAccountTrieNodesInterceptors(core.SovereignChainShardId)
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generatePeerAuthenticationInterceptor()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateHeartbeatInterceptor()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generatePeerShardInterceptor()
	if err != nil {
		return nil, nil, err
	}

	validatorInfoTopicID := common.ValidatorInfoTopic + sicf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	err = sicf.generateValidatorInfoInterceptor(validatorInfoTopicID)
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateSovereignExtendedHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	return sicf.mainContainer, sicf.fullArchiveContainer, nil
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateSovereignHeaderInterceptors() error {
	hdrFactory, err := interceptorFactory.NewInterceptedSovereignShardHeaderDataFactory(sicf.argInterceptorFactory)
	if err != nil {
		return err
	}
	return sicf.generateHeaderInterceptors(hdrFactory, core.SovereignChainShardId)
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateTxInterceptors() error {
	keys := make([]string, 0, 1)
	interceptorSlice := make([]process.Interceptor, 0, 1)

	identifierTx := factory.TransactionTopic + sicf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	interceptor, err := sicf.createOneTxInterceptor(identifierTx)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)
	return sicf.addInterceptorsToContainers(keys, interceptorSlice)
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateUnsignedTxsInterceptors() error {
	keys := make([]string, 0, 1)
	interceptorsSlice := make([]process.Interceptor, 0, 1)

	identifierScr := factory.UnsignedTransactionTopic + sicf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	interceptor, err := sicf.createOneUnsignedTxInterceptor(identifierScr)
	if err != nil {
		return err
	}

	keys = append(keys, identifierScr)
	interceptorsSlice = append(interceptorsSlice, interceptor)

	return sicf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateMiniBlocksInterceptors() error {
	keys := make([]string, 0, 1)
	interceptorsSlice := make([]process.Interceptor, 0, 1)

	identifierMiniBlocks := factory.MiniBlocksTopic + sicf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	interceptor, err := sicf.createOneSovereignMiniBlocksInterceptor(identifierMiniBlocks)
	if err != nil {
		return err
	}

	keys = append(keys, identifierMiniBlocks)
	interceptorsSlice = append(interceptorsSlice, interceptor)

	return sicf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (bicf *baseInterceptorsContainerFactory) createOneSovereignMiniBlocksInterceptor(topic string) (process.Interceptor, error) {
	miniBlockFactory, err := interceptorFactory.NewInterceptedSovereignMiniBlockDataFactory(bicf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	return bicf.baseCreateOneMiniBlocksInterceptor(miniBlockFactory, topic)
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateSovereignExtendedHeaderInterceptors() error {
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
		IncomingHeaderSubscriber: sicf.incomingHeaderSubscriber,
		HeadersPool:              sicf.dataPool.Headers(),
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
