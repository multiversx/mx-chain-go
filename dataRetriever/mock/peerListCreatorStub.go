package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type PeerListCreatorStub struct {
	PeerListCalled func() []p2p.PeerID
}

func (p PeerListCreatorStub) PeerList() []p2p.PeerID {
	return p.PeerListCalled()
}
