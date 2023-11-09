package disabled

import "github.com/multiversx/mx-chain-core-go/core"

// peerShardMapper represents the disabled structure of peerShardMapper
type peerShardMapper struct {
}

// NewPeerShardMapper returns default instance
func NewPeerShardMapper() *peerShardMapper {
	return &peerShardMapper{}
}

// GetLastKnownPeerID returns nothing
func (p *peerShardMapper) GetLastKnownPeerID(_ []byte) (core.PeerID, bool) {
	return "", false
}

// UpdatePeerIDPublicKeyPair does nothing
func (p *peerShardMapper) UpdatePeerIDPublicKeyPair(_ core.PeerID, _ []byte) {
}

// PutPeerIdShardId does nothing
func (p *peerShardMapper) PutPeerIdShardId(_ core.PeerID, _ uint32) {
}

// PutPeerIdSubType does nothing
func (p *peerShardMapper) PutPeerIdSubType(_ core.PeerID, _ core.P2PPeerSubType) {
}

// GetPeerInfo returns default instance
func (p *peerShardMapper) GetPeerInfo(_ core.PeerID) core.P2PPeerInfo {
	return core.P2PPeerInfo{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *peerShardMapper) IsInterfaceNil() bool {
	return p == nil
}
