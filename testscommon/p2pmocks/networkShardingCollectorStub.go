package p2pmocks

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// NetworkShardingCollectorStub -
type NetworkShardingCollectorStub struct {
	UpdatePeerIDPublicKeyPairCalled func(pid core.PeerID, pk []byte)
	UpdatePeerIDInfoCalled          func(pid core.PeerID, pk []byte, shardID uint32)
	PutPeerIdShardIdCalled          func(pid core.PeerID, shardId uint32)
	PutPeerIdSubTypeCalled          func(pid core.PeerID, peerSubType core.P2PPeerSubType)
	GetLastKnownPeerIDCalled        func(pk []byte) (core.PeerID, bool)
	GetPeerInfoCalled               func(pid core.PeerID) core.P2PPeerInfo
}

// UpdatePeerIDPublicKeyPair -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIDPublicKeyPair(pid core.PeerID, pk []byte) {
	if nscs.UpdatePeerIDPublicKeyPairCalled != nil {
		nscs.UpdatePeerIDPublicKeyPairCalled(pid, pk)
	}
}

// PutPeerIdShardId -
func (nscs *NetworkShardingCollectorStub) PutPeerIdShardId(pid core.PeerID, shardID uint32) {
	if nscs.PutPeerIdShardIdCalled != nil {
		nscs.PutPeerIdShardIdCalled(pid, shardID)
	}
}

// UpdatePeerIDInfo -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32) {
	if nscs.UpdatePeerIDInfoCalled != nil {
		nscs.UpdatePeerIDInfoCalled(pid, pk, shardID)
	}
}

// PutPeerIdSubType -
func (nscs *NetworkShardingCollectorStub) PutPeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	if nscs.PutPeerIdSubTypeCalled != nil {
		nscs.PutPeerIdSubTypeCalled(pid, peerSubType)
	}
}

// GetLastKnownPeerID -
func (nscs *NetworkShardingCollectorStub) GetLastKnownPeerID(pk []byte) (core.PeerID, bool) {
	if nscs.GetLastKnownPeerIDCalled != nil {
		return nscs.GetLastKnownPeerIDCalled(pk)
	}

	return "", false
}

// GetPeerInfo -
func (nscs *NetworkShardingCollectorStub) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	if nscs.GetPeerInfoCalled != nil {
		return nscs.GetPeerInfoCalled(pid)
	}

	return core.P2PPeerInfo{}
}

// IsInterfaceNil -
func (nscs *NetworkShardingCollectorStub) IsInterfaceNil() bool {
	return nscs == nil
}
