package libp2p

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func TestConnectableHost_ConnectToPeerWrongAddressShouldErr(t *testing.T) {
	uhs := &mock.ConnectableHostStub{}
	//we can safely use an upgraded instead of a real host as to not create another (useless) stub
	uh := NewConnectableHost(uhs)

	err := uh.ConnectToPeer(context.Background(), "invalid address")

	assert.NotNil(t, err)
}

func TestConnectableHost_ConnectToPeerShouldWork(t *testing.T) {
	wasCalled := false

	uhs := &mock.ConnectableHostStub{
		ConnectCalled: func(ctx context.Context, pi peer.AddrInfo) error {
			wasCalled = true
			return nil
		},
	}
	//we can safely use an upgraded instead of a real host as to not create another (useless) stub
	uh := NewConnectableHost(uhs)

	validAddress := "/ip4/82.5.34.12/tcp/23000/p2p/16Uiu2HAkyqtHSEJDkYhVWTtm9j58Mq5xQJgrApBYXMwS6sdamXuE"
	err := uh.ConnectToPeer(context.Background(), validAddress)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}
