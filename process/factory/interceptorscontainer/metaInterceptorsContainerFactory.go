package interceptorscontainer

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/throttler"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	processInterceptors "github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
)

var _ process.InterceptorsContainerFactory = (*metaInterceptorsContainerFactory)(nil)

// metaInterceptorsContainerFactory will handle the creation the interceptors container for metachain
type metaInterceptorsContainerFactory struct {
	*baseInterceptorsContainerFactory
}

// NewMetaInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewMetaInterceptorsContainerFactory(
	args CommonInterceptorsContainerFactoryArgs,
) (*metaInterceptorsContainerFactory, error) {
	err := checkBaseParams(
		args.CoreComponents,
		args.CryptoComponents,
		args.ShardCoordinator,
		args.Accounts,
		args.Store,
		args.DataPool,
		args.Messenger,
		args.NodesCoordinator,
		args.BlockBlackList,
		args.AntifloodHandler,
		args.WhiteListHandler,
		args.WhiteListerVerifiedTxs,
		args.PreferredPeersHolder,
		args.RequestHandler,
	)
	if err != nil {
		return nil, err
	}
	if args.SizeCheckDelta > 0 {
		sizeCheckMarshalizer := marshal.NewSizeCheckUnmarshalizer(
			args.CoreComponents.InternalMarshalizer(),
			args.SizeCheckDelta,
		)
		err = args.CoreComponents.SetInternalMarshalizer(sizeCheckMarshalizer)
		if err != nil {
			return nil, err
		}
	}

	if check.IfNil(args.TxFeeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.HeaderSigVerifier) {
		return nil, process.ErrNilHeaderSigVerifier
	}
	if check.IfNil(args.HeaderIntegrityVerifier) {
		return nil, process.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(args.EpochStartTrigger) {
		return nil, process.ErrNilEpochStartTrigger
	}
	if check.IfNil(args.ValidityAttester) {
		return nil, process.ErrNilValidityAttester
	}

	argInterceptorFactory := &interceptorFactory.ArgInterceptedDataFactory{
		CoreComponents:            args.CoreComponents,
		CryptoComponents:          args.CryptoComponents,
		ShardCoordinator:          args.ShardCoordinator,
		NodesCoordinator:          args.NodesCoordinator,
		FeeHandler:                args.TxFeeHandler,
		HeaderSigVerifier:         args.HeaderSigVerifier,
		HeaderIntegrityVerifier:   args.HeaderIntegrityVerifier,
		ValidityAttester:          args.ValidityAttester,
		EpochStartTrigger:         args.EpochStartTrigger,
		WhiteListerVerifiedTxs:    args.WhiteListerVerifiedTxs,
		ArgsParser:                args.ArgumentsParser,
		EnableSignTxWithHashEpoch: args.EnableSignTxWithHashEpoch,
	}

	container := containers.NewInterceptorsContainer()
	base := &baseInterceptorsContainerFactory{
		container:              container,
		shardCoordinator:       args.ShardCoordinator,
		messenger:              args.Messenger,
		store:                  args.Store,
		dataPool:               args.DataPool,
		nodesCoordinator:       args.NodesCoordinator,
		blockBlackList:         args.BlockBlackList,
		argInterceptorFactory:  argInterceptorFactory,
		maxTxNonceDeltaAllowed: args.MaxTxNonceDeltaAllowed,
		accounts:               args.Accounts,
		antifloodHandler:       args.AntifloodHandler,
		whiteListHandler:       args.WhiteListHandler,
		whiteListerVerifiedTxs: args.WhiteListerVerifiedTxs,
		preferredPeersHolder:   args.PreferredPeersHolder,
		hasher:                 args.CoreComponents.Hasher(),
		requestHandler:         args.RequestHandler,
	}

	icf := &metaInterceptorsContainerFactory{
		baseInterceptorsContainerFactory: base,
	}

	icf.globalThrottler, err = throttler.NewNumGoRoutinesThrottler(numGoRoutines)
	if err != nil {
		return nil, err
	}

	return icf, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (micf *metaInterceptorsContainerFactory) Create() (process.InterceptorsContainer, error) {
	err := micf.generateMetachainHeaderInterceptors()
	if err != nil {
		return nil, err
	}

	err = micf.generateShardHeaderInterceptors()
	if err != nil {
		return nil, err
	}

	err = micf.generateTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = micf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, err
	}

	err = micf.generateRewardTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = micf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, err
	}

	err = micf.generateTrieNodesInterceptors()
	if err != nil {
		return nil, err
	}

	return micf.container, nil
}

// AddShardTrieNodeInterceptors will add the shard trie node interceptors into the existing container
func (micf *metaInterceptorsContainerFactory) AddShardTrieNodeInterceptors(container process.InterceptorsContainer) error {
	if check.IfNil(container) {
		return process.ErrNilInterceptorContainer
	}

	shardC := micf.shardCoordinator

	keys := make([]string, 0)
	trieInterceptors := make([]process.Interceptor, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(i)
		interceptor, err := micf.createOneTrieNodesInterceptor(identifierTrieNodes)
		if err != nil {
			return err
		}

		keys = append(keys, identifierTrieNodes)
		trieInterceptors = append(trieInterceptors, interceptor)
	}

	return container.AddMultiple(keys, trieInterceptors)
}

//------- Shard header interceptors

func (micf *metaInterceptorsContainerFactory) generateShardHeaderInterceptors() error {
	shardC := micf.shardCoordinator
	noOfShards := shardC.NumberOfShards()
	keys := make([]string, noOfShards)
	interceptorsSlice := make([]process.Interceptor, noOfShards)

	//wire up to topics: shardBlocks_0_META, shardBlocks_1_META ...
	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierHeader := factory.ShardBlocksTopic + shardC.CommunicationIdentifier(idx)
		interceptor, err := micf.createOneShardHeaderInterceptor(identifierHeader)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierHeader
		interceptorsSlice[int(idx)] = interceptor
	}

	return micf.container.AddMultiple(keys, interceptorsSlice)
}

func (micf *metaInterceptorsContainerFactory) createOneShardHeaderInterceptor(topic string) (process.Interceptor, error) {
	//TODO implement other HeaderHandlerProcessValidator that will check the header's nonce
	// against blockchain's latest nonce - k finality
	hdrValidator, err := dataValidators.NewNilHeaderValidator()
	if err != nil {
		return nil, err
	}

	hdrFactory, err := interceptorFactory.NewInterceptedShardHeaderDataFactory(micf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:        micf.dataPool.Headers(),
		HdrValidator:   hdrValidator,
		BlockBlackList: micf.blockBlackList,
	}
	hdrProcessor, err := processor.NewHdrInterceptorProcessor(argProcessor)
	if err != nil {
		return nil, err
	}

	interceptor, err := processInterceptors.NewSingleDataInterceptor(
		processInterceptors.ArgSingleDataInterceptor{
			Topic:                topic,
			DataFactory:          hdrFactory,
			Processor:            hdrProcessor,
			Throttler:            micf.globalThrottler,
			AntifloodHandler:     micf.antifloodHandler,
			WhiteListRequest:     micf.whiteListHandler,
			CurrentPeerId:        micf.messenger.ID(),
			PreferredPeersHolder: micf.preferredPeersHolder,
		},
	)
	if err != nil {
		return nil, err
	}

	return micf.createTopicAndAssignHandler(topic, interceptor, true)
}

func (micf *metaInterceptorsContainerFactory) generateTrieNodesInterceptors() error {
	keys := make([]string, 0)
	trieInterceptors := make([]process.Interceptor, 0)

	identifierTrieNodes := factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	interceptor, err := micf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTrieNodes)
	trieInterceptors = append(trieInterceptors, interceptor)

	identifierTrieNodes = factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	interceptor, err = micf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTrieNodes)
	trieInterceptors = append(trieInterceptors, interceptor)

	return micf.container.AddMultiple(keys, trieInterceptors)
}

//------- Reward transactions interceptors

func (micf *metaInterceptorsContainerFactory) generateRewardTxInterceptors() error {
	shardC := micf.shardCoordinator

	noOfShards := shardC.NumberOfShards()

	keys := make([]string, noOfShards)
	interceptorSlice := make([]process.Interceptor, noOfShards)

	for idx := uint32(0); idx < noOfShards; idx++ {
		identifierScr := factory.RewardsTransactionTopic + shardC.CommunicationIdentifier(idx)
		interceptor, err := micf.createOneRewardTxInterceptor(identifierScr)
		if err != nil {
			return err
		}

		keys[int(idx)] = identifierScr
		interceptorSlice[int(idx)] = interceptor
	}

	return micf.container.AddMultiple(keys, interceptorSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (micf *metaInterceptorsContainerFactory) IsInterfaceNil() bool {
	return micf == nil
}
