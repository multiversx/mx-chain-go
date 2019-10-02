package heartbeat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//------- newHeartbeatMessageInfo

func TestNewHeartbeatMessageInfo_InvalidDurationShouldErr(t *testing.T) {
	t.Parallel()

	hbmi, err := newHeartbeatMessageInfo(
		0,
		false,
		time.Time{},
		func() time.Time {
			return time.Time{}
		},
	)

	assert.Nil(t, hbmi)
	assert.Equal(t, ErrInvalidMaxDurationPeerUnresponsive, err)
}

func TestNewHeartbeatMessageInfo_NilGetTimeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	hbmi, err := newHeartbeatMessageInfo(
		1,
		false,
		time.Time{},
		nil,
	)

	assert.Nil(t, hbmi)
	assert.Equal(t, ErrNilGetTimeHandler, err)
}

func TestNewHeartbeatMessageInfo_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	hbmi, err := newHeartbeatMessageInfo(
		1,
		false,
		time.Time{},
		func() time.Time {
			return time.Time{}
		},
	)

	assert.NotNil(t, hbmi)
	assert.Nil(t, err)
}

//------- HeartbeatReceived

func TestHeartbeatMessageInfo_HeartbeatReceivedShouldUpdate(t *testing.T) {
	t.Parallel()

	genesisTime := time.Unix(0, 1)
	incrementalTime := int64(0)
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(10),
		false,
		genesisTime,
		func() time.Time {
			incrementalTime += 10

			return time.Unix(0, incrementalTime)
		},
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	expectedTime := time.Unix(0, 20)
	hbmi.HeartbeatReceived(uint32(0), uint32(0), "v0.1", "undefined")
	assert.Equal(t, expectedTime, hbmi.timeStamp)
	assert.Equal(t, uint32(0), hbmi.receivedShardID)

	expectedTime = time.Unix(0, 30)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	assert.Equal(t, expectedTime, hbmi.timeStamp)
	assert.Equal(t, uint32(1), hbmi.receivedShardID)
}

func TestHeartbeatMessageInfo_HeartbeatUpdateFieldsShouldWork(t *testing.T) {
	t.Parallel()

	incrementalTime := int64(0)
	getTimeHandler := func() time.Time {
		incrementalTime += 10

		return time.Unix(0, incrementalTime)
	}
	genesisTime := time.Unix(0, 1)
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(1),
		false,
		genesisTime,
		getTimeHandler,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	expectedTime := time.Unix(0, 20)
	hbmi.HeartbeatReceived(uint32(0), uint32(3), "v0.1", "undefined")
	assert.Equal(t, expectedTime, hbmi.timeStamp)
}

func TestHeartbeatMessageInfo_HeartbeatShouldUpdateUpTime(t *testing.T) {
	t.Parallel()

	incrementalTime := int64(0)
	getTimeHandler := func() time.Time {
		incrementalTime += 10

		return time.Unix(0, incrementalTime)
	}
	genesisTime := time.Unix(0, 1)
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(100),
		false,
		genesisTime,
		getTimeHandler,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	// send heartbeat twice in order to calculate the duration between thm
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDownDuration := time.Duration(0)
	expectedUpDuration := time.Duration(10)
	assert.Equal(t, expectedUpDuration, hbmi.totalUpTime.Duration)
	assert.Equal(t, expectedDownDuration, hbmi.totalDownTime.Duration)
	expectedTime := time.Unix(0, 30)
	assert.Equal(t, expectedTime, hbmi.timeStamp)
}

func TestHeartbeatMessageInfo_HeartbeatShouldUpdateDownTime(t *testing.T) {
	t.Parallel()

	incrementalTime := int64(0)
	getTimeHandler := func() time.Time {
		incrementalTime += 10

		return time.Unix(0, incrementalTime)
	}
	genesisTime := time.Unix(0, 1)
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(1),
		false,
		genesisTime,
		getTimeHandler,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	// send heartbeat twice in order to calculate the duration between thm
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDownDuration := time.Duration(10)
	expectedUpDuration := time.Duration(0)
	assert.Equal(t, expectedDownDuration, hbmi.totalDownTime.Duration)
	assert.Equal(t, expectedUpDuration, hbmi.totalUpTime.Duration)
	expectedTime := time.Unix(0, 30)
	assert.Equal(t, expectedTime, hbmi.timeStamp)
}

func TestHeartbeatMessageInfo_HeartbeatBeforeGenesisShouldNotUpdateUpDownTime(t *testing.T) {
	t.Parallel()

	incrementalTime := int64(0)
	getTimeHandler := func() time.Time {
		incrementalTime += 10

		return time.Unix(0, incrementalTime)
	}
	genesisTime := time.Unix(0, 40)
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(100),
		false,
		genesisTime,
		getTimeHandler,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	// send heartbeat twice in order to calculate the duration between thm
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, hbmi.totalDownTime.Duration)
	assert.Equal(t, expectedDuration, hbmi.totalUpTime.Duration)
	expectedTime := time.Unix(0, 30)
	assert.Equal(t, expectedTime, hbmi.timeStamp)
}

func TestHeartbeatMessageInfo_HeartbeatEqualGenesisShouldHaveUpDownTimeZero(t *testing.T) {
	t.Parallel()

	incrementalTime := int64(0)
	getTimeHandler := func() time.Time {
		incrementalTime += 10

		return time.Unix(0, incrementalTime)
	}
	genesisTime := time.Unix(0, 30)
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(100),
		false,
		genesisTime,
		getTimeHandler,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	// send heartbeat twice in order to calculate the duration between thm
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	time.Sleep(10 * time.Millisecond)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, hbmi.totalUpTime.Duration)
	assert.Equal(t, expectedDuration, hbmi.totalDownTime.Duration)
	expectedTime := time.Unix(0, 30)
	assert.Equal(t, expectedTime, hbmi.timeStamp)
}
