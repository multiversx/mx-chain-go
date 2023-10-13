package chainSimulator

import (
	"errors"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/random"
)

const latency = time.Millisecond * 30

type workerPool interface {
	Submit(task func())
}

type testOnlySyncedBroadcastNetwork struct {
	*syncedBroadcastNetwork

	maxNumOfConnections       int
	maxNumOfBroadcasts        int
	workerPool                workerPool
	messagesFinished          chan bool
	mut                       sync.RWMutex
	peersWithMessagesReceived map[core.PeerID]struct{}
}

// NewTestOnlySyncedBroadcastNetwork -
func NewTestOnlySyncedBroadcastNetwork(maxNumOfConnections int, maxNumOfBroadcasts int, workerPool workerPool, messagesFinished chan bool) (*testOnlySyncedBroadcastNetwork, error) {
	if workerPool == nil {
		return nil, errors.New("nil worker pool")
	}

	return &testOnlySyncedBroadcastNetwork{
		syncedBroadcastNetwork:    NewSyncedBroadcastNetwork(),
		maxNumOfConnections:       maxNumOfConnections,
		maxNumOfBroadcasts:        maxNumOfBroadcasts,
		workerPool:                workerPool,
		messagesFinished:          messagesFinished,
		peersWithMessagesReceived: make(map[core.PeerID]struct{}),
	}, nil
}

// Broadcast -
func (network *testOnlySyncedBroadcastNetwork) Broadcast(pid core.PeerID, message p2p.MessageP2P) {
	peers, handlers := network.getPeersAndHandlers()
	totalNumOfPeers := len(peers)

	totalPeers := network.maxNumOfConnections
	if len(peers) < totalPeers {
		totalPeers = len(peers)
	}

	peers = peers[:totalPeers]
	handlers = handlers[:totalPeers]

	network.mut.Lock()
	network.peersWithMessagesReceived[pid] = struct{}{}
	currentLen := len(network.peersWithMessagesReceived)
	network.mut.Unlock()
	if currentLen == totalNumOfPeers {
		network.messagesFinished <- true
		return
	}

	selfIdx := 0
	for i, peer := range peers {
		if peer == pid {
			selfIdx = i
			break
		}
	}

	handlers = append(handlers[:selfIdx], handlers[selfIdx+1:]...)

	indexes := createIndexList(len(handlers))
	shuffledIndexes := random.FisherYatesShuffle(indexes, &random.ConcurrentSafeIntRandomizer{})

	totalBroadcasts := network.maxNumOfBroadcasts
	if len(shuffledIndexes) < network.maxNumOfBroadcasts {
		totalBroadcasts = len(shuffledIndexes)
	}
	for idx := 0; idx < totalBroadcasts; idx++ {
		randIdx := shuffledIndexes[idx]

		network.workerPool.Submit(func() {
			time.Sleep(latency)
			handlers[randIdx].receive(pid, message)
		})
	}
}

func createIndexList(listLength int) []int {
	indexes := make([]int, listLength)
	for i := 0; i < listLength; i++ {
		indexes[i] = i
	}

	return indexes
}
