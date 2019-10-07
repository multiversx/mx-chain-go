package heartbeat_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

//------- newHeartbeatMessageInfo

func TestNewHeartbeatMessageInfo_InvalidDurationShouldErr(t *testing.T) {
	t.Parallel()

	hbmi, err := heartbeat.NewHeartbeatMessageInfo(
		0,
		false,
		time.Time{},
		&mock.MockTimer{},
	)

	assert.Nil(t, hbmi)
	assert.Equal(t, heartbeat.ErrInvalidMaxDurationPeerUnresponsive, err)
}

func TestNewHeartbeatMessageInfo_NilGetTimeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	hbmi, err := heartbeat.NewHeartbeatMessageInfo(
		1,
		false,
		time.Time{},
		nil,
	)

	assert.Nil(t, hbmi)
	assert.Equal(t, heartbeat.ErrNilTimer, err)
}

func TestNewHeartbeatMessageInfo_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	hbmi, err := heartbeat.NewHeartbeatMessageInfo(
		1,
		false,
		time.Time{},
		&mock.MockTimer{},
	)

	assert.NotNil(t, hbmi)
	assert.Nil(t, err)
}

//------- HeartbeatReceived

func TestHeartbeatMessageInfo_HeartbeatReceivedShouldUpdate(t *testing.T) {
	t.Parallel()

	mockTimer := &mock.MockTimer{}
	genesisTime := mockTimer.Now()

	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		10*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, heartbeat.GetTimeStamp(hbmi))

	mockTimer.IncrementSeconds(1)

	expectedTime := time.Unix(1, 0)
	hbmi.HeartbeatReceived(uint32(0), uint32(0), "v0.1", "undefined")
	assert.Equal(t, expectedTime, heartbeat.GetTimeStamp(hbmi))
	assert.Equal(t, uint32(0), heartbeat.GetReceiverShardId(hbmi))

	mockTimer.IncrementSeconds(1)
	expectedTime = time.Unix(2, 0)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	assert.Equal(t, expectedTime, heartbeat.GetTimeStamp(hbmi))
	assert.Equal(t, uint32(1), heartbeat.GetReceiverShardId(hbmi))
}

func TestHeartbeatMessageInfo_HeartbeatUpdateFieldsShouldWork(t *testing.T) {
	t.Parallel()

	mockTimer := &mock.MockTimer{}
	genesisTime := mockTimer.Now()
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, heartbeat.GetTimeStamp(hbmi))

	mockTimer.IncrementSeconds(1)

	expectedTime := time.Unix(1, 0)
	expectedUptime := time.Duration(0)
	expectedDownTime := time.Duration(1 * time.Second)
	hbmi.HeartbeatReceived(uint32(0), uint32(3), "v0.1", "undefined")
	assert.Equal(t, expectedTime, heartbeat.GetTimeStamp(hbmi))
	assert.Equal(t, true, heartbeat.GetIsActive(hbmi))
	assert.Equal(t, expectedUptime, heartbeat.GetUpTimeDuration(hbmi))
	assert.Equal(t, expectedDownTime, heartbeat.GetDownTimeDuration(hbmi))
}

func TestHeartbeatMessageInfo_HeartbeatShouldUpdateUpDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := &mock.MockTimer{}
	genesisTime := mockTimer.Now()
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, heartbeat.GetTimeStamp(hbmi))

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDownDuration := time.Duration(1 * time.Second)
	expectedUpDuration := time.Duration(1 * time.Second)
	assert.Equal(t, expectedUpDuration, heartbeat.GetUpTimeDuration(hbmi))
	assert.Equal(t, expectedDownDuration, heartbeat.GetDownTimeDuration(hbmi))
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, heartbeat.GetTimeStamp(hbmi))
}

func TestHeartbeatMessageInfo_HeartbeatLongerDurationThanMaxShouldUpdateDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := &mock.MockTimer{}
	genesisTime := mockTimer.Now()
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		500*time.Millisecond,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, heartbeat.GetTimeStamp(hbmi))

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDownDuration := time.Duration(2 * time.Second)
	expectedUpDuration := time.Duration(0)
	assert.Equal(t, expectedDownDuration, heartbeat.GetDownTimeDuration(hbmi))
	assert.Equal(t, expectedUpDuration, heartbeat.GetUpTimeDuration(hbmi))
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, heartbeat.GetTimeStamp(hbmi))
}

func TestHeartbeatMessageInfo_HeartbeatBeforeGenesisShouldNotUpdateUpDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := &mock.MockTimer{}
	genesisTime := time.Unix(5, 0)
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, heartbeat.GetTimeStamp(hbmi))

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, heartbeat.GetDownTimeDuration(hbmi))
	assert.Equal(t, expectedDuration, heartbeat.GetUpTimeDuration(hbmi))
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, heartbeat.GetTimeStamp(hbmi))
}

func TestHeartbeatMessageInfo_HeartbeatEqualGenesisShouldHaveUpDownTimeZero(t *testing.T) {
	t.Parallel()

	mockTimer := &mock.MockTimer{}
	genesisTime := time.Unix(1, 0)
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, heartbeat.GetTimeStamp(hbmi))
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, heartbeat.GetUpTimeDuration(hbmi))
	assert.Equal(t, expectedDuration, heartbeat.GetDownTimeDuration(hbmi))
	expectedTime := time.Unix(1, 0)
	assert.Equal(t, expectedTime, heartbeat.GetTimeStamp(hbmi))
}
