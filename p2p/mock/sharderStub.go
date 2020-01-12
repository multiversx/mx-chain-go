package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/libp2p/go-libp2p-core/peer"
)

type SharderStub struct {
	ComputeEvictListCalled  func(pid peer.ID, connected []peer.ID) []peer.ID
	HasCalled               func(pid peer.ID, list []peer.ID) bool
	PeerShardResolverCalled func() p2p.PeerShardResolver
}

func (ss *SharderStub) ComputeEvictList(pid peer.ID, connected []peer.ID) []peer.ID {
	if ss.ComputeEvictListCalled != nil {
		return ss.ComputeEvictListCalled(pid, connected)
	}

	return make([]peer.ID, 0)
}

func (ss *SharderStub) Has(pid peer.ID, list []peer.ID) bool {
	if ss.HasCalled != nil {
		return ss.HasCalled(pid, list)
	}

	return false
}

func (ss *SharderStub) PeerShardResolver() p2p.PeerShardResolver {
	if ss.PeerShardResolverCalled != nil {
		return ss.PeerShardResolverCalled()
	}

	return nil
}

func (ss *SharderStub) IsInterfaceNil() bool {
	return ss == nil
}
