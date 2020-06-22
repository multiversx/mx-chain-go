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
}

// UpdatePeerIdPublicKey -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdPublicKey(pid core.PeerID, pk []byte) {
	nscs.UpdatePeerIdPublicKeyCalled(pid, pk)
}

// UpdatePublicKeyShardId -
func (nscs *NetworkShardingCollectorStub) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	nscs.UpdatePublicKeyShardIdCalled(pk, shardId)
}

// UpdatePeerIdShardId -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdShardId(pid core.PeerID, shardId uint32) {
	nscs.UpdatePeerIdShardIdCalled(pid, shardId)
}

// GetPeerInfo -
func (nscs *NetworkShardingCollectorStub) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	return nscs.GetPeerInfoCalled(pid)
}

// IsInterfaceNil -
func (nscs *NetworkShardingCollectorStub) IsInterfaceNil() bool {
	return nscs == nil
}
