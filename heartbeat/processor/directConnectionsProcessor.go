package processor

import (
	"context"
	"fmt"
	"sync"
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
	NodesCoordinator          NodesCoordinator
}

type directConnectionsProcessor struct {
	messenger                 p2p.Messenger
	marshaller                marshal.Marshalizer
	shardCoordinator          sharding.Coordinator
	delayBetweenNotifications time.Duration
	cancel                    func()
	nodesCoordinator          NodesCoordinator

	mutConnectedPeersMap sync.Mutex
	connectedPeersMap    map[core.PeerID]struct{}
}

// Connected is called whenever a new peer is connected
func (dcp *directConnectionsProcessor) Connected(pid core.PeerID, _ string) {
	dcp.mutConnectedPeersMap.Lock()
	dcp.connectedPeersMap[pid] = struct{}{}
	dcp.mutConnectedPeersMap.Unlock()
}

// Disconnected does nothing
func (dcp *directConnectionsProcessor) Disconnected(_ core.PeerID) {
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
		connectedPeersMap:         make(map[core.PeerID]struct{}),
		nodesCoordinator:          args.NodesCoordinator,
	}

	err = args.Messenger.AddPeerEventsHandler(dcp)
	if err != nil {
		return nil, err
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
	if check.IfNil(args.NodesCoordinator) {
		return heartbeat.ErrNilNodesCoordinator
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
	newPeers := dcp.getNewPeersEmptyingTheMap()
	if dcp.isCurrentNodeValidator() {
		log.Debug("directConnectionsProcessor.sendMessageToNewConnections current node is validator, will not send messages to connected peers")
		return
	}

	dcp.notifyNewPeers(newPeers)
}

func (dcp *directConnectionsProcessor) isCurrentNodeValidator() bool {
	currentBLSKey := dcp.nodesCoordinator.GetOwnPublicKey()
	_, _, err := dcp.nodesCoordinator.GetValidatorWithPublicKey(currentBLSKey)

	return err == nil
}

func (dcp *directConnectionsProcessor) getNewPeersEmptyingTheMap() []core.PeerID {
	dcp.mutConnectedPeersMap.Lock()
	defer dcp.mutConnectedPeersMap.Unlock()

	result := make([]core.PeerID, 0, len(dcp.connectedPeersMap))
	for pid := range dcp.connectedPeersMap {
		result = append(result, pid)
	}

	dcp.connectedPeersMap = make(map[core.PeerID]struct{})

	return result
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
		log.Trace("directConnectionsProcessor.notifyNewPeers sending message", "pid", newPeer.Pretty())

		errNotCritical := dcp.messenger.SendToConnectedPeer(common.ConnectionTopic, shardValidatorInfoBuff, newPeer)
		if errNotCritical != nil {
			log.Trace("directConnectionsProcessor.notifyNewPeers", "pid", newPeer.Pretty(), "error", errNotCritical)
			continue
		}

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
