package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// NetworkShardingCollectorStub -
type NetworkShardingCollectorStub struct {
	UpdatePeerIdPublicKeyCalled  func(pid core.PeerID, pk []byte)
	UpdatePublicKeyShardIdCalled func(pk []byte, shardId uint32)
	UpdatePeerIdShardIdCalled    func(pid core.PeerID, shardId uint32)
	GetPeerInfoCalled            func(pid core.PeerID) core.P2PPeerInfo
	UpdatePeerIdSubTypeCalled    func(pid core.PeerID, peerSubType core.P2PPeerSubType)
}

// UpdatePeerIdPublicKey -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdPublicKey(pid core.PeerID, pk []byte) {
	if nscs.UpdatePeerIdPublicKeyCalled != nil {
		nscs.UpdatePeerIdPublicKeyCalled(pid, pk)
	}
}

// UpdatePublicKeyShardId -
func (nscs *NetworkShardingCollectorStub) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	if nscs.UpdatePublicKeyShardIdCalled != nil {
		nscs.UpdatePublicKeyShardIdCalled(pk, shardId)
	}
}

// UpdatePeerIdShardId -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdShardId(pid core.PeerID, shardId uint32) {
	if nscs.UpdatePeerIdShardIdCalled != nil {
		nscs.UpdatePeerIdShardIdCalled(pid, shardId)
	}
}

// GetPeerInfo -
func (nscs *NetworkShardingCollectorStub) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	if nscs.GetPeerInfoCalled != nil {
		return nscs.GetPeerInfoCalled(pid)
	}
	return core.P2PPeerInfo{}
}

// UpdatePeerIdSubType -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	if nscs.UpdatePeerIdSubTypeCalled != nil {
		nscs.UpdatePeerIdSubTypeCalled(pid, peerSubType)
	}
}

// IsInterfaceNil -
func (nscs *NetworkShardingCollectorStub) IsInterfaceNil() bool {
	return nscs == nil
}
