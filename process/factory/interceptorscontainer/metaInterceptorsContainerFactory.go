package interceptorscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	processInterceptors "github.com/multiversx/mx-chain-go/process/interceptors"
	interceptorFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
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
		args.MainMessenger,
		args.FullArchiveMessenger,
		args.NodesCoordinator,
		args.BlockBlackList,
		args.AntifloodHandler,
		args.WhiteListHandler,
		args.WhiteListerVerifiedTxs,
		args.PreferredPeersHolder,
		args.RequestHandler,
		args.MainPeerShardMapper,
		args.FullArchivePeerShardMapper,
		args.HardforkTrigger,
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
	if check.IfNil(args.SignaturesHandler) {
		return nil, process.ErrNilSignaturesHandler
	}
	if check.IfNil(args.PeerSignatureHandler) {
		return nil, process.ErrNilPeerSignatureHandler
	}
	if args.HeartbeatExpiryTimespanInSec < minTimespanDurationInSec {
		return nil, process.ErrInvalidExpiryTimespan
	}

	argInterceptorFactory := &interceptorFactory.ArgInterceptedDataFactory{
		CoreComponents:               args.CoreComponents,
		CryptoComponents:             args.CryptoComponents,
		ShardCoordinator:             args.ShardCoordinator,
		NodesCoordinator:             args.NodesCoordinator,
		FeeHandler:                   args.TxFeeHandler,
		WhiteListerVerifiedTxs:       args.WhiteListerVerifiedTxs,
		HeaderSigVerifier:            args.HeaderSigVerifier,
		ValidityAttester:             args.ValidityAttester,
		HeaderIntegrityVerifier:      args.HeaderIntegrityVerifier,
		EpochStartTrigger:            args.EpochStartTrigger,
		ArgsParser:                   args.ArgumentsParser,
		PeerSignatureHandler:         args.PeerSignatureHandler,
		SignaturesHandler:            args.SignaturesHandler,
		HeartbeatExpiryTimespanInSec: args.HeartbeatExpiryTimespanInSec,
		PeerID:                       args.MainMessenger.ID(),
	}

	base := &baseInterceptorsContainerFactory{
		mainContainer:              containers.NewInterceptorsContainer(),
		fullArchiveContainer:       containers.NewInterceptorsContainer(),
		shardCoordinator:           args.ShardCoordinator,
		mainMessenger:              args.MainMessenger,
		fullArchiveMessenger:       args.FullArchiveMessenger,
		store:                      args.Store,
		dataPool:                   args.DataPool,
		nodesCoordinator:           args.NodesCoordinator,
		blockBlackList:             args.BlockBlackList,
		argInterceptorFactory:      argInterceptorFactory,
		maxTxNonceDeltaAllowed:     args.MaxTxNonceDeltaAllowed,
		accounts:                   args.Accounts,
		antifloodHandler:           args.AntifloodHandler,
		whiteListHandler:           args.WhiteListHandler,
		whiteListerVerifiedTxs:     args.WhiteListerVerifiedTxs,
		preferredPeersHolder:       args.PreferredPeersHolder,
		hasher:                     args.CoreComponents.Hasher(),
		requestHandler:             args.RequestHandler,
		mainPeerShardMapper:        args.MainPeerShardMapper,
		fullArchivePeerShardMapper: args.FullArchivePeerShardMapper,
		hardforkTrigger:            args.HardforkTrigger,
		nodeOperationMode:          args.NodeOperationMode,
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
func (micf *metaInterceptorsContainerFactory) Create() (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	err := micf.generateMetachainHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generateShardHeaderInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generateTxInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generateRewardTxInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generateTrieNodesInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generatePeerAuthenticationInterceptor()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generateHeartbeatInterceptor()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generatePeerShardInterceptor()
	if err != nil {
		return nil, nil, err
	}

	err = micf.generateValidatorInfoInterceptor()
	if err != nil {
		return nil, nil, err
	}

	return micf.mainContainer, micf.fullArchiveContainer, nil
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

	return micf.addInterceptorsToContainers(keys, interceptorsSlice)
}

func (micf *metaInterceptorsContainerFactory) createOneShardHeaderInterceptor(topic string) (process.Interceptor, error) {
	hdrFactory, err := interceptorFactory.NewInterceptedShardHeaderDataFactory(micf.argInterceptorFactory)
	if err != nil {
		return nil, err
	}

	argProcessor := &processor.ArgHdrInterceptorProcessor{
		Headers:        micf.dataPool.Headers(),
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
			CurrentPeerId:        micf.mainMessenger.ID(),
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

	return micf.addInterceptorsToContainers(keys, trieInterceptors)
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

	return micf.addInterceptorsToContainers(keys, interceptorSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (micf *metaInterceptorsContainerFactory) IsInterfaceNil() bool {
	return micf == nil
}
