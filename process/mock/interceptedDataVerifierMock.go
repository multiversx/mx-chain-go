package mock

import (
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptedDataVerifierStub -
type InterceptedDataVerifierStub struct {
	VerifyCalled func(interceptedData process.InterceptedData) error
}

// Verify -
func (idv *InterceptedDataVerifierStub) Verify(interceptedData process.InterceptedData) error {
	if idv.VerifyCalled != nil {
		return idv.VerifyCalled(interceptedData)
	}

	return nil
}

// IsInterfaceNil -
func (idv *InterceptedDataVerifierStub) IsInterfaceNil() bool {
	return idv == nil
}
