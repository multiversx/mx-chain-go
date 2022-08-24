package keysManagement

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// P2PIdentityGenerator defines the operations of an entity that can generate P2P identities
type P2PIdentityGenerator interface {
	CreateRandomP2PIdentity() ([]byte, core.PeerID, error)
	IsInterfaceNil() bool
}

// KeysHolder defines the operations of an entity that holds virtual identities for a node
type KeysHolder interface {
	AddVirtualPeer(privateKeyBytes []byte) error
	GetPrivateKey(pkBytes []byte) (crypto.PrivateKey, error)
	GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error)
	GetMachineID(pkBytes []byte) (string, error)
	IncrementRoundsWithoutReceivedMessages(pkBytes []byte) error
	ResetRoundsWithoutReceivedMessages(pkBytes []byte) error
	GetManagedKeysByCurrentNode() map[string]crypto.PrivateKey
	IsKeyManagedByCurrentNode(pkBytes []byte) bool
	IsKeyRegistered(pkBytes []byte) bool
	IsPidManagedByCurrentNode(pid core.PeerID) bool
	IsKeyValidator(pkBytes []byte) (bool, error)
	SetValidatorState(pkBytes []byte, state bool) error
	GetNextPeerAuthenticationTime(pkBytes []byte) (time.Time, error)
	SetNextPeerAuthenticationTime(pkBytes []byte, nextTime time.Time) error
	IsInterfaceNil() bool
}
