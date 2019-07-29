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

	assert.Equal(t, emptyTimestamp, hbmi.timeStamp)

	hbmi.HeartbeatReceived(uint32(0))
	assert.NotEqual(t, emptyTimestamp, hbmi.timeStamp)
	assert.Equal(t, uint32(0), hbmi.shardID)

	hbmi.HeartbeatReceived(uint32(1))
	assert.NotEqual(t, emptyTimestamp, hbmi.timeStamp)
	assert.Equal(t, uint32(1), hbmi.shardID)
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

	assert.Equal(t, emptyTimestamp, hbmi.timeStamp)

	hbmi.HeartbeatReceived(uint32(3))
	assert.NotEqual(t, emptyTimestamp, hbmi.timeStamp)
}
