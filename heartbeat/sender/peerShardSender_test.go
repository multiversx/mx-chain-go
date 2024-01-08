package sender

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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
		MainMessenger:             &p2pmocks.MessengerStub{},
		FullArchiveMessenger:      &p2pmocks.MessengerStub{},
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

	t.Run("nil main messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.MainMessenger = nil

		pss, err := NewPeerShardSender(args)
		assert.True(t, errors.Is(err, process.ErrNilMessenger))
		assert.Nil(t, pss)
	})
	t.Run("nil full archive messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.FullArchiveMessenger = nil

		pss, err := NewPeerShardSender(args)
		assert.True(t, errors.Is(err, process.ErrNilMessenger))
		assert.Nil(t, pss)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.Marshaller = nil

		pss, err := NewPeerShardSender(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.Nil(t, pss)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.ShardCoordinator = nil

		pss, err := NewPeerShardSender(args)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
		assert.Nil(t, pss)
	})
	t.Run("invalid time between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.TimeBetweenSends = time.Second - time.Nanosecond

		pss, err := NewPeerShardSender(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "TimeBetweenSends"))
		assert.Nil(t, pss)
	})
	t.Run("invalid threshold between sends should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.TimeThresholdBetweenSends = 0

		pss, err := NewPeerShardSender(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidThreshold))
		assert.True(t, strings.Contains(err.Error(), "TimeThresholdBetweenSends"))
		assert.Nil(t, pss)
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		args.NodesCoordinator = nil

		pss, err := NewPeerShardSender(args)
		assert.True(t, errors.Is(err, heartbeat.ErrNilNodesCoordinator))
		assert.Nil(t, pss)
	})
	t.Run("should work and validator should not broadcast", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		wasCalled := false
		args.MainMessenger = &p2pmocks.MessengerStub{
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
		assert.NotNil(t, pss)

		time.Sleep(3 * time.Second)
		_ = pss.Close()
		assert.False(t, wasCalled)
	})
	t.Run("should work and broadcast", func(t *testing.T) {
		t.Parallel()

		args := createMockArgPeerShardSender()
		expectedShard := fmt.Sprintf("%d", args.ShardCoordinator.SelfId())
		numOfCalls := uint32(0)
		args.MainMessenger = &p2pmocks.MessengerStub{
			BroadcastCalled: func(topic string, buff []byte) {
				shardInfo := &factory.PeerShard{}
				err := args.Marshaller.Unmarshal(shardInfo, buff)
				assert.Nil(t, err)
				assert.Equal(t, expectedShard, shardInfo.ShardId)
				assert.Equal(t, common.ConnectionTopic, topic)
				atomic.AddUint32(&numOfCalls, 1)
			},
		}
		args.FullArchiveMessenger = &p2pmocks.MessengerStub{
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
		assert.NotNil(t, pss)

		time.Sleep(3 * time.Second)
		_ = pss.Close()
		assert.Equal(t, uint32(2), atomic.LoadUint32(&numOfCalls)) // one call for each messenger
	})
}

func TestPeerShardSender_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var pss *peerShardSender
	assert.True(t, pss.IsInterfaceNil())

	pss, _ = NewPeerShardSender(createMockArgPeerShardSender())
	assert.False(t, pss.IsInterfaceNil())
}
