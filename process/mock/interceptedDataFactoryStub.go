package mock

import "github.com/ElrondNetwork/elrond-go/process"

type InterceptedDataFactoryStub struct {
	CreateCalled func(buff []byte) (process.InterceptedData, error)
}

func (idfs *InterceptedDataFactoryStub) Create(buff []byte) (process.InterceptedData, error) {
	return idfs.CreateCalled(buff)
}

func (idfs *InterceptedDataFactoryStub) IsInterfaceNil() bool {
	if idfs == nil {
		return true
	}
	return false
}
