package antiflood

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/floodPreventers"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

// DurationBootstrapingTime -
const DurationBootstrapingTime = 2 * time.Second

// FloodTheNetwork -
func FloodTheNetwork(peer p2p.Messenger, topic string, isFlooding *atomic.Value, messageSize uint64) {
	for {
		peer.BroadcastOnChannel(topic, topic, make([]byte, messageSize))

		if !isFlooding.Load().(bool) {
			return
		}
	}
}

// CreateTopicsAndMockInterceptors -
func CreateTopicsAndMockInterceptors(
	peers []p2p.Messenger,
	blacklistHandlers []floodPreventers.QuotaStatusHandler,
	topic string,
	peerMaxNumMessages uint32,
	peerMaxSize uint64,
) ([]*MessageProcessor, error) {

	interceptors := make([]*MessageProcessor, len(peers))

	for idx, p := range peers {
		err := p.CreateTopic(topic, true)
		if err != nil {
			return nil, fmt.Errorf("%w, pid: %s", err, p.ID())
		}

		cacherCfg := storageunit.CacheConfig{Capacity: 100, Type: storageunit.LRUCache, Shards: 1}
		antifloodPool, _ := storageunit.NewCache(cacherCfg)

		interceptors[idx] = newMessageProcessor()
		statusHandlers := []floodPreventers.QuotaStatusHandler{&nilQuotaStatusHandler{}}
		if len(blacklistHandlers) == len(peers) {
			statusHandlers = append(statusHandlers, blacklistHandlers[idx])
		}
		arg := floodPreventers.ArgQuotaFloodPreventer{
			Name:                      "test",
			Cacher:                    antifloodPool,
			StatusHandlers:            statusHandlers,
			BaseMaxNumMessagesPerPeer: peerMaxNumMessages,
			MaxTotalSizePerPeer:       peerMaxSize,
			PercentReserved:           0,
			IncreaseThreshold:         0,
			IncreaseFactor:            0,
		}
		interceptors[idx].FloodPreventer, err = floodPreventers.NewQuotaFloodPreventer(arg)
		if err != nil {
			return nil, err
		}

		err = p.RegisterMessageProcessor(topic, "test", interceptors[idx])
		if err != nil {
			return nil, fmt.Errorf("%w, pid: %s", err, p.ID())
		}
	}

	return interceptors, nil
}
