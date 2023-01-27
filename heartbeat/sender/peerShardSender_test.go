package sender

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

func createMockArgPeerShardSender() ArgPeerShardSender {
	return ArgPeerShardSender{
		Messenger:                 &p2pmocks.MessengerStub{},
		Marshaller:                &marshal.GogoProtoMarshalizer{},
		ShardCoordinator:          &testscommon.ShardsCoordinatorMock{},
		TimeBetweenSends:          time.Second,
		TimeThresholdBetweenSends: 0.1,
		NodesCoordinator: &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, errors.New("the node is an observer")
			},
		},
	}
}

func TestNewPeerShardSender(t *testing.T) {
	t.Parallel()

	t.Run("nil messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.Messenger = nil

		pss, err := NewPeerShardSender(args)
		assert.Equal(t, process.ErrNilMessenger, err)
		assert.True(t, check.IfNil(pss))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.Marshaller = nil

		pss, err := NewPeerShardSender(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(pss))
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.ShardCoordinator = nil

		pss, err := NewPeerShardSender(args)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
		assert.True(t, check.IfNil(pss))
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.TimeBetweenSends = time.Second - time.Nanosecond

		pss, err := NewPeerShardSender(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "TimeBetweenSends"))
		assert.True(t, check.IfNil(pss))
	})
	t.Run("invalid threshold between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.TimeThresholdBetweenSends = 0

		pss, err := NewPeerShardSender(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidThreshold))
		assert.True(t, strings.Contains(err.Error(), "TimeThresholdBetweenSends"))
		assert.True(t, check.IfNil(pss))
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.NodesCoordinator = nil

		pss, err := NewPeerShardSender(args)
		assert.True(t, errors.Is(err, heartbeat.ErrNilNodesCoordinator))
		assert.True(t, check.IfNil(pss))
	})
	t.Run("should work and validator should not broadcast", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		wasCalled := false
		args.Messenger = &p2pmocks.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				wasCalled = true
			},
		}
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, nil
			},
		}
		args.TimeBetweenSends = 2 * time.Second

		pss, _ := NewPeerShardSender(args)
		assert.False(t, check.IfNil(pss))

		time.Sleep(3 * time.Second)
		_ = pss.Close()
		assert.False(t, wasCalled)
	})
	t.Run("should work and broadcast", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		expectedShard := fmt.Sprintf("%d", args.ShardCoordinator.SelfId())
		numOfCalls := uint32(0)
		args.Messenger = &p2pmocks.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				shardInfo := &factory.PeerShard{}
				err := args.Marshaller.Unmarshal(shardInfo, buff)
				assert.Nil(t, err)
				assert.Equal(t, expectedShard, shardInfo.ShardId)
				assert.Equal(t, common.ConnectionTopic, topic)
				atomic.AddUint32(&numOfCalls, 1)
			},
		}
		args.TimeBetweenSends = 2 * time.Second

		pss, _ := NewPeerShardSender(args)
		assert.False(t, check.IfNil(pss))

		time.Sleep(3 * time.Second)
		_ = pss.Close()
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numOfCalls))
	})
}
