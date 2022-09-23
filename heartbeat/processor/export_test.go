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
		notifiedPeersMap:          make(map[core.PeerID]struct{}),
		nodesCoordinator:          args.NodesCoordinator,
	}

	return dcp, nil
}
