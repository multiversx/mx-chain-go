package factory

import (
	"context"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/disabled"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/floodPreventers"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

const outputReservedPercent = float32(0)

// NewP2POutputAntiFlood will return an instance of an output antiflood component based on the config
func NewP2POutputAntiFlood(
	ctx context.Context,
	antifloodConfigsHandler common.AntifloodConfigsHandler,
) (process.P2PAntifloodHandler, error) {
	if antifloodConfigsHandler.IsEnabled() {
		return initP2POutputAntiFlood(ctx, antifloodConfigsHandler)
	}

	return &disabled.AntiFlood{}, nil
}

func initP2POutputAntiFlood(
	ctx context.Context,
	antifloodConfigsHandler common.AntifloodConfigsHandler,
) (process.P2PAntifloodHandler, error) {
	currentConfig := antifloodConfigsHandler.GetCurrentConfig()

	cacheConfig := storageFactory.GetCacherFromConfig(currentConfig.Cache)
	antifloodCache, err := storageunit.NewCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	arg := floodPreventers.ArgQuotaFloodPreventer{
		Name:             antiflood.OutputIdentifier,
		Cacher:           antifloodCache,
		StatusHandlers:   make([]floodPreventers.QuotaStatusHandler, 0),
		AntifloodConfigs: antifloodConfigsHandler,
		ConfigFetcher: func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
			return config.FloodPreventerConfig{} // not needed on this flow
		},
	}

	floodPreventer, err := floodPreventers.NewQuotaFloodPreventer(arg)
	if err != nil {
		return nil, err
	}

	topicFloodPreventer := disabled.NewNilTopicFloodPreventer()
	startResettingTopicFloodPreventer(ctx, topicFloodPreventer, make([]config.TopicMaxMessagesConfig, 0), floodPreventer)

	return antiflood.NewP2PAntiflood(&disabled.PeerBlacklistCacher{}, topicFloodPreventer, floodPreventer)
}
