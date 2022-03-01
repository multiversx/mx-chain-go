package p2pmocks

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// NetworkShardingCollectorStub -
type NetworkShardingCollectorStub struct {
	UpdatePeerIDPublicKeyPairCalled func(pid core.PeerID, pk []byte)
	UpdatePeerIDInfoCalled          func(pid core.PeerID, pk []byte, shardID uint32)
	UpdatePeerIdSubTypeCalled       func(pid core.PeerID, peerSubType core.P2PPeerSubType)
	GetPeerIDCalled                 func(pk []byte) (*core.PeerID, bool)
	GetPeerInfoCalled               func(pid core.PeerID) core.P2PPeerInfo
}

// UpdatePeerIDPublicKeyPair -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIDPublicKeyPair(pid core.PeerID, pk []byte) {
	if nscs.UpdatePeerIDPublicKeyPairCalled != nil {
		nscs.UpdatePeerIDPublicKeyPairCalled(pid, pk)
	}
}

// UpdatePeerIDInfo -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32) {
	if nscs.UpdatePeerIDInfoCalled != nil {
		nscs.UpdatePeerIDInfoCalled(pid, pk, shardID)
	}
}

// UpdatePeerIdSubType -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	if nscs.UpdatePeerIdSubTypeCalled != nil {
		nscs.UpdatePeerIdSubTypeCalled(pid, peerSubType)
	}
}

// GetPeerID -
func (nscs *NetworkShardingCollectorStub) GetPeerID(pk []byte) (*core.PeerID, bool) {
	if nscs.GetPeerIDCalled != nil {
		return nscs.GetPeerIDCalled(pk)
	}

	return nil, false
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
