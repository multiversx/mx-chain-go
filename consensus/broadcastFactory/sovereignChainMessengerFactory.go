package broadcastFactory

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
)

type sovereignChainMessengerFactory struct {
}

func NewSovereignShardChainMessengerFactory() *sovereignChainMessengerFactory {
	return &sovereignChainMessengerFactory{}
}

func (f *sovereignChainMessengerFactory) CreateShardChainMessenger(args broadcast.ShardChainMessengerArgs) (consensus.BroadcastMessenger, error) {
	return broadcast.NewSovereignShardChainMessenger(args)
}

func (f *sovereignChainMessengerFactory) IsInterfaceNil() bool {
	return f == nil
}
