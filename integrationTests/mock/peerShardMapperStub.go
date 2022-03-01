package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// PeerShardMapperStub -
type PeerShardMapperStub struct {
	GetPeerIDCalled                 func(pk []byte) (*core.PeerID, bool)
	UpdatePeerIDPublicKeyPairCalled func(pid core.PeerID, pk []byte)
}

// UpdatePeerIDPublicKeyPair -
func (psms *PeerShardMapperStub) UpdatePeerIDPublicKeyPair(pid core.PeerID, pk []byte) {
	if psms.UpdatePeerIDPublicKeyPairCalled != nil {
		psms.UpdatePeerIDPublicKeyPairCalled(pid, pk)
	}
}

// GetPeerID -
func (psms *PeerShardMapperStub) GetPeerID(pk []byte) (*core.PeerID, bool) {
	if psms.GetPeerIDCalled != nil {
		return psms.GetPeerIDCalled(pk)
	}

	return nil, false
}

// GetPeerInfo -
func (psms *PeerShardMapperStub) GetPeerInfo(_ core.PeerID) core.P2PPeerInfo {
	return core.P2PPeerInfo{}
}

// IsInterfaceNil -
func (psms *PeerShardMapperStub) IsInterfaceNil() bool {
	return psms == nil
}
