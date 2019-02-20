package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type InterceptorsResolversFactoryStub struct {
	CreateInterceptorsCalled   func() error
	CreateResolversCalled      func() error
	InterceptorContainerCalled func() process.Container
	ResolverContainerCalled    func() process.ResolversContainer
}

func (irfs *InterceptorsResolversFactoryStub) CreateInterceptors() error {
	return irfs.CreateInterceptorsCalled()
}

func (irfs *InterceptorsResolversFactoryStub) CreateResolvers() error {
	return irfs.CreateResolversCalled()
}

func (irfs *InterceptorsResolversFactoryStub) InterceptorContainer() process.Container {
	return irfs.InterceptorContainerCalled()
}

func (irfs *InterceptorsResolversFactoryStub) ResolverContainer() process.ResolversContainer {
	return irfs.ResolverContainerCalled()
}
