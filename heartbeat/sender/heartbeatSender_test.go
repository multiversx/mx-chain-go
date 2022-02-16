package sender

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")

func createMockHeartbeatSenderArgs() ArgHeartbeatSender {
	return ArgHeartbeatSender{
		ArgBaseSender: ArgBaseSender{
			Messenger:                 &mock.MessengerStub{},
			Marshaller:                &mock.MarshallerMock{},
			Topic:                     "topic",
			TimeBetweenSends:          time.Second,
			TimeBetweenSendsWhenError: time.Second,
		},
		VersionNumber:        "v1",
		NodeDisplayName:      "node",
		Identity:             "identity",
		PeerSubType:          core.RegularPeer,
		CurrentBlockProvider: &mock.CurrentBlockProviderStub{},
	}
}

func TestNewHeartbeatSender(t *testing.T) {
	t.Parallel()

	t.Run("nil peer messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.Messenger = nil
		sender, err := NewHeartbeatSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilMessenger, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.Marshaller = nil
		sender, err := NewHeartbeatSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilMarshaller, err)
	})
	t.Run("empty topic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.Topic = ""
		sender, err := NewHeartbeatSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.TimeBetweenSends = time.Second - time.Nanosecond
		sender, err := NewHeartbeatSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "TimeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "TimeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.TimeBetweenSendsWhenError = time.Second - time.Nanosecond
		sender, err := NewHeartbeatSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "TimeBetweenSendsWhenError"))
	})
	t.Run("empty version number should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.VersionNumber = ""
		sender, err := NewHeartbeatSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptyVersionNumber, err)
	})
	t.Run("empty node display name should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.NodeDisplayName = ""
		sender, err := NewHeartbeatSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptyNodeDisplayName, err)
	})
	t.Run("empty identity should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.Identity = ""
		sender, err := NewHeartbeatSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptyIdentity, err)
	})
	t.Run("nil current block provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.CurrentBlockProvider = nil
		sender, err := NewHeartbeatSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilCurrentBlockProvider, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		sender, err := NewHeartbeatSender(args)

		assert.False(t, check.IfNil(sender))
		assert.Nil(t, err)
	})
}

func TestHeartbeatSender_Execute(t *testing.T) {
	t.Parallel()

	t.Run("execute errors, should set the error time duration value", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockHeartbeatSenderArgs()
		args.TimeBetweenSendsWhenError = time.Second * 3
		args.TimeBetweenSends = time.Second * 2
		args.Marshaller = &mock.MarshallerStub{
			MarshalHandler: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		sender, _ := NewHeartbeatSender(args)
		sender.timerHandler = &mock.TimerHandlerStub{
			CreateNewTimerCalled: func(duration time.Duration) {
				assert.Equal(t, args.TimeBetweenSendsWhenError, duration)
				wasCalled = true
			},
		}

		sender.Execute()
		assert.True(t, wasCalled)
	})
	t.Run("execute worked, should set the normal time duration value", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockHeartbeatSenderArgs()
		args.TimeBetweenSendsWhenError = time.Second * 3
		args.TimeBetweenSends = time.Second * 2
		sender, _ := NewHeartbeatSender(args)
		sender.timerHandler = &mock.TimerHandlerStub{
			CreateNewTimerCalled: func(duration time.Duration) {
				assert.Equal(t, args.TimeBetweenSends, duration)
				wasCalled = true
			},
		}

		sender.Execute()
		assert.True(t, wasCalled)
	})
}

func TestHeartbeatSender_execute(t *testing.T) {
	t.Parallel()

	t.Run("marshal returns error first time", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		args.Marshaller = &mock.MarshallerStub{
			MarshalHandler: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		sender, _ := NewHeartbeatSender(args)
		assert.False(t, check.IfNil(sender))

		err := sender.execute()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("marshal returns error second time", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		numOfCalls := 0
		args.Marshaller = &mock.MarshallerStub{
			MarshalHandler: func(obj interface{}) ([]byte, error) {
				if numOfCalls < 1 {
					numOfCalls++
					return []byte(""), nil
				}

				return nil, expectedErr
			},
		}
		sender, _ := NewHeartbeatSender(args)
		assert.False(t, check.IfNil(sender))

		err := sender.execute()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatSenderArgs()
		broadcastCalled := false
		args.Messenger = &mock.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				assert.Equal(t, args.Topic, topic)
				broadcastCalled = true
			},
		}

		args.CurrentBlockProvider = &mock.CurrentBlockProviderStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{}
			},
		}

		sender, _ := NewHeartbeatSender(args)
		assert.False(t, check.IfNil(sender))

		err := sender.execute()
		assert.Nil(t, err)
		assert.True(t, broadcastCalled)
		assert.Equal(t, uint64(1), args.CurrentBlockProvider.GetCurrentBlockHeader().GetNonce())
	})
}
