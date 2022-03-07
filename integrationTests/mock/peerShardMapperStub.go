package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// PeerShardMapperStub -
type PeerShardMapperStub struct {
	GetLastKnownPeerIDCalled        func(pk []byte) (*core.PeerID, bool)
	UpdatePeerIDPublicKeyPairCalled func(pid core.PeerID, pk []byte)
	UpdatePeerIdShardIdCalled       func(pid core.PeerID, shardID uint32)
	UpdatePeerIdSubTypeCalled       func(pid core.PeerID, peerSubType core.P2PPeerSubType)
}

// UpdatePeerIDPublicKeyPair -
func (psms *PeerShardMapperStub) UpdatePeerIDPublicKeyPair(pid core.PeerID, pk []byte) {
	if psms.UpdatePeerIDPublicKeyPairCalled != nil {
		psms.UpdatePeerIDPublicKeyPairCalled(pid, pk)
	}
}

// UpdatePeerIdShardId -
func (psms *PeerShardMapperStub) UpdatePeerIdShardId(pid core.PeerID, shardID uint32) {
	if psms.UpdatePeerIdShardIdCalled != nil {
		psms.UpdatePeerIdShardIdCalled(pid, shardID)
	}
}

// UpdatePeerIdSubType -
func (psms *PeerShardMapperStub) UpdatePeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	if psms.UpdatePeerIdSubTypeCalled != nil {
		psms.UpdatePeerIdSubTypeCalled(pid, peerSubType)
	}
}

// GetLastKnownPeerID -
func (psms *PeerShardMapperStub) GetLastKnownPeerID(pk []byte) (*core.PeerID, bool) {
	if psms.GetLastKnownPeerIDCalled != nil {
		return psms.GetLastKnownPeerIDCalled(pk)
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
