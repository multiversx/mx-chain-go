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

type sharderFactory struct {
	reconnecter        p2p.Reconnecter
	peerShardResolver  p2p.PeerShardResolver
	prioBits           uint32
	pid                peer.ID
	maxConnectionCount int
	maxIntraShard      int
	maxCrossShard      int
}

// NewSharderFactory creates a new instance of sharderFactory able to create Sharder instances
//TODO(iulian) improve the parameter passing here
//TODO(iulian) should not pass reconnecter but a config, make this a constructor function
func NewSharderFactory(
	reconnecter p2p.Reconnecter,
	peerShardResolver p2p.PeerShardResolver,
	prioBits uint32,
	pid peer.ID,
	maxConnectionCount int,
	maxIntraShard int,
	maxCrossShard int,
) *sharderFactory {

	return &sharderFactory{
		reconnecter:        reconnecter,
		peerShardResolver:  peerShardResolver,
		prioBits:           prioBits,
		pid:                pid,
		maxConnectionCount: maxConnectionCount,
		maxIntraShard:      maxIntraShard,
		maxCrossShard:      maxCrossShard,
	}
}

// Create creates new Sharder instances
//TODO(iulian) make a common interface out of all sharders implementations and replace interface{}
func (sf *sharderFactory) Create() (interface{}, error) {
	if check.IfNil(sf.reconnecter) {
		return nil, fmt.Errorf("%w for sharderFactory.Create", p2p.ErrIncompatibleMethodCalled)
	}

	switch sf.reconnecter.(type) {
	case p2p.ReconnecterWithPauseResumeAndWatchdog:
		log.Debug("using kadSharder")
		return networksharding.NewKadSharder(sf.prioBits, sf.peerShardResolver)
	default:
		log.Debug("using list-based kadSharder")
		return networksharding.NewListKadSharder(
			sf.peerShardResolver,
			sf.pid,
			sf.maxConnectionCount,
			sf.maxIntraShard,
			sf.maxCrossShard,
		)
	}
}
