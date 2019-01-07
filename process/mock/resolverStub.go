package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type ResolverStub struct {
	RequestDataCalled        func(rd process.RequestData) error
	SetResolverHandlerCalled func(func(rd process.RequestData) ([]byte, error))
	ResolverHandlerCalled    func() func(rd process.RequestData) ([]byte, error)
}

func (rs *ResolverStub) RequestData(rd process.RequestData) error {
	return rs.RequestDataCalled(rd)
}

func (rs *ResolverStub) SetResolverHandler(handler func(rd process.RequestData) ([]byte, error)) {
	rs.SetResolverHandlerCalled(handler)
}

func (rs *ResolverStub) ResolverHandler() func(rd process.RequestData) ([]byte, error) {
	return rs.ResolverHandlerCalled()
}
