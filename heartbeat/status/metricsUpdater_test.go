package status

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/heartbeat/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func createMockArgsMetricsUpdater() ArgsMetricsUpdater {
	return ArgsMetricsUpdater{
		PeerAuthenticationCacher:            testscommon.NewCacherMock(),
		HeartbeatMonitor:                    &mock.HeartbeatMonitorStub{},
		HeartbeatSenderInfoProvider:         &mock.HeartbeatSenderInfoProviderStub{},
		AppStatusHandler:                    &statusHandler.AppStatusHandlerStub{},
		TimeBetweenConnectionsMetricsUpdate: time.Second,
	}
}

func TestNewMetricsUpdater(t *testing.T) {
	t.Parallel()

	t.Run("nil peer authentication cacher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMetricsUpdater()
		args.PeerAuthenticationCacher = nil
		updater, err := NewMetricsUpdater(args)

		assert.Equal(t, heartbeat.ErrNilCacher, err)
		assert.True(t, check.IfNil(updater))
	})
	t.Run("nil heartbeat monitor should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMetricsUpdater()
		args.HeartbeatMonitor = nil
		updater, err := NewMetricsUpdater(args)

		assert.Equal(t, heartbeat.ErrNilHeartbeatMonitor, err)
		assert.True(t, check.IfNil(updater))
	})
	t.Run("nil heartbeat sender info provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMetricsUpdater()
		args.HeartbeatSenderInfoProvider = nil
		updater, err := NewMetricsUpdater(args)

		assert.Equal(t, heartbeat.ErrNilHeartbeatSenderInfoProvider, err)
		assert.True(t, check.IfNil(updater))
	})
	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMetricsUpdater()
		args.AppStatusHandler = nil
		updater, err := NewMetricsUpdater(args)

		assert.Equal(t, heartbeat.ErrNilAppStatusHandler, err)
		assert.True(t, check.IfNil(updater))
	})
	t.Run("invalid TimeBetweenConnectionsMetricsUpdate should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMetricsUpdater()
		args.TimeBetweenConnectionsMetricsUpdate = time.Second - time.Nanosecond
		updater, err := NewMetricsUpdater(args)

		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, check.IfNil(updater))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMetricsUpdater()
		updater, err := NewMetricsUpdater(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(updater))

		_ = updater.Close()
	})
}

func TestMetricsUpdater_Close(t *testing.T) {
	t.Parallel()

	setUin64Called := atomic.Flag{}
	args := createMockArgsMetricsUpdater()
	args.AppStatusHandler = &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			setUin64Called.SetValue(true)
		},
	}
	updater, _ := NewMetricsUpdater(args)

	time.Sleep(time.Second*3 + time.Millisecond*500)
	assert.True(t, setUin64Called.IsSet())

	err := updater.Close()
	assert.Nil(t, err)

	time.Sleep(time.Second)
	setUin64Called.SetValue(false)

	time.Sleep(time.Second*3 + time.Millisecond*500)
	assert.False(t, setUin64Called.IsSet())
}

func TestMetricsUpdater_updateMetrics(t *testing.T) {
	t.Parallel()

	args := createMockArgsMetricsUpdater()
	t.Run("should send connection metrics", func(t *testing.T) {
		_ = args.PeerAuthenticationCacher.Put([]byte("key1"), "key1", 0)
		_ = args.PeerAuthenticationCacher.Put([]byte("key2"), "key2", 0)
		_ = args.PeerAuthenticationCacher.Put([]byte("key3"), "key2", 0)

		args.HeartbeatMonitor = &mock.HeartbeatMonitorStub{
			GetHeartbeatsCalled: func() []data.PubKeyHeartbeat {
				return []data.PubKeyHeartbeat{
					{
						IsActive: false,
					},
					{
						IsActive: true,
						PeerType: string(common.EligibleList),
					},
					{
						IsActive: true,
						PeerType: string(common.WaitingList),
					},
					{
						IsActive: true,
						PeerType: string(common.JailedList),
					},
					{
						IsActive: true,
						PeerType: string(common.ObserverList),
					},
					{
						IsActive: true,
						PeerType: string(common.NewList),
					},
				}
			},
		}
		testUpdaterForConnectionMetrics(t, args)
	})
	t.Run("should send sender metrics", func(t *testing.T) {
		t.Run("eligible node", func(t *testing.T) {
			args.HeartbeatSenderInfoProvider = &mock.HeartbeatSenderInfoProviderStub{
				GetCurrentNodeTypeCalled: func() (string, core.P2PPeerSubType, error) {
					return string(common.EligibleList), core.FullHistoryObserver, nil
				},
			}
			testUpdaterForSenderMetrics(
				t,
				args,
				string(common.EligibleList),
				string(core.NodeTypeValidator),
				core.FullHistoryObserver)
		})
		t.Run("waiting node", func(t *testing.T) {
			args.HeartbeatSenderInfoProvider = &mock.HeartbeatSenderInfoProviderStub{
				GetCurrentNodeTypeCalled: func() (string, core.P2PPeerSubType, error) {
					return string(common.WaitingList), core.FullHistoryObserver, nil
				},
			}
			testUpdaterForSenderMetrics(
				t,
				args,
				string(common.WaitingList),
				string(core.NodeTypeValidator),
				core.FullHistoryObserver)
		})
		t.Run("observer node", func(t *testing.T) {
			args.HeartbeatSenderInfoProvider = &mock.HeartbeatSenderInfoProviderStub{
				GetCurrentNodeTypeCalled: func() (string, core.P2PPeerSubType, error) {
					return string(common.ObserverList), core.FullHistoryObserver, nil
				},
			}
			testUpdaterForSenderMetrics(
				t,
				args,
				string(common.ObserverList),
				string(core.NodeTypeObserver),
				core.FullHistoryObserver)
		})
	})
	t.Run("GetSenderInfo errors", func(t *testing.T) {
		args.HeartbeatSenderInfoProvider = &mock.HeartbeatSenderInfoProviderStub{
			GetCurrentNodeTypeCalled: func() (string, core.P2PPeerSubType, error) {
				return "", 0, errors.New("expected error")
			},
		}
		args.AppStatusHandler = &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				switch key {
				case common.MetricNodeType, common.MetricPeerType, common.MetricPeerSubType:
					assert.Fail(t, "should have not set status metrics")
				}
			},
		}

		updater, _ := NewMetricsUpdaterWithoutGoRoutineStart(args)

		updater.updateMetrics()
	})
}

func TestMetricsUpdater_MetricLiveValidatorNodesUpdatesDirectly(t *testing.T) {
	t.Parallel()

	args := createMockArgsMetricsUpdater()

	wasCalled := atomic.Flag{}
	args.AppStatusHandler = &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			switch key {
			case common.MetricLiveValidatorNodes:
				assert.Equal(t, uint64(1), value)
				wasCalled.SetValue(true)
			}
		},
	}
	updater, _ := NewMetricsUpdaterWithoutGoRoutineStart(args)
	time.Sleep(time.Second)
	updater.peerAuthenticationCacher.Put([]byte("key1"), "value1", 0)
	time.Sleep(time.Second)
	assert.True(t, wasCalled.IsSet())
}

func testUpdaterForConnectionMetrics(tb testing.TB, args ArgsMetricsUpdater) {
	args.AppStatusHandler = &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			switch key {
			case common.MetricNumIntraShardValidatorNodes:
				assert.Equal(tb, uint64(2), value)
			case common.MetricConnectedNodes:
				assert.Equal(tb, uint64(5), value)
			case common.MetricLiveValidatorNodes:
				assert.Equal(tb, uint64(3), value)
			}
		},
	}
	updater, _ := NewMetricsUpdaterWithoutGoRoutineStart(args)

	updater.updateMetrics()
}

func testUpdaterForSenderMetrics(
	tb testing.TB,
	args ArgsMetricsUpdater,
	peerType string,
	nodeType string,
	peerSubType core.P2PPeerSubType,
) {
	args.AppStatusHandler = &statusHandler.AppStatusHandlerStub{
		SetStringValueHandler: func(key string, value string) {
			switch key {
			case common.MetricNodeType:
				assert.Equal(tb, nodeType, value)
			case common.MetricPeerType:
				assert.Equal(tb, peerType, value)
			case common.MetricPeerSubType:
				assert.Equal(tb, peerSubType.String(), value)
			}
		},
	}
	updater, _ := NewMetricsUpdaterWithoutGoRoutineStart(args)

	updater.updateMetrics()
}
