package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/libp2p/go-libp2p-core/peer"
)

var log = logger.GetOrCreate("p2p/networksharding/factory")

// ArgsSharderFactory represents the argument for the sharder factory
type ArgsSharderFactory struct {
	PeerShardResolver       p2p.PeerShardResolver
	Pid                     peer.ID
	MaxConnectionCount      uint32
	MaxIntraShardValidators uint32
	MaxCrossShardValidators uint32
	MaxIntraShardObservers  uint32
	MaxCrossShardObservers  uint32
	MaxSeeders              uint32
	Type                    string
}

// NewSharder creates new Sharder instances
func NewSharder(arg ArgsSharderFactory) (p2p.Sharder, error) {
	switch arg.Type {
	case p2p.ListsSharder:
		log.Debug("using lists sharder",
			"MaxConnectionCount", arg.MaxConnectionCount,
			"MaxIntraShardValidators", arg.MaxIntraShardValidators,
			"MaxCrossShardValidators", arg.MaxCrossShardValidators,
			"MaxIntraShardObservers", arg.MaxIntraShardObservers,
			"MaxCrossShardObservers", arg.MaxCrossShardObservers,
			"MaxSeeders", arg.MaxSeeders,
		)

		argListsSharder := networksharding.ArgListsSharder{
			PeerResolver:            arg.PeerShardResolver,
			SelfPeerId:              arg.Pid,
			MaxPeerCount:            arg.MaxConnectionCount,
			MaxIntraShardValidators: arg.MaxIntraShardValidators,
			MaxCrossShardValidators: arg.MaxCrossShardValidators,
			MaxIntraShardObservers:  arg.MaxIntraShardObservers,
			MaxCrossShardObservers:  arg.MaxCrossShardObservers,
			MaxSeeders:              arg.MaxSeeders,
		}
		return networksharding.NewListsSharder(argListsSharder)
	case p2p.OneListSharder:
		log.Debug("using one list sharder",
			"MaxConnectionCount", arg.MaxConnectionCount,
		)
		return networksharding.NewOneListSharder(
			arg.Pid,
			int(arg.MaxConnectionCount),
		)
	case p2p.NilListSharder:
		log.Debug("using nil list sharder")
		return networksharding.NewNilListSharder(), nil
	default:
		return nil, fmt.Errorf("%w when selecting sharder: unknown %s value", p2p.ErrInvalidValue, arg.Type)
	}
}
