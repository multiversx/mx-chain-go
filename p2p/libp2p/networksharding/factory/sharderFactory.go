package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/libp2p/go-libp2p-core/peer"
)

var log = logger.GetOrCreate("p2p/networksharding/factory")

// ArgsSharderFactory represents the argument for the sharder factory
type ArgsSharderFactory struct {
	PeerShardResolver   p2p.PeerShardResolver
	PrioBits            uint32
	Pid                 peer.ID
	MaxConnectionCount  int
	MaxIntraShard       int
	MaxCrossShard       int
	PercentageObservers int
	Type                string
}

// NewSharder creates new Sharder instances
func NewSharder(arg ArgsSharderFactory) (p2p.CommonSharder, error) {
	switch arg.Type {
	case p2p.PrioBitsSharder:
		log.Debug("using prio bits sharder")
		return networksharding.NewPrioBitsSharder(arg.PrioBits, arg.PeerShardResolver)
	case p2p.SimplePrioBitsSharder:
		log.Debug("using simple prio bits sharder")
		return &networksharding.SimplePrioBitsSharder{}, nil
	case p2p.ListsSharder:
		log.Debug("using lists sharder")
		return networksharding.NewListsSharder(
			arg.PeerShardResolver,
			arg.Pid,
			arg.MaxConnectionCount,
			arg.MaxIntraShard,
			arg.MaxCrossShard,
			arg.PercentageObservers,
		)
	case p2p.OneListSharder:
		log.Debug("using one list sharder")
		return networksharding.NewOneListSharder(
			arg.Pid,
			arg.MaxConnectionCount,
		)
	case p2p.NilListSharder:
		log.Debug("using nil list sharder")
		return networksharding.NewNilListSharder(), nil
	default:
		return nil, fmt.Errorf("%w when selecting sharder: unknown %s value", p2p.ErrInvalidValue, arg.Type)
	}
}
