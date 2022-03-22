package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgConnectionsProcessor represents the arguments for the connections processor
type ArgConnectionsProcessor struct {
	Messenger                 p2p.Messenger
	Marshaller                marshal.Marshalizer
	ShardCoordinator          sharding.Coordinator
	DelayBetweenNotifications time.Duration
}

type connectionsProcessor struct {
	messenger                 p2p.Messenger
	marshaller                marshal.Marshalizer
	shardCoordinator          sharding.Coordinator
	delayBetweenNotifications time.Duration
	notifiedPeersMap          map[core.PeerID]struct{}
	cancel                    func()
}

// NewConnectionsProcessor creates a new instance of connectionsProcessor
func NewConnectionsProcessor(args ArgConnectionsProcessor) (*connectionsProcessor, error) {
	err := checkArgConnectionsProcessor(args)
	if err != nil {
		return nil, err
	}

	cp := &connectionsProcessor{
		messenger:                 args.Messenger,
		marshaller:                args.Marshaller,
		shardCoordinator:          args.ShardCoordinator,
		delayBetweenNotifications: args.DelayBetweenNotifications,
		notifiedPeersMap:          make(map[core.PeerID]struct{}),
	}

	var ctx context.Context
	ctx, cp.cancel = context.WithCancel(context.Background())

	go cp.startProcessLoop(ctx)

	return cp, nil
}

func checkArgConnectionsProcessor(args ArgConnectionsProcessor) error {
	if check.IfNil(args.Messenger) {
		return process.ErrNilMessenger
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if args.DelayBetweenNotifications < minDelayBetweenRequests {
		return fmt.Errorf("%w for DelayBetweenNotifications, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.DelayBetweenNotifications, minDelayBetweenRequests)
	}

	return nil
}

func (cp *connectionsProcessor) startProcessLoop(ctx context.Context) {
	timer := time.NewTimer(cp.delayBetweenNotifications)
	defer timer.Stop()

	for {
		timer.Reset(cp.delayBetweenNotifications)

		select {
		case <-timer.C:
			cp.sendMessageToNewConnections()
		case <-ctx.Done():
			log.Debug("closing connectionsProcessor go routine")
			return
		}
	}
}

func (cp *connectionsProcessor) sendMessageToNewConnections() {
	connectedPeers := cp.messenger.ConnectedPeers()
	newPeers := cp.computeNewPeers(connectedPeers)
	cp.notifyNewPeers(newPeers)
}

func (cp *connectionsProcessor) computeNewPeers(connectedPeers []core.PeerID) []core.PeerID {
	newPeers := make([]core.PeerID, 0)

	for _, connectedPeer := range connectedPeers {
		_, wasNotified := cp.notifiedPeersMap[connectedPeer]
		if !wasNotified {
			newPeers = append(newPeers, connectedPeer)
		}
	}

	return newPeers
}

func (cp *connectionsProcessor) notifyNewPeers(newPeers []core.PeerID) {
	cp.notifiedPeersMap = make(map[core.PeerID]struct{})

	shardValidatorInfo := message.ShardValidatorInfo{
		ShardId: cp.shardCoordinator.SelfId(),
	}

	shardValidatorInfoBuff, err := cp.marshaller.Marshal(shardValidatorInfo)
	if err != nil {
		return
	}

	for _, newPeer := range newPeers {
		errNotCritical := cp.messenger.SendToConnectedPeer(common.ConnectionTopic, shardValidatorInfoBuff, newPeer)
		if errNotCritical != nil {
			// todo replace with log.trace
			log.Info("connectionsProcessor.notifyNewPeers", "pid", newPeer.Pretty(), "error", errNotCritical)
			continue
		}

		cp.notifiedPeersMap[newPeer] = struct{}{}
	}
}

// Close triggers the closing of the internal goroutine
func (cp *connectionsProcessor) Close() error {
	log.Debug("closing connectionsProcessor...")
	cp.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (cp *connectionsProcessor) IsInterfaceNil() bool {
	return cp == nil
}
