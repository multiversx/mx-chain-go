package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.InterceptedHeaderIntegrityVerifier = (*headerIntegrityVerifier)(nil)

type headerIntegrityVerifier struct {
}

// NewHeaderIntegrityVerifier returns a new instance of headerIntegrityVerifier
func NewHeaderIntegrityVerifier() *headerIntegrityVerifier {
	return &headerIntegrityVerifier{}
}

// Verify returns nil as this is a disabled implementation
func (h *headerIntegrityVerifier) Verify(header data.HeaderHandler, referenceChainID []byte) error {
	return nil
}

// IsInterfaceNil returns if there is no value under the interface
func (h *headerIntegrityVerifier) IsInterfaceNil() bool {
	return h == nil
}
