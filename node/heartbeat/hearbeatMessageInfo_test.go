package heartbeat

import (
	mock2 "github.com/ElrondNetwork/elrond-go/node/heartbeat/mock"
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

	mockTimer := mock2.MockTimer{}
	genesisTime := mockTimer.Now()

	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(10*time.Second),
		false,
		genesisTime,
		mockTimer.Now,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	mockTimer.IncrementSeconds(1)

	expectedTime := time.Unix(1, 0)
	hbmi.HeartbeatReceived(uint32(0), uint32(0), "v0.1", "undefined")
	assert.Equal(t, expectedTime, hbmi.timeStamp)
	assert.Equal(t, uint32(0), hbmi.receivedShardID)

	mockTimer.IncrementSeconds(1)
	expectedTime = time.Unix(2, 0)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	assert.Equal(t, expectedTime, hbmi.timeStamp)
	assert.Equal(t, uint32(1), hbmi.receivedShardID)
}

func TestHeartbeatMessageInfo_HeartbeatUpdateFieldsShouldWork(t *testing.T) {
	t.Parallel()

	mockTimer := mock2.MockTimer{}
	genesisTime := mockTimer.Now()
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(100*time.Second),
		false,
		genesisTime,
		mockTimer.Now,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	mockTimer.IncrementSeconds(1)

	expectedTime := time.Unix(1, 0)
	expectedUptime := time.Duration(0)
	expectedDownTime := time.Duration(1 * time.Second)
	hbmi.HeartbeatReceived(uint32(0), uint32(3), "v0.1", "undefined")
	assert.Equal(t, expectedTime, hbmi.timeStamp)
	assert.Equal(t, true, hbmi.isActive)
	assert.Equal(t, expectedUptime, hbmi.totalUpTime.Duration)
	assert.Equal(t, expectedDownTime, hbmi.totalDownTime.Duration)
}

func TestHeartbeatMessageInfo_HeartbeatShouldUpdateUpDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := mock2.MockTimer{}
	genesisTime := mockTimer.Now()
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(100*time.Second),
		false,
		genesisTime,
		mockTimer.Now,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDownDuration := time.Duration(1 * time.Second)
	expectedUpDuration := time.Duration(1 * time.Second)
	assert.Equal(t, expectedUpDuration, hbmi.totalUpTime.Duration)
	assert.Equal(t, expectedDownDuration, hbmi.totalDownTime.Duration)
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, hbmi.timeStamp)
}

func TestHeartbeatMessageInfo_HeartbeatLongerDurationThanMaxShouldUpdateDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := mock2.MockTimer{}
	genesisTime := mockTimer.Now()
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(500*time.Millisecond),
		false,
		genesisTime,
		mockTimer.Now,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDownDuration := time.Duration(2 * time.Second)
	expectedUpDuration := time.Duration(0)
	assert.Equal(t, expectedDownDuration, hbmi.totalDownTime.Duration)
	assert.Equal(t, expectedUpDuration, hbmi.totalUpTime.Duration)
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, hbmi.timeStamp)
}

func TestHeartbeatMessageInfo_HeartbeatBeforeGenesisShouldNotUpdateUpDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := mock2.MockTimer{}
	genesisTime := time.Unix(5, 0)
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(100*time.Second),
		false,
		genesisTime,
		mockTimer.Now,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, hbmi.totalDownTime.Duration)
	assert.Equal(t, expectedDuration, hbmi.totalUpTime.Duration)
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, hbmi.timeStamp)
}

func TestHeartbeatMessageInfo_HeartbeatEqualGenesisShouldHaveUpDownTimeZero(t *testing.T) {
	t.Parallel()

	mockTimer := mock2.MockTimer{}
	genesisTime := time.Unix(1, 0)
	hbmi, _ := newHeartbeatMessageInfo(
		time.Duration(100*time.Second),
		false,
		genesisTime,
		mockTimer.Now,
	)

	assert.Equal(t, genesisTime, hbmi.timeStamp)
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, hbmi.totalUpTime.Duration)
	assert.Equal(t, expectedDuration, hbmi.totalDownTime.Duration)
	expectedTime := time.Unix(1, 0)
	assert.Equal(t, expectedTime, hbmi.timeStamp)
}
