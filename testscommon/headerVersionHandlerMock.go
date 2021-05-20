package testscommon

import "github.com/ElrondNetwork/elrond-go/data"

// HeaderVersionHandlerMock -
type HeaderVersionHandlerMock struct {
	GetVersionCalled     func(epoch uint32) string
	VerifyCalled         func(hdr data.HeaderHandler) error
	IsInterfaceNilCalled func() bool
}

// GetVersion -
func (hvm *HeaderVersionHandlerMock) GetVersion(epoch uint32) string {
	if hvm.GetVersionCalled != nil {
		return hvm.GetVersionCalled(epoch)
	}
	return "*"
}

// Verify -
func (hvm *HeaderVersionHandlerMock) Verify(hdr data.HeaderHandler) error {
	if hvm.VerifyCalled != nil {
		return hvm.VerifyCalled(hdr)
	}
	return nil
}

// IsInterfaceNil -
func (hvm *HeaderVersionHandlerMock) IsInterfaceNil() bool {
	return hvm == nil
}
