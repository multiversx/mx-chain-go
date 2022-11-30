package keysManagement

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// P2PIdentityGenerator defines the operations of an entity that can generate P2P identities
type P2PIdentityGenerator interface {
	CreateRandomP2PIdentity() ([]byte, core.PeerID, error)
	IsInterfaceNil() bool
}

// ManagedPeersHolder defines the operations of an entity that holds managed identities for a node
type ManagedPeersHolder interface {
	GetPrivateKey(pkBytes []byte) (crypto.PrivateKey, error)
	GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error)
	GetNameAndIdentity(pkBytes []byte) (string, string, error)
	IncrementRoundsWithoutReceivedMessages(pkBytes []byte)
	ResetRoundsWithoutReceivedMessages(pkBytes []byte)
	IsKeyManagedByCurrentNode(pkBytes []byte) bool
	IsInterfaceNil() bool
}
