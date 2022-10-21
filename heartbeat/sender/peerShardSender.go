package sender

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/processor"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/factory"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const minDelayBetweenSends = time.Second

// ArgPeerShardSender represents the arguments for the peer shard sender
type ArgPeerShardSender struct {
	Messenger             p2p.Messenger
	Marshaller            marshal.Marshalizer
	ShardCoordinator      sharding.Coordinator
	TimeBetweenSends      time.Duration
	ThresholdBetweenSends float64
	NodesCoordinator      processor.NodesCoordinator
}

type peerShardSender struct {
	messenger             p2p.Messenger
	marshaller            marshal.Marshalizer
	shardCoordinator      sharding.Coordinator
	timeBetweenSends      time.Duration
	thresholdBetweenSends float64
	nodesCoordinator      processor.NodesCoordinator
	cancel                func()
}

// NewPeerShardSender creates a new instance of peerShardSender
func NewPeerShardSender(args ArgPeerShardSender) (*peerShardSender, error) {
	err := checkArgPeerShardSender(args)
	if err != nil {
		return nil, err
	}

	pss := &peerShardSender{
		messenger:             args.Messenger,
		marshaller:            args.Marshaller,
		shardCoordinator:      args.ShardCoordinator,
		timeBetweenSends:      args.TimeBetweenSends,
		thresholdBetweenSends: args.ThresholdBetweenSends,
		nodesCoordinator:      args.NodesCoordinator,
	}

	var ctx context.Context
	ctx, pss.cancel = context.WithCancel(context.Background())

	go pss.startSendingShard(ctx)

	return pss, nil
}

func checkArgPeerShardSender(args ArgPeerShardSender) error {
	if check.IfNil(args.Messenger) {
		return process.ErrNilMessenger
	}
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if args.TimeBetweenSends < minDelayBetweenSends {
		return fmt.Errorf("%w for TimeBetweenSends, provided %d, min expected %d",
			heartbeat.ErrInvalidTimeDuration, args.TimeBetweenSends, minDelayBetweenSends)
	}
	if args.ThresholdBetweenSends < minThresholdBetweenSends || args.ThresholdBetweenSends > maxThresholdBetweenSends {
		return fmt.Errorf("%w for ThresholdBetweenSends, receieved %f, min allowed %f, max allowed %f",
			heartbeat.ErrInvalidThreshold, args.ThresholdBetweenSends, minThresholdBetweenSends, maxThresholdBetweenSends)
	}
	if check.IfNil(args.NodesCoordinator) {
		return heartbeat.ErrNilNodesCoordinator
	}

	return nil
}

func (pss *peerShardSender) startSendingShard(ctx context.Context) {
	duration := pss.computeRandomDuration(pss.timeBetweenSends)
	timer := time.NewTimer(duration)
	defer timer.Stop()

	log.Debug("starting peer shard sender's goroutine")
	for {
		duration = pss.computeRandomDuration(pss.timeBetweenSends)
		timer.Reset(duration)

		select {
		case <-timer.C:
			pss.broadcastShard()
		case <-ctx.Done():
			log.Debug("closing peerShardSender go routine")
			return
		}
	}
}

func (pss *peerShardSender) computeRandomDuration(baseDuration time.Duration) time.Duration {
	timeBetweenSendsInNano := baseDuration.Nanoseconds()
	maxThreshold := float64(timeBetweenSendsInNano) * pss.thresholdBetweenSends
	randThreshold := randomizer.Intn(int(maxThreshold))

	ret := time.Duration(timeBetweenSendsInNano + int64(randThreshold))
	return ret
}

func (pss *peerShardSender) broadcastShard() {
	if pss.isCurrentNodeValidator() {
		log.Debug("peerShardSender.sendShard current node is validator, will not send messages to connected peers")
		return
	}

	peerShard := &factory.PeerShard{
		ShardId: fmt.Sprintf("%d", pss.shardCoordinator.SelfId()),
	}

	peerShardBuff, err := pss.marshaller.Marshal(peerShard)
	if err != nil {
		return
	}

	log.Debug("broadcast peer shard", "shard", peerShard.ShardId)
	pss.messenger.Broadcast(common.ConnectionTopic, peerShardBuff)
}

func (pss *peerShardSender) isCurrentNodeValidator() bool {
	currentBLSKey := pss.nodesCoordinator.GetOwnPublicKey()
	_, _, err := pss.nodesCoordinator.GetValidatorWithPublicKey(currentBLSKey)

	return err == nil
}

// Close triggers the closing of the internal goroutine
func (pss *peerShardSender) Close() error {
	log.Debug("closing shardSender...")
	pss.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pss *peerShardSender) IsInterfaceNil() bool {
	return pss == nil
}
