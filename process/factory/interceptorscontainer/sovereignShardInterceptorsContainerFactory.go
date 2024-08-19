package interceptorscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
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

	err = sicf.generateHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateTrieNodesInterceptors()
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

	err = sicf.generateValidatorInfoInterceptor()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateSovereignHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	return sicf.mainContainer, sicf.fullArchiveContainer, nil
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateTxInterceptors() error {
	shardC := sicf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, 0, noOfShards)
	interceptorSlice := make([]process.Interceptor, 0, noOfShards)

	// tx interceptor for metachain topic
	identifierTx := factory.TransactionTopic + shardC.CommunicationIdentifier(core.SovereignChainShardId)

	interceptor, err := sicf.createOneTxInterceptor(identifierTx)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)

	return sicf.addInterceptorsToContainers(keys, interceptorSlice)
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateUnsignedTxsInterceptors() error {
	shardC := sicf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, 0, noOfShards)
	interceptorsSlice := make([]process.Interceptor, 0, noOfShards)

	identifierScr := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(core.SovereignChainShardId)
	interceptor, err := sicf.createOneUnsignedTxInterceptor(identifierScr)
	if err != nil {
		return err
	}

	keys = append(keys, identifierScr)
	interceptorsSlice = append(interceptorsSlice, interceptor)

	return sicf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateHeaderInterceptors() error {
	shardC := sicf.shardCoordinator

	hdrFactory, err := interceptorFactory.NewInterceptedShardHeaderDataFactory(sicf.argInterceptorFactory)
	if err != nil {
		return err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:        sicf.dataPool.Headers(),
		BlockBlackList: sicf.blockBlackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return err
	}

	// compose header shard topic, for example: shardBlocks_0_META
	identifierHdr := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(core.SovereignChainShardId)

	// only one intrashard header topic
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

func (sicf *sovereignShardInterceptorsContainerFactory) generateMiniBlocksInterceptors() error {
	shardC := sicf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	interceptorsSlice := make([]process.Interceptor, noOfShards)

	identifierMiniBlocks := factory.MiniBlocksTopic + shardC.CommunicationIdentifier(core.SovereignChainShardId)

	interceptor, err := sicf.createOneMiniBlocksInterceptor(identifierMiniBlocks)
	if err != nil {
		return err
	}

	keys[0] = identifierMiniBlocks
	interceptorsSlice[0] = interceptor

	return sicf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (sicf *sovereignShardInterceptorsContainerFactory) generateTrieNodesInterceptors() error {
	keys := make([]string, 0)
	trieInterceptors := make([]process.Interceptor, 0)

	identifierTrieNodes := factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
	interceptor, err := sicf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTrieNodes)
	trieInterceptors = append(trieInterceptors, interceptor)

	identifierTrieNodes = factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
	interceptor, err = sicf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTrieNodes)
	trieInterceptors = append(trieInterceptors, interceptor)

	return sicf.addInterceptorsToContainers(keys, trieInterceptors)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sicf *sovereignShardInterceptorsContainerFactory) IsInterfaceNil() bool {
	return sicf == nil
}
