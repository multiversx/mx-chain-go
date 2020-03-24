package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// PeerShardResolverStub -
type PeerShardResolverStub struct {
	GetPeerInfoCalled func(pid p2p.PeerID) core.P2PPeerInfo
}

// GetPeerInfo -
func (psrs *PeerShardResolverStub) GetPeerInfo(pid p2p.PeerID) core.P2PPeerInfo {
	return psrs.GetPeerInfoCalled(pid)
}

// IsInterfaceNil -
func (psrs *PeerShardResolverStub) IsInterfaceNil() bool {
	return psrs == nil
}
