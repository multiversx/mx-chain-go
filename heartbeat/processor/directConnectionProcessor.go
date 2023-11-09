package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/sharding"
)

const minTimeToReadDirectConnections = time.Second

// ArgsDirectConnectionProcessor is the argument DTO for the NewDirectConnectionProcessor constructor function
type ArgsDirectConnectionProcessor struct {
	TimeToReadDirectConnections time.Duration
	Messenger                   heartbeat.P2PMessenger
	PeerShardMapper             heartbeat.PeerShardMapper
	ShardCoordinator            sharding.Coordinator
	BaseIntraShardTopic         string
	BaseCrossShardTopic         string
}

type directConnectionProcessor struct {
	messenger                   heartbeat.P2PMessenger
	peerShardMapper             heartbeat.PeerShardMapper
	shardCoordinator            sharding.Coordinator
	timeToReadDirectConnections time.Duration
	cancel                      func()
	intraShardTopic             string
	baseCrossShardTopic         string
}

// NewDirectConnectionProcessor will create a new direct connection processor instance
func NewDirectConnectionProcessor(args ArgsDirectConnectionProcessor) (*directConnectionProcessor, error) {
	err := checkArgsDirectConnectionProcessor(args)
	if err != nil {
		return nil, err
	}

	intraShardTopic := args.BaseIntraShardTopic +
		args.ShardCoordinator.CommunicationIdentifier(args.ShardCoordinator.SelfId())
	processor := &directConnectionProcessor{
		messenger:                   args.Messenger,
		peerShardMapper:             args.PeerShardMapper,
		shardCoordinator:            args.ShardCoordinator,
		timeToReadDirectConnections: args.TimeToReadDirectConnections,
		intraShardTopic:             intraShardTopic,
		baseCrossShardTopic:         args.BaseCrossShardTopic,
	}

	var ctx context.Context
	ctx, processor.cancel = context.WithCancel(context.Background())
	go processor.processLoop(ctx)

	return processor, nil
}

func checkArgsDirectConnectionProcessor(args ArgsDirectConnectionProcessor) error {
	if check.IfNil(args.Messenger) {
		return heartbeat.ErrNilMessenger
	}
	if check.IfNil(args.PeerShardMapper) {
		return heartbeat.ErrNilPeerShardMapper
	}
	if check.IfNil(args.ShardCoordinator) {
		return heartbeat.ErrNilShardCoordinator
	}
	if len(args.BaseIntraShardTopic) == 0 {
		return fmt.Errorf("%w for the BaseIntraShardTopic field", heartbeat.ErrInvalidValue)
	}
	if len(args.BaseCrossShardTopic) == 0 {
		return fmt.Errorf("%w for the BaseCrossShardTopic field", heartbeat.ErrInvalidValue)
	}
	if args.TimeToReadDirectConnections < minTimeToReadDirectConnections {
		return fmt.Errorf("%w for the TimeToReadDirectConnections field, minimum %v, got %v",
			heartbeat.ErrInvalidValue, minTimeToReadDirectConnections, args.TimeToReadDirectConnections)
	}

	return nil
}

func (processor *directConnectionProcessor) processLoop(ctx context.Context) {
	timer := time.NewTimer(processor.timeToReadDirectConnections)
	for {
		timer.Reset(processor.timeToReadDirectConnections)

		select {
		case <-ctx.Done():
			log.Debug("directConnectionProcessor is closing...")
			return
		case <-timer.C:
			processor.processDirectConnections()
		}
	}
}

func (processor *directConnectionProcessor) processDirectConnections() {
	mapIntraShardPeers := processor.processIntraShardPeers()

	processor.processCrossShardPeers(mapIntraShardPeers)
}

func (processor *directConnectionProcessor) processIntraShardPeers() map[core.PeerID]struct{} {
	mapIntraShardPeers := make(map[core.PeerID]struct{})
	intraShardPeers := processor.messenger.ConnectedPeersOnTopic(processor.intraShardTopic)
	for _, peer := range intraShardPeers {
		processor.peerShardMapper.PutPeerIdShardId(peer, processor.shardCoordinator.SelfId())
		mapIntraShardPeers[peer] = struct{}{}
	}

	return mapIntraShardPeers
}

func (processor *directConnectionProcessor) processCrossShardPeers(intraShardPeers map[core.PeerID]struct{}) {
	for i := uint32(0); i < processor.shardCoordinator.NumberOfShards(); i++ {
		if i == processor.shardCoordinator.SelfId() {
			continue
		}

		processor.processCrossShardPeersOnShard(i, intraShardPeers)
	}

	if processor.shardCoordinator.SelfId() != common.MetachainShardId {
		processor.processCrossShardPeersOnShard(common.MetachainShardId, intraShardPeers)
	}
}

func (processor *directConnectionProcessor) processCrossShardPeersOnShard(shardID uint32, intraShardPeers map[core.PeerID]struct{}) {
	// TODO create a component that can handle the topic creation & management
	identifier := processor.baseCrossShardTopic + processor.shardCoordinator.CommunicationIdentifier(shardID)
	crossShardPeers := processor.messenger.ConnectedPeersOnTopic(identifier)
	for _, peer := range crossShardPeers {
		_, isIntraShard := intraShardPeers[peer]
		if isIntraShard {
			continue
		}

		processor.peerShardMapper.PutPeerIdShardId(peer, shardID)
	}
}

// Close will end the direct connection processor main loop
func (processor *directConnectionProcessor) Close() error {
	processor.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (processor *directConnectionProcessor) IsInterfaceNil() bool {
	return processor == nil
}
