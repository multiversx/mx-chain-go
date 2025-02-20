package bootstrap

import (
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	interceptorsFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
)

type epochStartSovereignSyncer struct {
	*epochStartMetaSyncer
}

// internal constructor
func newEpochStartSovereignSyncer(args ArgsNewEpochStartMetaSyncer) (*epochStartSovereignSyncer, error) {
	baseSyncer, err := newEpochStartMetaSyncer(args)
	if err != nil {
		return nil, err
	}

	topicProvider := newSovereignTopicProvider()
	baseSyncer.epochStartTopicProviderHandler = topicProvider
	baseSyncer.singleDataInterceptor, err = createShardSingleDataInterceptor(args)
	if err != nil {
		return nil, err
	}

	return &epochStartSovereignSyncer{
		epochStartMetaSyncer: baseSyncer,
	}, nil
}

func createShardSingleDataInterceptor(args ArgsNewEpochStartMetaSyncer) (process.Interceptor, error) {
	argsInterceptedDataFactory := createArgsInterceptedDataFactory(args)
	interceptedMetaHdrDataFactory, err := interceptorsFactory.NewInterceptedSovereignShardHeaderDataFactory(&argsInterceptedDataFactory)
	if err != nil {
		return nil, err
	}

	return interceptors.NewSingleDataInterceptor(
		interceptors.ArgSingleDataInterceptor{
			Topic:                factory.ShardBlocksTopic,
			DataFactory:          interceptedMetaHdrDataFactory,
			Processor:            args.MetaBlockProcessor,
			Throttler:            disabled.NewThrottler(),
			AntifloodHandler:     disabled.NewAntiFloodHandler(),
			WhiteListRequest:     args.WhitelistHandler,
			CurrentPeerId:        args.Messenger.ID(),
			PreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		},
	)
}
