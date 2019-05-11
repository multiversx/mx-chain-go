package libp2p_test

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewLibp2pContext_NilContextShoulsErr(t *testing.T) {
	lctx, err := libp2p.NewLibp2pContext(nil, &mock.ConnectableHostStub{})

	assert.Nil(t, lctx)
	assert.Equal(t, p2p.ErrNilContext, err)
}

func TestNewLibp2pContext_NilHostShoulsErr(t *testing.T) {
	lctx, err := libp2p.NewLibp2pContext(context.Background(), nil)

	assert.Nil(t, lctx)
	assert.Equal(t, p2p.ErrNilHost, err)
}

func TestNewLibp2pContext_OkValsShouldWork(t *testing.T) {
	lctx, err := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})

	assert.NotNil(t, lctx)
	assert.Nil(t, err)
}

func TestLibp2pContext_Context(t *testing.T) {
	ctx := context.Background()

	lctx, _ := libp2p.NewLibp2pContext(ctx, &mock.ConnectableHostStub{})

	assert.True(t, ctx == lctx.Context())
}

func TestLibp2pContext_Host(t *testing.T) {
	h := &mock.ConnectableHostStub{}

	lctx, _ := libp2p.NewLibp2pContext(context.Background(), h)

	assert.True(t, h == lctx.Host())
}
