package factory

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/libp2p/go-libp2p-core/peer"
)

var log = logger.GetOrCreate("p2p/networksharding/factory")

// ArgsSharderFactory represents the argument for the sharder factory
type ArgsSharderFactory struct {
	PeerShardResolver    p2p.PeerShardResolver
	Pid                  peer.ID
	P2pConfig            config.P2PConfig
	PreferredPeersHolder p2p.PreferredPeersHolderHandler
	Type                 string
}

// NewSharder creates new Sharder instances
func NewSharder(arg ArgsSharderFactory) (p2p.Sharder, error) {
	switch arg.Type {
	case p2p.ListsSharder:
		log.Debug("using lists sharder",
			"MaxConnectionCount", arg.P2pConfig.Sharding.TargetPeerCount,
			"MaxIntraShardValidators", arg.P2pConfig.Sharding.MaxIntraShardValidators,
			"MaxCrossShardValidators", arg.P2pConfig.Sharding.MaxCrossShardValidators,
			"MaxIntraShardObservers", arg.P2pConfig.Sharding.MaxIntraShardObservers,
			"MaxCrossShardObservers", arg.P2pConfig.Sharding.MaxCrossShardObservers,
			"MaxFullHistoryObservers", arg.P2pConfig.Sharding.MaxFullHistoryObservers,
			"MaxSeeders", arg.P2pConfig.Sharding.MaxSeeders,
		)
		argListsSharder := networksharding.ArgListsSharder{
			PeerResolver:         arg.PeerShardResolver,
			SelfPeerId:           arg.Pid,
			P2pConfig:            arg.P2pConfig,
			PreferredPeersHolder: arg.PreferredPeersHolder,
		}
		return networksharding.NewListsSharder(argListsSharder)
	case p2p.OneListSharder:
		log.Debug("using one list sharder",
			"MaxConnectionCount", arg.P2pConfig.Sharding.TargetPeerCount,
		)
		return networksharding.NewOneListSharder(
			arg.Pid,
			int(arg.P2pConfig.Sharding.TargetPeerCount),
		)
	case p2p.NilListSharder:
		log.Debug("using nil list sharder")
		return networksharding.NewNilListSharder(), nil
	default:
		return nil, fmt.Errorf("%w when selecting sharder: unknown %s value", p2p.ErrInvalidValue, arg.Type)
	}
}
