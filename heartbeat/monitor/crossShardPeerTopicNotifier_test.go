package monitor

import (
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgsCrossShardPeerTopicNotifier() ArgsCrossShardPeerTopicNotifier {
	return ArgsCrossShardPeerTopicNotifier{
		ShardCoordinator: &testscommon.ShardsCoordinatorMock{
			NoShards:     3,
			CurrentShard: 1,
		},
		PeerShardMapper: &mock.PeerShardMapperStub{},
	}
}

func TestNewCrossShardPeerTopicNotifier(t *testing.T) {
	t.Parallel()

	t.Run("nil sharding coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.ShardCoordinator = nil

		notifier, err := NewCrossShardPeerTopicNotifier(args)
		assert.True(t, check.IfNil(notifier))
		assert.Equal(t, heartbeat.ErrNilShardCoordinator, err)
	})
	t.Run("nil peer shard mapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = nil

		notifier, err := NewCrossShardPeerTopicNotifier(args)
		assert.True(t, check.IfNil(notifier))
		assert.Equal(t, heartbeat.ErrNilPeerShardMapper, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()

		notifier, err := NewCrossShardPeerTopicNotifier(args)
		assert.False(t, check.IfNil(notifier))
		assert.Nil(t, err)
	})
}

func TestCrossShardPeerTopicNotifier_NewPeerFound(t *testing.T) {
	t.Parallel()

	testTopic := "test"
	t.Run("global topic should not notice", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Fail(t, "should have not called PutPeerIdShardId")
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		notifier.NewPeerFound("pid", "random topic")
	})
	t.Run("intra-shard topic should not notice", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Fail(t, "should have not called PutPeerIdShardId")
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + core.CommunicationIdentifierBetweenShards(0, 0)
		notifier.NewPeerFound("pid", topic)
	})
	t.Run("cross-shard topic but not relevant to current node should not notice", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Fail(t, "should have not called PutPeerIdShardId")
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + core.CommunicationIdentifierBetweenShards(0, 2)
		notifier.NewPeerFound("pid", topic)
	})
	t.Run("first shard ID is a NaN should not notice", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Fail(t, "should have not called PutPeerIdShardId")
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + "_NaN_1"
		notifier.NewPeerFound("pid", topic)
	})
	t.Run("second shard ID is a NaN should not notice", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Fail(t, "should have not called PutPeerIdShardId")
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + "_1_NaN"
		notifier.NewPeerFound("pid", topic)
	})
	t.Run("second shard ID is a negative value should not notice", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Fail(t, "should have not called PutPeerIdShardId")
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + "_1_-1"
		notifier.NewPeerFound("pid", topic)
	})
	t.Run("second shard ID is an out of range value should not notice", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Fail(t, "should have not called PutPeerIdShardId")
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + "_1_4"
		notifier.NewPeerFound("pid", topic)
	})
	t.Run("same shard IDs should not notice", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Fail(t, "should have not called PutPeerIdShardId")
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + "_0_0"
		notifier.NewPeerFound("pid", topic)
	})
	t.Run("cross-shard between 0 and 1 should notice", func(t *testing.T) {
		t.Parallel()

		expectedPid := core.PeerID("pid")
		notifiedShardID := uint32(math.MaxUint32)
		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Equal(t, pid, expectedPid)
				notifiedShardID = shardId
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + core.CommunicationIdentifierBetweenShards(0, 1)
		notifier.NewPeerFound("pid", topic)
		assert.Equal(t, uint32(0), notifiedShardID)
	})
	t.Run("cross-shard between 1 and 2 should notice", func(t *testing.T) {
		t.Parallel()

		expectedPid := core.PeerID("pid")
		notifiedShardID := uint32(math.MaxUint32)
		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Equal(t, pid, expectedPid)
				notifiedShardID = shardId
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + core.CommunicationIdentifierBetweenShards(1, 2)
		notifier.NewPeerFound("pid", topic)
		assert.Equal(t, uint32(2), notifiedShardID)
	})
	t.Run("cross-shard between 1 and META should notice", func(t *testing.T) {
		t.Parallel()

		expectedPid := core.PeerID("pid")
		notifiedShardID := uint32(math.MaxUint32)
		args := createMockArgsCrossShardPeerTopicNotifier()
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Equal(t, pid, expectedPid)
				notifiedShardID = shardId
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + core.CommunicationIdentifierBetweenShards(1, common.MetachainShardId)
		notifier.NewPeerFound("pid", topic)
		assert.Equal(t, common.MetachainShardId, notifiedShardID)
	})
	t.Run("cross-shard between META and 1 should notice", func(t *testing.T) {
		t.Parallel()

		expectedPid := core.PeerID("pid")
		notifiedShardID := uint32(math.MaxUint32)
		args := createMockArgsCrossShardPeerTopicNotifier()
		args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
			NoShards:     3,
			CurrentShard: common.MetachainShardId,
		}
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Equal(t, pid, expectedPid)
				notifiedShardID = shardId
			},
		}

		notifier, _ := NewCrossShardPeerTopicNotifier(args)
		topic := testTopic + core.CommunicationIdentifierBetweenShards(common.MetachainShardId, 1)
		notifier.NewPeerFound("pid", topic)
		assert.Equal(t, uint32(1), notifiedShardID)
	})
}

func BenchmarkCrossShardPeerTopicNotifier_NewPeerFound(b *testing.B) {
	args := createMockArgsCrossShardPeerTopicNotifier()
	notifier, _ := NewCrossShardPeerTopicNotifier(args)

	for i := 0; i < b.N; i++ {
		switch i % 6 {
		case 0:
			notifier.NewPeerFound("pid", "global")
		case 2:
			notifier.NewPeerFound("pid", "intrashard_1")
		case 3:
			notifier.NewPeerFound("pid", "crossshard_1_2")
		case 4:
			notifier.NewPeerFound("pid", "crossshard_1_META")
		case 5:
			notifier.NewPeerFound("pid", "crossshard_META_1")
		case 6:
			notifier.NewPeerFound("pid", "crossshard_2_META")
		}
	}
}
