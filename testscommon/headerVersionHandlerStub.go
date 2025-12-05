package testscommon

import "github.com/multiversx/mx-chain-core-go/data"

// HeaderVersionHandlerStub -
type HeaderVersionHandlerStub struct {
	GetVersionCalled     func(epoch uint32, round uint64) string
	VerifyCalled         func(hdr data.HeaderHandler) error
	IsInterfaceNilCalled func() bool
}

// GetVersion -
func (hvm *HeaderVersionHandlerStub) GetVersion(epoch uint32, round uint64) string {
	if hvm.GetVersionCalled != nil {
		return hvm.GetVersionCalled(epoch, round)
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
