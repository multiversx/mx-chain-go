package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgCrossShardStatusProcessor represents the arguments for the cross shard status processor
type ArgCrossShardStatusProcessor struct {
	Messenger            p2p.Messenger
	PeerShardMapper      process.PeerShardMapper
	ShardCoordinator     sharding.Coordinator
	DelayBetweenRequests time.Duration
}

type crossShardStatusProcessor struct {
	messenger            p2p.Messenger
	peerShardMapper      process.PeerShardMapper
	shardCoordinator     sharding.Coordinator
	delayBetweenRequests time.Duration
	cancel               func()
	// todo remove this - tests only
	LatestKnownPeers map[string][]core.PeerID
}

// NewCrossShardStatusProcessor creates a new instance of crossShardStatusProcessor
func NewCrossShardStatusProcessor(args ArgCrossShardStatusProcessor) (*crossShardStatusProcessor, error) {
	err := checkArgsCrossShardStatusProcessor(args)
	if err != nil {
		return nil, err
	}

	cssp := &crossShardStatusProcessor{
		messenger:            args.Messenger,
		peerShardMapper:      args.PeerShardMapper,
		shardCoordinator:     args.ShardCoordinator,
		delayBetweenRequests: args.DelayBetweenRequests,
	}

	var ctx context.Context
	ctx, cssp.cancel = context.WithCancel(context.Background())

	go cssp.startProcessLoop(ctx)

	return cssp, nil
}

func checkArgsCrossShardStatusProcessor(args ArgCrossShardStatusProcessor) error {
	if check.IfNil(args.Messenger) {
		return process.ErrNilMessenger
	}
	if check.IfNil(args.PeerShardMapper) {
		return process.ErrNilPeerShardMapper
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if args.DelayBetweenRequests < minDelayBetweenRequests {
		return fmt.Errorf("%w for DelayBetweenRequests, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.DelayBetweenRequests, minDelayBetweenRequests)
	}

	return nil
}

func (cssp *crossShardStatusProcessor) startProcessLoop(ctx context.Context) {
	timer := time.NewTimer(cssp.delayBetweenRequests)

	defer func() {
		cssp.cancel()
		timer.Stop()
	}()

	requestedTopicsMap := cssp.computeTopicsMap()

	for {
		timer.Reset(cssp.delayBetweenRequests)

		select {
		case <-timer.C:
			cssp.updatePeersInfo(requestedTopicsMap)
		case <-ctx.Done():
			log.Debug("closing crossShardStatusProcessor go routine")
			return
		}
	}
}

func (cssp *crossShardStatusProcessor) computeTopicsMap() map[uint32]string {
	requestedTopicsMap := make(map[uint32]string, 0)

	numOfShards := cssp.shardCoordinator.NumberOfShards()
	for shard := uint32(0); shard < numOfShards; shard++ {
		topicIdentifier := factory.TransactionTopic + cssp.shardCoordinator.CommunicationIdentifier(shard)
		requestedTopicsMap[shard] = topicIdentifier
	}

	metaIdentifier := factory.TransactionTopic + cssp.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	requestedTopicsMap[core.MetachainShardId] = metaIdentifier

	selfShard := cssp.shardCoordinator.SelfId()
	delete(requestedTopicsMap, selfShard)

	return requestedTopicsMap
}

func (cssp *crossShardStatusProcessor) updatePeersInfo(requestedTopicsMap map[uint32]string) {
	cssp.LatestKnownPeers = make(map[string][]core.PeerID, 0)

	intraShardPeersMap := cssp.getIntraShardConnectedPeers()

	for shard, topic := range requestedTopicsMap {
		connectedPids := cssp.messenger.ConnectedPeersOnTopic(topic)
		for _, pid := range connectedPids {
			_, fromSameShard := intraShardPeersMap[pid]
			if fromSameShard {
				continue
			}

			cssp.peerShardMapper.UpdatePeerIdShardId(pid, shard)

			// todo remove this - tests only
			cssp.LatestKnownPeers[topic] = append(cssp.LatestKnownPeers[topic], pid)
		}
	}
}

func (cssp *crossShardStatusProcessor) getIntraShardConnectedPeers() map[core.PeerID]struct{} {
	selfShard := cssp.shardCoordinator.SelfId()
	intraShardTopic := factory.TransactionTopic + cssp.shardCoordinator.CommunicationIdentifier(selfShard)
	intraShardPeers := cssp.messenger.ConnectedPeersOnTopic(intraShardTopic)

	intraShardPeersMap := make(map[core.PeerID]struct{}, 0)
	for _, pid := range intraShardPeers {
		intraShardPeersMap[pid] = struct{}{}
	}

	return intraShardPeersMap
}

// GetLatestKnownPeers - todo remove this - tests only
func (cssp *crossShardStatusProcessor) GetLatestKnownPeers() map[string][]core.PeerID {
	return cssp.LatestKnownPeers
}

// Close closes the internal goroutine
func (cssp *crossShardStatusProcessor) Close() error {
	log.Debug("closing crossShardStatusProcessor...")
	cssp.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under interface
func (cssp *crossShardStatusProcessor) IsInterfaceNil() bool {
	return cssp == nil
}
