package requestHandlers

import (
	"github.com/multiversx/mx-chain-go/process"
)

type resolverRequestHandlerFactory struct {
}

// NewResolverRequestHandlerFactory creates a new resolver request handler factory for chain run type normal
func NewResolverRequestHandlerFactory() (RequestHandlerCreator, error) {
	return &resolverRequestHandlerFactory{}, nil
}

// CreateRequestHandler creates a RequestHandler for chain run type normal
func (rrh *resolverRequestHandlerFactory) CreateRequestHandler(resolverRequestArgs RequestHandlerArgs) (process.RequestHandler, error) {
	return NewResolverRequestHandler(
		resolverRequestArgs.RequestersFinder,
		resolverRequestArgs.RequestedItemsHandler,
		resolverRequestArgs.WhiteListHandler,
		resolverRequestArgs.MaxTxsToRequest,
		resolverRequestArgs.ShardID,
		resolverRequestArgs.RequestInterval,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rrh *resolverRequestHandlerFactory) IsInterfaceNil() bool {
	return rrh == nil
}
