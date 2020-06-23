package factory

import (
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
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

const defaultSpan = 300 * time.Second
const fastReactingIdentifier = "fast_reacting"
const slowReactingIdentifier = "slow_reacting"
const outOfSpecsIdentifier = "out_of_specs"
const outputIdentifier = "output"

// NewP2PAntiFloodAndBlackList will return instances of antiflood and blacklist, based on the config
func NewP2PAntiFloodAndBlackList(
	config config.Config,
	statusHandler core.AppStatusHandler,
	currentPid core.PeerID,
) (process.P2PAntifloodHandler, process.PeerBlackListCacher, process.TimeCacher, error) {
	if check.IfNil(statusHandler) {
		return nil, nil, nil, p2p.ErrNilStatusHandler
	}
	if config.Antiflood.Enabled {
		return initP2PAntiFloodAndBlackList(config, statusHandler, currentPid)
	}

	return &disabled.AntiFlood{}, &disabled.PeerBlacklistCacher{}, &disabled.TimeCache{}, nil
}

func initP2PAntiFloodAndBlackList(
	mainConfig config.Config,
	statusHandler core.AppStatusHandler,
	currentPid core.PeerID,
) (process.P2PAntifloodHandler, process.PeerBlackListCacher, process.TimeCacher, error) {
	cache := timecache.NewTimeCache(defaultSpan)
	p2pPeerBlackList, err := timecache.NewPeerTimeCache(cache)
	if err != nil {
		return nil, nil, nil, err
	}

	publicKeysCache := timecache.NewTimeCache(defaultSpan)

	fastReactingFloodPreventer, err := createFloodPreventer(
		mainConfig.Antiflood.FastReacting,
		mainConfig.Antiflood.Cache,
		statusHandler,
		fastReactingIdentifier,
		p2pPeerBlackList,
		currentPid,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w when creating fast reacting flood preventer", err)
	}

	slowReactingFloodPreventer, err := createFloodPreventer(
		mainConfig.Antiflood.SlowReacting,
		mainConfig.Antiflood.Cache,
		statusHandler,
		slowReactingIdentifier,
		p2pPeerBlackList,
		currentPid,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w when creating fast reacting flood preventer", err)
	}

	outOfSpecsFloodPreventer, err := createFloodPreventer(
		mainConfig.Antiflood.OutOfSpecs,
		mainConfig.Antiflood.Cache,
		statusHandler,
		outOfSpecsIdentifier,
		p2pPeerBlackList,
		currentPid,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w when creating out of specs flood preventer", err)
	}

	topicFloodPreventer, err := floodPreventers.NewTopicFloodPreventer(mainConfig.Antiflood.Topic.DefaultMaxMessagesPerSec)
	if err != nil {
		return nil, nil, nil, err
	}

	topicMaxMessages := mainConfig.Antiflood.Topic.MaxMessages
	setMaxMessages(topicFloodPreventer, topicMaxMessages)

	p2pAntiflood, err := antiflood.NewP2PAntiflood(
		p2pPeerBlackList,
		topicFloodPreventer,
		fastReactingFloodPreventer,
		slowReactingFloodPreventer,
		outOfSpecsFloodPreventer,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	startResettingTopicFloodPreventer(topicFloodPreventer, topicMaxMessages)
	startSweepingTimeCaches(p2pPeerBlackList, publicKeysCache)

	return p2pAntiflood, p2pPeerBlackList, publicKeysCache, nil
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

func startSweepingTimeCaches(p2pPeerBlackList process.PeerBlackListCacher, publicKeysCache process.TimeCacher) {
	go func() {
		for {
			time.Sleep(durationSweepP2PBlacklist)
			p2pPeerBlackList.Sweep()
			publicKeysCache.Sweep()
		}
	}()
}

func createFloodPreventer(
	floodPreventerConfig config.FloodPreventerConfig,
	antifloodCacheConfig config.CacheConfig,
	statusHandler core.AppStatusHandler,
	quotaIdentifier string,
	blackListHandler process.PeerBlackListCacher,
	selfPid core.PeerID,
) (process.FloodPreventer, error) {
	cacheConfig := storageFactory.GetCacherFromConfig(antifloodCacheConfig)
	blackListCache, err := storageUnit.NewCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	blackListProcessor, err := blackList.NewP2PBlackListProcessor(
		blackListCache,
		blackListHandler,
		floodPreventerConfig.BlackList.ThresholdNumMessagesPerInterval,
		floodPreventerConfig.BlackList.ThresholdSizePerInterval,
		floodPreventerConfig.BlackList.NumFloodingRounds,
		time.Duration(floodPreventerConfig.BlackList.PeerBanDurationInSeconds)*time.Second,
		quotaIdentifier,
		selfPid,
	)
	if err != nil {
		return nil, err
	}

	antifloodCache, err := storageUnit.NewCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	quotaProcessor, err := p2pQuota.NewP2PQuotaProcessor(statusHandler, quotaIdentifier)
	if err != nil {
		return nil, err
	}

	basePeerMaxMessagesPerInterval := floodPreventerConfig.PeerMaxInput.BaseMessagesPerInterval
	peerMaxTotalSizePerInterval := floodPreventerConfig.PeerMaxInput.TotalSizePerInterval
	reservedPercent := floodPreventerConfig.ReservedPercent

	argFloodPreventer := floodPreventers.ArgQuotaFloodPreventer{
		Name:                      quotaIdentifier,
		Cacher:                    antifloodCache,
		StatusHandlers:            []floodPreventers.QuotaStatusHandler{quotaProcessor, blackListProcessor},
		BaseMaxNumMessagesPerPeer: basePeerMaxMessagesPerInterval,
		MaxTotalSizePerPeer:       peerMaxTotalSizePerInterval,
		PercentReserved:           reservedPercent,
		IncreaseThreshold:         floodPreventerConfig.PeerMaxInput.IncreaseFactor.Threshold,
		IncreaseFactor:            floodPreventerConfig.PeerMaxInput.IncreaseFactor.Factor,
	}
	floodPreventer, err := floodPreventers.NewQuotaFloodPreventer(argFloodPreventer)
	if err != nil {
		return nil, err
	}

	log.Debug("started antiflood & blacklist component",
		"type", quotaIdentifier,
		"interval in seconds", floodPreventerConfig.IntervalInSeconds,
		"base peerMaxMessagesPerInterval", basePeerMaxMessagesPerInterval,
		"peerMaxTotalSizePerInterval", core.ConvertBytes(peerMaxTotalSizePerInterval),
		"peerBanDurationInSeconds", floodPreventerConfig.BlackList.PeerBanDurationInSeconds,
		"thresholdNumMessagesPerSecond", floodPreventerConfig.BlackList.ThresholdNumMessagesPerInterval,
		"thresholdSizePerSecond", floodPreventerConfig.BlackList.ThresholdSizePerInterval,
		"numFloodingRounds", floodPreventerConfig.BlackList.NumFloodingRounds,
		"increase threshold", floodPreventerConfig.PeerMaxInput.IncreaseFactor.Threshold,
		"increase factor", floodPreventerConfig.PeerMaxInput.IncreaseFactor.Factor,
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
