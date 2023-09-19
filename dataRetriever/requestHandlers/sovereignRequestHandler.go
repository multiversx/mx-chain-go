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

// TODO: Implement this in MX-14496

// RequestExtendedShardHeaderByNonce method asks for extended shard header from the connected peers by nonce
func (srrh *sovereignResolverRequestHandler) RequestExtendedShardHeaderByNonce(_ uint64) {
	//TODO: This method should be implemented for sovereign chain
}

// RequestExtendedShardHeader method asks for extended shard header from the connected peers by nonce
func (srrh *sovereignResolverRequestHandler) RequestExtendedShardHeader(_ []byte) {
	//TODO: This method should be implemented for sovereign chain
}
