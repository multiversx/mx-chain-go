package processor

import (
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createMockArgConnectionsProcessor() ArgConnectionsProcessor {
	return ArgConnectionsProcessor{
		Messenger:                 &p2pmocks.MessengerStub{},
		Marshaller:                &mock.MarshallerStub{},
		ShardCoordinator:          &mock.ShardCoordinatorMock{},
		DelayBetweenNotifications: time.Second,
	}
}

func TestNewConnectionsProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsProcessor()
		args.Messenger = nil

		cp, err := NewConnectionsProcessor(args)
		assert.Equal(t, process.ErrNilMessenger, err)
		assert.True(t, check.IfNil(cp))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsProcessor()
		args.Marshaller = nil

		cp, err := NewConnectionsProcessor(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(cp))
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsProcessor()
		args.ShardCoordinator = nil

		cp, err := NewConnectionsProcessor(args)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
		assert.True(t, check.IfNil(cp))
	})
	t.Run("invalid delay should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgConnectionsProcessor()
		args.DelayBetweenNotifications = time.Second - time.Nanosecond

		cp, err := NewConnectionsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "DelayBetweenNotifications"))
		assert.True(t, check.IfNil(cp))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cp, err := NewConnectionsProcessor(createMockArgConnectionsProcessor())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(cp))
	})
	t.Run("should work and process once", func(t *testing.T) {
		t.Parallel()

		providedConnectedPeers := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5", "pid6"}
		args := createMockArgConnectionsProcessor()
		expectedShard := args.ShardCoordinator.SelfId()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				shardValidatorInfo := message.ShardValidatorInfo{}
				err := args.Marshaller.Unmarshal(shardValidatorInfo, buff)
				assert.Nil(t, err)
				assert.Equal(t, expectedShard, shardValidatorInfo.ShardId)

				return nil
			},
		}
		args.Messenger = &p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return providedConnectedPeers
			},
		}
		args.DelayBetweenNotifications = 2 * time.Second

		cp, _ := NewConnectionsProcessor(args)
		assert.False(t, check.IfNil(cp))

		time.Sleep(3 * time.Second)
		_ = cp.Close()

		notifiedPeersSlice := make([]core.PeerID, 0)
		for peerInMap := range cp.notifiedPeersMap {
			notifiedPeersSlice = append(notifiedPeersSlice, peerInMap)
		}

		sort.Slice(notifiedPeersSlice, func(i, j int) bool {
			return notifiedPeersSlice[i] < notifiedPeersSlice[j]
		})
		assert.Equal(t, providedConnectedPeers, notifiedPeersSlice)
	})
}

func Test_connectionsProcessor_computeNewPeers(t *testing.T) {
	t.Parallel()

	t.Run("no peers connected", func(t *testing.T) {
		t.Parallel()

		cp, _ := NewConnectionsProcessor(createMockArgConnectionsProcessor())
		assert.False(t, check.IfNil(cp))
		_ = cp.Close() // avoid concurrency issues on notifiedPeersMap

		providedNotifiedPeersMap := make(map[core.PeerID]struct{})
		providedNotifiedPeersMap["pid1"] = struct{}{}
		providedNotifiedPeersMap["pid2"] = struct{}{}

		cp.notifiedPeersMap = providedNotifiedPeersMap

		newPeers := cp.computeNewPeers(nil)
		assert.Equal(t, 0, len(newPeers))
	})
	t.Run("some connected peers are new", func(t *testing.T) {
		t.Parallel()

		cp, _ := NewConnectionsProcessor(createMockArgConnectionsProcessor())
		assert.False(t, check.IfNil(cp))
		_ = cp.Close() // avoid concurrency issues on notifiedPeersMap

		providedNotifiedPeersMap := make(map[core.PeerID]struct{})
		providedNotifiedPeersMap["pid1"] = struct{}{}
		providedNotifiedPeersMap["pid2"] = struct{}{}

		cp.notifiedPeersMap = providedNotifiedPeersMap

		connectedPeers := []core.PeerID{"pid2", "pid3"}
		newPeers := cp.computeNewPeers(connectedPeers)

		assert.Equal(t, []core.PeerID{"pid3"}, newPeers)
	})
	t.Run("all connected peers are new", func(t *testing.T) {
		t.Parallel()

		cp, _ := NewConnectionsProcessor(createMockArgConnectionsProcessor())
		assert.False(t, check.IfNil(cp))

		connectedPeers := []core.PeerID{"pid3", "pid4"}
		newPeers := cp.computeNewPeers(connectedPeers)

		assert.Equal(t, connectedPeers, newPeers)
	})
}

func Test_connectionsProcessor_notifyNewPeers(t *testing.T) {
	t.Parallel()

	t.Run("marshal returns error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockArgConnectionsProcessor()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				wasCalled = true
				return nil
			},
		}
		args.Marshaller = &mock.MarshallerStub{
			MarshalHandler: func(obj interface{}) ([]byte, error) {
				return nil, errors.New("error")
			},
		}

		cp, _ := NewConnectionsProcessor(args)
		assert.False(t, check.IfNil(cp))

		cp.notifyNewPeers(nil)
		assert.False(t, wasCalled)
	})
	t.Run("no new peers", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockArgConnectionsProcessor()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				wasCalled = true
				return nil
			},
		}

		cp, _ := NewConnectionsProcessor(args)
		assert.False(t, check.IfNil(cp))

		cp.notifyNewPeers(nil)
		assert.False(t, wasCalled)
	})
	t.Run("send returns error", func(t *testing.T) {
		t.Parallel()

		providedPeer := core.PeerID("pid")
		args := createMockArgConnectionsProcessor()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				assert.Equal(t, common.ConnectionTopic, topic)
				assert.Equal(t, providedPeer, peerID)
				return errors.New("error")
			},
		}

		cp, _ := NewConnectionsProcessor(args)
		assert.False(t, check.IfNil(cp))
		_ = cp.Close() // avoid concurrency issues on notifiedPeersMap

		cp.notifyNewPeers([]core.PeerID{providedPeer})
		assert.Equal(t, 0, len(cp.notifiedPeersMap))
	})
	t.Run("send returns error only after 4th call", func(t *testing.T) {
		t.Parallel()

		providedConnectedPeers := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5", "pid6"}
		counter := 0
		args := createMockArgConnectionsProcessor()
		expectedShard := args.ShardCoordinator.SelfId()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				shardValidatorInfo := message.ShardValidatorInfo{}
				err := args.Marshaller.Unmarshal(shardValidatorInfo, buff)
				assert.Nil(t, err)
				assert.Equal(t, expectedShard, shardValidatorInfo.ShardId)

				counter++
				if counter > 4 {
					return errors.New("error")
				}

				return nil
			},
		}

		cp, _ := NewConnectionsProcessor(args)
		assert.False(t, check.IfNil(cp))
		_ = cp.Close() // avoid concurrency issues on notifiedPeersMap

		cp.notifyNewPeers(providedConnectedPeers)
		assert.Equal(t, 4, len(cp.notifiedPeersMap))
	})
}
