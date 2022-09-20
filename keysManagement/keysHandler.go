package keysManagement

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// keysHandler will manage all keys available on the node either in single signer mode or multi key mode
// TODO (next PR) add implementation & unit tests
type keysHandler struct {
}

// NewKeysHandler will create a new instance of type keysHandler
func NewKeysHandler() *keysHandler {
	return &keysHandler{}
}

// GetHandledPrivateKey will return the correct private key by using the provided pkBytes to select from a stored one
// Returns the current private key if the pkBytes is not handled by the current node
func (handler *keysHandler) GetHandledPrivateKey(pkBytes []byte) crypto.PrivateKey {
	return nil
}

// GetP2PIdentity returns the associated p2p identity with the provided public key bytes: the private key and the peer ID
func (handler *keysHandler) GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error) {
	return make([]byte, 0), "", nil
}

// IsKeyManagedByCurrentNode will return if the provided key is a managed one and the current node should use it
func (handler *keysHandler) IsKeyManagedByCurrentNode(pkBytes []byte) bool {
	return false
}

// IncrementRoundsWithoutReceivedMessages increments the provided rounds without received messages counter on the provided public key
func (handler *keysHandler) IncrementRoundsWithoutReceivedMessages(pkBytes []byte) {
}

// GetAssociatedPid will return the associated peer ID from the provided public key bytes. Will search in the managed keys mapping
// if the public key is not the original public key of the node
func (handler *keysHandler) GetAssociatedPid(pkBytes []byte) core.PeerID {
	return ""
}

// IsOriginalPublicKeyOfTheNode returns true if the provided public key bytes are the original ones used by the node
func (handler *keysHandler) IsOriginalPublicKeyOfTheNode(pkBytes []byte) bool {
	return true
}

// UpdatePublicKeyLiveness update the provided public key liveness if the provided pid is not managed by the current node
func (handler *keysHandler) UpdatePublicKeyLiveness(pkBytes []byte, pid core.PeerID) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *keysHandler) IsInterfaceNil() bool {
	return handler == nil
}
