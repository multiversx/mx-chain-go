package processor

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

const testBaseIntraShardTopic = "intra"
const testBaseCrossShardTopic = "cross"

func createMockArgsDirectConnectionProcessor() ArgsDirectConnectionProcessor {
	return ArgsDirectConnectionProcessor{
		TimeToReadDirectConnections: time.Second,
		Messenger:                   &p2pmocks.MessengerStub{},
		PeerShardMapper:             &mock.PeerShardMapperStub{},
		ShardCoordinator:            testscommon.NewMultiShardsCoordinatorMock(3),
		BaseIntraShardTopic:         testBaseIntraShardTopic,
		BaseCrossShardTopic:         testBaseCrossShardTopic,
	}
}

func TestNewDirectConnectionProcessor(t *testing.T) {
	t.Parallel()

	t.Run("invalid TimeToReadDirectConnections should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsDirectConnectionProcessor()
		args.TimeToReadDirectConnections = time.Second - 1
		processor, err := NewDirectConnectionProcessor(args)

		assert.True(t, check.IfNil(processor))
		assert.ErrorIs(t, err, heartbeat.ErrInvalidValue)
		assert.ErrorContains(t, err, "TimeToReadDirectConnections")
	})
	t.Run("nil Messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsDirectConnectionProcessor()
		args.Messenger = nil
		processor, err := NewDirectConnectionProcessor(args)

		assert.True(t, check.IfNil(processor))
		assert.Equal(t, heartbeat.ErrNilMessenger, err)
	})
	t.Run("nil PeerShardMapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsDirectConnectionProcessor()
		args.PeerShardMapper = nil
		processor, err := NewDirectConnectionProcessor(args)

		assert.True(t, check.IfNil(processor))
		assert.Equal(t, heartbeat.ErrNilPeerShardMapper, err)
	})
	t.Run("nil ShardCoordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsDirectConnectionProcessor()
		args.ShardCoordinator = nil
		processor, err := NewDirectConnectionProcessor(args)

		assert.True(t, check.IfNil(processor))
		assert.Equal(t, heartbeat.ErrNilShardCoordinator, err)
	})
	t.Run("empty BaseIntraShardTopic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsDirectConnectionProcessor()
		args.BaseIntraShardTopic = ""
		processor, err := NewDirectConnectionProcessor(args)

		assert.True(t, check.IfNil(processor))
		assert.ErrorIs(t, err, heartbeat.ErrInvalidValue)
		assert.ErrorContains(t, err, "BaseIntraShardTopic")
	})
	t.Run("empty BaseCrossShardTopic should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsDirectConnectionProcessor()
		args.BaseCrossShardTopic = ""
		processor, err := NewDirectConnectionProcessor(args)

		assert.True(t, check.IfNil(processor))
		assert.ErrorIs(t, err, heartbeat.ErrInvalidValue)
		assert.ErrorContains(t, err, "BaseCrossShardTopic")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsDirectConnectionProcessor()
		processor, err := NewDirectConnectionProcessor(args)

		assert.False(t, check.IfNil(processor))
		assert.Nil(t, err)
	})
}

func TestDirectConnectionProcessor_ProcessingShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("node is in shard 1, process should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &testscommon.ShardsCoordinatorMock{
			CurrentShard: 1,
			NoShards:     3,
		}
		directConnections := map[string][]core.PeerID{
			testBaseIntraShardTopic + "_1":      {"pid1", "pid2", "pid3"},
			testBaseCrossShardTopic + "_0_1":    {"pid1", "pid4", "pid2", "pid3", "pid5"}, // only pid4 & pid5 are pure cross shard
			testBaseCrossShardTopic + "_1_2":    {"pid6", "pid2", "pid3", "pid7", "pid1"}, // only pid6 & pid7 are pure cross shard
			testBaseCrossShardTopic + "_1_META": {"pid8", "pid9", "pid1", "pid2", "pid3"}, // only pid8 & pid9 are pure cross shard
		}
		args := createMockArgsDirectConnectionProcessor()
		args.ShardCoordinator = shardCoordinator
		args.Messenger = &p2pmocks.MessengerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
				peers, found := directConnections[topic]
				if !found {
					assert.Fail(t, fmt.Sprintf("connected peers called on an unsupported topic: %s", topic))
				}

				return peers
			},
		}
		mutResultMap := sync.RWMutex{}
		resultMap := make(map[uint32][]core.PeerID)
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				mutResultMap.Lock()
				resultMap[shardId] = append(resultMap[shardId], pid)
				mutResultMap.Unlock()
			},
		}

		processor, _ := NewDirectConnectionProcessor(args)
		time.Sleep(time.Second + time.Millisecond*500) // wait for one loop
		_ = processor.Close()

		mutResultMap.RLock()
		defer mutResultMap.RUnlock()

		assert.Equal(t, 4, len(resultMap))
		assert.Equal(t, []core.PeerID{"pid4", "pid5"}, resultMap[0])
		assert.Equal(t, []core.PeerID{"pid1", "pid2", "pid3"}, resultMap[1])
		assert.Equal(t, []core.PeerID{"pid6", "pid7"}, resultMap[2])
		assert.Equal(t, []core.PeerID{"pid8", "pid9"}, resultMap[common.MetachainShardId])
	})
	t.Run("node is in shard meta, process should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := &testscommon.ShardsCoordinatorMock{
			CurrentShard: common.MetachainShardId,
			NoShards:     3,
		}
		directConnections := map[string][]core.PeerID{
			testBaseIntraShardTopic + "_META":   {"pid1", "pid2", "pid3"},
			testBaseCrossShardTopic + "_0_META": {"pid1", "pid4", "pid2", "pid3", "pid5"}, // only pid4 & pid5 are pure cross shard
			testBaseCrossShardTopic + "_1_META": {"pid6", "pid2", "pid3", "pid7", "pid1"}, // only pid6 & pid7 are pure cross shard
			testBaseCrossShardTopic + "_2_META": {"pid8", "pid9", "pid1", "pid2", "pid3"}, // only pid8 & pid9 are pure cross shard
		}
		args := createMockArgsDirectConnectionProcessor()
		args.ShardCoordinator = shardCoordinator
		args.Messenger = &p2pmocks.MessengerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
				peers, found := directConnections[topic]
				if !found {
					assert.Fail(t, fmt.Sprintf("connected peers called on an unsupported topic: %s", topic))
				}

				return peers
			},
		}
		mutResultMap := sync.RWMutex{}
		resultMap := make(map[uint32][]core.PeerID)
		args.PeerShardMapper = &mock.PeerShardMapperStub{
			PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				mutResultMap.Lock()
				resultMap[shardId] = append(resultMap[shardId], pid)
				mutResultMap.Unlock()
			},
		}

		processor, _ := NewDirectConnectionProcessor(args)
		time.Sleep(time.Second + time.Millisecond*500) // wait for one loop
		_ = processor.Close()

		mutResultMap.RLock()
		defer mutResultMap.RUnlock()

		assert.Equal(t, 4, len(resultMap))
		assert.Equal(t, []core.PeerID{"pid4", "pid5"}, resultMap[0])
		assert.Equal(t, []core.PeerID{"pid6", "pid7"}, resultMap[1])
		assert.Equal(t, []core.PeerID{"pid8", "pid9"}, resultMap[2])
		assert.Equal(t, []core.PeerID{"pid1", "pid2", "pid3"}, resultMap[common.MetachainShardId])
	})
}

func TestDirectConnectionProcessor_Close(t *testing.T) {
	t.Parallel()

	shardCoordinator := &testscommon.ShardsCoordinatorMock{
		CurrentShard: 1,
		NoShards:     3,
	}
	directConnections := map[string][]core.PeerID{
		testBaseIntraShardTopic + "_1": {"pid1"},
	}
	args := createMockArgsDirectConnectionProcessor()
	args.ShardCoordinator = shardCoordinator
	args.Messenger = &p2pmocks.MessengerStub{
		ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
			return directConnections[topic]
		},
	}
	mutResultMap := sync.RWMutex{}
	resultMap := make(map[uint32][]core.PeerID)
	args.PeerShardMapper = &mock.PeerShardMapperStub{
		PutPeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
			mutResultMap.Lock()
			resultMap[shardId] = append(resultMap[shardId], pid)
			mutResultMap.Unlock()
		},
	}

	processor, _ := NewDirectConnectionProcessor(args)
	numLoops := 3
	time.Sleep(time.Second*time.Duration(numLoops) + time.Millisecond*500) // wait for numLoops + some extra to eliminate edge-cases
	_ = processor.Close()

	mutResultMap.RLock()
	assert.Equal(t, 1, len(resultMap))
	assert.Equal(t, numLoops, len(resultMap[1])) // PutPeerIdShardIdCalled was called for numLoops times
	mutResultMap.RUnlock()

	time.Sleep(time.Second * 2) // wait for two loops but the processor is closed, should not call PutPeerIdShardIdCalled again

	mutResultMap.RLock()
	assert.Equal(t, 1, len(resultMap))
	assert.Equal(t, numLoops, len(resultMap[1])) // PutPeerIdShardIdCalled was called for numLoops times
	mutResultMap.RUnlock()
}
