package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

// SharderStub -
type SharderStub struct {
	ComputeEvictListCalled     func(pidList []peer.ID) []peer.ID
	HasCalled                  func(pid peer.ID, list []peer.ID) bool
	SetPeerShardResolverCalled func(psp p2p.PeerShardResolver) error
}

// ComputeEvictionList -
func (ss *SharderStub) ComputeEvictionList(pidList []peer.ID) []peer.ID {
	if ss.ComputeEvictListCalled != nil {
		return ss.ComputeEvictListCalled(pidList)
	}

	return make([]peer.ID, 0)
}

// Has -
func (ss *SharderStub) Has(pid peer.ID, list []peer.ID) bool {
	if ss.HasCalled != nil {
		return ss.HasCalled(pid, list)
	}

	return false
}

// SetPeerShardResolver -
func (ss *SharderStub) SetPeerShardResolver(psp p2p.PeerShardResolver) error {
	if ss.SetPeerShardResolverCalled != nil {
		return ss.SetPeerShardResolverCalled(psp)
	}

	return nil
}

// IsInterfaceNil -
func (ss *SharderStub) IsInterfaceNil() bool {
	return ss == nil
}
