package interceptorscontainer

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	interceptorFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
)

var _ process.InterceptorsContainerFactory = (*shardInterceptorsContainerFactory)(nil)

// shardInterceptorsContainerFactory will handle the creation the interceptors container for shards
type shardInterceptorsContainerFactory struct {
	*baseInterceptorsContainerFactory
}

// NewShardInterceptorsContainerFactory is responsible for creating a new interceptors factory object
func NewShardInterceptorsContainerFactory(
	args ShardInterceptorsContainerFactoryArgs,
) (*shardInterceptorsContainerFactory, error) {
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
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
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
		EpochNotifier:             args.EpochNotifier,
	}

	container := containers.NewInterceptorsContainer()
	base := &baseInterceptorsContainerFactory{
		container:              container,
		accounts:               args.Accounts,
		shardCoordinator:       args.ShardCoordinator,
		messenger:              args.Messenger,
		store:                  args.Store,
		dataPool:               args.DataPool,
		nodesCoordinator:       args.NodesCoordinator,
		argInterceptorFactory:  argInterceptorFactory,
		blockBlackList:         args.BlockBlackList,
		maxTxNonceDeltaAllowed: args.MaxTxNonceDeltaAllowed,
		antifloodHandler:       args.AntifloodHandler,
		whiteListHandler:       args.WhiteListHandler,
		whiteListerVerifiedTxs: args.WhiteListerVerifiedTxs,
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
func (sicf *shardInterceptorsContainerFactory) Create() (process.InterceptorsContainer, error) {
	err := sicf.generateTxInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateUnsignedTxsInterceptorsForShard()
	if err != nil {
		return nil, err
	}

	err = sicf.generateRewardTxInterceptor()
	if err != nil {
		return nil, err
	}

	err = sicf.generateHeaderInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateMiniBlocksInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateMetachainHeaderInterceptors()
	if err != nil {
		return nil, err
	}

	err = sicf.generateTrieNodesInterceptors()
	if err != nil {
		return nil, err
	}

	return sicf.container, nil
}

//------- Unsigned transactions interceptors

func (sicf *shardInterceptorsContainerFactory) generateUnsignedTxsInterceptorsForShard() error {
	err := sicf.generateUnsignedTxsInterceptors()
	if err != nil {
		return err
	}

	shardC := sicf.shardCoordinator
	identifierTx := factory.UnsignedTransactionTopic + shardC.CommunicationIdentifier(core.MetachainShardId)

	interceptor, err := sicf.createOneUnsignedTxInterceptor(identifierTx)
	if err != nil {
		return err
	}

	return sicf.container.Add(identifierTx, interceptor)
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

	return sicf.container.AddMultiple(keys, interceptorsSlice)
}

//------- Reward transactions interceptors

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

	return sicf.container.AddMultiple(keys, interceptorSlice)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sicf *shardInterceptorsContainerFactory) IsInterfaceNil() bool {
	return sicf == nil
}
