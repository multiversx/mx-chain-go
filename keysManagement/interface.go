package keysManagement

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// P2PIdentityGenerator defines the operations of an entity that can generate P2P identities
type P2PIdentityGenerator interface {
	CreateRandomP2PIdentity() ([]byte, core.PeerID, error)
	IsInterfaceNil() bool
}
