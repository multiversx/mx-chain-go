package factory

import (
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/process"
)

// RequestHandlerFactoryStub -
type RequestHandlerFactoryStub struct {
	CreateRequestHandlerCalled func(args requestHandlers.RequestHandlerArgs) (process.RequestHandler, error)
}

// NewRequestHandlerFactoryStub -
func NewRequestHandlerFactoryStub() *RequestHandlerFactoryStub {
	return &RequestHandlerFactoryStub{}
}

// CreateRequestHandler -
func (r *RequestHandlerFactoryStub) CreateRequestHandler(args requestHandlers.RequestHandlerArgs) (process.RequestHandler, error) {
	if r.CreateRequestHandlerCalled != nil {
		return r.CreateRequestHandlerCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (r *RequestHandlerFactoryStub) IsInterfaceNil() bool {
	return false
}
