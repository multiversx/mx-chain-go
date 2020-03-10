package factory

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/blackList"
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
) (process.P2PAntifloodHandler, p2p.BlacklistHandler, error) {
	if check.IfNil(statusHandler) {
		return nil, nil, p2p.ErrNilStatusHandler
	}
	if config.Antiflood.Enabled {
		return initP2PAntiFloodAndBlackList(config, statusHandler)
	}

	return &disabledAntiFlood{}, &disabledBlacklistHandler{}, nil
}

func initP2PAntiFloodAndBlackList(
	mainConfig config.Config,
	statusHandler core.AppStatusHandler,
) (process.P2PAntifloodHandler, p2p.BlacklistHandler, error) {
	cacheConfig := storageFactory.GetCacherFromConfig(mainConfig.Antiflood.Cache)
	antifloodCache, err := storageUnit.NewCache(cacheConfig.Type, cacheConfig.Size, cacheConfig.Shards)
	if err != nil {
		return nil, nil, err
	}

	blackListCache, err := storageUnit.NewCache(cacheConfig.Type, cacheConfig.Size, cacheConfig.Shards)
	if err != nil {
		return nil, nil, err
	}

	peerMaxMessagesPerSecond := mainConfig.Antiflood.PeerMaxMessagesPerSecond
	peerMaxTotalSizePerSecond := mainConfig.Antiflood.PeerMaxTotalSizePerSecond
	maxMessagesPerSecond := mainConfig.Antiflood.MaxMessagesPerSecond
	maxTotalSizePerSecond := mainConfig.Antiflood.MaxTotalSizePerSecond

	quotaProcessor, err := p2pQuota.NewP2PQuotaProcessor(statusHandler)
	if err != nil {
		return nil, nil, err
	}

	peerBanInSeconds := mainConfig.Antiflood.BlackList.PeerBanDurationInSeconds
	if peerBanInSeconds == 0 {
		return nil, nil, fmt.Errorf("Antiflood.BlackList.PeerBanDurationInSeconds should be greater than 0")
	}

	p2pPeerBlackList := timecache.NewTimeCache(time.Second * time.Duration(peerBanInSeconds))
	blackListProcessor, err := blackList.NewP2PBlackListProcessor(
		blackListCache,
		p2pPeerBlackList,
		mainConfig.Antiflood.BlackList.ThresholdNumMessagesPerSecond,
		mainConfig.Antiflood.BlackList.ThresholdSizePerSecond,
		mainConfig.Antiflood.BlackList.NumFloodingRounds,
	)
	if err != nil {
		return nil, nil, err
	}

	floodPreventer, err := floodPreventers.NewQuotaFloodPreventer(
		antifloodCache,
		[]floodPreventers.QuotaStatusHandler{quotaProcessor, blackListProcessor},
		peerMaxMessagesPerSecond,
		peerMaxTotalSizePerSecond,
		maxMessagesPerSecond,
		maxTotalSizePerSecond,
	)
	if err != nil {
		return nil, nil, err
	}

	topicFloodPreventer, err := floodPreventers.NewTopicFloodPreventer(mainConfig.Antiflood.Topic.DefaultMaxMessagesPerSec)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("started antiflood & blacklist components",
		"peerMaxMessagesPerSecond", peerMaxMessagesPerSecond,
		"peerMaxTotalSizePerSecond", core.ConvertBytes(peerMaxTotalSizePerSecond),
		"maxMessagesPerSecond", maxMessagesPerSecond,
		"maxTotalSizePerSecond", core.ConvertBytes(maxTotalSizePerSecond),
		"peerBanDurationInSeconds", peerBanInSeconds,
		"thresholdNumMessagesPerSecond", mainConfig.Antiflood.BlackList.ThresholdNumMessagesPerSecond,
		"thresholdSizePerSecond", mainConfig.Antiflood.BlackList.ThresholdSizePerSecond,
		"numFloodingRounds", mainConfig.Antiflood.BlackList.NumFloodingRounds,
	)

	topicMaxMessages := mainConfig.Antiflood.Topic.MaxMessages
	setMaxMessages(topicFloodPreventer, topicMaxMessages)

	p2pAntiflood, err := antiflood.NewP2PAntiflood(floodPreventer, topicFloodPreventer)
	if err != nil {
		return nil, nil, err
	}

	startResettingFloodPreventers(floodPreventer, topicFloodPreventer, topicMaxMessages)
	startSweepingP2PPeerBlackList(p2pPeerBlackList)

	return p2pAntiflood, p2pPeerBlackList, nil
}

func setMaxMessages(topicFloodPreventer p2p.TopicFloodPreventer, topicMaxMessages []config.TopicMaxMessagesConfig) {
	for _, topicMaxMsg := range topicMaxMessages {
		topicFloodPreventer.SetMaxMessagesForTopic(topicMaxMsg.Topic, topicMaxMsg.NumMessagesPerSec)
	}
}

func startResettingFloodPreventers(
	floodPreventer p2p.FloodPreventer,
	topicFloodPreventer p2p.TopicFloodPreventer,
	topicMaxMessages []config.TopicMaxMessagesConfig,
) {
	localTopicMaxMessages := make([]config.TopicMaxMessagesConfig, len(topicMaxMessages))
	copy(localTopicMaxMessages, topicMaxMessages)

	go func() {
		for {
			time.Sleep(time.Second)
			floodPreventer.Reset()
			for _, topicMaxMsg := range localTopicMaxMessages {
				topicFloodPreventer.ResetForTopic(topicMaxMsg.Topic)
			}
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
