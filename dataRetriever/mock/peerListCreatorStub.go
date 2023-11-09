package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// PeerListCreatorStub -
type PeerListCreatorStub struct {
	CrossShardPeerListCalled func() []core.PeerID
	IntraShardPeerListCalled func() []core.PeerID
}

// CrossShardPeerList -
func (p *PeerListCreatorStub) CrossShardPeerList() []core.PeerID {
	if p.CrossShardPeerListCalled != nil {
		return p.CrossShardPeerListCalled()
	}
	return make([]core.PeerID, 0)
}

// IntraShardPeerList -
func (p *PeerListCreatorStub) IntraShardPeerList() []core.PeerID {
	if p.IntraShardPeerListCalled != nil {
		return p.IntraShardPeerListCalled()
	}
	return make([]core.PeerID, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *PeerListCreatorStub) IsInterfaceNil() bool {
	return p == nil
}
