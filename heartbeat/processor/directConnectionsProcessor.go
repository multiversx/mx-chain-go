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

// ArgDirectConnectionsProcessor represents the arguments for the direct connections processor
type ArgDirectConnectionsProcessor struct {
	Messenger                 p2p.Messenger
	Marshaller                marshal.Marshalizer
	ShardCoordinator          sharding.Coordinator
	DelayBetweenNotifications time.Duration
}

type directConnectionsProcessor struct {
	messenger                 p2p.Messenger
	marshaller                marshal.Marshalizer
	shardCoordinator          sharding.Coordinator
	delayBetweenNotifications time.Duration
	notifiedPeersMap          map[core.PeerID]struct{}
	cancel                    func()
}

// NewDirectConnectionsProcessor creates a new instance of directConnectionsProcessor
func NewDirectConnectionsProcessor(args ArgDirectConnectionsProcessor) (*directConnectionsProcessor, error) {
	err := checkArgDirectConnectionsProcessor(args)
	if err != nil {
		return nil, err
	}

	dcp := &directConnectionsProcessor{
		messenger:                 args.Messenger,
		marshaller:                args.Marshaller,
		shardCoordinator:          args.ShardCoordinator,
		delayBetweenNotifications: args.DelayBetweenNotifications,
		notifiedPeersMap:          make(map[core.PeerID]struct{}),
	}

	var ctx context.Context
	ctx, dcp.cancel = context.WithCancel(context.Background())

	go dcp.startProcessLoop(ctx)

	return dcp, nil
}

func checkArgDirectConnectionsProcessor(args ArgDirectConnectionsProcessor) error {
	if check.IfNil(args.Messenger) {
		return process.ErrNilMessenger
	}
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
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

func (dcp *directConnectionsProcessor) startProcessLoop(ctx context.Context) {
	timer := time.NewTimer(dcp.delayBetweenNotifications)
	defer timer.Stop()

	for {
		timer.Reset(dcp.delayBetweenNotifications)

		select {
		case <-timer.C:
			dcp.sendMessageToNewConnections()
		case <-ctx.Done():
			log.Debug("closing directConnectionsProcessor go routine")
			return
		}
	}
}

func (dcp *directConnectionsProcessor) sendMessageToNewConnections() {
	connectedPeers := dcp.messenger.ConnectedPeers()
	newPeers := dcp.computeNewPeers(connectedPeers)
	dcp.notifyNewPeers(newPeers)
	dcp.recreateNotifiedPeers(connectedPeers)
}

func (dcp *directConnectionsProcessor) computeNewPeers(connectedPeers []core.PeerID) []core.PeerID {
	newPeers := make([]core.PeerID, 0)

	for _, connectedPeer := range connectedPeers {
		_, wasNotified := dcp.notifiedPeersMap[connectedPeer]
		if !wasNotified {
			newPeers = append(newPeers, connectedPeer)
		}
	}

	return newPeers
}

func (dcp *directConnectionsProcessor) notifyNewPeers(newPeers []core.PeerID) {
	shardValidatorInfo := &message.DirectConnectionInfo{
		ShardId: fmt.Sprintf("%d", dcp.shardCoordinator.SelfId()),
	}

	shardValidatorInfoBuff, err := dcp.marshaller.Marshal(shardValidatorInfo)
	if err != nil {
		return
	}

	for _, newPeer := range newPeers {
		errNotCritical := dcp.messenger.SendToConnectedPeer(common.ConnectionTopic, shardValidatorInfoBuff, newPeer)
		if errNotCritical != nil {
			log.Trace("directConnectionsProcessor.notifyNewPeers", "pid", newPeer.Pretty(), "error", errNotCritical)
			continue
		}

		dcp.notifiedPeersMap[newPeer] = struct{}{}
	}
}

func (dcp *directConnectionsProcessor) recreateNotifiedPeers(connectedPeers []core.PeerID) {
	dcp.notifiedPeersMap = make(map[core.PeerID]struct{})

	for _, peer := range connectedPeers {
		dcp.notifiedPeersMap[peer] = struct{}{}
	}
}

// Close triggers the closing of the internal goroutine
func (dcp *directConnectionsProcessor) Close() error {
	log.Debug("closing directConnectionsProcessor...")
	dcp.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dcp *directConnectionsProcessor) IsInterfaceNil() bool {
	return dcp == nil
}
