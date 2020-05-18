package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// InterceptorProcessorStub -
type InterceptorProcessorStub struct {
	ValidateCalled func(data process.InterceptedData) error
	SaveCalled     func(data process.InterceptedData) error
}

// Validate -
func (ips *InterceptorProcessorStub) Validate(data process.InterceptedData, _ p2p.PeerID) error {
	return ips.ValidateCalled(data)
}

// Save -
func (ips *InterceptorProcessorStub) Save(data process.InterceptedData, _ p2p.PeerID) error {
	return ips.SaveCalled(data)
}

// IsInterfaceNil -
func (ips *InterceptorProcessorStub) IsInterfaceNil() bool {
	return ips == nil
}
