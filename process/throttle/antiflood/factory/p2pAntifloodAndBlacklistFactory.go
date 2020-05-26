package factory

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/blackList"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/disabled"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/floodPreventers"
	"github.com/ElrondNetwork/elrond-go/statusHandler/p2pQuota"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
)

var durationSweepP2PBlacklist = time.Second * 5
var log = logger.GetOrCreate("p2p/antiflood/factory")

// NewP2PAntiFloodAndBlackList will return instances of antiflood and blacklist, based on the config
func NewP2PAntiFloodAndBlackList(
	config config.Config,
	statusHandler core.AppStatusHandler,
) (process.P2PAntifloodHandler, process.BlackListHandler, error) {
	if check.IfNil(statusHandler) {
		return nil, nil, p2p.ErrNilStatusHandler
	}
	if config.Antiflood.Enabled {
		return initP2PAntiFloodAndBlackList(config, statusHandler)
	}

	return &disabled.AntiFlood{}, &disabled.BlacklistHandler{}, nil
}

func initP2PAntiFloodAndBlackList(
	mainConfig config.Config,
	statusHandler core.AppStatusHandler,
) (process.P2PAntifloodHandler, process.BlackListHandler, error) {
	p2pPeerBlackList := timecache.NewTimeCache(300 * time.Second)

	fastReactingFloodPreventer, err := createFloodPreventer(
		mainConfig.Antiflood.FastReacting,
		mainConfig.Antiflood.Cache,
		statusHandler,
		"fast_reacting",
		p2pPeerBlackList,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%w when creating fast reacting flood preventer", err)
	}

	slowReactingFloodPreventer, err := createFloodPreventer(
		mainConfig.Antiflood.SlowReacting,
		mainConfig.Antiflood.Cache,
		statusHandler,
		"slow_reacting",
		p2pPeerBlackList,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%w when creating fast reacting flood preventer", err)
	}

	topicFloodPreventer, err := floodPreventers.NewTopicFloodPreventer(mainConfig.Antiflood.Topic.DefaultMaxMessagesPerSec)
	if err != nil {
		return nil, nil, err
	}

	topicMaxMessages := mainConfig.Antiflood.Topic.MaxMessages
	setMaxMessages(topicFloodPreventer, topicMaxMessages)

	p2pAntiflood, err := antiflood.NewP2PAntiflood(
		p2pPeerBlackList,
		topicFloodPreventer,
		fastReactingFloodPreventer,
		slowReactingFloodPreventer,
	)
	if err != nil {
		return nil, nil, err
	}

	startResettingTopicFloodPreventer(topicFloodPreventer, topicMaxMessages)
	startSweepingP2PPeerBlackList(p2pPeerBlackList)

	return p2pAntiflood, p2pPeerBlackList, nil
}

func setMaxMessages(topicFloodPreventer process.TopicFloodPreventer, topicMaxMessages []config.TopicMaxMessagesConfig) {
	for _, topicMaxMsg := range topicMaxMessages {
		topicFloodPreventer.SetMaxMessagesForTopic(topicMaxMsg.Topic, topicMaxMsg.NumMessagesPerSec)
	}
}

func startResettingTopicFloodPreventer(
	topicFloodPreventer process.TopicFloodPreventer,
	topicMaxMessages []config.TopicMaxMessagesConfig,
	floodPreventers ...process.FloodPreventer,
) {
	localTopicMaxMessages := make([]config.TopicMaxMessagesConfig, len(topicMaxMessages))
	copy(localTopicMaxMessages, topicMaxMessages)

	go func() {
		for {
			time.Sleep(time.Second)
			for _, fp := range floodPreventers {
				fp.Reset()
			}
			for _, topicMaxMsg := range localTopicMaxMessages {
				topicFloodPreventer.ResetForTopic(topicMaxMsg.Topic)
			}
			topicFloodPreventer.ResetForNotRegisteredTopics()
		}
	}()
}

func startSweepingP2PPeerBlackList(p2pPeerBlackList process.BlackListHandler) {
	go func() {
		for {
			time.Sleep(durationSweepP2PBlacklist)
			p2pPeerBlackList.Sweep()
		}
	}()
}

func createFloodPreventer(
	floodPreventerConfig config.FloodPreventerConfig,
	antifloodCacheConfig config.CacheConfig,
	statusHandler core.AppStatusHandler,
	quotaIdentifier string,
	blackListHandler process.BlackListHandler,
) (process.FloodPreventer, error) {
	cacheConfig := storageFactory.GetCacherFromConfig(antifloodCacheConfig)
	blackListCache, err := storageUnit.NewCache(cacheConfig.Type, cacheConfig.Size, cacheConfig.Shards)
	if err != nil {
		return nil, err
	}

	blackListProcessor, err := blackList.NewP2PBlackListProcessor(
		blackListCache,
		blackListHandler,
		floodPreventerConfig.BlackList.ThresholdNumMessagesPerInterval,
		floodPreventerConfig.BlackList.ThresholdSizePerInterval,
		floodPreventerConfig.BlackList.NumFloodingRounds,
		time.Duration(floodPreventerConfig.IntervalInSeconds)*time.Second,
	)
	if err != nil {
		return nil, err
	}

	antifloodCache, err := storageUnit.NewCache(cacheConfig.Type, cacheConfig.Size, cacheConfig.Shards)
	if err != nil {
		return nil, err
	}

	quotaProcessor, err := p2pQuota.NewP2PQuotaProcessor(statusHandler, quotaIdentifier)
	if err != nil {
		return nil, err
	}

	peerMaxMessagesPerSecond := floodPreventerConfig.PeerMaxInput.MessagesPerInterval
	peerMaxTotalSizePerSecond := floodPreventerConfig.PeerMaxInput.TotalSizePerInterval
	reservedPercent := floodPreventerConfig.ReservedPercent

	floodPreventer, err := floodPreventers.NewQuotaFloodPreventer(
		antifloodCache,
		[]floodPreventers.QuotaStatusHandler{quotaProcessor, blackListProcessor},
		peerMaxMessagesPerSecond,
		peerMaxTotalSizePerSecond,
		reservedPercent,
	)
	if err != nil {
		return nil, err
	}

	log.Debug("started antiflood & blacklist component",
		"type", quotaIdentifier,
		"interval in seconds", floodPreventerConfig.IntervalInSeconds,
		"peerMaxMessagesPerInterval", peerMaxMessagesPerSecond,
		"peerMaxTotalSizePerInterval", core.ConvertBytes(peerMaxTotalSizePerSecond),
		"peerBanDurationInSeconds", floodPreventerConfig.BlackList.PeerBanDurationInSeconds,
		"thresholdNumMessagesPerSecond", floodPreventerConfig.BlackList.ThresholdNumMessagesPerInterval,
		"thresholdSizePerSecond", floodPreventerConfig.BlackList.ThresholdSizePerInterval,
		"numFloodingRounds", floodPreventerConfig.BlackList.NumFloodingRounds,
	)

	go func() {
		wait := time.Duration(floodPreventerConfig.IntervalInSeconds) * time.Second

		for {
			time.Sleep(wait)
			floodPreventer.Reset()
		}
	}()

	return floodPreventer, nil
}
