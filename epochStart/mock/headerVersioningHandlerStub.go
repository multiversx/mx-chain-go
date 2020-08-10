package mock

import "github.com/ElrondNetwork/elrond-go/data"

// HeaderVersioningHandlerStub -
type HeaderVersioningHandlerStub struct {
	VerifyCalled     func(header data.HeaderHandler) error
	GetVersionCalled func(epoch uint32) string
}

// Verify -
func (hvhs *HeaderVersioningHandlerStub) Verify(header data.HeaderHandler) error {
	if hvhs.VerifyCalled != nil {
		return hvhs.VerifyCalled(header)
	}

	return nil
}

// GetVersion -
func (hvhs *HeaderVersioningHandlerStub) GetVersion(epoch uint32) string {
	if hvhs.GetVersionCalled != nil {
		return hvhs.GetVersionCalled(epoch)
	}

	return "version"
}

// IsInterfaceNil -
func (hvhs *HeaderVersioningHandlerStub) IsInterfaceNil() bool {
	return hvhs == nil
}
