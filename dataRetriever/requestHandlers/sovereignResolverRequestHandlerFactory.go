package requestHandlers

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignResolverRequestHandlerFactory struct {
	resolverRequestFactoryHandler RequestHandlerCreator
}

// NewSovereignResolverRequestHandlerFactory creates a new resolver request handler factory for chain run type sovereign
func NewSovereignResolverRequestHandlerFactory(re RequestHandlerCreator) (RequestHandlerCreator, error) {
	if check.IfNil(re) {
		return nil, errors.ErrNilResolverRequestFactoryHandler
	}
	return &sovereignResolverRequestHandlerFactory{
		resolverRequestFactoryHandler: re,
	}, nil
}

// CreateRequestHandler creates a RequestHandler for chain run type sovereign
func (rrh *sovereignResolverRequestHandlerFactory) CreateRequestHandler(resolverRequestArgs RequestHandlerArgs) (process.RequestHandler, error) {
	requestHandler, err := rrh.resolverRequestFactoryHandler.CreateRequestHandler(resolverRequestArgs)
	if err != nil {
		return nil, err
	}

	resolverRequester, ok := requestHandler.(*resolverRequestHandler)
	if !ok {
		return nil, errors.ErrInvalidTypeConversion
	}

	return NewSovereignResolverRequestHandler(resolverRequester)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rrh *sovereignResolverRequestHandlerFactory) IsInterfaceNil() bool {
	return rrh == nil
}
