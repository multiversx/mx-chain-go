package sender

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/stretchr/testify/assert"
)

func createMockSenderArgs() ArgSender {
	return ArgSender{
		Messenger:                          &mock.MessengerStub{},
		Marshaller:                         &mock.MarshallerMock{},
		PeerAuthenticationTopic:            "pa-topic",
		HeartbeatTopic:                     "hb-topic",
		PeerAuthenticationTimeBetweenSends: time.Second,
		PeerAuthenticationTimeBetweenSendsWhenError: time.Second,
		HeartbeatTimeBetweenSends:                   time.Second,
		HeartbeatTimeBetweenSendsWhenError:          time.Second,
		VersionNumber:                               "v1",
		NodeDisplayName:                             "node",
		Identity:                                    "identity",
		PeerSubType:                                 core.RegularPeer,
		CurrentBlockProvider:                        &mock.CurrentBlockProviderStub{},
		PeerSignatureHandler:                        &mock.PeerSignatureHandlerStub{},
		PrivateKey:                                  &mock.PrivateKeyStub{},
		RedundancyHandler:                           &mock.RedundancyHandlerStub{},
	}
}

func TestNewSender(t *testing.T) {
	t.Parallel()

	t.Run("nil peer messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.Messenger = nil
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilMessenger, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.Marshaller = nil
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilMarshaller, err)
	})
	t.Run("empty peer auth topic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerAuthenticationTopic = ""
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("empty heartbeat topic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.HeartbeatTopic = ""
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("invalid peer auth time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerAuthenticationTimeBetweenSends = time.Second - time.Nanosecond
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid peer auth time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerAuthenticationTimeBetweenSendsWhenError = time.Second - time.Nanosecond
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.HeartbeatTimeBetweenSends = time.Second - time.Nanosecond
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.HeartbeatTimeBetweenSendsWhenError = time.Second - time.Nanosecond
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("empty version number should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.VersionNumber = ""
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptyVersionNumber, err)
	})
	t.Run("empty node display name should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.NodeDisplayName = ""
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptyNodeDisplayName, err)
	})
	t.Run("empty identity should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.Identity = ""
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrEmptyIdentity, err)
	})
	t.Run("nil current block provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.CurrentBlockProvider = nil
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilCurrentBlockProvider, err)
	})
	t.Run("nil peer signature handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerSignatureHandler = nil
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilPeerSignatureHandler, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PrivateKey = nil
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilPrivateKey, err)
	})
	t.Run("nil redundancy handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.RedundancyHandler = nil
		sender, err := NewSender(args)

		assert.Nil(t, sender)
		assert.Equal(t, heartbeat.ErrNilRedundancyHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		sender, err := NewSender(args)

		assert.False(t, check.IfNil(sender))
		assert.Nil(t, err)
	})
}

func TestSender_Close(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	args := createMockSenderArgs()
	sender, _ := NewSender(args)
	sender.Close()
}
