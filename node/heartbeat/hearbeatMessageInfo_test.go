package heartbeat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//------ newHeartbeatMessageInfo

func TestNewHeartbeatMessageInfo_InvalidDurationShouldErr(t *testing.T) {
	t.Parallel()

	hbmi, err := newHeartbeatMessageInfo(0)

	assert.Nil(t, hbmi)
	assert.Equal(t, ErrInvalidMaxDurationPeerUnresponsive, err)
}

func TestNewHeartbeatMessageInfo_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	hbmi, err := newHeartbeatMessageInfo(1)

	assert.NotNil(t, hbmi)
	assert.Nil(t, err)
}

func TestHeartbeatMessageInfo_HeartbeatReceivedFirstTimeForAddressShouldWork(t *testing.T) {
	t.Parallel()

	hbmi, _ := newHeartbeatMessageInfo(time.Duration(10))
	hbmi.getTimeHandler = func() time.Time {
		return time.Unix(0, 1)
	}
	p2pAddr := "p2p address"

	hbmi.HeartbeatReceived(p2pAddr)
	heartbeats := hbmi.GetPeerHeartbeats()

	expectedHeartBeat := PeerHeartbeat{
		P2PAddress:      p2pAddr,
		TimeStamp:       time.Unix(0, 1),
		MaxInactiveTime: Duration{Duration: 0},
		IsActive:        true,
	}

	assert.Equal(t, 1, len(heartbeats))
	assert.Equal(t, expectedHeartBeat, heartbeats[0])
}

func TestHeartbeatMessageInfo_HeartbeatReceivedShouldUpdate(t *testing.T) {
	t.Parallel()

	hbmi, _ := newHeartbeatMessageInfo(time.Duration(10))
	incrementalTime := int64(0)
	hbmi.getTimeHandler = func() time.Time {
		if incrementalTime < 2 {
			incrementalTime++
		}
		return time.Unix(0, incrementalTime)
	}
	p2pAddr := "p2p address"

	hbmi.HeartbeatReceived(p2pAddr)
	hbmi.HeartbeatReceived(p2pAddr)
	heartbeats := hbmi.GetPeerHeartbeats()

	expectedHeartBeat := PeerHeartbeat{
		P2PAddress:      p2pAddr,
		TimeStamp:       time.Unix(0, 2),
		MaxInactiveTime: Duration{Duration: 1},
		IsActive:        true,
	}

	assert.Equal(t, 1, len(heartbeats))
	assert.Equal(t, expectedHeartBeat, heartbeats[0])
}

func TestHeartbeatMessageInfo_HeartbeatSweepShouldUpdate(t *testing.T) {
	t.Parallel()

	hbmi, _ := newHeartbeatMessageInfo(time.Duration(1))
	incrementalTime := int64(0)
	hbmi.getTimeHandler = func() time.Time {
		tReturned := time.Unix(0, incrementalTime)
		incrementalTime += 10

		return tReturned
	}
	p2pAddr := "p2p address"

	hbmi.HeartbeatReceived(p2pAddr)
	heartbeats := hbmi.GetPeerHeartbeats()

	expectedHeartBeat := PeerHeartbeat{
		P2PAddress:      p2pAddr,
		TimeStamp:       time.Unix(0, 0),
		MaxInactiveTime: Duration{Duration: 10},
		IsActive:        false,
	}

	assert.Equal(t, 1, len(heartbeats))
	assert.Equal(t, expectedHeartBeat, heartbeats[0])
}
