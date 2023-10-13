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
	Stopped() bool
}

type testOnlySyncedBroadcastNetwork struct {
	*syncedBroadcastNetwork

	numOfBroadcasts           int
	workerPool                workerPool
	messagesFinished          chan bool
	mut                       sync.RWMutex
	peersWithMessagesReceived map[core.PeerID]struct{}
}

// NewTestOnlySyncedBroadcastNetwork -
func NewTestOnlySyncedBroadcastNetwork(numOfBroadcasts int, workerPool workerPool, messagesFinished chan bool) (*testOnlySyncedBroadcastNetwork, error) {
	if workerPool == nil {
		return nil, errors.New("nil worker pool")
	}

	return &testOnlySyncedBroadcastNetwork{
		syncedBroadcastNetwork:    NewSyncedBroadcastNetwork(),
		numOfBroadcasts:           numOfBroadcasts,
		workerPool:                workerPool,
		messagesFinished:          messagesFinished,
		peersWithMessagesReceived: make(map[core.PeerID]struct{}),
	}, nil
}

// Broadcast -
func (network *testOnlySyncedBroadcastNetwork) Broadcast(pid core.PeerID, message p2p.MessageP2P) {
	peers, handlers := network.getPeersAndHandlers()
	totalNumOfPeers := len(peers)

	selfIdx := 0
	for i, peer := range peers {
		if peer == pid {
			selfIdx = i
			break
		}
	}

	// remove self from list, not calling ProcessReceivedMessage for self
	handlers = append(handlers[:selfIdx], handlers[selfIdx+1:]...)
	peers = append(peers[:selfIdx], peers[selfIdx+1:]...)

	indexes := createIndexList(len(handlers))
	shuffledIndexes := random.FisherYatesShuffle(indexes, &random.ConcurrentSafeIntRandomizer{})

	totalBroadcasts := network.numOfBroadcasts
	if len(shuffledIndexes) < network.numOfBroadcasts {
		totalBroadcasts = len(shuffledIndexes)
	}
	for idx := 0; idx < totalBroadcasts; idx++ {
		randIdx := shuffledIndexes[idx]

		if !network.workerPool.Stopped() {
			network.workerPool.Submit(func() {
				time.Sleep(latency)

				handlers[randIdx].receive(pid, message)

				network.mut.Lock()
				defer network.mut.Unlock()
				network.peersWithMessagesReceived[peers[randIdx]] = struct{}{}
				currentLen := len(network.peersWithMessagesReceived)

				if currentLen == totalNumOfPeers {
					go func() { network.messagesFinished <- true }()
				}
			})
		}

	}
}

func createIndexList(listLength int) []int {
	indexes := make([]int, listLength)
	for i := 0; i < listLength; i++ {
		indexes[i] = i
	}

	return indexes
}
