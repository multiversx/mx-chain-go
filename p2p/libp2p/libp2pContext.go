package libp2p

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

// Libp2pContext holds the context for the wrappers over libp2p implementation
type Libp2pContext struct {
	ctx      context.Context
	connHost ConnectableHost
}

// NewLibp2pContext constructs a new Libp2pContext object
func NewLibp2pContext(ctx context.Context, connHost ConnectableHost) (*Libp2pContext, error) {
	if ctx == nil {
		return nil, p2p.ErrNilContext
	}

	if connHost == nil || connHost.IsInterfaceNil() {
		return nil, p2p.ErrNilHost
	}

	return &Libp2pContext{
		ctx:      ctx,
		connHost: connHost,
	}, nil
}

// Context returns the context associated with this struct
func (lctx *Libp2pContext) Context() context.Context {
	return lctx.ctx
}

// Host returns the connectable host
func (lctx *Libp2pContext) Host() ConnectableHost {
	return lctx.connHost
}

// IsInterfaceNil returns true if there is no value under the interface
func (lctx *Libp2pContext) IsInterfaceNil() bool {
	if lctx == nil {
		return true
	}
	return false
}
