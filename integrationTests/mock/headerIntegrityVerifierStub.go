package mock

import "github.com/ElrondNetwork/elrond-go/data"

// HeaderIntegrityVerifierStub -
type HeaderIntegrityVerifierStub struct {
	VerifyCalled     func(header data.HeaderHandler) error
	GetVersionCalled func(epoch uint32) string
}

// Verify -
func (h *HeaderIntegrityVerifierStub) Verify(header data.HeaderHandler) error {
	if h.VerifyCalled != nil {
		return h.VerifyCalled(header)
	}

	return nil
}

// GetVersion -
func (h *HeaderIntegrityVerifierStub) GetVersion(epoch uint32) string {
	if h.GetVersionCalled != nil {
		return h.GetVersionCalled(epoch)
	}

	return "version"
}

// IsInterfaceNil -
func (h *HeaderIntegrityVerifierStub) IsInterfaceNil() bool {
	return h == nil
}
