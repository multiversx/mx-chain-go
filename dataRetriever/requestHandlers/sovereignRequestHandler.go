package requestHandlers

import "github.com/multiversx/mx-chain-go/process"

type sovereignResolverRequestHandler struct {
	*resolverRequestHandler
}

// NewSovereignResolverRequestHandler creates a sovereignRequestHandler interface implementation with request functions
func NewSovereignResolverRequestHandler(resolverRequestHandler *resolverRequestHandler) (*sovereignResolverRequestHandler, error) {
	if resolverRequestHandler == nil {
		return nil, process.ErrNilRequestHandler
	}

	srrh := &sovereignResolverRequestHandler{
		resolverRequestHandler,
	}

	return srrh, nil
}

// RequestExtendedShardHeaderByNonce method asks for extended shard header from the connected peers by nonce
func (srrh *sovereignResolverRequestHandler) RequestExtendedShardHeaderByNonce(nonce uint64) {
	//TODO: This method should be implemented for sovereign chain
}
