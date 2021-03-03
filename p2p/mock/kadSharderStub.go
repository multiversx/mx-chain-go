package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

// KadSharderStub -
type KadSharderStub struct {
	ComputeEvictListCalled     func(pidList []peer.ID) []peer.ID
	HasCalled                  func(pid peer.ID, list []peer.ID) bool
	SetPeerShardResolverCalled func(psp p2p.PeerShardResolver) error
	SetSeedersCalled           func(addresses []string)
	IsSeederCalled             func(pid core.PeerID) bool
}

// ComputeEvictionList -
func (kss *KadSharderStub) ComputeEvictionList(pidList []peer.ID) []peer.ID {
	if kss.ComputeEvictListCalled != nil {
		return kss.ComputeEvictListCalled(pidList)
	}

	return make([]peer.ID, 0)
}

// Has -
func (kss *KadSharderStub) Has(pid peer.ID, list []peer.ID) bool {
	if kss.HasCalled != nil {
		return kss.HasCalled(pid, list)
	}

	return false
}

// SetPeerShardResolver -
func (kss *KadSharderStub) SetPeerShardResolver(psp p2p.PeerShardResolver) error {
	if kss.SetPeerShardResolverCalled != nil {
		return kss.SetPeerShardResolverCalled(psp)
	}

	return nil
}

// SetSeeders -
func (kss *KadSharderStub) SetSeeders(addresses []string) {
	if kss.SetSeedersCalled != nil {
		kss.SetSeedersCalled(addresses)
	}
}

// IsSeeder -
func (kss *KadSharderStub) IsSeeder(pid core.PeerID) bool {
	if kss.IsSeederCalled != nil {
		return kss.IsSeederCalled(pid)
	}

	return false
}

// IsInterfaceNil -
func (kss *KadSharderStub) IsInterfaceNil() bool {
	return kss == nil
}
