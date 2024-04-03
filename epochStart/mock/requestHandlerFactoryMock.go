package mock

import (
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// RequestHandlerFactoryMock -
type RequestHandlerFactoryMock struct {
	CreateRequestHandlerCalled func(args requestHandlers.RequestHandlerArgs) (process.RequestHandler, error)
}

// CreateRequestHandler -
func (r *RequestHandlerFactoryMock) CreateRequestHandler(args requestHandlers.RequestHandlerArgs) (process.RequestHandler, error) {
	if r.CreateRequestHandlerCalled != nil {
		return r.CreateRequestHandlerCalled(args)
	}
	return &testscommon.RequestHandlerStub{}, nil
}

// IsInterfaceNil -
func (r *RequestHandlerFactoryMock) IsInterfaceNil() bool {
	return r == nil
}
