package mock

import "github.com/multiversx/mx-chain-core-go/core"

// PeerShardMapperStub -
type PeerShardMapperStub struct {
	GetLastKnownPeerIDCalled        func(pk []byte) (core.PeerID, bool)
	GetPeerInfoCalled               func(pid core.PeerID) core.P2PPeerInfo
	UpdatePeerIdPublicKeyCalled     func(pid core.PeerID, pk []byte)
	UpdatePublicKeyShardIdCalled    func(pk []byte, shardId uint32)
	PutPeerIdShardIdCalled          func(pid core.PeerID, shardId uint32)
	UpdatePeerIDPublicKeyPairCalled func(pid core.PeerID, pk []byte)
	PutPeerIdSubTypeCalled          func(pid core.PeerID, peerSubType core.P2PPeerSubType)
}

// GetLastKnownPeerID -
func (psms *PeerShardMapperStub) GetLastKnownPeerID(pk []byte) (core.PeerID, bool) {
	if psms.GetLastKnownPeerIDCalled != nil {
		return psms.GetLastKnownPeerIDCalled(pk)
	}

	return "", false
}

// GetPeerInfo -
func (psms *PeerShardMapperStub) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	if psms.GetPeerInfoCalled != nil {
		return psms.GetPeerInfoCalled(pid)
	}

	return core.P2PPeerInfo{}
}

// UpdatePeerIDPublicKeyPair -
func (psms *PeerShardMapperStub) UpdatePeerIDPublicKeyPair(pid core.PeerID, pk []byte) {
	if psms.UpdatePeerIDPublicKeyPairCalled != nil {
		psms.UpdatePeerIDPublicKeyPairCalled(pid, pk)
	}
}

// UpdatePeerIdPublicKey -
func (psms *PeerShardMapperStub) UpdatePeerIdPublicKey(pid core.PeerID, pk []byte) {
	if psms.UpdatePeerIdPublicKeyCalled != nil {
		psms.UpdatePeerIdPublicKeyCalled(pid, pk)
	}
}

// UpdatePublicKeyShardId -
func (psms *PeerShardMapperStub) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	if psms.UpdatePublicKeyShardIdCalled != nil {
		psms.UpdatePublicKeyShardIdCalled(pk, shardId)
	}
}

// PutPeerIdShardId -
func (psms *PeerShardMapperStub) PutPeerIdShardId(pid core.PeerID, shardId uint32) {
	if psms.PutPeerIdShardIdCalled != nil {
		psms.PutPeerIdShardIdCalled(pid, shardId)
	}
}

// PutPeerIdSubType -
func (psms *PeerShardMapperStub) PutPeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	if psms.PutPeerIdSubTypeCalled != nil {
		psms.PutPeerIdSubTypeCalled(pid, peerSubType)
	}
}

// IsInterfaceNil -
func (psms *PeerShardMapperStub) IsInterfaceNil() bool {
	return psms == nil
}
