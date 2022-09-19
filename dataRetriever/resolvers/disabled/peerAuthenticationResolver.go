package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type peerAuthenticatorResolver struct {
}

// NewDisabledPeerAuthenticatorResolver creates a new disabled peer authentication resolver instance
func NewDisabledPeerAuthenticatorResolver() *peerAuthenticatorResolver {
	return &peerAuthenticatorResolver{}
}

// RequestDataFromHash does nothing and returns nil
func (res *peerAuthenticatorResolver) RequestDataFromHash(_ []byte, _ uint32) error {
	return nil
}

// RequestDataFromHashArray does nothing and returns nil
func (res *peerAuthenticatorResolver) RequestDataFromHashArray(_ [][]byte, _ uint32) error {
	return nil
}

// ProcessReceivedMessage does nothing and returns nil
func (res *peerAuthenticatorResolver) ProcessReceivedMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	return nil
}

// SetResolverDebugHandler does nothing and returns nil
func (res *peerAuthenticatorResolver) SetResolverDebugHandler(_ dataRetriever.ResolverDebugHandler) error {
	return nil
}

// SetNumPeersToQuery does nothing
func (res *peerAuthenticatorResolver) SetNumPeersToQuery(_ int, _ int) {
}

// NumPeersToQuery returns 0 and 0
func (res *peerAuthenticatorResolver) NumPeersToQuery() (int, int) {
	return 0, 0
}

// Close does nothing and returns nil
func (res *peerAuthenticatorResolver) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *peerAuthenticatorResolver) IsInterfaceNil() bool {
	return res == nil
}
