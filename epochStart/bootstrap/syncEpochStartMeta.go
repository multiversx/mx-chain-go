package bootstrap

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	interceptorsFactory "github.com/ElrondNetwork/elrond-go/process/interceptors/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ epochStart.StartOfEpochMetaSyncer = (*epochStartMetaSyncer)(nil)

type epochStartMetaSyncer struct {
	requestHandler        RequestHandler
	messenger             Messenger
	marshalizer           marshal.Marshalizer
	hasher                hashing.Hasher
	singleDataInterceptor process.Interceptor
	metaBlockProcessor    EpochStartMetaBlockInterceptorProcessor
}

// ArgsNewEpochStartMetaSyncer -
type ArgsNewEpochStartMetaSyncer struct {
	CoreComponentsHolder    process.CoreComponentsHolder
	CryptoComponentsHolder  process.CryptoComponentsHolder
	RequestHandler          RequestHandler
	Messenger               Messenger
	ShardCoordinator        sharding.Coordinator
	EconomicsData           process.EconomicsDataHandler
	WhitelistHandler        process.WhiteListHandler
	StartInEpochConfig      config.EpochStartConfig
	ArgsParser              process.ArgumentsParser
	HeaderIntegrityVerifier process.HeaderIntegrityVerifier
	MetaBlockProcessor      EpochStartMetaBlockInterceptorProcessor
}

// NewEpochStartMetaSyncer will return a new instance of epochStartMetaSyncer
func NewEpochStartMetaSyncer(args ArgsNewEpochStartMetaSyncer) (*epochStartMetaSyncer, error) {
	if check.IfNil(args.CoreComponentsHolder) {
		return nil, epochStart.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.CryptoComponentsHolder) {
		return nil, epochStart.ErrNilCryptoComponentsHolder
	}
	if check.IfNil(args.CoreComponentsHolder.AddressPubKeyConverter()) {
		return nil, epochStart.ErrNilPubkeyConverter
	}
	if check.IfNil(args.HeaderIntegrityVerifier) {
		return nil, epochStart.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(args.MetaBlockProcessor) {
		return nil, epochStart.ErrNilMetablockProcessor
	}

	e := &epochStartMetaSyncer{
		requestHandler:     args.RequestHandler,
		messenger:          args.Messenger,
		marshalizer:        args.CoreComponentsHolder.InternalMarshalizer(),
		hasher:             args.CoreComponentsHolder.Hasher(),
		metaBlockProcessor: args.MetaBlockProcessor,
	}

	argsInterceptedDataFactory := interceptorsFactory.ArgInterceptedDataFactory{
		CoreComponents:          args.CoreComponentsHolder,
		CryptoComponents:        args.CryptoComponentsHolder,
		ShardCoordinator:        args.ShardCoordinator,
		NodesCoordinator:        disabled.NewNodesCoordinator(),
		FeeHandler:              args.EconomicsData,
		HeaderSigVerifier:       disabled.NewHeaderSigVerifier(),
		HeaderIntegrityVerifier: args.HeaderIntegrityVerifier,
		ValidityAttester:        disabled.NewValidityAttester(),
		EpochStartTrigger:       disabled.NewEpochStartTrigger(),
		ArgsParser:              args.ArgsParser,
	}

	interceptedMetaHdrDataFactory, err := interceptorsFactory.NewInterceptedMetaHeaderDataFactory(&argsInterceptedDataFactory)
	if err != nil {
		return nil, err
	}

	e.singleDataInterceptor, err = interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:            factory.MetachainBlocksTopic,
			DataFactory:      interceptedMetaHdrDataFactory,
			Processor:        args.MetaBlockProcessor,
			Throttler:        disabled.NewThrottler(),
			AntifloodHandler: disabled.NewAntiFloodHandler(),
			WhiteListRequest: args.WhitelistHandler,
			CurrentPeerId:    args.Messenger.ID(),
		},
	)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// SyncEpochStartMeta syncs the latest epoch start metablock
func (e *epochStartMetaSyncer) SyncEpochStartMeta(timeToWait time.Duration) (data.MetaHeaderHandler, error) {
	err := e.initTopicForEpochStartMetaBlockInterceptor()
	if err != nil {
		return nil, err
	}
	defer func() {
		e.resetTopicsAndInterceptors()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeToWait)
	mb, errConsensusNotReached := e.metaBlockProcessor.GetEpochStartMetaBlock(ctx)
	cancel()

	if errConsensusNotReached != nil {
		return nil, errConsensusNotReached
	}

	return mb, nil
}

func (e *epochStartMetaSyncer) resetTopicsAndInterceptors() {
	err := e.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Trace("error unregistering message processors", "error", err)
	}
}

func (e *epochStartMetaSyncer) initTopicForEpochStartMetaBlockInterceptor() error {
	err := e.messenger.CreateTopic(factory.MetachainBlocksTopic, true)
	if err != nil {
		log.Warn("error messenger create topic", "error", err)
		return err
	}

	e.resetTopicsAndInterceptors()
	err = e.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, e.singleDataInterceptor)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (e *epochStartMetaSyncer) IsInterfaceNil() bool {
	return e == nil
}
