package mock

import "github.com/ElrondNetwork/elrond-go/data/block"

type PeerChangesHandler struct {
	PeerChangesCalled       func() []block.PeerData
	VerifyPeerChangesCalled func(peerChanges []block.PeerData) error
}

func (p *PeerChangesHandler) PeerChanges() []block.PeerData {
	if p.PeerChangesCalled != nil {
		return p.PeerChangesCalled()
	}
	return nil
}

func (p *PeerChangesHandler) VerifyPeerChanges(peerChanges []block.PeerData) error {
	if p.VerifyPeerChangesCalled != nil {
		return p.VerifyPeerChangesCalled(peerChanges)
	}
	return nil
}

func (p *PeerChangesHandler) IsInterfaceNil() bool {
	if p == nil {
		return true
	}
	return false
}
