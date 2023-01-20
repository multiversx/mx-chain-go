package monitor

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/heartbeat/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/assert"
)

func createMockHeartbeatV2MonitorArgs() ArgHeartbeatV2Monitor {
	return ArgHeartbeatV2Monitor{
		Cache:                         testscommon.NewCacherMock(),
		PubKeyConverter:               &testscommon.PubkeyConverterMock{},
		Marshaller:                    &marshallerMock.MarshalizerMock{},
		MaxDurationPeerUnresponsive:   time.Second * 3,
		HideInactiveValidatorInterval: time.Second * 5,
		ShardId:                       0,
		PeerTypeProvider:              &mock.PeerTypeProviderStub{},
	}
}

func createHeartbeatMessage(active bool, publicKeyBytes []byte) *heartbeat.HeartbeatV2 {
	crtTime := time.Now()
	providedAgeInSec := int64(1)
	messageTimestamp := crtTime.Unix() - providedAgeInSec

	if !active {
		messageTimestamp = crtTime.Unix() - int64(60)
	}

	payload := heartbeat.Payload{
		Timestamp: messageTimestamp,
	}

	marshaller := marshallerMock.MarshalizerMock{}
	payloadBytes, _ := marshaller.Marshal(payload)
	return &heartbeat.HeartbeatV2{
		Payload:            payloadBytes,
		VersionNumber:      "v01",
		NodeDisplayName:    "node name",
		Identity:           "identity",
		Nonce:              0,
		PeerSubType:        0,
		Pubkey:             publicKeyBytes,
		NumTrieNodesSynced: 150,
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
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		args.PeerTypeProvider = &mock.PeerTypeProviderStub{
			ComputeForPubKeyCalled: func(pubKey []byte) (common.PeerType, uint32, error) {
				return common.EligibleList, 0, nil
			},
		}
		monitor, err := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))
		assert.Nil(t, err)

		pid1 := []byte("validator peer id")
		message := createHeartbeatMessage(true, []byte("public key"))
		args.Cache.Put(pid1, message, message.Size())
	})
}

func TestHeartbeatV2Monitor_parseMessage(t *testing.T) {
	t.Parallel()

	t.Run("wrong message type should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		_, err := monitor.parseMessage("pid", "dummy msg")
		assert.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		message := createHeartbeatMessage(true, []byte("public key"))
		message.Payload = []byte("dummy payload")
		_, err := monitor.parseMessage("pid", message)
		assert.NotNil(t, err)
	})
	t.Run("skippable message should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2MonitorArgs()
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		message := createHeartbeatMessage(false, []byte("public key"))
		_, err := monitor.parseMessage("pid", message)
		assert.True(t, errors.Is(err, heartbeat.ErrShouldSkipValidator))
	})
	t.Run("should work, peer type provider returns error", func(t *testing.T) {
		t.Parallel()

		providedPkBytes := []byte("provided pk")
		providedPkBytesFromMessage := []byte("provided pk message")
		args := createMockHeartbeatV2MonitorArgs()
		args.PeerTypeProvider = &mock.PeerTypeProviderStub{
			ComputeForPubKeyCalled: func(pubKey []byte) (common.PeerType, uint32, error) {
				return "", 0, errors.New("some error")
			},
		}
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		message := createHeartbeatMessage(true, providedPkBytes)
		message.Pubkey = providedPkBytesFromMessage
		providedPid := core.PeerID("pid")
		providedMap := map[string]struct{}{
			providedPid.Pretty(): {},
		}
		hb, err := monitor.parseMessage(providedPid, message)
		assert.Nil(t, err)
		checkResults(t, message, *hb, true, providedMap, 0)
		assert.Equal(t, 0, len(providedMap))
		assert.Equal(t, string(common.ObserverList), hb.PeerType)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedPkBytes := []byte("provided pk")
		args := createMockHeartbeatV2MonitorArgs()
		expectedPeerType := common.EligibleList
		args.PeerTypeProvider = &mock.PeerTypeProviderStub{
			ComputeForPubKeyCalled: func(pubKey []byte) (common.PeerType, uint32, error) {
				return expectedPeerType, 0, nil
			},
		}
		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		message := createHeartbeatMessage(true, providedPkBytes)
		providedPid := core.PeerID("pid")
		providedMap := map[string]struct{}{
			providedPid.Pretty(): {},
		}
		hb, err := monitor.parseMessage(providedPid, message)
		assert.Nil(t, err)
		checkResults(t, message, *hb, true, providedMap, 0)
		assert.Equal(t, 0, len(providedMap))
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
	messageTime := time.Unix(messageTimestamp, 0)

	msgAge := monitor.getMessageAge(crtTime, messageTime)
	assert.Equal(t, providedAgeInSec, int64(msgAge.Seconds()))
}

func TestHeartbeatV2Monitor_isActive(t *testing.T) {
	t.Parallel()

	args := createMockHeartbeatV2MonitorArgs()
	monitor, _ := NewHeartbeatV2Monitor(args)
	assert.False(t, check.IfNil(monitor))
	messageTimestamp := int64(-10)
	messageTime := time.Unix(messageTimestamp, 0)

	// negative age should not be active
	assert.False(t, monitor.isActive(monitor.getMessageAge(time.Now(), messageTime)))
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
		providedStatuses := []bool{true, true, false}
		numOfMessages := len(providedStatuses)
		providedPids := make(map[string]struct{}, numOfMessages)
		providedMessages := make([]*heartbeat.HeartbeatV2, numOfMessages)
		for i := 0; i < numOfMessages; i++ {
			pid := core.PeerID(fmt.Sprintf("%s%d", "pid", i))
			providedPids[pid.Pretty()] = struct{}{}
			providedMessages[i] = createHeartbeatMessage(providedStatuses[i], pid.Bytes())

			args.Cache.Put(pid.Bytes(), providedMessages[i], providedMessages[i].Size())
		}

		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		heartbeats := monitor.GetHeartbeats()
		assert.Equal(t, len(providedStatuses)-1, len(heartbeats))
		assert.Equal(t, len(providedStatuses)-1, args.Cache.Len()) // faulty message was removed from cache
		for i := 0; i < len(heartbeats); i++ {
			checkResults(t, providedMessages[i], heartbeats[i], providedStatuses[i], providedPids, 1)
		}
		assert.Equal(t, 1, len(providedPids)) // one message is skipped
	})
	t.Run("should remove all inactive messages except from the latest message", func(t *testing.T) {
		t.Parallel()
		args := createMockHeartbeatV2MonitorArgs()
		args.HideInactiveValidatorInterval = time.Minute
		providedStatuses := []bool{false, false, false}
		numOfMessages := len(providedStatuses)
		providedPids := make(map[string]struct{}, numOfMessages)
		providedMessages := make([]*heartbeat.HeartbeatV2, numOfMessages)
		for i := 0; i < numOfMessages; i++ {
			pid := core.PeerID(fmt.Sprintf("%s%d", "pid", i))
			providedPids[pid.Pretty()] = struct{}{}
			pkBytes := []byte("same pk")

			providedMessages[i] = createHeartbeatMessage(providedStatuses[i], pkBytes)
			payload := heartbeat.Payload{
				Timestamp: time.Now().Unix() - 30 + int64(i), // the last message will be the latest, so it will be returned
			}

			marshaller := marshallerMock.MarshalizerMock{}
			payloadBytes, _ := marshaller.Marshal(payload)
			providedMessages[i].Payload = payloadBytes

			args.Cache.Put(pid.Bytes(), providedMessages[i], providedMessages[i].Size())
		}

		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		heartbeats := monitor.GetHeartbeats()
		assert.Equal(t, 1, len(heartbeats))
		assert.Equal(t, 1, args.Cache.Len())
		checkResults(t, providedMessages[2], heartbeats[0], providedStatuses[2], providedPids, 0)
		assert.Equal(t, 2, len(providedPids)) // 1 inactive was removed from the heartbeat list
	})
	t.Run("should remove all inactive messages if one is active", func(t *testing.T) {
		t.Parallel()
		args := createMockHeartbeatV2MonitorArgs()
		args.HideInactiveValidatorInterval = time.Minute
		providedStatuses := []bool{true, false, false}
		numOfMessages := len(providedStatuses)
		providedPids := make(map[string]struct{}, numOfMessages)
		providedMessages := make([]*heartbeat.HeartbeatV2, numOfMessages)
		for i := 0; i < numOfMessages; i++ {
			pid := core.PeerID(fmt.Sprintf("%s%d", "pid", i))
			providedPids[pid.Pretty()] = struct{}{}
			pkBytes := []byte("same pk")

			providedMessages[i] = createHeartbeatMessage(providedStatuses[i], pkBytes)
			if i != 0 {
				// 1, 2 ... are inactive messages
				payload := heartbeat.Payload{
					Timestamp: time.Now().Unix() - 30 + int64(i), // the last message will be the latest, so it will be returned
				}

				marshaller := marshallerMock.MarshalizerMock{}
				payloadBytes, _ := marshaller.Marshal(payload)
				providedMessages[i].Payload = payloadBytes
			}

			args.Cache.Put(pid.Bytes(), providedMessages[i], providedMessages[i].Size())
		}

		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		heartbeats := monitor.GetHeartbeats()
		assert.Equal(t, 1, len(heartbeats))
		assert.Equal(t, 1, args.Cache.Len())
		checkResults(t, providedMessages[0], heartbeats[0], providedStatuses[0], providedPids, 1)
		assert.Equal(t, 2, len(providedPids)) // 1 inactive was removed from the heartbeat list
	})
	t.Run("nil message in cache should remove", func(t *testing.T) {
		t.Parallel()
		args := createMockHeartbeatV2MonitorArgs()
		args.HideInactiveValidatorInterval = time.Minute
		args.Cache.Put([]byte("pid"), nil, 0)

		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		heartbeats := monitor.GetHeartbeats()
		assert.Equal(t, 0, len(heartbeats))
		assert.Equal(t, 0, args.Cache.Len())
	})
	t.Run("should work with 2 public keys on different pids and one public key", func(t *testing.T) {
		t.Parallel()
		args := createMockHeartbeatV2MonitorArgs()
		providedStatuses := []bool{true, true, true}
		numOfMessages := len(providedStatuses)
		providedPids := make(map[string]struct{}, numOfMessages)
		providedMessages := make([]*heartbeat.HeartbeatV2, numOfMessages)
		for i := 0; i < numOfMessages; i++ {
			pid := core.PeerID(fmt.Sprintf("%s%d", "pid", i))
			providedPids[pid.Pretty()] = struct{}{}

			pkBytes := []byte("same pk")
			if i == 0 {
				pkBytes = pid.Bytes()
			}

			providedMessages[i] = createHeartbeatMessage(providedStatuses[i], pkBytes)

			args.Cache.Put(pid.Bytes(), providedMessages[i], providedMessages[i].Size())
		}

		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		heartbeats := monitor.GetHeartbeats()
		assert.Equal(t, 2, len(heartbeats)) // 2 have the same pk
		for i := 0; i < 2; i++ {
			numInstances := uint64(1)
			if i > 0 {
				numInstances = 2
			}
			checkResults(t, providedMessages[i], heartbeats[i], providedStatuses[i], providedPids, numInstances)
		}
		assert.Equal(t, 1, len(providedPids)) // 1 active was removed from the heartbeat list
	})
	t.Run("should choose the 'smaller' pid if multiple active messages are found (sort should work)", func(t *testing.T) {
		t.Parallel()
		args := createMockHeartbeatV2MonitorArgs()
		providedStatuses := []bool{true, true, true}
		numOfMessages := len(providedStatuses)
		providedPids := make(map[string]struct{}, numOfMessages)
		providedMessages := make([]*heartbeat.HeartbeatV2, numOfMessages)
		for i := numOfMessages - 1; i >= 0; i-- {
			pid := core.PeerID(fmt.Sprintf("%s%d", "pid", i))
			providedPids[pid.Pretty()] = struct{}{}

			pkBytes := []byte("same pk")
			providedMessages[i] = createHeartbeatMessage(providedStatuses[i], pkBytes)

			args.Cache.Put(pid.Bytes(), providedMessages[i], providedMessages[i].Size())
		}

		monitor, _ := NewHeartbeatV2Monitor(args)
		assert.False(t, check.IfNil(monitor))

		heartbeats := monitor.GetHeartbeats()
		assert.Equal(t, 1, len(heartbeats))
		checkResults(t, providedMessages[0], heartbeats[0], providedStatuses[0], providedPids, 3)
	})
}

func checkResults(t *testing.T, message *heartbeat.HeartbeatV2, hb data.PubKeyHeartbeat, isActive bool, providedPids map[string]struct{}, numInstances uint64) {
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
	assert.Equal(t, message.NumTrieNodesSynced, hb.NumTrieNodesReceived)
}
