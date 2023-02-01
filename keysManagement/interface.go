package keysManagement

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// ManagedPeersHolder defines the operations of an entity that holds managed identities for a node
type ManagedPeersHolder interface {
	AddManagedPeer(privateKeyBytes []byte) error
	GetPrivateKey(pkBytes []byte) (crypto.PrivateKey, error)
	GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error)
	GetMachineID(pkBytes []byte) (string, error)
	GetNameAndIdentity(pkBytes []byte) (string, string, error)
	IncrementRoundsWithoutReceivedMessages(pkBytes []byte)
	ResetRoundsWithoutReceivedMessages(pkBytes []byte)
	GetManagedKeysByCurrentNode() map[string]crypto.PrivateKey
	IsKeyManagedByCurrentNode(pkBytes []byte) bool
	IsKeyRegistered(pkBytes []byte) bool
	IsPidManagedByCurrentNode(pid core.PeerID) bool
	IsKeyValidator(pkBytes []byte) bool
	SetValidatorState(pkBytes []byte, state bool)
	GetNextPeerAuthenticationTime(pkBytes []byte) (time.Time, error)
	SetNextPeerAuthenticationTime(pkBytes []byte, nextTime time.Time)
	IsMultiKeyMode() bool
	IsInterfaceNil() bool
}
