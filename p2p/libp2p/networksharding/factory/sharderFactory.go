package factory

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/libp2p/go-libp2p-core/peer"
)

var log = logger.GetOrCreate("p2p/networksharding/factory")

// ArgsSharderFactory represents the argument for the sharder factory
type ArgsSharderFactory struct {
	PeerShardResolver       p2p.PeerShardResolver
	Pid                     peer.ID
	MaxConnectionCount      int
	MaxIntraShardValidators int
	MaxCrossShardValidators int
	MaxIntraShardObservers  int
	MaxCrossShardObservers  int
	MaxFullHistoryObservers int
	Type                    string
}

// NewSharder creates new Sharder instances
func NewSharder(arg ArgsSharderFactory) (p2p.CommonSharder, error) {
	switch arg.Type {
	case p2p.ListsSharder:
		log.Debug("using lists sharder",
			"MaxConnectionCount", arg.MaxConnectionCount,
			"MaxIntraShardValidators", arg.MaxIntraShardValidators,
			"MaxCrossShardValidators", arg.MaxCrossShardValidators,
			"MaxIntraShardObservers", arg.MaxIntraShardObservers,
			"MaxCrossShardObservers", arg.MaxCrossShardObservers,
			"MaxFullHistoryObservers", arg.MaxFullHistoryObservers,
		)
		return networksharding.NewListsSharder(
			arg.PeerShardResolver,
			arg.Pid,
			arg.MaxConnectionCount,
			arg.MaxIntraShardValidators,
			arg.MaxCrossShardValidators,
			arg.MaxIntraShardObservers,
			arg.MaxCrossShardObservers,
			arg.MaxFullHistoryObservers,
		)
	case p2p.OneListSharder:
		log.Debug("using one list sharder",
			"MaxConnectionCount", arg.MaxConnectionCount,
		)
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
