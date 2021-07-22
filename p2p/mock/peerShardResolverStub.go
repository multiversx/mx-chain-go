package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// PeerShardResolverStub -
type PeerShardResolverStub struct {
	GetPeerInfoCalled func(pid core.PeerID) core.P2PPeerInfo
}

// GetPeerInfo -
func (psrs *PeerShardResolverStub) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	return psrs.GetPeerInfoCalled(pid)
}

// IsInterfaceNil -
func (psrs *PeerShardResolverStub) IsInterfaceNil() bool {
	return psrs == nil
}
