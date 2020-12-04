package discovery_test

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func createStubNetwork() network.Network {
	return &mock.NetworkStub{
		PeersCall: func() []peer.ID {
			return make([]peer.ID, 0)
		},
	}
}

func TestNewHostWithConnectionManagement_NilHostShouldErr(t *testing.T) {
	t.Parallel()

	hwcm, err := discovery.NewHostWithConnectionManagement(nil, &mock.KadSharderStub{})

	assert.True(t, check.IfNil(hwcm))
	assert.Equal(t, p2p.ErrNilHost, err)
}

func TestNewHostWithConnectionManagement_NilSharderShouldErr(t *testing.T) {
	t.Parallel()

	hwcm, err := discovery.NewHostWithConnectionManagement(&mock.ConnectableHostStub{}, nil)

	assert.True(t, check.IfNil(hwcm))
	assert.Equal(t, p2p.ErrNilSharder, err)
}

func TestNewHostWithConnectionManagement_ShouldWork(t *testing.T) {
	t.Parallel()

	hwcm, err := discovery.NewHostWithConnectionManagement(&mock.ConnectableHostStub{}, &mock.KadSharderStub{})

	assert.False(t, check.IfNil(hwcm))
	assert.Nil(t, err)
}

//------- Connect

func TestHostWithConnectionManagement_ConnectWithSharderNotEvictedShouldCallConnect(t *testing.T) {
	t.Parallel()

	connectCalled := false
	hwcm, _ := discovery.NewHostWithConnectionManagement(
		&mock.ConnectableHostStub{
			ConnectCalled: func(_ context.Context, _ peer.AddrInfo) error {
				connectCalled = true
				return nil
			},
			NetworkCalled: func() network.Network {
				return createStubNetwork()
			},
		},
		&mock.KadSharderStub{
			ComputeEvictListCalled: func(pidList []peer.ID) []peer.ID {
				return make([]peer.ID, 0)
			},
			HasCalled: func(pid peer.ID, list []peer.ID) bool {
				return false
			},
		},
	)

	_ = hwcm.Connect(context.Background(), peer.AddrInfo{})

	assert.True(t, connectCalled)
}

func TestHostWithConnectionManagement_ConnectWithSharderEvictedShouldNotCallConnect(t *testing.T) {
	t.Parallel()

	connectCalled := false
	hwcm, _ := discovery.NewHostWithConnectionManagement(
		&mock.ConnectableHostStub{
			ConnectCalled: func(_ context.Context, _ peer.AddrInfo) error {
				connectCalled = true
				return nil
			},
			NetworkCalled: func() network.Network {
				return createStubNetwork()
			},
		},
		&mock.KadSharderStub{
			ComputeEvictListCalled: func(pidList []peer.ID) []peer.ID {
				return make([]peer.ID, 0)
			},
			HasCalled: func(pid peer.ID, list []peer.ID) bool {
				return true
			},
		},
	)

	_ = hwcm.Connect(context.Background(), peer.AddrInfo{})

	assert.False(t, connectCalled)
}
