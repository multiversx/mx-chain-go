package libp2p

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/libp2p/go-libp2p-peer"
)

// peersOnChannel manages peers on topics
// it buffers the data and refresh the peers list continuously (in refreshInterval intervals)
type peersOnChannel struct {
	mutPeers    sync.RWMutex
	peers       map[string][]p2p.PeerID
	lastUpdated map[string]time.Time

	refreshInterval   time.Duration
	ttlInterval       time.Duration
	fetchPeersHandler func(topic string) []peer.ID
	getTimeHandler    func() time.Time
}

// newPeersOnChannel returns a new peersOnChannel object
func newPeersOnChannel(
	fetchPeersHandler func(topic string) []peer.ID,
	refreshInterval time.Duration,
	ttlInterval time.Duration,
) (*peersOnChannel, error) {

	if fetchPeersHandler == nil {
		return nil, p2p.ErrNilFetchPeersOnTopicHandler
	}
	if refreshInterval == 0 {
		return nil, p2p.ErrInvalidDurationProvided
	}
	if ttlInterval == 0 {
		return nil, p2p.ErrInvalidDurationProvided
	}

	poc := &peersOnChannel{
		peers:             make(map[string][]p2p.PeerID),
		lastUpdated:       make(map[string]time.Time),
		refreshInterval:   refreshInterval,
		ttlInterval:       ttlInterval,
		fetchPeersHandler: fetchPeersHandler,
	}
	poc.getTimeHandler = poc.clockTime

	go poc.refreshPeersOnAllKnownTopics()

	return poc, nil
}

func (poc *peersOnChannel) clockTime() time.Time {
	return time.Now()
}

// ConnectedPeersOnChannel returns the known peers on a topic
// if the list was not initialized, it will trigger a manual fetch
func (poc *peersOnChannel) ConnectedPeersOnChannel(topic string) []p2p.PeerID {
	poc.mutPeers.RLock()
	peers := poc.peers[topic]
	poc.mutPeers.RUnlock()

	if peers != nil {
		return peers
	}

	return poc.refreshPeersOnTopic(topic)
}

// updateConnectedPeersOnTopic updates the connected peers on a topic and the last update timestamp
func (poc *peersOnChannel) updateConnectedPeersOnTopic(topic string, connectedPeers []p2p.PeerID) {
	poc.mutPeers.Lock()
	poc.peers[topic] = connectedPeers
	poc.lastUpdated[topic] = poc.clockTime()
	poc.mutPeers.Unlock()
}

// refreshPeersOnAllKnownTopics iterates each topic, fetching its last timestamp
// it the timestamp + ttlInterval < time.Now, will trigger a fetch of connected peers on topic
func (poc *peersOnChannel) refreshPeersOnAllKnownTopics() {
	for {
		listTopicsToBeRefreshed := make([]string, 0)

		//build required topic list
		poc.mutPeers.RLock()
		for topic, lastRefreshed := range poc.lastUpdated {
			needsToBeRefreshed := poc.getTimeHandler().Sub(lastRefreshed) > poc.ttlInterval
			if needsToBeRefreshed {
				listTopicsToBeRefreshed = append(listTopicsToBeRefreshed, topic)
			}
		}
		poc.mutPeers.RUnlock()

		for _, topic := range listTopicsToBeRefreshed {
			_ = poc.refreshPeersOnTopic(topic)
		}

		time.Sleep(poc.refreshInterval)
	}
}

// refreshPeersOnTopic
func (poc *peersOnChannel) refreshPeersOnTopic(topic string) []p2p.PeerID {
	list := poc.fetchPeersHandler(topic)
	connectedPeers := make([]p2p.PeerID, len(list))
	for i, pid := range list {
		connectedPeers[i] = p2p.PeerID(pid)
	}

	poc.updateConnectedPeersOnTopic(topic, connectedPeers)
	return connectedPeers
}
