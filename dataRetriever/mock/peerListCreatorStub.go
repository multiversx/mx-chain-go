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
	return p.CrossShardPeerListCalled()
}

// IntraShardPeerList -
func (p *PeerListCreatorStub) IntraShardPeerList() []core.PeerID {
	return p.IntraShardPeerListCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *PeerListCreatorStub) IsInterfaceNil() bool {
	return p == nil
}
