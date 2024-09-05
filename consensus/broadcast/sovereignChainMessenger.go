package broadcast

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignChainMessenger struct {
	*shardChainMessenger
}

// NewSovereignShardChainMessenger creates a new sovereign shard chain messenger
func NewSovereignShardChainMessenger(
	args ShardChainMessengerArgs,
) (*sovereignChainMessenger, error) {
	err := checkShardChainNilParameters(args)
	if err != nil {
		return nil, err
	}

	cm := &commonMessenger{
		marshalizer:          args.Marshalizer,
		hasher:               args.Hasher,
		messenger:            args.Messenger,
		shardCoordinator:     args.ShardCoordinator,
		peerSignatureHandler: args.PeerSignatureHandler,
		keysHandler:          args.KeysHandler,
	}

	dbbArgs := &ArgsDelayedBlockBroadcaster{
		InterceptorsContainer: args.InterceptorsContainer,
		HeadersSubscriber:     args.HeadersSubscriber,
		LeaderCacheSize:       args.MaxDelayCacheSize,
		ValidatorCacheSize:    args.MaxValidatorDelayCacheSize,
		ShardCoordinator:      args.ShardCoordinator,
		AlarmScheduler:        args.AlarmScheduler,
	}

	dbb, err := NewSovereignDelayedBlockBroadcaster(dbbArgs)
	if err != nil {
		return nil, err
	}

	cm.delayedBlockBroadcaster = dbb

	scm := &sovereignChainMessenger{
		shardChainMessenger: &shardChainMessenger{
			commonMessenger: cm,
		},
	}

	err = dbb.SetBroadcastHandlers(scm.BroadcastMiniBlocks, scm.BroadcastTransactions, scm.BroadcastHeader)
	if err != nil {
		return nil, err
	}

	return scm, nil
}

// BroadcastBlock will send on in-shard headers topic and on in-shard miniblocks topic the header and block body
func (scm *sovereignChainMessenger) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	if check.IfNil(blockBody) {
		return spos.ErrNilBody
	}

	err := blockBody.IntegrityAndValidity()
	if err != nil {
		return err
	}

	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	b := blockBody.(*block.Body)
	msgBlockBody, err := scm.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	headerIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	selfIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)

	scm.messenger.Broadcast(factory.ShardBlocksTopic+headerIdentifier, msgHeader)
	scm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastHeader will send on in-shard headers topic the header
func (scm *sovereignChainMessenger) BroadcastHeader(header data.HeaderHandler, pkBytes []byte) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	shardIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	scm.broadcast(factory.ShardBlocksTopic+shardIdentifier, msgHeader, pkBytes)

	return nil
}
