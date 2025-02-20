package broadcast

import (
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/sharding"
)

type sovereignChainMessenger struct {
	*shardChainMessenger
}

// ArgsSovereignShardChainMessenger defines a struct placeholder for args needed to create a sovereign shard chain messenger
type ArgsSovereignShardChainMessenger struct {
	DelayedBroadcaster   delayedBroadcaster
	Marshaller           marshal.Marshalizer
	Hasher               hashing.Hasher
	ShardCoordinator     sharding.Coordinator
	Messenger            consensus.P2PMessenger
	PeerSignatureHandler crypto.PeerSignatureHandler
	KeysHandler          consensus.KeysHandler
}

// NewSovereignShardChainMessenger creates a new sovereign shard chain messenger
func NewSovereignShardChainMessenger(
	args ArgsSovereignShardChainMessenger,
) (*sovereignChainMessenger, error) {
	err := checkSovArgs(args)
	if err != nil {
		return nil, err
	}

	scm := &sovereignChainMessenger{
		shardChainMessenger: &shardChainMessenger{
			commonMessenger: &commonMessenger{
				marshalizer:             args.Marshaller,
				hasher:                  args.Hasher,
				messenger:               args.Messenger,
				shardCoordinator:        args.ShardCoordinator,
				peerSignatureHandler:    args.PeerSignatureHandler,
				keysHandler:             args.KeysHandler,
				delayedBlockBroadcaster: args.DelayedBroadcaster,
			},
		},
	}

	scm.broadcasterFilterHandler = scm

	err = scm.delayedBlockBroadcaster.SetBroadcastHandlers(scm.BroadcastMiniBlocks, scm.BroadcastTransactions, scm.BroadcastHeader)
	if err != nil {
		return nil, err
	}

	return scm, nil
}

func checkSovArgs(
	args ArgsSovereignShardChainMessenger,
) error {
	if check.IfNil(args.Marshaller) {
		return spos.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return spos.ErrNilHasher
	}
	if check.IfNil(args.Messenger) {
		return spos.ErrNilMessenger
	}
	if check.IfNil(args.ShardCoordinator) {
		return spos.ErrNilShardCoordinator
	}
	if check.IfNil(args.PeerSignatureHandler) {
		return spos.ErrNilPeerSignatureHandler
	}
	if check.IfNil(args.KeysHandler) {
		return ErrNilKeysHandler
	}
	if args.DelayedBroadcaster == nil {
		return errNilDelayedShardBroadCaster
	}

	return nil
}

// BroadcastBlock will send on in-shard headers topic and on in-shard miniblocks topic the header and block body
func (scm *sovereignChainMessenger) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	broadCastData, err := scm.getBroadCastBlockData(blockBody, header)
	if err != nil {
		return err
	}

	identifier := scm.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)

	scm.messenger.Broadcast(factory.ShardBlocksTopic+identifier, broadCastData.marshalledHeader)
	scm.messenger.Broadcast(factory.MiniBlocksTopic+identifier, broadCastData.marshalledBody)

	return nil
}

// BroadcastHeader will send on in-shard headers topic the header
func (scm *sovereignChainMessenger) BroadcastHeader(header data.HeaderHandler, pkBytes []byte) error {
	shardIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	return scm.broadcastHeader(header, pkBytes, shardIdentifier)
}

func (scm *sovereignChainMessenger) shouldSkipShard(shardID uint32) bool {
	return shardID != core.SovereignChainShardId
}
func (scm *sovereignChainMessenger) shouldSkipTopic(topic string) bool {
	return strings.Contains(topic, fmt.Sprintf("%d", core.MainChainShardId))
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *sovereignChainMessenger) IsInterfaceNil() bool {
	return scm == nil
}
