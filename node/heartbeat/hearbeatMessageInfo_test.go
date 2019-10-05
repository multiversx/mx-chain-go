package heartbeat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//------ newHeartbeatMessageInfo
func TestNewHeartbeatMessageInfo_InvalidDurationShouldErr(t *testing.T) {
	t.Parallel()

	hbmi, err := newHeartbeatMessageInfo(0, false)

	assert.Nil(t, hbmi)
	assert.Equal(t, ErrInvalidMaxDurationPeerUnresponsive, err)
}

func TestNewHeartbeatMessageInfo_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	hbmi, err := newHeartbeatMessageInfo(1, false)

	assert.NotNil(t, hbmi)
	assert.Nil(t, err)
}

func TestHeartbeatMessageInfo_HeartbeatReceivedShouldUpdate(t *testing.T) {
	t.Parallel()

	hbmi, _ := newHeartbeatMessageInfo(time.Duration(10), false)
	incrementalTime := int64(0)
	hbmi.getTimeHandler = func() time.Time {
		if incrementalTime < 2 {
			incrementalTime++
		}
		return time.Unix(0, incrementalTime)
	}

	assert.Equal(t, emptyTimestamp, hbmi.timeStamp)

	hbmi.HeartbeatReceived(uint32(0), uint32(0), "v0.1", "undefined")
	assert.NotEqual(t, emptyTimestamp, hbmi.timeStamp)
	assert.Equal(t, uint32(0), hbmi.receivedShardID)

	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	assert.NotEqual(t, emptyTimestamp, hbmi.timeStamp)
	assert.Equal(t, uint32(1), hbmi.receivedShardID)
}

func TestHeartbeatMessageInfo_HeartbeatUpdateFieldsShouldWork(t *testing.T) {
	t.Parallel()

	hbmi, _ := newHeartbeatMessageInfo(time.Duration(1), false)
	incrementalTime := int64(0)
	hbmi.getTimeHandler = func() time.Time {
		tReturned := time.Unix(0, incrementalTime)
		incrementalTime += 10

		return tReturned
	}

	assert.Equal(t, emptyTimestamp, hbmi.timeStamp)

	hbmi.HeartbeatReceived(uint32(0), uint32(3), "v0.1", "undefined")
	assert.NotEqual(t, emptyTimestamp, hbmi.timeStamp)
}

func TestHeartbeatMessageInfo_HeartbeatShouldUpdateUpTime(t *testing.T) {
	t.Parallel()

	hbmi, _ := newHeartbeatMessageInfo(time.Duration(10), false)
	incrementalTime := int64(0)
	hbmi.getTimeHandler = func() time.Time {
		tReturned := time.Unix(0, incrementalTime)
		incrementalTime += 1

		return tReturned
	}

	assert.Equal(t, emptyTimestamp, hbmi.timeStamp)

	// send heartbeat twice in order to calculate the duration between thm
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	time.Sleep(10 * time.Millisecond)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	assert.True(t, hbmi.totalUpTime.Duration > time.Duration(0))
	assert.NotEqual(t, emptyTimestamp, hbmi.timeStamp)
}
