package mock

import (
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptedDataVerifierFactoryStub -
type InterceptedDataVerifierFactoryStub struct {
	CreateCalled func(topic string) (process.InterceptedDataVerifier, error)
}

// Create -
func (idvfs *InterceptedDataVerifierFactoryStub) Create(topic string) (process.InterceptedDataVerifier, error) {
	if idvfs.CreateCalled != nil {
		return idvfs.CreateCalled(topic)
	}

	return nil, nil
}

// IsInterfaceNil -
func (idvfs *InterceptedDataVerifierFactoryStub) IsInterfaceNil() bool {
	return idvfs == nil
}
