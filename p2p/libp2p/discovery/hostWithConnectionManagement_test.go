package discovery_test

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
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

func createMockArgsHostWithConnectionManagement() discovery.ArgsHostWithConnectionManagement {
	return discovery.ArgsHostWithConnectionManagement{
		ConnectableHost:    &mock.ConnectableHostStub{},
		Sharder:            &mock.KadSharderStub{},
		ConnectionsWatcher: &mock.ConnectionsWatcherStub{},
	}
}

func TestNewHostWithConnectionManagement(t *testing.T) {
	t.Parallel()

	t.Run("nil connectable host should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsHostWithConnectionManagement()
		args.ConnectableHost = nil
		hwcm, err := discovery.NewHostWithConnectionManagement(args)

		assert.True(t, check.IfNil(hwcm))
		assert.Equal(t, p2p.ErrNilHost, err)
	})
	t.Run("nil sharder should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsHostWithConnectionManagement()
		args.Sharder = nil
		hwcm, err := discovery.NewHostWithConnectionManagement(args)

		assert.True(t, check.IfNil(hwcm))
		assert.Equal(t, p2p.ErrNilSharder, err)
	})
	t.Run("nil connection watcher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsHostWithConnectionManagement()
		args.ConnectionsWatcher = nil
		hwcm, err := discovery.NewHostWithConnectionManagement(args)

		assert.True(t, check.IfNil(hwcm))
		assert.Equal(t, p2p.ErrNilConnectionsWatcher, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsHostWithConnectionManagement()
		hwcm, err := discovery.NewHostWithConnectionManagement(args)

		assert.False(t, check.IfNil(hwcm))
		assert.Nil(t, err)
	})
}

// ------- Connect

func TestHostWithConnectionManagement_ConnectWithSharderNotEvictedShouldCallConnect(t *testing.T) {
	t.Parallel()

	connectCalled := false
	args := createMockArgsHostWithConnectionManagement()
	args.ConnectableHost = &mock.ConnectableHostStub{
		ConnectCalled: func(_ context.Context, _ peer.AddrInfo) error {
			connectCalled = true
			return nil
		},
		NetworkCalled: func() network.Network {
			return createStubNetwork()
		},
	}
	args.Sharder = &mock.KadSharderStub{
		ComputeEvictListCalled: func(pidList []peer.ID) []peer.ID {
			return make([]peer.ID, 0)
		},
		HasCalled: func(pid peer.ID, list []peer.ID) bool {
			return false
		},
	}
	newKnownConnectionCalled := false
	args.ConnectionsWatcher = &mock.ConnectionsWatcherStub{
		NewKnownConnectionCalled: func(pid core.PeerID, connection string) {
			newKnownConnectionCalled = true
		},
	}
	hwcm, _ := discovery.NewHostWithConnectionManagement(args)

	_ = hwcm.Connect(context.Background(), peer.AddrInfo{})

	assert.True(t, connectCalled)
	assert.True(t, newKnownConnectionCalled)
}

func TestHostWithConnectionManagement_ConnectWithSharderEvictedShouldNotCallConnect(t *testing.T) {
	t.Parallel()

	connectCalled := false
	args := createMockArgsHostWithConnectionManagement()
	args.ConnectableHost = &mock.ConnectableHostStub{
		ConnectCalled: func(_ context.Context, _ peer.AddrInfo) error {
			connectCalled = true
			return nil
		},
		NetworkCalled: func() network.Network {
			return createStubNetwork()
		},
	}
	args.Sharder = &mock.KadSharderStub{
		ComputeEvictListCalled: func(pidList []peer.ID) []peer.ID {
			return make([]peer.ID, 0)
		},
		HasCalled: func(pid peer.ID, list []peer.ID) bool {
			return true
		},
	}
	newKnownConnectionCalled := false
	args.ConnectionsWatcher = &mock.ConnectionsWatcherStub{
		NewKnownConnectionCalled: func(pid core.PeerID, connection string) {
			newKnownConnectionCalled = true
		},
	}
	hwcm, _ := discovery.NewHostWithConnectionManagement(args)

	_ = hwcm.Connect(context.Background(), peer.AddrInfo{})

	assert.False(t, connectCalled)
	assert.True(t, newKnownConnectionCalled)
}
