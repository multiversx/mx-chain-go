package interceptorscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	interceptorFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
)

var _ process.InterceptorsContainerFactory = (*shardInterceptorsContainerFactory)(nil)

// shardInterceptorsContainerFactory will handle the creation the interceptors container for shards
type shardInterceptorsContainerFactory struct {
	*baseInterceptorsContainerFactory
}

// NewShardInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewShardInterceptorsContainerFactory(
	args CommonInterceptorsContainerFactoryArgs,
) (*shardInterceptorsContainerFactory, error) {
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
		intMarshalizer := marshal.NewSizeCheckUnmarshalizer(args.CoreComponents.InternalMarshalizer(), args.SizeCheckDelta)
		err = args.CoreComponents.SetInternalMarshalizer(intMarshalizer)
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
	if check.IfNil(args.ValidityAttester) {
		return nil, process.ErrNilValidityAttester
	}
	if check.IfNil(args.EpochStartTrigger) {
		return nil, process.ErrNilEpochStartTrigger
	}
	if check.IfNil(args.PreferredPeersHolder) {
		return nil, process.ErrNilPreferredPeersHolder
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
		HeaderSigVerifier:            args.HeaderSigVerifier,
		HeaderIntegrityVerifier:      args.HeaderIntegrityVerifier,
		ValidityAttester:             args.ValidityAttester,
		EpochStartTrigger:            args.EpochStartTrigger,
		WhiteListerVerifiedTxs:       args.WhiteListerVerifiedTxs,
		ArgsParser:                   args.ArgumentsParser,
		PeerSignatureHandler:         args.PeerSignatureHandler,
		SignaturesHandler:            args.SignaturesHandler,
		HeartbeatExpiryTimespanInSec: args.HeartbeatExpiryTimespanInSec,
		PeerID:                       args.MainMessenger.ID(),
	}

	base := &baseInterceptorsContainerFactory{
		mainContainer:              containers.NewInterceptorsContainer(),
		fullArchiveContainer:       containers.NewInterceptorsContainer(),
		accounts:                   args.Accounts,
		shardCoordinator:           args.ShardCoordinator,
		mainMessenger:              args.MainMessenger,
		fullArchiveMessenger:       args.FullArchiveMessenger,
		store:                      args.Store,
		dataPool:                   args.DataPool,
		nodesCoordinator:           args.NodesCoordinator,
		argInterceptorFactory:      argInterceptorFactory,
		blockBlackList:             args.BlockBlackList,
		maxTxNonceDeltaAllowed:     args.MaxTxNonceDeltaAllowed,
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

	icf := &shardInterceptorsContainerFactory{
		baseInterceptorsContainerFactory: base,
	}

	icf.globalThrottler, err = throttler.NewNumGoRoutinesThrottler(numGoRoutines)
	if err != nil {
		return nil, err
	}

	return icf, nil
}

// Create returns an interceptor container that will hold all interceptors in the system
func (sicf *shardInterceptorsContainerFactory) Create() (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	err := sicf.generateTxInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateUnsignedTxsInterceptors()
	if err != nil {
		return nil, nil, err
	}

	err = sicf.generateRewardTxInterceptor()
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

	err = sicf.generateMetachainHeaderInterceptors()
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

	return sicf.mainContainer, sicf.fullArchiveContainer, nil
}

func (sicf *shardInterceptorsContainerFactory) generateTrieNodesInterceptors() error {
	shardC := sicf.shardCoordinator

	keys := make([]string, 0)
	interceptorsSlice := make([]process.Interceptor, 0)

	identifierTrieNodes := factory.AccountTrieNodesTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	interceptor, err := sicf.createOneTrieNodesInterceptor(identifierTrieNodes)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTrieNodes)
	interceptorsSlice = append(interceptorsSlice, interceptor)

	return sicf.addInterceptorsToContainers(keys, interceptorsSlice)
}

// ------- Reward transactions interceptors

func (sicf *shardInterceptorsContainerFactory) generateRewardTxInterceptor() error {
	shardC := sicf.shardCoordinator

	keys := make([]string, 0)
	interceptorSlice := make([]process.Interceptor, 0)

	identifierTx := factory.RewardsTransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)
	interceptor, err := sicf.createOneRewardTxInterceptor(identifierTx)
	if err != nil {
		return err
	}

	keys = append(keys, identifierTx)
	interceptorSlice = append(interceptorSlice, interceptor)

	return sicf.addInterceptorsToContainers(keys, interceptorSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sicf *shardInterceptorsContainerFactory) IsInterfaceNil() bool {
	return sicf == nil
}
