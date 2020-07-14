package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// NetworkShardingCollectorStub -
type NetworkShardingCollectorStub struct {
	UpdatePeerIdPublicKeyCalled       func(pid core.PeerID, pk []byte)
	UpdatePublicKeyShardIdCalled      func(pk []byte, shardId uint32)
	UpdatePeerIdShardIdCalled         func(pid core.PeerID, shardId uint32)
	UpdatePublicKeyPIDSignatureCalled func(pk []byte, pid []byte, signature []byte)
	GetPidAndSignatureFromPkCalled    func(pk []byte) (pid []byte, signature []byte)
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

// UpdatePublicKeyPIDSignature -
func (nscs *NetworkShardingCollectorStub) UpdatePublicKeyPIDSignature(pk []byte, pid []byte, signature []byte) {
	nscs.UpdatePublicKeyPIDSignatureCalled(pk, pid, signature)
}

// GetPidAndSignatureFromPk -
func (nscs *NetworkShardingCollectorStub) GetPidAndSignatureFromPk(pk []byte) (pid []byte, signature []byte) {
	return nscs.GetPidAndSignatureFromPkCalled(pk)
}

// IsInterfaceNil -
func (nscs *NetworkShardingCollectorStub) IsInterfaceNil() bool {
	return nscs == nil
}
