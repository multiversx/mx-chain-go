package processor

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	p2pFactory "github.com/ElrondNetwork/elrond-go/p2p/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

func createMockArgDirectConnectionsProcessor() ArgDirectConnectionsProcessor {
	return ArgDirectConnectionsProcessor{
		Messenger:                 &p2pmocks.MessengerStub{},
		Marshaller:                &marshal.GogoProtoMarshalizer{},
		ShardCoordinator:          &testscommon.ShardsCoordinatorMock{},
		DelayBetweenNotifications: time.Second,
		NodesCoordinator: &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, errors.New("the node is an observer")
			},
		},
	}
}

func TestNewDirectConnectionsProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectConnectionsProcessor()
		args.Messenger = nil

		cp, err := NewDirectConnectionsProcessor(args)
		assert.Equal(t, process.ErrNilMessenger, err)
		assert.True(t, check.IfNil(cp))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectConnectionsProcessor()
		args.Marshaller = nil

		cp, err := NewDirectConnectionsProcessor(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(cp))
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectConnectionsProcessor()
		args.ShardCoordinator = nil

		cp, err := NewDirectConnectionsProcessor(args)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
		assert.True(t, check.IfNil(cp))
	})
	t.Run("invalid delay should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectConnectionsProcessor()
		args.DelayBetweenNotifications = time.Second - time.Nanosecond

		cp, err := NewDirectConnectionsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "DelayBetweenNotifications"))
		assert.True(t, check.IfNil(cp))
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDirectConnectionsProcessor()
		args.NodesCoordinator = nil

		cp, err := NewDirectConnectionsProcessor(args)
		assert.True(t, errors.Is(err, heartbeat.ErrNilNodesCoordinator))
		assert.True(t, check.IfNil(cp))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cp, err := NewDirectConnectionsProcessor(createMockArgDirectConnectionsProcessor())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(cp))
	})
	t.Run("should work and process once", func(t *testing.T) {
		t.Parallel()

		providedConnectedPeers := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5", "pid6"}
		notifiedPeers := make([]core.PeerID, 0)
		var mutNotifiedPeers sync.RWMutex
		args := createMockArgDirectConnectionsProcessor()
		expectedShard := fmt.Sprintf("%d", args.ShardCoordinator.SelfId())
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				mutNotifiedPeers.Lock()
				defer mutNotifiedPeers.Unlock()

				shardValidatorInfo := &p2pFactory.DirectConnectionInfo{}
				err := args.Marshaller.Unmarshal(shardValidatorInfo, buff)
				assert.Nil(t, err)
				assert.Equal(t, expectedShard, shardValidatorInfo.ShardId)

				notifiedPeers = append(notifiedPeers, peerID)
				return nil
			},
			ConnectedPeersCalled: func() []core.PeerID {
				return providedConnectedPeers
			},
		}
		args.DelayBetweenNotifications = 2 * time.Second

		cp, _ := NewDirectConnectionsProcessor(args)
		assert.False(t, check.IfNil(cp))

		time.Sleep(3 * time.Second)
		_ = cp.Close()

		mutNotifiedPeers.Lock()
		defer mutNotifiedPeers.Unlock()

		sort.Slice(notifiedPeers, func(i, j int) bool {
			return notifiedPeers[i] < notifiedPeers[j]
		})
		assert.Equal(t, providedConnectedPeers, notifiedPeers)
	})
}

func Test_directConnectionsProcessor_computeNewPeers(t *testing.T) {
	t.Parallel()

	t.Run("no peers connected", func(t *testing.T) {
		t.Parallel()

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(createMockArgDirectConnectionsProcessor())
		assert.False(t, check.IfNil(cp))

		providedNotifiedPeersMap := make(map[core.PeerID]struct{})
		providedNotifiedPeersMap["pid1"] = struct{}{}
		providedNotifiedPeersMap["pid2"] = struct{}{}

		cp.notifiedPeersMap = providedNotifiedPeersMap

		newPeers := cp.computeNewPeers(nil)
		assert.Equal(t, 0, len(newPeers))
	})
	t.Run("some connected peers are new", func(t *testing.T) {
		t.Parallel()

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(createMockArgDirectConnectionsProcessor())
		assert.False(t, check.IfNil(cp))

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

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(createMockArgDirectConnectionsProcessor())
		assert.False(t, check.IfNil(cp))

		connectedPeers := []core.PeerID{"pid3", "pid4"}
		newPeers := cp.computeNewPeers(connectedPeers)

		assert.Equal(t, connectedPeers, newPeers)
	})
}

func Test_directConnectionsProcessor_notifyNewPeers(t *testing.T) {
	t.Parallel()

	t.Run("marshal returns error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockArgDirectConnectionsProcessor()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				wasCalled = true
				return nil
			},
		}
		args.Marshaller = &testscommon.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, errors.New("error")
			},
		}

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(args)
		assert.False(t, check.IfNil(cp))

		cp.notifyNewPeers(nil)
		assert.False(t, wasCalled)
	})
	t.Run("no new peers", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockArgDirectConnectionsProcessor()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				wasCalled = true
				return nil
			},
		}

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(args)
		assert.False(t, check.IfNil(cp))

		cp.notifyNewPeers(nil)
		assert.False(t, wasCalled)
	})
	t.Run("send returns error", func(t *testing.T) {
		t.Parallel()

		providedPeer := core.PeerID("pid")
		args := createMockArgDirectConnectionsProcessor()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				assert.Equal(t, common.ConnectionTopic, topic)
				assert.Equal(t, providedPeer, peerID)
				return errors.New("error")
			},
		}

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(args)
		assert.False(t, check.IfNil(cp))

		cp.notifyNewPeers([]core.PeerID{providedPeer})
		assert.Equal(t, 0, len(cp.notifiedPeersMap))
	})
	t.Run("send returns error only after 4th call", func(t *testing.T) {
		t.Parallel()

		providedConnectedPeers := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5", "pid6"}
		counter := 0
		args := createMockArgDirectConnectionsProcessor()
		expectedShard := fmt.Sprintf("%d", args.ShardCoordinator.SelfId())
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				shardValidatorInfo := &p2pFactory.DirectConnectionInfo{}
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

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(args)
		assert.False(t, check.IfNil(cp))

		cp.notifyNewPeers(providedConnectedPeers)
		assert.Equal(t, 4, len(cp.notifiedPeersMap))
	})
}

func TestDirectConnectionsProcessor_sendMessageToNewConnections(t *testing.T) {
	t.Parallel()

	t.Run("current node is validator, should not send messages", func(t *testing.T) {
		t.Parallel()

		sentToPeers := make(map[core.PeerID]int)
		args := createMockArgDirectConnectionsProcessor()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				sentToPeers[peerID]++
				return nil
			},
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"pid1", "pid2"}
			},
		}
		args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
				return nil, 0, nil
			},
		}

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(args)

		numSends := 10
		for i := 0; i < numSends; i++ {
			cp.sendMessageToNewConnections()
			assert.Equal(t, 0, len(sentToPeers))
		}
	})
	t.Run("same peers should send only once", func(t *testing.T) {
		t.Parallel()

		sentToPeers := make(map[core.PeerID]int)
		args := createMockArgDirectConnectionsProcessor()
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				sentToPeers[peerID]++
				return nil
			},
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{"pid1", "pid2"}
			},
		}

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(args)

		numSends := 10
		for i := 0; i < numSends; i++ {
			cp.sendMessageToNewConnections()
			assert.Equal(t, 2, len(sentToPeers))
			assert.Equal(t, 1, sentToPeers["pid1"])
			assert.Equal(t, 1, sentToPeers["pid2"])
		}
	})
	t.Run("same peers should send only once, new peer should also send once", func(t *testing.T) {
		t.Parallel()

		sentToPeers := make(map[core.PeerID]int)
		args := createMockArgDirectConnectionsProcessor()
		connectedPeers := []core.PeerID{"pid1", "pid2"}
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				sentToPeers[peerID]++
				return nil
			},
			ConnectedPeersCalled: func() []core.PeerID {
				return connectedPeers
			},
		}

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(args)

		numSends := 10
		for i := 0; i < numSends; i++ {
			cp.sendMessageToNewConnections()
			assert.Equal(t, 2, len(sentToPeers))
			assert.Equal(t, 1, sentToPeers["pid1"])
			assert.Equal(t, 1, sentToPeers["pid2"])
		}

		connectedPeers = append(connectedPeers, "pid3")
		for i := 0; i < numSends; i++ {
			cp.sendMessageToNewConnections()
			assert.Equal(t, 3, len(sentToPeers))
			assert.Equal(t, 1, sentToPeers["pid1"])
			assert.Equal(t, 1, sentToPeers["pid2"])
			assert.Equal(t, 1, sentToPeers["pid3"])
		}
	})
	t.Run("same peers should send only once, old peer should be removed from map and resend when re-connecting", func(t *testing.T) {
		t.Parallel()

		sentToPeers := make(map[core.PeerID]int)
		args := createMockArgDirectConnectionsProcessor()
		connectedPeers := []core.PeerID{"pid1", "pid2"}
		args.Messenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				sentToPeers[peerID]++
				return nil
			},
			ConnectedPeersCalled: func() []core.PeerID {
				return connectedPeers
			},
		}

		cp, _ := NewDirectConnectionsProcessorNoGoRoutine(args)

		numSends := 10
		for i := 0; i < numSends; i++ {
			cp.sendMessageToNewConnections()
			assert.Equal(t, 2, len(sentToPeers))
			assert.Equal(t, 1, sentToPeers["pid1"])
			assert.Equal(t, 1, sentToPeers["pid2"])
		}

		// pid2 got disconnected
		connectedPeers = []core.PeerID{"pid1"}
		for i := 0; i < numSends; i++ {
			cp.sendMessageToNewConnections()
			assert.Equal(t, 2, len(sentToPeers))
			assert.Equal(t, 1, sentToPeers["pid1"])
			assert.Equal(t, 1, sentToPeers["pid2"])

			assert.Equal(t, 1, len(cp.notifiedPeersMap))
		}

		connectedPeers = append(connectedPeers, "pid2")
		for i := 0; i < numSends; i++ {
			cp.sendMessageToNewConnections()
			assert.Equal(t, 2, len(sentToPeers))
			assert.Equal(t, 1, sentToPeers["pid1"])
			assert.Equal(t, 2, sentToPeers["pid2"]) // one before disconnection, one after reconnection

			assert.Equal(t, 2, len(cp.notifiedPeersMap))
		}
	})
}
