package processor

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createMockArgCrossShardStatusProcessor() ArgCrossShardStatusProcessor {
	return ArgCrossShardStatusProcessor{
		Messenger:            &p2pmocks.MessengerStub{},
		PeerShardMapper:      &p2pmocks.NetworkShardingCollectorStub{},
		ShardCoordinator:     &mock.ShardCoordinatorStub{},
		DelayBetweenRequests: time.Second,
	}
}

func TestNewCrossShardStatusProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgCrossShardStatusProcessor()
		args.Messenger = nil

		processor, err := NewCrossShardStatusProcessor(args)
		assert.True(t, check.IfNil(processor))
		assert.Equal(t, process.ErrNilMessenger, err)
	})
	t.Run("nil peer shard mapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgCrossShardStatusProcessor()
		args.PeerShardMapper = nil

		processor, err := NewCrossShardStatusProcessor(args)
		assert.True(t, check.IfNil(processor))
		assert.Equal(t, process.ErrNilPeerShardMapper, err)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgCrossShardStatusProcessor()
		args.ShardCoordinator = nil

		processor, err := NewCrossShardStatusProcessor(args)
		assert.True(t, check.IfNil(processor))
		assert.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("invalid delay between requests should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgCrossShardStatusProcessor()
		args.DelayBetweenRequests = time.Second - time.Nanosecond

		processor, err := NewCrossShardStatusProcessor(args)
		assert.True(t, check.IfNil(processor))
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "DelayBetweenRequests"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedSuffix := "test"
		expectedNumberOfShards := uint32(1)
		args := createMockArgCrossShardStatusProcessor()
		args.ShardCoordinator = &mock.ShardCoordinatorStub{
			NumberOfShardsCalled: func() uint32 {
				return expectedNumberOfShards
			},
			CommunicationIdentifierCalled: func(destShardID uint32) string {
				return expectedSuffix
			},
		}

		providedPid := core.PeerID("provided pid")
		args.Messenger = &p2pmocks.MessengerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
				return []core.PeerID{providedPid}
			},
		}

		args.PeerShardMapper = &p2pmocks.NetworkShardingCollectorStub{
			UpdatePeerIdShardIdCalled: func(pid core.PeerID, shardId uint32) {
				assert.Equal(t, providedPid, pid)
			},
		}

		processor, err := NewCrossShardStatusProcessor(args)
		assert.False(t, check.IfNil(processor))
		assert.Nil(t, err)

		// for coverage, to make sure a loop is finished
		time.Sleep(args.DelayBetweenRequests * 2)

		// close the internal go routine
		err = processor.Close()
		assert.Nil(t, err)

		topicsMap := processor.computeTopicsMap()
		assert.Equal(t, expectedNumberOfShards, uint32(len(topicsMap)))

		metaTopic, ok := topicsMap[core.MetachainShardId]
		assert.True(t, ok)
		assert.Equal(t, factory.TransactionTopic+expectedSuffix, metaTopic)

		delete(topicsMap, core.MetachainShardId)

		expectedTopic := factory.TransactionTopic + expectedSuffix
		for _, shardTopic := range topicsMap {
			assert.Equal(t, expectedTopic, shardTopic)
		}
	})
}
