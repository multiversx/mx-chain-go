package mock

import (
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptedDataVerifierMock -
type InterceptedDataVerifierMock struct {
	VerifyCalled       func(interceptedData process.InterceptedData, topic string) error
	MarkVerifiedCalled func(interceptedData process.InterceptedData, topic string)
}

// Verify -
func (idv *InterceptedDataVerifierMock) Verify(interceptedData process.InterceptedData, topic string) error {
	if idv.VerifyCalled != nil {
		return idv.VerifyCalled(interceptedData, topic)
	}

	return nil
}

// MarkVerified -
func (idv *InterceptedDataVerifierMock) MarkVerified(interceptedData process.InterceptedData, topic string) {
	if idv.MarkVerifiedCalled != nil {
		idv.MarkVerifiedCalled(interceptedData, topic)
	}
}

// IsInterfaceNil -
func (idv *InterceptedDataVerifierMock) IsInterfaceNil() bool {
	return idv == nil
}
