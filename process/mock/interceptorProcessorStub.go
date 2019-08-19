package mock

import "github.com/ElrondNetwork/elrond-go/process"

type InterceptorProcessorStub struct {
	CheckValidForProcessingCalled func(data process.InterceptedData) error
	ProcessInteceptedDataCalled   func(data process.InterceptedData) error
}

func (ips InterceptorProcessorStub) CheckValidForProcessing(data process.InterceptedData) error {
	return ips.CheckValidForProcessingCalled(data)
}

func (ips InterceptorProcessorStub) ProcessInteceptedData(data process.InterceptedData) error {
	return ips.ProcessInteceptedDataCalled(data)
}
