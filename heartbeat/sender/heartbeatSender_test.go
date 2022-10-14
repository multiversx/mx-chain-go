package sender

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createMockHeartbeatSenderArgs(argBase argBaseSender) argHeartbeatSender {
	return argHeartbeatSender{
		argBaseSender:        argBase,
		versionNumber:        "v1",
		nodeDisplayName:      "node",
		identity:             "identity",
		peerSubType:          core.RegularPeer,
		currentBlockProvider: &mock.CurrentBlockProviderStub{},
		peerTypeProvider:     &mock.PeerTypeProviderStub{},
	}
}

func TestNewHeartbeatSender(t *testing.T) {
	t.Parallel()

	t.Run("nil peer messenger should error", func(t *testing.T) {
		t.Parallel()

		argBase := createMockBaseArgs()
		argBase.messenger = nil
		args := createMockHeartbeatSenderArgs(argBase)
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilMessenger, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		argBase := createMockBaseArgs()
		argBase.marshaller = nil
		args := createMockHeartbeatSenderArgs(argBase)
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilMarshaller, err)
	})
	t.Run("empty topic should error", func(t *testing.T) {
		t.Parallel()

		argBase := createMockBaseArgs()
		argBase.topic = ""
		args := createMockHeartbeatSenderArgs(argBase)
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		argBase := createMockBaseArgs()
		argBase.timeBetweenSends = time.Second - time.Nanosecond
		args := createMockHeartbeatSenderArgs(argBase)
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends when error should error", func(t *testing.T) {
		t.Parallel()

		argBase := createMockBaseArgs()
		argBase.timeBetweenSendsWhenError = time.Second - time.Nanosecond
		args := createMockHeartbeatSenderArgs(argBase)
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		args.privKey = nil
		senderInstance, err := newHeartbeatSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrNilPrivateKey, err)
	})
	t.Run("nil redundancy handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		args.redundancyHandler = nil
		senderInstance, err := newHeartbeatSender(args)

		assert.True(t, check.IfNil(senderInstance))
		assert.Equal(t, heartbeat.ErrNilRedundancyHandler, err)
	})
	t.Run("version number too long should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		args.versionNumber = string(make([]byte, 150))
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
		assert.True(t, strings.Contains(err.Error(), "versionNumber"))
	})
	t.Run("node display name too long should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		args.nodeDisplayName = string(make([]byte, 150))
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
		assert.True(t, strings.Contains(err.Error(), "nodeDisplayName"))
	})
	t.Run("identity too long should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		args.identity = string(make([]byte, 150))
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
		assert.True(t, strings.Contains(err.Error(), "identity"))
	})
	t.Run("nil current block provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		args.currentBlockProvider = nil
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilCurrentBlockProvider, err)
	})
	t.Run("threshold too small should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		args.thresholdBetweenSends = 0.001
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidThreshold))
		assert.True(t, strings.Contains(err.Error(), "thresholdBetweenSends"))
	})
	t.Run("threshold too big should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		args.thresholdBetweenSends = 1.001
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidThreshold))
		assert.True(t, strings.Contains(err.Error(), "thresholdBetweenSends"))
	})
	t.Run("nil peer type provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		args.peerTypeProvider = nil
		senderInstance, err := newHeartbeatSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilPeerTypeProvider, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs(createMockBaseArgs())
		senderInstance, err := newHeartbeatSender(args)

		assert.False(t, check.IfNil(senderInstance))
		assert.Nil(t, err)
	})
}

func TestHeartbeatSender_Execute(t *testing.T) {
	t.Parallel()

	t.Run("execute errors, should set the error time duration value", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		argsBase := createMockBaseArgs()
		argsBase.timeBetweenSendsWhenError = time.Second * 3
		argsBase.timeBetweenSends = time.Second * 2
		argsBase.marshaller = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}

		args := createMockHeartbeatSenderArgs(argsBase)
		senderInstance, _ := newHeartbeatSender(args)
		senderInstance.timerHandler = &mock.TimerHandlerStub{
			CreateNewTimerCalled: func(duration time.Duration) {
				assert.Equal(t, argsBase.timeBetweenSendsWhenError, duration)
				wasCalled = true
			},
		}

		senderInstance.Execute()
		assert.True(t, wasCalled)
	})
	t.Run("execute worked, should set the normal time duration value", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		argsBase := createMockBaseArgs()
		argsBase.timeBetweenSendsWhenError = time.Second * 3
		argsBase.timeBetweenSends = time.Second * 2

		args := createMockHeartbeatSenderArgs(argsBase)
		senderInstance, _ := newHeartbeatSender(args)
		senderInstance.timerHandler = &mock.TimerHandlerStub{
			CreateNewTimerCalled: func(duration time.Duration) {
				floatTBS := float64(argsBase.timeBetweenSends.Nanoseconds())
				maxDuration := floatTBS + floatTBS*argsBase.thresholdBetweenSends
				assert.True(t, time.Duration(maxDuration) > duration)
				assert.True(t, argsBase.timeBetweenSends <= duration)
				wasCalled = true
			},
		}

		senderInstance.Execute()
		assert.True(t, wasCalled)
	})
}

func TestHeartbeatSender_execute(t *testing.T) {
	t.Parallel()

	t.Run("marshal returns error first time", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		argsBase.marshaller = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}

		args := createMockHeartbeatSenderArgs(argsBase)
		senderInstance, _ := newHeartbeatSender(args)
		assert.False(t, check.IfNil(senderInstance))

		err := senderInstance.execute()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("marshal returns error second time", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		numOfCalls := 0
		argsBase.marshaller = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				if numOfCalls < 1 {
					numOfCalls++
					return []byte(""), nil
				}

				return nil, expectedErr
			},
		}

		args := createMockHeartbeatSenderArgs(argsBase)
		senderInstance, _ := newHeartbeatSender(args)
		assert.False(t, check.IfNil(senderInstance))

		err := senderInstance.execute()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockBaseArgs()
		broadcastCalled := false
		argsBase.messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Equal(t, argsBase.topic, topic)
				recoveredMessage := &heartbeat.HeartbeatV2{}
				err := argsBase.marshaller.Unmarshal(recoveredMessage, buff)
				assert.Nil(t, err)
				pk := argsBase.privKey.GeneratePublic()
				pkBytes, _ := pk.ToByteArray()
				assert.Equal(t, pkBytes, recoveredMessage.Pubkey)
				broadcastCalled = true
			},
		}

		args := createMockHeartbeatSenderArgs(argsBase)

		args.currentBlockProvider = &mock.CurrentBlockProviderStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{}
			},
		}

		providedPeerType := common.EligibleList
		args.peerTypeProvider = &mock.PeerTypeProviderStub{
			ComputeForPubKeyCalled: func(pubKey []byte) (common.PeerType, uint32, error) {
				return providedPeerType, 0, nil
			},
		}

		senderInstance, _ := newHeartbeatSender(args)
		assert.False(t, check.IfNil(senderInstance))

		err := senderInstance.execute()
		assert.Nil(t, err)
		assert.True(t, broadcastCalled)
		assert.Equal(t, uint64(1), args.currentBlockProvider.GetCurrentBlockHeader().GetNonce())
	})
}

func TestHeartbeatSender_getSenderInfo(t *testing.T) {
	t.Parallel()

	args := createMockHeartbeatSenderArgs(createMockBaseArgs())
	args.peerSubType = core.FullHistoryObserver
	args.peerTypeProvider = &mock.PeerTypeProviderStub{
		ComputeForPubKeyCalled: func(pubKey []byte) (common.PeerType, uint32, error) {
			return common.EligibleList, 0, nil
		},
	}
	senderInstance, _ := newHeartbeatSender(args)

	peerType, subType, err := senderInstance.getSenderInfo()
	assert.Nil(t, err)
	assert.Equal(t, string(common.EligibleList), peerType)
	assert.Equal(t, core.FullHistoryObserver, subType)
}
