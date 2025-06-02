package mock

import (
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptedDataVerifierMock -
type InterceptedDataVerifierMock struct {
	VerifyCalled func(interceptedData process.InterceptedData) error
}

// Verify -
func (idv *InterceptedDataVerifierMock) Verify(interceptedData process.InterceptedData) error {
	if idv.VerifyCalled != nil {
		return idv.VerifyCalled(interceptedData)
	}

	return nil
}

// IsInterfaceNil -
func (idv *InterceptedDataVerifierMock) IsInterfaceNil() bool {
	return idv == nil
}
