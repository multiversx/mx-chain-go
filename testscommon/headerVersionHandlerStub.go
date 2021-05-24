package testscommon

import "github.com/ElrondNetwork/elrond-go/data"

// HeaderVersionHandlerStub -
type HeaderVersionHandlerStub struct {
	GetVersionCalled     func(epoch uint32) string
	VerifyCalled         func(hdr data.HeaderHandler) error
	IsInterfaceNilCalled func() bool
}

// GetVersion -
func (hvm *HeaderVersionHandlerStub) GetVersion(epoch uint32) string {
	if hvm.GetVersionCalled != nil {
		return hvm.GetVersionCalled(epoch)
	}
	return "*"
}

// Verify -
func (hvm *HeaderVersionHandlerStub) Verify(hdr data.HeaderHandler) error {
	if hvm.VerifyCalled != nil {
		return hvm.VerifyCalled(hdr)
	}
	return nil
}

// IsInterfaceNil -
func (hvm *HeaderVersionHandlerStub) IsInterfaceNil() bool {
	return hvm == nil
}
