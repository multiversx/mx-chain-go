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
	PeerShardResolver  p2p.PeerShardResolver
	PrioBits           uint32
	Pid                peer.ID
	MaxConnectionCount int
	MaxIntraShard      int
	MaxCrossShard      int
	Type               string
}

// NewSharder creates new Sharder instances
func NewSharder(arg ArgsSharderFactory) (p2p.CommonSharder, error) {
	switch arg.Type {
	case p2p.SharderVariantPrioBits:
		log.Debug("using kadSharder with prio bits")
		return networksharding.NewKadSharder(arg.PrioBits, arg.PeerShardResolver)
	case p2p.SharderVariantWithLists:
		log.Debug("using list-based kadSharder")
		return networksharding.NewListKadSharder(
			arg.PeerShardResolver,
			arg.Pid,
			arg.MaxConnectionCount,
			arg.MaxIntraShard,
			arg.MaxCrossShard,
		)
	case p2p.NoSharderWithLists:
		log.Debug("using list-based no kadSharder")
		return networksharding.NewListNoKadSharder(
			arg.Pid,
			arg.MaxConnectionCount,
		)
	case p2p.DisabledSharder:
		log.Debug("kadSharder disabled")
		return networksharding.NewDisabledSharder(), nil
	default:
		return nil, fmt.Errorf("%w when selecting sharder: unknown %s value", p2p.ErrInvalidValue, arg.Type)
	}
}
