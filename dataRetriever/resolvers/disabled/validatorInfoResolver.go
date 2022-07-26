package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type validatorInfoResolver struct {
}

// NewDisabledValidatorInfoResolver creates a new disabled validator info resolver instance
func NewDisabledValidatorInfoResolver() *validatorInfoResolver {
	return &validatorInfoResolver{}
}

// RequestDataFromHash does nothing and returns nil
func (res *validatorInfoResolver) RequestDataFromHash(_ []byte, _ uint32) error {
	return nil
}

// RequestDataFromHashArray does nothing and returns nil
func (res *validatorInfoResolver) RequestDataFromHashArray(_ [][]byte, _ uint32) error {
	return nil
}

// ProcessReceivedMessage does nothing and returns nil
func (res *validatorInfoResolver) ProcessReceivedMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	return nil
}

// SetResolverDebugHandler does nothing and returns nil
func (res *validatorInfoResolver) SetResolverDebugHandler(_ dataRetriever.ResolverDebugHandler) error {
	return nil
}

// SetNumPeersToQuery does nothing
func (res *validatorInfoResolver) SetNumPeersToQuery(_ int, _ int) {
}

// NumPeersToQuery returns 0 and 0
func (res *validatorInfoResolver) NumPeersToQuery() (int, int) {
	return 0, 0
}

// Close does nothing and returns nil
func (res *validatorInfoResolver) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (res *validatorInfoResolver) IsInterfaceNil() bool {
	return res == nil
}
