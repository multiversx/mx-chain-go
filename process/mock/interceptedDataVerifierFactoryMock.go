package mock

import (
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptedDataVerifierFactoryMock -
type InterceptedDataVerifierFactoryMock struct {
	CreateCalled func(topic string) (process.InterceptedDataVerifier, error)
}

// Create -
func (idvfs *InterceptedDataVerifierFactoryMock) Create(topic string) (process.InterceptedDataVerifier, error) {
	if idvfs.CreateCalled != nil {
		return idvfs.CreateCalled(topic)
	}

	return &InterceptedDataVerifierMock{}, nil
}

// Close -
func (idvfs *InterceptedDataVerifierFactoryMock) Close() error {
	return nil
}

// IsInterfaceNil -
func (idvfs *InterceptedDataVerifierFactoryMock) IsInterfaceNil() bool {
	return idvfs == nil
}
