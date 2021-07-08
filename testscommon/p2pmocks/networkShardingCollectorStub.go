package p2pmocks

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// NetworkShardingCollectorStub -
type NetworkShardingCollectorStub struct {
	UpdatePeerIDInfoCalled    func(pid core.PeerID, pk []byte, shardID uint32)
	UpdatePeerIdSubTypeCalled func(pid core.PeerID, peerSubType core.P2PPeerSubType)
	GetPeerInfoCalled         func(pid core.PeerID) core.P2PPeerInfo
}

// UpdatePeerIDInfo -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32) {
	if nscs.UpdatePeerIDInfoCalled != nil {
		nscs.UpdatePeerIDInfoCalled(pid, pk, shardID)
	}
}

// UpdatePeerIdSubType
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	if nscs.UpdatePeerIdSubTypeCalled != nil {
		nscs.UpdatePeerIdSubTypeCalled(pid, peerSubType)
	}
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
