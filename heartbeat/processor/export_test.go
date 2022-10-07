package processor

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// NewDirectConnectionsProcessorNoGoRoutine creates a new instance of directConnectionsProcessor but does not start the goroutine
func NewDirectConnectionsProcessorNoGoRoutine(args ArgDirectConnectionsProcessor) (*directConnectionsProcessor, error) {
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

	return dcp, nil
}

func (dcp *directConnectionsProcessor) getNewPeers() []core.PeerID {
	dcp.mutConnectedPeersMap.Lock()
	defer dcp.mutConnectedPeersMap.Unlock()

	result := make([]core.PeerID, 0, len(dcp.connectedPeersMap))
	for pid := range dcp.connectedPeersMap {
		result = append(result, pid)
	}

	return result
}
