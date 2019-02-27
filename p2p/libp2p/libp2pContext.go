package libp2p

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// Libp2pContext holds the context for the wrappers over libp2p implementation
type Libp2pContext struct {
	ctx     context.Context
	updHost UpgradedHost
}

// NewLibp2pContext constructs a new Libp2pContext object
func NewLibp2pContext(ctx context.Context, updHost UpgradedHost) (*Libp2pContext, error) {
	if ctx == nil {
		return nil, p2p.ErrNilContext
	}

	if updHost == nil {
		return nil, p2p.ErrNilHost
	}

	return &Libp2pContext{
		ctx:     ctx,
		updHost: updHost,
	}, nil
}

// Context returns the context associated with this struct
func (lctx *Libp2pContext) Context() context.Context {
	return lctx.ctx
}

// Host returns the upgraded host
func (lctx *Libp2pContext) Host() UpgradedHost {
	return lctx.updHost
}
