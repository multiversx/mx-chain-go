package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
)

// InterceptorProcessorStub -
type InterceptorProcessorStub struct {
	ValidateCalled func(data process.InterceptedData) error
	SaveCalled     func(data process.InterceptedData) error
}

// Validate -
func (ips *InterceptorProcessorStub) Validate(data process.InterceptedData, _ core.PeerID) error {
	return ips.ValidateCalled(data)
}

// Save -
func (ips *InterceptorProcessorStub) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	return ips.SaveCalled(data)
}

// RegisterHandler -
func (ips *InterceptorProcessorStub) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
}

// IsInterfaceNil -
func (ips *InterceptorProcessorStub) IsInterfaceNil() bool {
	return ips == nil
}
