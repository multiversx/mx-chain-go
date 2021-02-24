package componentHandler

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgument() ArgHeartbeat {
	arg := ArgHeartbeat{
		HeartbeatConfig: config.HeartbeatConfig{
			MinTimeToWaitBetweenBroadcastsInSec: 2,
			MaxTimeToWaitBetweenBroadcastsInSec: 3,
			DurationToConsiderUnresponsiveInSec: 10,
			HeartbeatRefreshIntervalInSec:       1,
			HideInactiveValidatorIntervalInSec:  20,
		},
		PrefsConfig: config.PreferencesConfig{
			DestinationShardAsObserver: "0",
			NodeDisplayName:            "node name",
			Identity:                   "identity",
		},
		Marshalizer:              &mock.MarshalizerMock{},
		Messenger:                &mock.MessengerStub{},
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		NodesCoordinator:         &mock.NodesCoordinatorMock{},
		AppStatusHandler:         &mock.AppStatusHandlerStub{},
		Storer:                   mock.NewStorerMock(),
		ValidatorStatistics:      &mock.ValidatorStatisticsStub{},
		PeerSignatureHandler:     &mock.PeerSignatureHandler{},
		PrivKey:                  &mock.PrivateKeyStub{},
		HardforkTrigger:          &mock.HardforkTriggerStub{},
		AntifloodHandler:         &mock.P2PAntifloodHandlerStub{},
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		EpochStartTrigger:        &mock.EpochStartTriggerStub{},
		EpochStartRegistration:   &mock.EpochStartNotifierStub{},
		Timer:                    mock.NewTimerMock(),
		GenesisTime:              time.Time{},
		VersionNumber:            "v0.0.0",
		PeerShardMapper:          &mock.NetworkShardingCollectorStub{},
		SizeCheckDelta:           0,
		ValidatorsProvider:       &mock.ValidatorsProviderStub{},
		CurrentBlockProvider:     &mock.CurrentBlockProviderStub{},
		RedundancyHandler:        &mock.RedundancyHandlerStub{},
	}

	return arg
}

//------- NewHeartbeatHandler

func TestNewHeartbeatHandler_DurationToConsiderUnresponsiveInSec(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.HeartbeatConfig.DurationToConsiderUnresponsiveInSec = 0
	hbh, err := NewHeartbeatHandler(arg)

	assert.True(t, check.IfNil(hbh))
	assert.Equal(t, heartbeat.ErrInvalidDurationToConsiderUnresponsiveInSec, err)
}

func TestNewHeartbeatHandler_MaxTimeToWaitBetweenBroadcastsInSec(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.HeartbeatConfig.MaxTimeToWaitBetweenBroadcastsInSec = 0
	hbh, err := NewHeartbeatHandler(arg)

	assert.True(t, check.IfNil(hbh))
	assert.Equal(t, heartbeat.ErrNegativeMaxTimeToWaitBetweenBroadcastsInSec, err)
}

func TestNewHeartbeatHandler_MinTimeToWaitBetweenBroadcastsInSec(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.HeartbeatConfig.MinTimeToWaitBetweenBroadcastsInSec = 0
	hbh, err := NewHeartbeatHandler(arg)

	assert.True(t, check.IfNil(hbh))
	assert.Equal(t, heartbeat.ErrNegativeMinTimeToWaitBetweenBroadcastsInSec, err)
}

func TestNewHeartbeatHandler_InvalidMaxTimeToWaitBetweenBroadcastsInSec(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.HeartbeatConfig.MaxTimeToWaitBetweenBroadcastsInSec = 2
	arg.HeartbeatConfig.MinTimeToWaitBetweenBroadcastsInSec = 3
	hbh, err := NewHeartbeatHandler(arg)

	assert.True(t, check.IfNil(hbh))
	assert.True(t, errors.Is(err, heartbeat.ErrWrongValues))
}

func TestNewHeartbeatHandler_InvalidDurationToConsiderUnresponsiveInSec(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.HeartbeatConfig.DurationToConsiderUnresponsiveInSec = 2
	arg.HeartbeatConfig.MaxTimeToWaitBetweenBroadcastsInSec = 3
	hbh, err := NewHeartbeatHandler(arg)

	assert.True(t, check.IfNil(hbh))
	assert.True(t, errors.Is(err, heartbeat.ErrWrongValues))
}

func TestNewHeartbeatHandler_NilMessenger(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Messenger = nil
	hbh, err := NewHeartbeatHandler(arg)

	assert.True(t, check.IfNil(hbh))
	assert.Equal(t, heartbeat.ErrNilMessenger, err)
}

func TestNewHeartbeatHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	hbh, err := NewHeartbeatHandler(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(hbh))
	require.NotNil(t, hbh.Monitor())
	require.NotNil(t, hbh.Sender())

	// wait for the sending go routine start
	time.Sleep(time.Second)

	err = hbh.Close()
	assert.Nil(t, err)

	// let the sending go routine finish
	time.Sleep(time.Second)
}

//TODO(next PR) add more tests
