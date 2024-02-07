package keysManagement

import (
	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
)

// GetRoundsOfInactivity -
func (pInfo *peerInfo) GetRoundsOfInactivity() int {
	pInfo.mutChangeableData.RLock()
	defer pInfo.mutChangeableData.RUnlock()

	return pInfo.handler.RoundsOfInactivity()
}

// Pid -
func (pInfo *peerInfo) Pid() core.PeerID {
	return pInfo.pid
}

// P2pPrivateKeyBytes -
func (pInfo *peerInfo) P2pPrivateKeyBytes() []byte {
	return pInfo.p2pPrivateKeyBytes
}

// PrivateKey -
func (pInfo *peerInfo) PrivateKey() crypto.PrivateKey {
	return pInfo.privateKey
}

// MachineID -
func (pInfo *peerInfo) MachineID() string {
	return pInfo.machineID
}

// NodeName -
func (pInfo *peerInfo) NodeName() string {
	return pInfo.nodeName
}

// NodeIdentity -
func (pInfo *peerInfo) NodeIdentity() string {
	return pInfo.nodeIdentity
}

// GetPeerInfo -
func (holder *managedPeersHolder) GetPeerInfo(pkBytes []byte) *peerInfo {
	return holder.getPeerInfo(pkBytes)
}

// ManagedPeersHolder -
func (handler *keysHandler) ManagedPeersHolder() common.ManagedPeersHolder {
	return handler.managedPeersHolder
}

// PrivateKey -
func (handler *keysHandler) PrivateKey() crypto.PrivateKey {
	return handler.privateKey
}

// PublicKey -
func (handler *keysHandler) PublicKey() crypto.PublicKey {
	return handler.publicKey
}

// PublicKeyBytes -
func (handler *keysHandler) PublicKeyBytes() []byte {
	return handler.publicKeyBytes
}

// Pid -
func (handler *keysHandler) Pid() core.PeerID {
	return handler.pid
}
