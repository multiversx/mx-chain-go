package broadcastFactory

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
)

type shardChainMessengerFactory struct {
}

// NewShardChainMessengerFactory creates a shard messenger factory
func NewShardChainMessengerFactory() *shardChainMessengerFactory {
	return &shardChainMessengerFactory{}
}

// CreateShardChainMessenger creates a shard messenger for regular chain run type
func (f *shardChainMessengerFactory) CreateShardChainMessenger(args broadcast.ShardChainMessengerArgs) (consensus.BroadcastMessenger, error) {
	return broadcast.NewShardChainMessenger(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *shardChainMessengerFactory) IsInterfaceNil() bool {
	return f == nil
}
