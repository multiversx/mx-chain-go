package mock

import "github.com/ElrondNetwork/elrond-go/process"

type InterceptorProcessorStub struct {
	ValidateCalled func(data process.InterceptedData) error
	SaveCalled     func(data process.InterceptedData) error
}

func (ips *InterceptorProcessorStub) Validate(data process.InterceptedData) error {
	return ips.ValidateCalled(data)
}

func (ips *InterceptorProcessorStub) Save(data process.InterceptedData) error {
	return ips.SaveCalled(data)
}

func (ips *InterceptorProcessorStub) IsInterfaceNil() bool {
	if ips == nil {
		return true
	}
	return false
}
