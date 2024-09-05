package broadcastFactory

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
)

type shardChainMessengerFactory struct {
}

func NewShardChainMessengerFactory() *shardChainMessengerFactory {
	return &shardChainMessengerFactory{}
}

func (f *shardChainMessengerFactory) CreateShardChainMessenger(args broadcast.ShardChainMessengerArgs) (consensus.BroadcastMessenger, error) {
	return broadcast.NewShardChainMessenger(args)
}

func (f *shardChainMessengerFactory) IsInterfaceNil() bool {
	return f == nil
}
