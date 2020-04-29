package heartbeat_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

const dummyPeerType = "dummy peer type"
const dummyIdentity = "dummy identity"
const dummyNodeDisplayName = "dummy node display name"

//------- newHeartbeatMessageInfo

func TestNewHeartbeatMessageInfo_InvalidDurationShouldErr(t *testing.T) {
	t.Parallel()

	hbmi, err := heartbeat.NewHeartbeatMessageInfo(
		0,
		dummyPeerType,
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
		dummyPeerType,
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
		dummyPeerType,
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
		dummyPeerType,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	mockTimer.IncrementSeconds(1)

	expectedTime := time.Unix(1, 0)
	hbmi.HeartbeatReceived(uint32(0), uint32(0), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
	assert.Equal(t, uint32(0), hbmi.GetReceiverShardId())

	mockTimer.IncrementSeconds(1)
	expectedTime = time.Unix(2, 0)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
	assert.Equal(t, uint32(1), hbmi.GetReceiverShardId())
}

func TestHeartbeatMessageInfo_HeartbeatUpdateFieldsShouldWork(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := mockTimer.Now()
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		dummyPeerType,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	mockTimer.IncrementSeconds(1)

	expectedTime := time.Unix(1, 0)
	expectedUptime := time.Duration(0)
	expectedDownTime := 1 * time.Second
	hbmi.HeartbeatReceived(uint32(0), uint32(3), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
	assert.Equal(t, true, hbmi.GetIsActive())
	assert.Equal(t, expectedUptime, hbmi.GetTotalUpTime())
	assert.Equal(t, expectedDownTime, hbmi.GetTotalDownTime())
}

func TestHeartbeatMessageInfo_HeartbeatShouldUpdateUpDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := mockTimer.Now()
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		dummyPeerType,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)

	expectedDownDuration := 1 * time.Second
	expectedUpDuration := 1 * time.Second
	assert.Equal(t, expectedUpDuration, hbmi.GetTotalUpTime())
	assert.Equal(t, expectedDownDuration, hbmi.GetTotalDownTime())
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
		"eligible",
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)

	expectedDownDuration := 1500 * time.Millisecond
	expectedUpDuration := maxUnresponsiveTime
	assert.Equal(t, expectedDownDuration, hbmi.GetTotalDownTime())
	assert.Equal(t, expectedUpDuration, hbmi.GetTotalUpTime())
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
}

func TestHeartbeatMessageInfo_HeartbeatBeforeGenesisShouldNotUpdateUpDownTime(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := time.Unix(5, 0)
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		dummyPeerType,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())

	// send heartbeat twice in order to calculate the duration between thm
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(2), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, hbmi.GetTotalDownTime())
	assert.Equal(t, expectedDuration, hbmi.GetTotalUpTime())
	expectedTime := time.Unix(2, 0)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
}

func TestHeartbeatMessageInfo_HeartbeatEqualGenesisShouldHaveUpDownTimeZero(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := time.Unix(1, 0)
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		dummyPeerType,
		genesisTime,
		mockTimer,
	)

	assert.Equal(t, genesisTime, hbmi.GetTimeStamp())
	mockTimer.IncrementSeconds(1)
	hbmi.HeartbeatReceived(uint32(0), uint32(1), "v0.1", dummyNodeDisplayName, dummyIdentity, dummyPeerType)

	expectedDuration := time.Duration(0)
	assert.Equal(t, expectedDuration, hbmi.GetTotalUpTime())
	assert.Equal(t, expectedDuration, hbmi.GetTotalDownTime())
	expectedTime := time.Unix(1, 0)
	assert.Equal(t, expectedTime, hbmi.GetTimeStamp())
}

func TestHeartbeatMessageInfo_GetIsValidator_NotValidatorShouldReturnFalse(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := time.Unix(1, 0)
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		dummyPeerType,
		genesisTime,
		mockTimer,
	)

	assert.False(t, hbmi.GetIsValidator())
}

func TestHeartbeatMessageInfo_GetIsValidator_PeerTypeEligibleShouldReturnTrue(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := time.Unix(1, 0)
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		string(core.EligibleList),
		genesisTime,
		mockTimer,
	)

	assert.True(t, hbmi.GetIsValidator())
}

func TestHeartbeatMessageInfo_GetIsValidator_PeerTypeWaitingShouldReturnTrue(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := time.Unix(1, 0)
	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		100*time.Second,
		string(core.WaitingList),
		genesisTime,
		mockTimer,
	)

	assert.True(t, hbmi.GetIsValidator())
}

//------- UpdateShardAndPeerType

func TestHeartbeatMessageInfo_Update(t *testing.T) {
	t.Parallel()

	mockTimer := mock.NewMockTimer()
	genesisTime := mockTimer.Now()

	hbmi, _ := heartbeat.NewHeartbeatMessageInfo(
		10*time.Second,
		dummyPeerType,
		genesisTime,
		mockTimer,
	)

	computedShardId := uint32(0)
	peerType := dummyPeerType

	hbmi.HeartbeatReceived(computedShardId, uint32(0), "v0.1", dummyNodeDisplayName, dummyIdentity, peerType)
	assert.Equal(t, computedShardId, hbmi.GetComputedShardId())
	assert.Equal(t, peerType, hbmi.GetPeerType())

	computedShardId = uint32(2)
	peerType = "new peer type"
	hbmi.UpdateShardAndPeerType(computedShardId, peerType)
	assert.Equal(t, computedShardId, hbmi.GetComputedShardId())
	assert.Equal(t, peerType, hbmi.GetPeerType())
}
