package disabled

import "github.com/ElrondNetwork/elrond-go-core/core"

// peerShardMapper -
type peerShardMapper struct {
}

// NewPeerShardMapper -
func NewPeerShardMapper() *peerShardMapper {
	return &peerShardMapper{}
}

// UpdatePeerIDPublicKeyPair -
func (p *peerShardMapper) UpdatePeerIDPublicKeyPair(_ core.PeerID, _ []byte) {
}

// GetPeerInfo -
func (p *peerShardMapper) GetPeerInfo(_ core.PeerID) core.P2PPeerInfo {
	return core.P2PPeerInfo{}
}

// IsInterfaceNil -
func (p *peerShardMapper) IsInterfaceNil() bool {
	return p == nil
}
