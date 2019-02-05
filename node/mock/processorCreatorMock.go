package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// ProcessorCreatorMock is a mock implementation of ProcessorFactory
type ProcessorCreatorMock struct {
	CreateInterceptorsCalled   func() error
	CreateResolversCalled      func() error
	InterceptorContainerCalled func() process.InterceptorContainer
	ResolverContainerCalled    func() process.ResolverContainer
}

// CreateInterceptors is a mock function for creating interceptors
func (p *ProcessorCreatorMock) CreateInterceptors() error {
	return p.CreateInterceptorsCalled()
}

// CreateResolvers is a mock function for creating resolvers
func (p *ProcessorCreatorMock) CreateResolvers() error {
	return p.CreateResolversCalled()
}

// InterceptorContainer is a mock getter for the interceptor container
func (p *ProcessorCreatorMock) InterceptorContainer() process.InterceptorContainer {
	return p.InterceptorContainerCalled()
}

// ResolverContainer is a mock getter for the resolver container
func (p *ProcessorCreatorMock) ResolverContainer() process.ResolverContainer {
	return p.ResolverContainerCalled()
}
