package mock

import "github.com/ElrondNetwork/elrond-go/data"

// HeaderIntegrityVerifierStub -
type HeaderIntegrityVerifierStub struct {
	VerifyCalled func(header data.HeaderHandler) error
}

// Verify -
func (h *HeaderIntegrityVerifierStub) Verify(header data.HeaderHandler) error {
	if h.VerifyCalled != nil {
		return h.VerifyCalled(header)
	}

	return nil
}

// IsInterfaceNil -
func (h *HeaderIntegrityVerifierStub) IsInterfaceNil() bool {
	return h == nil
}
