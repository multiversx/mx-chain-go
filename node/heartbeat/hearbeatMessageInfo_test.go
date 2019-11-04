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
		mock.NewMockTimer(),
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
		mock.NewMockTimer(),
	)

	assert.NotNil(t, hbmi)
	assert.Nil(t, err)
}

//------- HeartbeatReceived

func TestHeartbeatMessageInfo_HeartbeatReceivedShouldUpdate(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := mockTimer.Now()

	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		10*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	mockTimer.IncrementSeconds(1)

	expectedTime := time.Unix(1, 0)
	hbmi.HeartbeatReceived(uint32(0), uint32(0), "v0.1", "undefined")
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
	assert.Equal(t, uint32(0), hbmi.GetReceiverShardId())

	mockTimer.IncrementSeconds(1)
	expectedTime = time.Unix(2, 0)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
	assert.Equal(t, uint32(1), hbmi.GetReceiverShardId())
}

func TestHeartbeatMessageInfo_HeartbeatUpdateFieldsShouldWork(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := mockTimer.Now()
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	mockTimer.IncrementSeconds(1)

	expectedTime := time.Unix(1, 0)
	expectedUptime := time.Duration(0)
	expectedDownTime := time.Duration(1 * time.Second)
	hbmi.HeartbeatReceived(uint32(0), uint32(3), "v0.1", "undefined")
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
	assert.Equal(t, true, hbmi.GetIsActive())
	assert.Equal(t, expectedUptime, hbmi.GetTotalUpTime().Duration)
	assert.Equal(t, expectedDownTime, hbmi.GetTotalDownTime().Duration)
}

func TestHeartbeatMessageInfo_HeartbeatShouldUpdateUpDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := mockTimer.Now()
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDownDuration := time.Duration(1 * time.Second)
	expectedUpDuration := time.Duration(1 * time.Second)
	assert.Equal(t, expectedUpDuration, hbmi.GetTotalUpTime().Duration)
	assert.Equal(t, expectedDownDuration, hbmi.GetTotalDownTime().Duration)
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
}

func TestHeartbeatMessageInfo_HeartbeatLongerDurationThanMaxShouldUpdateDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := mockTimer.Now()
	maxUnresponsiveTime := 500 * time.Millisecond
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		maxUnresponsiveTime,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDownDuration := time.Duration(1500 * time.Millisecond)
	expectedUpDuration := time.Duration(maxUnresponsiveTime)
	assert.Equal(t, expectedDownDuration, hbmi.GetTotalDownTime().Duration)
	assert.Equal(t, expectedUpDuration, hbmi.GetTotalUpTime().Duration)
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
}

func TestHeartbeatMessageInfo_HeartbeatBeforeGenesisShouldNotUpdateUpDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := time.Unix(5, 0)
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", "undefined")

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, hbmi.GetTotalDownTime().Duration)
	assert.Equal(t, expectedDuration, hbmi.GetTotalUpTime().Duration)
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
}

func TestHeartbeatMessageInfo_HeartbeatEqualGenesisShouldHaveUpDownTimeZero(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := time.Unix(1, 0)
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		false,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", "undefined")

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, hbmi.GetTotalUpTime().Duration)
	assert.Equal(t, expectedDuration, hbmi.GetTotalDownTime().Duration)
	expectedTime := time.Unix(1, 0)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
}
