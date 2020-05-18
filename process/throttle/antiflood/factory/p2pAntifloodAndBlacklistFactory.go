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
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/floodPreventers"
	"github.com/ElrondNetwork/elrond-go/statusHandler/p2pQuota"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
)

var log = logger.GetOrCreate("p2p/antiflood/factory")

// AntiFloodComponents holds the handlers returned when creating antiflood and blacklist mechanisms
type AntiFloodComponents struct {
	AntiFloodHandler    process.P2PAntifloodHandler
	BlacklistHandler    process.BlackListHandler
	FloodPreventer      process.FloodPreventer
	TopicFloodPreventer process.TopicFloodPreventer
}

// NewP2PAntiFloodAndBlackList will return instances of antiflood and blacklist, based on the config
func NewP2PAntiFloodAndBlackList(
	config config.Config,
	statusHandler core.AppStatusHandler,
) (*AntiFloodComponents, error) {
	if check.IfNil(statusHandler) {
		return nil, p2p.ErrNilStatusHandler
	}
	if config.Antiflood.Enabled {
		return initP2PAntiFloodAndBlackList(config, statusHandler)
	}

	return &AntiFloodComponents{
		AntiFloodHandler:    &disabledAntiFlood{},
		BlacklistHandler:    &disabledBlacklistHandler{},
		FloodPreventer:      &disabledFloodPreventer{},
		TopicFloodPreventer: &disabledTopicFloodPreventer{},
	}, nil
}

func initP2PAntiFloodAndBlackList(
	mainConfig config.Config,
	statusHandler core.AppStatusHandler,
) (*AntiFloodComponents, error) {
	cacheConfig := storageFactory.GetCacherFromConfig(mainConfig.Antiflood.Cache)
	antifloodCache, err := storageUnit.NewCache(cacheConfig.Type, cacheConfig.Size, cacheConfig.Shards)
	if err != nil {
		return &AntiFloodComponents{}, err
	}

	blackListCache, err := storageUnit.NewCache(cacheConfig.Type, cacheConfig.Size, cacheConfig.Shards)
	if err != nil {
		return &AntiFloodComponents{}, err
	}

	peerMaxMessagesPerSecond := mainConfig.Antiflood.PeerMaxInput.MessagesPerSecond
	peerMaxTotalSizePerSecond := mainConfig.Antiflood.PeerMaxInput.TotalSizePerSecond
	maxMessagesPerSecond := mainConfig.Antiflood.NetworkMaxInput.MessagesPerSecond
	maxTotalSizePerSecond := mainConfig.Antiflood.NetworkMaxInput.TotalSizePerSecond

	quotaProcessor, err := p2pQuota.NewP2PQuotaProcessor(statusHandler)
	if err != nil {
		return &AntiFloodComponents{}, err
	}

	peerBanInSeconds := mainConfig.Antiflood.BlackList.PeerBanDurationInSeconds
	if peerBanInSeconds == 0 {
		return &AntiFloodComponents{}, fmt.Errorf("Antiflood.BlackList.PeerBanDurationInSeconds should be greater than 0")
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
		return &AntiFloodComponents{}, err
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
		return &AntiFloodComponents{}, err
	}

	topicFloodPreventer, err := floodPreventers.NewTopicFloodPreventer(mainConfig.Antiflood.Topic.DefaultMaxMessagesPerSec)
	if err != nil {
		return &AntiFloodComponents{}, err
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
		return &AntiFloodComponents{}, err
	}

	return &AntiFloodComponents{
		AntiFloodHandler:    p2pAntiflood,
		BlacklistHandler:    p2pPeerBlackList,
		FloodPreventer:      floodPreventer,
		TopicFloodPreventer: topicFloodPreventer,
	}, nil
}

func setMaxMessages(topicFloodPreventer process.TopicFloodPreventer, topicMaxMessages []config.TopicMaxMessagesConfig) {
	for _, topicMaxMsg := range topicMaxMessages {
		topicFloodPreventer.SetMaxMessagesForTopic(topicMaxMsg.Topic, topicMaxMsg.NumMessagesPerSec)
	}
}
