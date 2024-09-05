package sposFactory

import (
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
)

type BroadCastShardMessengerFactoryHandler interface {
	CreateShardChainMessenger(args broadcast.ShardChainMessengerArgs) (consensus.BroadcastMessenger, error)
	IsInterfaceNil() bool
}
