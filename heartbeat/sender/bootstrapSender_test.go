package sender

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/heartbeat/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createMockBootstrapSenderArgs() ArgBootstrapSender {
	return ArgBootstrapSender{
		Messenger:                          &p2pmocks.MessengerStub{},
		Marshaller:                         &marshallerMock.MarshalizerMock{},
		HeartbeatTopic:                     "hb-topic",
		HeartbeatTimeBetweenSends:          time.Second,
		HeartbeatTimeBetweenSendsWhenError: time.Second,
		HeartbeatTimeThresholdBetweenSends: 0.1,
		VersionNumber:                      "v1",
		NodeDisplayName:                    "node",
		Identity:                           "identity",
		PeerSubType:                        core.RegularPeer,
		CurrentBlockProvider:               &mock.CurrentBlockProviderStub{},
		PrivateKey:                         &cryptoMocks.PrivateKeyStub{},
		RedundancyHandler:                  &mock.RedundancyHandlerStub{},
		PeerTypeProvider:                   &mock.PeerTypeProviderStub{},
		TrieSyncStatisticsProvider:         &testscommon.SizeSyncStatisticsHandlerStub{},
	}
}

func TestNewBootstrapSender(t *testing.T) {
	t.Parallel()

	t.Run("nil peer messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.Messenger = nil
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilMessenger, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.Marshaller = nil
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilMarshaller, err)
	})
	t.Run("empty heartbeat topic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.HeartbeatTopic = ""
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.HeartbeatTimeBetweenSends = time.Second - time.Nanosecond
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends when error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.HeartbeatTimeBetweenSendsWhenError = time.Second - time.Nanosecond
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("version number too long should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.VersionNumber = string(make([]byte, 150))
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
		assert.True(t, strings.Contains(err.Error(), "versionNumber"))
	})
	t.Run("node display name too long should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.NodeDisplayName = string(make([]byte, 150))
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
		assert.True(t, strings.Contains(err.Error(), "nodeDisplayName"))
	})
	t.Run("identity name too long should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.Identity = string(make([]byte, 150))
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
		assert.True(t, strings.Contains(err.Error(), "identity"))
	})
	t.Run("nil current block provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.CurrentBlockProvider = nil
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilCurrentBlockProvider, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.PrivateKey = nil
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilPrivateKey, err)
	})
	t.Run("nil redundancy handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.RedundancyHandler = nil
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilRedundancyHandler, err)
	})
	t.Run("nil peer type provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.PeerTypeProvider = nil
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilPeerTypeProvider, err)
	})
	t.Run("nil trie sync statistics provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		args.TrieSyncStatisticsProvider = nil
		senderInstance, err := NewBootstrapSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilTrieSyncStatisticsProvider, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockBootstrapSenderArgs()
		senderInstance, err := NewBootstrapSender(args)

		assert.False(t, check.IfNil(senderInstance))
		assert.Nil(t, err)
	})
}

func TestBootstrapSender_Close(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	args := createMockBootstrapSenderArgs()
	senderInstance, _ := NewBootstrapSender(args)
	err := senderInstance.Close()
	assert.Nil(t, err)
}
