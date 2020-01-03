package libp2p

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
)

// Libp2pContext holds the context for the wrappers over libp2p implementation
type Libp2pContext struct {
	ctx                     context.Context
	connHost                ConnectableHost
	mutChangeableComponents sync.RWMutex
	blacklistHandler        p2p.BlacklistHandler
}

// NewLibp2pContext constructs a new Libp2pContext object
func NewLibp2pContext(ctx context.Context, connHost ConnectableHost) (*Libp2pContext, error) {
	if ctx == nil {
		return nil, p2p.ErrNilContext
	}
	if check.IfNil(connHost) {
		return nil, p2p.ErrNilHost
	}

	return &Libp2pContext{
		ctx:              ctx,
		connHost:         connHost,
		blacklistHandler: &mock.NilBlacklistHandler{},
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

// SetPeerBlacklist sets the peer black list handler
func (lctx *Libp2pContext) SetPeerBlacklist(blacklistHandler p2p.BlacklistHandler) error {
	if check.IfNil(blacklistHandler) {
		return p2p.ErrNilPeerBlacklistHandler
	}

	lctx.mutChangeableComponents.Lock()
	lctx.blacklistHandler = blacklistHandler
	lctx.mutChangeableComponents.Unlock()

	return nil
}

func (lctx *Libp2pContext) PeerBlacklist() p2p.BlacklistHandler {
	lctx.mutChangeableComponents.RLock()
	defer lctx.mutChangeableComponents.RUnlock()

	return lctx.blacklistHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (lctx *Libp2pContext) IsInterfaceNil() bool {
	return lctx == nil
}
