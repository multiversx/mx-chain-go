package factory

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/disabled"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/floodPreventers"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

const outputReservedPercent = float32(0)

// NewP2POutputAntiFlood will return an instance of an output antiflood component based on the config
func NewP2POutputAntiFlood(ctx context.Context, mainConfig config.Config) (process.P2PAntifloodHandler, error) {
	if mainConfig.Antiflood.Enabled {
		return initP2POutputAntiFlood(ctx, mainConfig)
	}

	return &disabled.AntiFlood{}, nil
}

func initP2POutputAntiFlood(ctx context.Context, mainConfig config.Config) (process.P2PAntifloodHandler, error) {
	cacheConfig := storageFactory.GetCacherFromConfig(mainConfig.Antiflood.Cache)
	antifloodCache, err := storageUnit.NewCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	basePeerMaxMessagesPerInterval := mainConfig.Antiflood.PeerMaxOutput.BaseMessagesPerInterval
	peerMaxTotalSizePerInterval := mainConfig.Antiflood.PeerMaxOutput.TotalSizePerInterval
	arg := floodPreventers.ArgQuotaFloodPreventer{
		Name:                      outputIdentifier,
		Cacher:                    antifloodCache,
		StatusHandlers:            make([]floodPreventers.QuotaStatusHandler, 0),
		BaseMaxNumMessagesPerPeer: basePeerMaxMessagesPerInterval,
		MaxTotalSizePerPeer:       peerMaxTotalSizePerInterval,
		PercentReserved:           outputReservedPercent,
		IncreaseThreshold:         0,
		IncreaseFactor:            0,
	}

	floodPreventer, err := floodPreventers.NewQuotaFloodPreventer(arg)
	if err != nil {
		return nil, err
	}

	topicFloodPreventer := disabled.NewNilTopicFloodPreventer()
	startResettingTopicFloodPreventer(ctx, topicFloodPreventer, make([]config.TopicMaxMessagesConfig, 0), floodPreventer)

	return antiflood.NewP2PAntiflood(&disabled.PeerBlacklistCacher{}, topicFloodPreventer, floodPreventer)
}
