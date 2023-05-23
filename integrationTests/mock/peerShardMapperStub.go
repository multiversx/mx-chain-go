package mock

import "github.com/multiversx/mx-chain-core-go/core"

// PeerShardMapperStub -
type PeerShardMapperStub struct {
	GetLastKnownPeerIDCalled        func(pk []byte) (core.PeerID, bool)
	UpdatePeerIDPublicKeyPairCalled func(pid core.PeerID, pk []byte)
	PutPeerIdShardIdCalled          func(pid core.PeerID, shardID uint32)
	PutPeerIdSubTypeCalled          func(pid core.PeerID, peerSubType core.P2PPeerSubType)
	UpdatePeerIDInfoCalled          func(pid core.PeerID, pk []byte, shardID uint32)
}

// UpdatePeerIDInfo -
func (psms *PeerShardMapperStub) UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32) {
	if psms.UpdatePeerIDInfoCalled != nil {
		psms.UpdatePeerIDInfoCalled(pid, pk, shardID)
	}
}

// UpdatePeerIDPublicKeyPair -
func (psms *PeerShardMapperStub) UpdatePeerIDPublicKeyPair(pid core.PeerID, pk []byte) {
	if psms.UpdatePeerIDPublicKeyPairCalled != nil {
		psms.UpdatePeerIDPublicKeyPairCalled(pid, pk)
	}
}

// PutPeerIdShardId -
func (psms *PeerShardMapperStub) PutPeerIdShardId(pid core.PeerID, shardID uint32) {
	if psms.PutPeerIdShardIdCalled != nil {
		psms.PutPeerIdShardIdCalled(pid, shardID)
	}
}

// PutPeerIdSubType -
func (psms *PeerShardMapperStub) PutPeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	if psms.PutPeerIdSubTypeCalled != nil {
		psms.PutPeerIdSubTypeCalled(pid, peerSubType)
	}
}

// GetLastKnownPeerID -
func (psms *PeerShardMapperStub) GetLastKnownPeerID(pk []byte) (core.PeerID, bool) {
	if psms.GetLastKnownPeerIDCalled != nil {
		return psms.GetLastKnownPeerIDCalled(pk)
	}

	return "", false
}

// GetPeerInfo -
func (psms *PeerShardMapperStub) GetPeerInfo(_ core.PeerID) core.P2PPeerInfo {
	return core.P2PPeerInfo{}
}

// IsInterfaceNil -
func (psms *PeerShardMapperStub) IsInterfaceNil() bool {
	return psms == nil
}
