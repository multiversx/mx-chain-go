package sender

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/heartbeat/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockSenderArgs() ArgSender {
	return ArgSender{
		MainMessenger:                               &p2pmocks.MessengerStub{},
		FullArchiveMessenger:                        &p2pmocks.MessengerStub{},
		Marshaller:                                  &marshallerMock.MarshalizerMock{},
		PeerAuthenticationTopic:                     "pa-topic",
		HeartbeatTopic:                              "hb-topic",
		PeerAuthenticationTimeBetweenSends:          time.Second,
		PeerAuthenticationTimeBetweenSendsWhenError: time.Second,
		PeerAuthenticationTimeThresholdBetweenSends: 0.1,
		HeartbeatTimeBetweenSends:                   time.Second,
		HeartbeatTimeBetweenSendsWhenError:          time.Second,
		HeartbeatTimeThresholdBetweenSends:          0.1,
		BaseVersionNumber:                           "v1-base",
		VersionNumber:                               "v1",
		NodeDisplayName:                             "node",
		Identity:                                    "identity",
		PeerSubType:                                 core.RegularPeer,
		CurrentBlockProvider:                        &mock.CurrentBlockProviderStub{},
		PeerSignatureHandler:                        &cryptoMocks.PeerSignatureHandlerStub{},
		PrivateKey:                                  &cryptoMocks.PrivateKeyStub{},
		RedundancyHandler:                           &mock.RedundancyHandlerStub{},
		NodesCoordinator:                            &shardingMocks.NodesCoordinatorStub{},
		HardforkTrigger:                             &testscommon.HardforkTriggerStub{},
		HardforkTimeBetweenSends:                    time.Second,
		HardforkTriggerPubKey:                       providedHardforkPubKey,
		PeerTypeProvider:                            &mock.PeerTypeProviderStub{},
		ManagedPeersHolder:                          &testscommon.ManagedPeersHolderStub{},
		PeerAuthenticationTimeBetweenChecks:         time.Second,
		ShardCoordinator:                            createShardCoordinatorInShard(0),
	}
}

func TestNewSender(t *testing.T) {
	t.Parallel()

	t.Run("nil main messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.MainMessenger = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrNilMessenger))
	})
	t.Run("nil full archive messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.FullArchiveMessenger = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrNilMessenger))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.Marshaller = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilMarshaller, err)
	})
	t.Run("empty peer auth topic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerAuthenticationTopic = ""
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("empty heartbeat topic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.HeartbeatTopic = ""
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrEmptySendTopic, err)
	})
	t.Run("invalid peer auth time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerAuthenticationTimeBetweenSends = time.Second - time.Nanosecond
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid peer auth time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerAuthenticationTimeBetweenSendsWhenError = time.Second - time.Nanosecond
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.HeartbeatTimeBetweenSends = time.Second - time.Nanosecond
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSends"))
		assert.False(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("invalid time between sends when error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.HeartbeatTimeBetweenSendsWhenError = time.Second - time.Nanosecond
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSendsWhenError"))
	})
	t.Run("version number too long should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.VersionNumber = string(make([]byte, 150))
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
		assert.True(t, strings.Contains(err.Error(), "versionNumber"))
	})
	t.Run("node display name too long should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.NodeDisplayName = string(make([]byte, 150))
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
		assert.True(t, strings.Contains(err.Error(), "nodeDisplayName"))
	})
	t.Run("identity name too long should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.Identity = string(make([]byte, 150))
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrPropertyTooLong))
		assert.True(t, strings.Contains(err.Error(), "identity"))
	})
	t.Run("nil current block provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.CurrentBlockProvider = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilCurrentBlockProvider, err)
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.NodesCoordinator = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilNodesCoordinator, err)
	})
	t.Run("nil peer signature handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerSignatureHandler = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilPeerSignatureHandler, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PrivateKey = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilPrivateKey, err)
	})
	t.Run("nil redundancy handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.RedundancyHandler = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilRedundancyHandler, err)
	})
	t.Run("nil hardfork trigger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.HardforkTrigger = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilHardforkTrigger, err)
	})
	t.Run("invalid time between hardforks should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.HardforkTimeBetweenSends = time.Second - time.Nanosecond
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "hardforkTimeBetweenSends"))
	})
	t.Run("invalid hardfork pub key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.HardforkTriggerPubKey = make([]byte, 0)
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "hardfork"))
	})
	t.Run("nil peer type provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerTypeProvider = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.Equal(t, heartbeat.ErrNilPeerTypeProvider, err)
	})
	t.Run("nil managed peers holder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.ManagedPeersHolder = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrNilManagedPeersHolder))
	})
	t.Run("invalid time between checks should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.PeerAuthenticationTimeBetweenChecks = time.Second - time.Nanosecond
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenChecks"))
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		args.ShardCoordinator = nil
		senderInstance, err := NewSender(args)

		assert.Nil(t, senderInstance)
		assert.True(t, errors.Is(err, heartbeat.ErrNilShardCoordinator))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockSenderArgs()
		senderInstance, err := NewSender(args)

		assert.NotNil(t, senderInstance)
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
	senderInstance, _ := NewSender(args)
	err := senderInstance.Close()
	assert.Nil(t, err)
}

func TestSender_GetCurrentNodeTypeShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	args := createMockSenderArgs()
	senderInstance, err := NewSender(args)
	require.Nil(t, err)

	_, _, err = senderInstance.GetCurrentNodeType()
	assert.Nil(t, err)

	_ = senderInstance.Close()
}

func TestSender_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var senderInstance *sender
	assert.True(t, senderInstance.IsInterfaceNil())

	senderInstance, _ = NewSender(createMockSenderArgs())
	assert.False(t, senderInstance.IsInterfaceNil())
}
