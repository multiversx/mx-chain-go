package monitor

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	processMocks "github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func createMockHeartbeatV2MonitorArgs() ArgHeartbeatV2Monitor {
	return ArgHeartbeatV2Monitor{
		Cache:                               testscommon.NewCacherMock(),
		PubKeyConverter:                     &testscommon.PubkeyConverterMock{},
		Marshaller:                          &testscommon.MarshalizerMock{},
		PeerShardMapper:                     &p2pmocks.NetworkShardingCollectorStub{},
		MaxDurationPeerUnresponsive:         time.Second * 3,
		HideInactiveValidatorInterval:       time.Second * 5,
		ShardId:                             0,
		PeerTypeProvider:                    &mock.PeerTypeProviderStub{},
		AppStatusHandler:                    &statusHandler.AppStatusHandlerStub{},
		TimeBetweenConnectionsMetricsUpdate: time.Second * 5,
	}
}

func createHeartbeatMessage(active bool) *heartbeat.HeartbeatV2 {
	crtTime := time.Now()
	providedAgeInSec := int64(1)
	messageTimestamp := crtTime.Unix() - providedAgeInSec

	if !active {
		messageTimestamp = crtTime.Unix() - int64(60)
	}

	payload := heartbeat.Payload{
		Timestamp: messageTimestamp,
	}

	marshaller := testscommon.MarshalizerMock{}
	payloadBytes, _ := marshaller.Marshal(payload)
	return &heartbeat.HeartbeatV2{
		Payload:         payloadBytes,
		VersionNumber:   "v01",
		NodeDisplayName: "node name",
		Identity:        "identity",
		Nonce:           0,
		PeerSubType:     0,
		Pubkey:          []byte("public key"),
	}
}

func TestNewHeartbeatV2Monitor(t *testing.T) {
	t.Parallel()

	t.Run("nil cache should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.Cache = nil
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.True(t, check.IfNil(monitor))
		assert.Equal(t, heartbeat.ErrNilCacher, err)
	})
	t.Run("nil pub key converter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.PubKeyConverter = nil
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.True(t, check.IfNil(monitor))
		assert.Equal(t, heartbeat.ErrNilPubkeyConverter, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.Marshaller = nil
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.True(t, check.IfNil(monitor))
		assert.Equal(t, heartbeat.ErrNilMarshaller, err)
	})
	t.Run("nil peer shard mapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.PeerShardMapper = nil
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.True(t, check.IfNil(monitor))
		assert.Equal(t, heartbeat.ErrNilPeerShardMapper, err)
	})
	t.Run("invalid max duration peer unresponsive should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.MaxDurationPeerUnresponsive = time.Second - time.Nanosecond
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.True(t, check.IfNil(monitor))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "MaxDurationPeerUnresponsive"))
	})
	t.Run("invalid hide inactive validator interval should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.HideInactiveValidatorInterval = time.Second - time.Nanosecond
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.True(t, check.IfNil(monitor))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "HideInactiveValidatorInterval"))
	})
	t.Run("nil peer type provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.PeerTypeProvider = nil
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.True(t, check.IfNil(monitor))
		assert.Equal(t, heartbeat.ErrNilPeerTypeProvider, err)
	})
	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.AppStatusHandler = nil
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.True(t, check.IfNil(monitor))
		assert.Equal(t, heartbeat.ErrNilAppStatusHandler, err)
	})
	t.Run("invalid time between connections metrics should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.TimeBetweenConnectionsMetricsUpdate = time.Second - time.Nanosecond
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.True(t, check.IfNil(monitor))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "TimeBetweenConnectionsMetricsUpdate"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.TimeBetweenConnectionsMetricsUpdate = time.Second * 3
		args.PeerTypeProvider = &mock.PeerTypeProviderStub{
			ComputeForPubKeyCalled: func(pubKey []byte) (common.PeerType, uint32, error) {
				return common.EligibleList, 0, nil
			},
		}
		counterSetUInt64ValueHandler := 0
		args.AppStatusHandler = &statusHandler.AppStatusHandlerStub{
			SetUInt64ValueHandler: func(key string, value uint64) {
				switch counterSetUInt64ValueHandler {
				case 0:
					assert.Equal(t, common.MetricLiveValidatorNodes, key)
					assert.Equal(t, 1, int(value))
				case 1:
					assert.Equal(t, common.MetricConnectedNodes, key)
					assert.Equal(t, 1, int(value))
				default:
					assert.Fail(t, "only 1 node in test")
				}
				counterSetUInt64ValueHandler++
			},
		}
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))
		assert.Nil(t, err)

		pid1 := []byte("validator peer id")
		message := createHeartbeatMessage(true)
		args.Cache.Put(pid1, message, message.Size())

		assert.Nil(t, monitor.Close())
	})
}

func TestHeartbeatV2Monitor_parseMessage(t *testing.T) {
	t.Parallel()

	t.Run("wrong message type should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		_, err := monitor.parseMessage("pid", "dummy msg", nil)
		assert.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		message := createHeartbeatMessage(true)
		message.Payload = []byte("dummy payload")
		_, err := monitor.parseMessage("pid", message, nil)
		assert.NotNil(t, err)
	})
	t.Run("skippable message should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.PeerShardMapper = &processMocks.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				return core.P2PPeerInfo{
					PeerType: core.UnknownPeer,
				}
			},
		}
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		message := createHeartbeatMessage(false)
		_, err := monitor.parseMessage("pid", message, nil)
		assert.Equal(t, heartbeat.ErrShouldSkipValidator, err)
	})
	t.Run("should work, peer type provider returns error", func(t *testing.T) {
		t.Parallel()

		providedPkBytes := []byte("provided pk")
		providedPkBytesFromMessage := []byte("provided pk message")
		args := createMockHeartbeatV2MonitorArgs()
		args.PeerShardMapper = &processMocks.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				return core.P2PPeerInfo{
					PkBytes: providedPkBytes,
				}
			},
		}
		args.PeerTypeProvider = &mock.PeerTypeProviderStub{
			ComputeForPubKeyCalled: func(pubKey []byte) (common.PeerType, uint32, error) {
				return "", 0, errors.New("some error")
			},
		}
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		numInstances := make(map[string]uint64)
		message := createHeartbeatMessage(true)
		message.Pubkey = providedPkBytesFromMessage
		providedPid := core.PeerID("pid")
		providedMap := map[string]struct{}{
			providedPid.Pretty(): {},
		}
		hb, err := monitor.parseMessage(providedPid, message, numInstances)
		assert.Nil(t, err)
		checkResults(t, *message, hb, true, providedMap, 0)
		assert.Equal(t, 0, len(providedMap))
		pkFromMsg := args.PubKeyConverter.Encode(providedPkBytesFromMessage)
		entries, ok := numInstances[pkFromMsg]
		assert.True(t, ok)
		assert.Equal(t, uint64(1), entries)
		assert.Equal(t, string(common.ObserverList), hb.PeerType)

		pkFromPSM := args.PubKeyConverter.Encode(providedPkBytes)
		_, ok = numInstances[pkFromPSM]
		assert.False(t, ok)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedPkBytes := []byte("provided pk")
		args := createMockHeartbeatV2MonitorArgs()
		args.PeerShardMapper = &processMocks.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				return core.P2PPeerInfo{
					PkBytes: providedPkBytes,
				}
			},
		}
		expectedPeerType := common.EligibleList
		args.PeerTypeProvider = &mock.PeerTypeProviderStub{
			ComputeForPubKeyCalled: func(pubKey []byte) (common.PeerType, uint32, error) {
				return expectedPeerType, 0, nil
			},
		}
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		numInstances := make(map[string]uint64)
		message := createHeartbeatMessage(true)
		providedPid := core.PeerID("pid")
		providedMap := map[string]struct{}{
			providedPid.Pretty(): {},
		}
		hb, err := monitor.parseMessage(providedPid, message, numInstances)
		assert.Nil(t, err)
		checkResults(t, *message, hb, true, providedMap, 0)
		assert.Equal(t, 0, len(providedMap))
		pk := args.PubKeyConverter.Encode(providedPkBytes)
		entries, ok := numInstances[pk]
		assert.True(t, ok)
		assert.Equal(t, uint64(1), entries)
		assert.Equal(t, string(expectedPeerType), hb.PeerType)
	})
}

func TestHeartbeatV2Monitor_getMessageAge(t *testing.T) {
	t.Parallel()

	args := createMockHeartbeatV2MonitorArgs()
	monitor, _ := NewHeartbeatV2Monitor(args)
	assert.False(t, check.IfNil(monitor))

	crtTime := time.Now()
	providedAgeInSec := int64(args.MaxDurationPeerUnresponsive.Seconds() - 1)
	messageTimestamp := crtTime.Unix() - providedAgeInSec

	msgAge := monitor.getMessageAge(crtTime, messageTimestamp)
	assert.Equal(t, providedAgeInSec, int64(msgAge.Seconds()))
}

func TestHeartbeatV2Monitor_isActive(t *testing.T) {
	t.Parallel()

	args := createMockHeartbeatV2MonitorArgs()
	monitor, _ := NewHeartbeatV2Monitor(args)
	assert.False(t, check.IfNil(monitor))

	// negative age should not be active
	assert.False(t, monitor.isActive(monitor.getMessageAge(time.Now(), -10)))
	// one sec old message should be active
	assert.True(t, monitor.isActive(time.Second))
	// too old messages should not be active
	assert.False(t, monitor.isActive(args.MaxDurationPeerUnresponsive+time.Second))
}

func TestHeartbeatV2Monitor_shouldSkipMessage(t *testing.T) {
	t.Parallel()

	args := createMockHeartbeatV2MonitorArgs()
	monitor, _ := NewHeartbeatV2Monitor(args)
	assert.False(t, check.IfNil(monitor))

	// active
	assert.False(t, monitor.shouldSkipMessage(time.Second))
	// inactive but should not hide yet
	assert.False(t, monitor.shouldSkipMessage(args.HideInactiveValidatorInterval-time.Second))
	// inactive and too old should be hidden
	assert.True(t, monitor.shouldSkipMessage(args.HideInactiveValidatorInterval+time.Second))
}

func TestHeartbeatV2Monitor_GetHeartbeats(t *testing.T) {
	t.Parallel()

	t.Run("should work - one of the messages should be skipped", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.PeerShardMapper = &processMocks.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				return core.P2PPeerInfo{
					PkBytes:  pid.Bytes(),
					PeerType: core.ObserverPeer,
				}
			},
		}
		providedStatuses := []bool{true, true, false}
		numOfMessages := len(providedStatuses)
		providedPids := make(map[string]struct{}, numOfMessages)
		providedMessages := make([]*heartbeat.HeartbeatV2, numOfMessages)
		for i := 0; i < numOfMessages; i++ {
			pid := core.PeerID(fmt.Sprintf("%s%d", "pid", i))
			providedPids[pid.Pretty()] = struct{}{}
			providedMessages[i] = createHeartbeatMessage(providedStatuses[i])

			args.Cache.Put(pid.Bytes(), providedMessages[i], providedMessages[i].Size())
		}

		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		heartbeats := monitor.GetHeartbeats()
		assert.Equal(t, len(providedStatuses)-1, len(heartbeats))
		assert.Equal(t, len(providedStatuses)-1, args.Cache.Len()) // faulty message was removed from cache
		for i := 0; i < len(heartbeats); i++ {
			checkResults(t, *providedMessages[i], heartbeats[i], providedStatuses[i], providedPids, 1)
		}
		assert.Equal(t, 1, len(providedPids)) // one message is skipped
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		args := createMockHeartbeatV2MonitorArgs()
		providedStatuses := []bool{true, true, true}
		numOfMessages := len(providedStatuses)
		providedPids := make(map[string]struct{}, numOfMessages)
		providedMessages := make([]*heartbeat.HeartbeatV2, numOfMessages)
		for i := 0; i < numOfMessages; i++ {
			pid := core.PeerID(fmt.Sprintf("%s%d", "pid", i))
			providedPids[pid.Pretty()] = struct{}{}
			providedMessages[i] = createHeartbeatMessage(providedStatuses[i])

			args.Cache.Put(pid.Bytes(), providedMessages[i], providedMessages[i].Size())
		}
		counter := 0
		args.PeerShardMapper = &processMocks.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				// Only first entry is unique, then all should have same pk
				var info core.P2PPeerInfo
				if counter == 0 {
					info = core.P2PPeerInfo{
						PkBytes: pid.Bytes(),
					}
				} else {
					info = core.P2PPeerInfo{
						PkBytes: []byte("same pk"),
					}
				}
				counter++
				return info
			},
		}

		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		heartbeats := monitor.GetHeartbeats()
		assert.Equal(t, args.Cache.Len(), len(heartbeats))
		for i := 0; i < numOfMessages; i++ {
			numInstances := uint64(1)
			if i > 0 {
				numInstances = 2
			}
			checkResults(t, *providedMessages[i], heartbeats[i], providedStatuses[i], providedPids, numInstances)
		}
		assert.Equal(t, 0, len(providedPids))
	})
}

func checkResults(t *testing.T, message heartbeat.HeartbeatV2, hb data.PubKeyHeartbeat, isActive bool, providedPids map[string]struct{}, numInstances uint64) {
	assert.Equal(t, isActive, hb.IsActive)
	assert.Equal(t, message.VersionNumber, hb.VersionNumber)
	assert.Equal(t, message.NodeDisplayName, hb.NodeDisplayName)
	assert.Equal(t, message.Identity, hb.Identity)
	assert.Equal(t, message.Nonce, hb.Nonce)
	assert.Equal(t, message.PeerSubType, hb.PeerSubType)
	assert.Equal(t, numInstances, hb.NumInstances)
	_, ok := providedPids[hb.PidString]
	assert.True(t, ok)
	delete(providedPids, hb.PidString)
}
