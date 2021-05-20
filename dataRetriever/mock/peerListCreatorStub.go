package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// PeerListCreatorStub -
type PeerListCreatorStub struct {
	PeerListCalled           func() []core.PeerID
	IntraShardPeerListCalled func() []core.PeerID
	FullHistoryListCalled    func() []core.PeerID
}

// PeerList -
func (p *PeerListCreatorStub) PeerList() []core.PeerID {
	return p.PeerListCalled()
}

// IntraShardPeerList -
func (p *PeerListCreatorStub) IntraShardPeerList() []core.PeerID {
	return p.IntraShardPeerListCalled()
}

// FullHistoryList -
func (p *PeerListCreatorStub) FullHistoryList() []core.PeerID {
	return p.FullHistoryListCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *PeerListCreatorStub) IsInterfaceNil() bool {
	return p == nil
}
