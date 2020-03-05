package factory

import (
	"math"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/floodPreventers"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// NewP2POutputAntiFlood will return an instance of an output antiflood component based on the config
func NewP2POutputAntiFlood(mainConfig config.Config) (process.P2PAntifloodHandler, error) {
	if mainConfig.Antiflood.Enabled {
		return initP2POutputAntiFlood(mainConfig)
	}

	return &disabledAntiFlood{}, nil
}

func initP2POutputAntiFlood(mainConfig config.Config) (process.P2PAntifloodHandler, error) {
	cacheConfig := storageFactory.GetCacherFromConfig(mainConfig.Antiflood.Cache)
	antifloodCache, err := storageUnit.NewCache(cacheConfig.Type, cacheConfig.Size, cacheConfig.Shards)

	peerMaxMessagesPerSecond := mainConfig.Antiflood.PeerMaxMessagesPerSecond
	peerMaxTotalSizePerSecond := mainConfig.Antiflood.PeerMaxTotalSizePerSecond
	floodPreventer, err := floodPreventers.NewQuotaFloodPreventer(
		antifloodCache,
		make([]floodPreventers.QuotaStatusHandler, 0),
		peerMaxMessagesPerSecond,
		peerMaxTotalSizePerSecond,
		math.MaxUint32,
		math.MaxUint32,
	)
	if err != nil {
		return nil, err
	}

	topicFloodPreventer := floodPreventers.NewNilTopicFloodPreventer()
	startResettingFloodPreventers(floodPreventer, topicFloodPreventer, make([]config.TopicMaxMessagesConfig, 0))

	return antiflood.NewP2PAntiflood(floodPreventer, topicFloodPreventer)
}
