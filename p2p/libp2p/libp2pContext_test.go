package libp2p_test

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewLibp2pContext_NilContextShoulsErr(t *testing.T) {
	t.Parallel()

	lctx, err := libp2p.NewLibp2pContext(nil, &mock.ConnectableHostStub{})

	assert.True(t, check.IfNil(lctx))
	assert.Equal(t, p2p.ErrNilContext, err)
}

func TestNewLibp2pContext_NilHostShoulsErr(t *testing.T) {
	t.Parallel()

	lctx, err := libp2p.NewLibp2pContext(context.Background(), nil)

	assert.True(t, check.IfNil(lctx))
	assert.Equal(t, p2p.ErrNilHost, err)
}

func TestNewLibp2pContext_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	lctx, err := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})

	assert.False(t, check.IfNil(lctx))
	assert.Nil(t, err)
}

func TestLibp2pContext_Context(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	lctx, _ := libp2p.NewLibp2pContext(ctx, &mock.ConnectableHostStub{})

	assert.True(t, ctx == lctx.Context())
}

func TestLibp2pContext_Host(t *testing.T) {
	t.Parallel()

	h := &mock.ConnectableHostStub{}
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), h)

	assert.True(t, h == lctx.Host())
}

func TestLibp2pContext_SetPeerBlacklistNilPeerBlacklistShouldErr(t *testing.T) {
	t.Parallel()

	lctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	err := lctx.SetPeerBlacklist(nil)

	assert.Equal(t, p2p.ErrNilPeerBlacklistHandler, err)
}

func TestLibp2pContext_GetSetBlacklistHandlerShouldWork(t *testing.T) {
	t.Parallel()

	lctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	npbh := &mock.NilBlacklistHandler{}

	err := lctx.SetPeerBlacklist(npbh)
	assert.Nil(t, err)

	recoveredNpbh := lctx.PeerBlacklist()
	assert.True(t, npbh == recoveredNpbh)
}
