package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

// PeerListCreatorStub -
type PeerListCreatorStub struct {
	PeerListCalled           func() []p2p.PeerID
	IntraShardPeerListCalled func() []p2p.PeerID
}

// PeerList -
func (p *PeerListCreatorStub) PeerList() []p2p.PeerID {
	return p.PeerListCalled()
}

// IntraShardPeerList -
func (p *PeerListCreatorStub) IntraShardPeerList() []p2p.PeerID {
	return p.IntraShardPeerListCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *PeerListCreatorStub) IsInterfaceNil() bool {
	return p == nil
}
