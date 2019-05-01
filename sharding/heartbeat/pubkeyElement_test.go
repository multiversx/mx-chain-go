package heartbeat_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/sharding/heartbeat"
	"github.com/stretchr/testify/assert"
)

func TestPubkeyElement_HeartbeatArrivedFirstTimeForAddressShouldWork(t *testing.T) {
	t.Parallel()

	pe := heartbeat.NewPubkeyElement(time.Duration(10))
	pe.SetTimeGetter(func() time.Time {
		return time.Unix(0, 1)
	})
	p2pAddr := "p2p address"

	pe.HeartbeatArrived(p2pAddr)
	heartbeats := pe.GetPeerHeartbeats()

	expectedHeartBeat := heartbeat.PeerHeartbeat{
		P2PAddress:      p2pAddr,
		TimeStamp:       time.Unix(0, 1),
		MaxInactiveTime: 0,
		IsActive:        true,
	}

	assert.Equal(t, 1, len(heartbeats))
	assert.Equal(t, expectedHeartBeat, heartbeats[0])
}

func TestPubkeyElement_HeartbeatArrivedShouldUpdate(t *testing.T) {
	t.Parallel()

	pe := heartbeat.NewPubkeyElement(time.Duration(10))
	incrementalTime := int64(0)
	pe.SetTimeGetter(func() time.Time {
		if incrementalTime < 2 {
			incrementalTime++
		}
		return time.Unix(0, incrementalTime)
	})
	p2pAddr := "p2p address"

	pe.HeartbeatArrived(p2pAddr)
	pe.HeartbeatArrived(p2pAddr)
	heartbeats := pe.GetPeerHeartbeats()

	expectedHeartBeat := heartbeat.PeerHeartbeat{
		P2PAddress:      p2pAddr,
		TimeStamp:       time.Unix(0, 2),
		MaxInactiveTime: 1,
		IsActive:        true,
	}

	assert.Equal(t, 1, len(heartbeats))
	assert.Equal(t, expectedHeartBeat, heartbeats[0])
}

func TestPubkeyElement_HeartbeatSweepShouldUpdate(t *testing.T) {
	t.Parallel()

	pe := heartbeat.NewPubkeyElement(time.Duration(1))
	incrementalTime := int64(0)
	pe.SetTimeGetter(func() time.Time {
		tReturned := time.Unix(0, incrementalTime)
		incrementalTime += 10

		return tReturned
	})
	p2pAddr := "p2p address"

	pe.HeartbeatArrived(p2pAddr)
	heartbeats := pe.GetPeerHeartbeats()

	expectedHeartBeat := heartbeat.PeerHeartbeat{
		P2PAddress:      p2pAddr,
		TimeStamp:       time.Unix(0, 0),
		MaxInactiveTime: 10,
		IsActive:        false,
	}

	assert.Equal(t, 1, len(heartbeats))
	assert.Equal(t, expectedHeartBeat, heartbeats[0])
}
