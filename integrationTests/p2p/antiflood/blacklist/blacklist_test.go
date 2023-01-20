package blacklist

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/integrationTests/p2p/antiflood"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/blackList"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/floodPreventers"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

var log = logger.GetOrCreate("integrationtests/p2p/antiflooding/blacklist")

// TestAntifloodAndBlacklistWithNumMessages tests what happens if a peer decides to send a large number of messages
// all originating from its peer ID
// All directed peers should add the flooder peer to their black lists and disconnect from it. Further connections
// of the flooder to the flooded peers are no longer possible.
func TestAntifloodAndBlacklistWithNumMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:INFO,p2p:ERROR")
	defer func() {
		_ = logger.SetLogLevel("*:INFO")
	}()

	peers, err := integrationTests.CreateFixedNetworkOf8Peers()
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	// node 3 is connected to 0, 2, 4 and 6 (check integrationTests.CreateFixedNetworkOf7Peers function)
	// large number of broadcast messages from 3 might flood above-mentioned peers but should not flood 5 and 7

	topic := "test_topic"
	broadcastMessageDuration := time.Second * 2
	peerMaxNumProcessMessages := uint32(5)
	maxMessageSize := uint64(1 << 20) // 1MB

	blacklistProcessors, blacklistHandlers := createBlacklistHandlersAndProcessors(
		peers,
		peerMaxNumProcessMessages*2,
		maxMessageSize*2,
		2,
	)
	interceptors, err := antiflood.CreateTopicsAndMockInterceptors(
		peers,
		blacklistProcessors,
		topic,
		peerMaxNumProcessMessages,
		maxMessageSize,
	)
	applyBlacklistComponents(peers, blacklistHandlers)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(antiflood.DurationBootstrapingTime)

	flooderIdx := 3
	floodedIdxes := []int{0, 2, 4, 6}
	floodedIdxesConnections := make([]int, len(floodedIdxes))
	for i, idx := range floodedIdxes {
		floodedIdxesConnections[i] = len(peers[idx].ConnectedPeers())
	}

	// flooder will deactivate its flooding mechanism as to be able to flood the network
	interceptors[flooderIdx].FloodPreventer = nil

	go resetAntifloodStatsOnInterceptors(interceptors)

	fmt.Println("flooding the network")
	isFlooding := atomic.Value{}
	isFlooding.Store(true)
	go antiflood.FloodTheNetwork(peers[flooderIdx], topic, &isFlooding, 10)

	time.Sleep(broadcastMessageDuration)

	isFlooding.Store(false)
	fmt.Println("flooding the network stopped")
	printConnected(peers)

	fmt.Println("waiting for peers to eliminate the flooding peer")
	time.Sleep(time.Second * 10)

	printConnected(peers)
	testConnections(t, peers, flooderIdx, floodedIdxes, floodedIdxesConnections)
	fmt.Println("flooding peer wants to reconnect to the flooded peers (will fail)")

	reConnectFloodingPeer(peers, flooderIdx, floodedIdxes)
	time.Sleep(time.Second * 5)
	printConnected(peers)
	testConnections(t, peers, flooderIdx, floodedIdxes, floodedIdxesConnections)
}

func resetAntifloodStatsOnInterceptors(interceptors []*antiflood.MessageProcessor) {
	for {
		time.Sleep(time.Second)

		for _, interceptor := range interceptors {
			if interceptor.FloodPreventer == nil {
				continue
			}
			interceptor.FloodPreventer.Reset()
		}
	}
}

func testConnections(
	t *testing.T,
	peers []p2p.Messenger,
	flooderIdx int,
	floodedIdxes []int,
	floodedIdxesConnections []int,
) {
	// flooder has 0 connections
	assert.Equal(t, 0, len(peers[flooderIdx].ConnectedPeers()))

	// flooded peers have initial connection - 1 (they eliminated the flooder)
	for i, idx := range floodedIdxes {
		assert.Equal(t, floodedIdxesConnections[i]-1, len(peers[idx].ConnectedPeers()))
	}
}

func reConnectFloodingPeer(peers []p2p.Messenger, flooderIdx int, floodedIdxes []int) {
	flooder := peers[flooderIdx]
	for _, idx := range floodedIdxes {
		_ = flooder.ConnectToPeer(peers[idx].Addresses()[0])
	}
}

func applyBlacklistComponents(peers []p2p.Messenger, blacklistHandler []process.PeerBlackListCacher) {
	for idx, peer := range peers {
		pde, _ := blackList.NewPeerDenialEvaluator(
			blacklistHandler[idx],
			&testscommon.TimeCacheStub{},
			&mock.PeerShardMapperStub{},
		)

		_ = peer.SetPeerDenialEvaluator(pde)
	}
}

func createBlacklistHandlersAndProcessors(
	peers []p2p.Messenger,
	thresholdNumReceived uint32,
	thresholdSizeReceived uint64,
	maxFloodingRounds uint32,
) ([]floodPreventers.QuotaStatusHandler, []process.PeerBlackListCacher) {
	var err error

	blacklistProcessors := make([]floodPreventers.QuotaStatusHandler, len(peers))
	blacklistCachers := make([]process.PeerBlackListCacher, len(peers))
	for i := range peers {
		blacklistCache, _ := cache.NewLRUCache(5000)
		blacklistCachers[i], _ = cache.NewPeerTimeCache(cache.NewTimeCache(time.Minute * 5))

		blacklistProcessors[i], err = blackList.NewP2PBlackListProcessor(
			blacklistCache,
			blacklistCachers[i],
			thresholdNumReceived,
			thresholdSizeReceived,
			maxFloodingRounds,
			time.Minute*5,
			"",
			peers[i].ID(),
		)
		log.LogIfError(err)
	}
	return blacklistProcessors, blacklistCachers
}

func printConnected(peers []p2p.Messenger) {
	fmt.Println("Connected peers:")
	for idx, peer := range peers {
		fmt.Printf("%s, index %d has %d connections\n",
			peer.ID().Pretty(), idx, len(peer.ConnectedPeers()))
	}
	fmt.Println()
}
