package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding"
	"github.com/libp2p/go-libp2p-core/peer"
)

var log = logger.GetOrCreate("p2p/networksharding/factory")

// ArgsSharderFactory represents the argument for the sharder factory
//TODO(iulian) should not pass reconnecter but a config,
type ArgsSharderFactory struct {
	Reconnecter        p2p.Reconnecter
	PeerShardResolver  p2p.PeerShardResolver
	PrioBits           uint32
	Pid                peer.ID
	MaxConnectionCount int
	MaxIntraShard      int
	MaxCrossShard      int
}

// NewSharder creates new Sharder instances
//TODO(iulian) make a common interface out of all sharders implementations and replace interface{}
func NewSharder(arg ArgsSharderFactory) (p2p.CommonSharder, error) {
	if check.IfNil(arg.Reconnecter) {
		return nil, fmt.Errorf("%w for sharderFactory.Create", p2p.ErrIncompatibleMethodCalled)
	}

	switch arg.Reconnecter.(type) {
	case p2p.ReconnecterWithPauseResumeAndWatchdog:
		log.Debug("using kadSharder")
		return networksharding.NewKadSharder(arg.PrioBits, arg.PeerShardResolver)
	default:
		log.Debug("using list-based kadSharder")
		return networksharding.NewListKadSharder(
			arg.PeerShardResolver,
			arg.Pid,
			arg.MaxConnectionCount,
			arg.MaxIntraShard,
			arg.MaxCrossShard,
		)
	}
}
