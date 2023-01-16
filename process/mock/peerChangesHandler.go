package mock

import "github.com/multiversx/mx-chain-core-go/data/block"

// PeerChangesHandler -
type PeerChangesHandler struct {
	PeerChangesCalled       func() []block.PeerData
	VerifyPeerChangesCalled func(peerChanges []block.PeerData) error
}

// PeerChanges -
func (p *PeerChangesHandler) PeerChanges() []block.PeerData {
	if p.PeerChangesCalled != nil {
		return p.PeerChangesCalled()
	}
	return nil
}

// VerifyPeerChanges -
func (p *PeerChangesHandler) VerifyPeerChanges(peerChanges []block.PeerData) error {
	if p.VerifyPeerChangesCalled != nil {
		return p.VerifyPeerChangesCalled(peerChanges)
	}
	return nil
}

// IsInterfaceNil -
func (p *PeerChangesHandler) IsInterfaceNil() bool {
	return p == nil
}
