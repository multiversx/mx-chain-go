package libp2p

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func TestNewPeersOnChannel_NilPeersRatingHandlerShouldErr(t *testing.T) {
	t.Parallel()

	poc, err := newPeersOnChannel(nil, nil, 1, 1)

	assert.Nil(t, poc)
	assert.Equal(t, p2p.ErrNilPeersRatingHandler, err)
}

func TestNewPeersOnChannel_NilFetchPeersHandlerShouldErr(t *testing.T) {
	t.Parallel()

	poc, err := newPeersOnChannel(&p2pmocks.PeersRatingHandlerStub{}, nil, 1, 1)

	assert.Nil(t, poc)
	assert.Equal(t, p2p.ErrNilFetchPeersOnTopicHandler, err)
}

func TestNewPeersOnChannel_InvalidRefreshIntervalShouldErr(t *testing.T) {
	t.Parallel()

	poc, err := newPeersOnChannel(
		&p2pmocks.PeersRatingHandlerStub{},
		func(topic string) []peer.ID {
			return nil
		},
		0,
		1)

	assert.Nil(t, poc)
	assert.Equal(t, p2p.ErrInvalidDurationProvided, err)
}

func TestNewPeersOnChannel_InvalidTTLIntervalShouldErr(t *testing.T) {
	t.Parallel()

	poc, err := newPeersOnChannel(
		&p2pmocks.PeersRatingHandlerStub{},
		func(topic string) []peer.ID {
			return nil
		},
		1,
		0)

	assert.Nil(t, poc)
	assert.Equal(t, p2p.ErrInvalidDurationProvided, err)
}

func TestNewPeersOnChannel_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	poc, err := newPeersOnChannel(
		&p2pmocks.PeersRatingHandlerStub{},
		func(topic string) []peer.ID {
			return nil
		},
		1,
		1)

	assert.NotNil(t, poc)
	assert.Nil(t, err)
}

func TestPeersOnChannel_ConnectedPeersOnChannelMissingTopicShouldTriggerFetchAndReturn(t *testing.T) {
	t.Parallel()

	retPeerIDs := []peer.ID{"peer1", "peer2"}
	testTopic := "test_topic"
	wasFetchCalled := atomic.Value{}
	wasFetchCalled.Store(false)

	poc, _ := newPeersOnChannel(
		&p2pmocks.PeersRatingHandlerStub{},
		func(topic string) []peer.ID {
			if topic == testTopic {
				wasFetchCalled.Store(true)
				return retPeerIDs
			}
			return nil
		},
		time.Second,
		time.Second,
	)

	peers := poc.ConnectedPeersOnChannel(testTopic)

	assert.True(t, wasFetchCalled.Load().(bool))
	for idx, pid := range retPeerIDs {
		assert.Equal(t, []byte(pid), peers[idx].Bytes())
	}
}

func TestPeersOnChannel_ConnectedPeersOnChannelFindTopicShouldReturn(t *testing.T) {
	t.Parallel()

	retPeerIDs := []core.PeerID{"peer1", "peer2"}
	testTopic := "test_topic"
	wasFetchCalled := atomic.Value{}
	wasFetchCalled.Store(false)

	poc, _ := newPeersOnChannel(
		&p2pmocks.PeersRatingHandlerStub{},
		func(topic string) []peer.ID {
			wasFetchCalled.Store(true)
			return nil
		},
		time.Second,
		time.Second,
	)
	// manually put peers
	poc.mutPeers.Lock()
	poc.peers[testTopic] = retPeerIDs
	poc.mutPeers.Unlock()

	peers := poc.ConnectedPeersOnChannel(testTopic)

	assert.False(t, wasFetchCalled.Load().(bool))
	for idx, pid := range retPeerIDs {
		assert.Equal(t, []byte(pid), peers[idx].Bytes())
	}
}

func TestPeersOnChannel_RefreshShouldBeDone(t *testing.T) {
	t.Parallel()

	retPeerIDs := []core.PeerID{"peer1", "peer2"}
	testTopic := "test_topic"
	wasFetchCalled := atomic.Value{}
	wasFetchCalled.Store(false)

	refreshInterval := time.Millisecond * 100
	ttlInterval := time.Duration(2)

	poc, _ := newPeersOnChannel(
		&p2pmocks.PeersRatingHandlerStub{},
		func(topic string) []peer.ID {
			wasFetchCalled.Store(true)
			return nil
		},
		refreshInterval,
		ttlInterval,
	)
	poc.getTimeHandler = func() time.Time {
		return time.Unix(0, 4)
	}
	// manually put peers
	poc.mutPeers.Lock()
	poc.peers[testTopic] = retPeerIDs
	poc.lastUpdated[testTopic] = time.Unix(0, 1)
	poc.mutPeers.Unlock()

	// To be 100% sure it triggers at least once, the time.sleep should be at least twice the refreshInterval.
	// The reason behind this is that, when instantiating newPeersOnChannel, the go routine starts and it begins
	// to iterate in the lastUpdated map. Maybe the instruction
	//  poc.lastUpdated[testTopic] = time.Unix(0, 1)
	// will be executed after the first for range iteration and thus, causing a time.sleep in the
	// peersOnChannel.refreshPeersOnAllKnownTopics loop
	time.Sleep(refreshInterval * 2)

	assert.True(t, wasFetchCalled.Load().(bool))
	poc.mutPeers.Lock()
	assert.Empty(t, poc.peers[testTopic])
	poc.mutPeers.Unlock()
}
