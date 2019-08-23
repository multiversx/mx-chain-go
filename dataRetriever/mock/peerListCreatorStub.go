package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type PeerListCreatorStub struct {
	PeerListCalled func() []p2p.PeerID
}

func (p PeerListCreatorStub) PeerList() []p2p.PeerID {
	return p.PeerListCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *PeerListCreatorStub) IsInterfaceNil() bool {
	if p == nil {
		return true
	}
	return false
}
