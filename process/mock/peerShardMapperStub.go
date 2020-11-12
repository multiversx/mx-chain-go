package mock

import "github.com/ElrondNetwork/elrond-go/core"

// PeerShardMapperStub -
type PeerShardMapperStub struct {
	GetPeerInfoCalled            func(pid core.PeerID) core.P2PPeerInfo
	UpdatePeerIdPublicKeyCalled  func(pid core.PeerID, pk []byte)
	UpdatePublicKeyShardIdCalled func(pk []byte, shardId uint32)
	UpdatePeerIdShardIdCalled    func(pid core.PeerID, shardId uint32)
	UpdatePeerIdSubTypeCalled    func(pid core.PeerID, peerSubType core.P2PPeerSubType)
}

// GetPeerInfo -
func (psms *PeerShardMapperStub) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	if psms.GetPeerInfoCalled != nil {
		return psms.GetPeerInfoCalled(pid)
	}

	return core.P2PPeerInfo{}
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

// UpdatePeerIdShardId -
func (psms *PeerShardMapperStub) UpdatePeerIdShardId(pid core.PeerID, shardId uint32) {
	if psms.UpdatePeerIdShardIdCalled != nil {
		psms.UpdatePeerIdShardIdCalled(pid, shardId)
	}
}

// UpdatePeerIdSubType
func (psms *PeerShardMapperStub) UpdatePeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	psms.UpdatePeerIdSubTypeCalled(pid, peerSubType)
}

// IsInterfaceNil -
func (psms *PeerShardMapperStub) IsInterfaceNil() bool {
	return psms == nil
}
