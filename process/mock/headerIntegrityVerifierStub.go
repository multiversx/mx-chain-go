package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// HeaderIntegrityVerifierStub -
type HeaderIntegrityVerifierStub struct {
	VerifyCalled func(header data.HeaderHandler, referenceChainID []byte) error
}

// Verify -
func (h *HeaderIntegrityVerifierStub) Verify(header data.HeaderHandler, referenceChainID []byte) error {
	if h.VerifyCalled != nil {
		return h.VerifyCalled(header, referenceChainID)
	}

	return nil
}

// IsInterfaceNil -
func (h *HeaderIntegrityVerifierStub) IsInterfaceNil() bool {
	return h == nil
}
