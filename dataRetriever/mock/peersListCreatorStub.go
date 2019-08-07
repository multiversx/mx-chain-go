package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type PeersListCreatorStub struct {
	PeersListCalled func() []p2p.PeerID
}

func (p PeersListCreatorStub) PeersList() []p2p.PeerID {
	return p.PeersListCalled()
}
