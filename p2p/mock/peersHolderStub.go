package mock

// PeersHolderStub -
type PeersHolderStub struct {
	AddCalled                   func(publicKey string, peerID string)
	GetPeerIDForPublicKeyCalled func(publicKey string) (string, bool)
	GetPublicKeyForPeerIDCalled func(peerID string) (string, bool)
	DeletePublicKeyCalled       func(pubKey string) bool
	DeletePeerIDCalled          func(pID string) bool
	LenCalled                   func() int
	ClearCalled                 func()
}

// Add -
func (p *PeersHolderStub) Add(publicKey string, peerID string) {
	p.AddCalled(publicKey, peerID)
}

// GetPeerIDForPublicKey -
func (p *PeersHolderStub) GetPeerIDForPublicKey(publicKey string) (string, bool) {
	return p.GetPeerIDForPublicKeyCalled(publicKey)
}

// GetPublicKeyForPeerID -
func (p *PeersHolderStub) GetPublicKeyForPeerID(peerID string) (string, bool) {
	return p.GetPublicKeyForPeerIDCalled(peerID)
}

// DeletePublicKey -
func (p *PeersHolderStub) DeletePublicKey(pubKey string) bool {
	return p.DeletePublicKeyCalled(pubKey)
}

// DeletePeerID -
func (p *PeersHolderStub) DeletePeerID(pID string) bool {
	return p.DeletePeerIDCalled(pID)
}

// Len -
func (p *PeersHolderStub) Len() int {
	return p.LenCalled()
}

// Clear -
func (p *PeersHolderStub) Clear() {
	p.ClearCalled()
}

// IsInterfaceNil -
func (p *PeersHolderStub) IsInterfaceNil() bool {
	return p == nil
}
