package connectionMonitor

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/stretchr/testify/assert"
)

func TestNilConnectionMonitor_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced")
		}
	}()

	ncm := &NilConnectionMonitor{}

	assert.False(t, check.IfNil(ncm))
	ncm.OpenedStream(nil, nil)
	ncm.ListenClose(nil, nil)
	ncm.Listen(nil, nil)
	ncm.Disconnected(nil, nil)
	ncm.ClosedStream(nil, nil)
	ncm.Connected(nil, nil)
}

func TestNilConnectionMonitor_SetThresholdMinConnectedPeers(t *testing.T) {
	t.Parallel()

	ncm := &NilConnectionMonitor{}
	threshold := 10
	ncm.SetThresholdMinConnectedPeers(threshold, nil)

	assert.Equal(t, threshold, ncm.ThresholdMinConnectedPeers())
}

func TestNilConnectionMonitor_IsConnectedToTheNetwork(t *testing.T) {
	t.Parallel()

	ncm := &NilConnectionMonitor{}
	threshold := 10
	ncm.SetThresholdMinConnectedPeers(threshold, nil)

	connections := 5
	netw := &mock.NetworkStub{
		ConnsCalled: func() []network.Conn {
			return make([]network.Conn, connections)
		},
	}

	assert.False(t, ncm.IsConnectedToTheNetwork(netw))
	ncm.SetThresholdMinConnectedPeers(connections, nil)
	assert.True(t, ncm.IsConnectedToTheNetwork(netw))
}
