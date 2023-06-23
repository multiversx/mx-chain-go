package requestHandlers

import (
	"github.com/multiversx/mx-chain-go/process"
)

type resolverRequestHandlerFactory struct {
}

// NewResolverRequestHandlerFactory creates a new resolver request handler factory for chain run type normal
func NewResolverRequestHandlerFactory() (ResolverRequestFactoryHandler, error) {
	return &resolverRequestHandlerFactory{}, nil
}

// CreateResolverRequestHandler creates a RequestHandler for chain run type normal
func (rrh *resolverRequestHandlerFactory) CreateResolverRequestHandler(resolverRequestArgs ResolverRequestArgs) (process.RequestHandler, error) {
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
