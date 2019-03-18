package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	libp2p2 "github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/mock"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

var timeoutWaitResponses = time.Second * 2

func createDummyHost() libp2p2.ConnectableHost {
	netw := mocknet.New(context.Background())

	h, _ := netw.GenPeer()
	return libp2p2.NewConnectableHost(h)
}

func TestNewMdnsDiscoverer_ShouldSetValues(t *testing.T) {
	serviceTag := "srvcTag"
	interval := time.Duration(time.Second * 4)

	mdns := discovery.NewMdnsPeerDiscoverer(interval, serviceTag)

	assert.Equal(t, serviceTag, mdns.ServiceTag())
	assert.Equal(t, interval, mdns.RefreshInterval())
}

//------- Bootstrap

func TestMdnsPeerDiscoverer_BootstrapCalledWithoutContextAppliedShouldErr(t *testing.T) {
	serviceTag := "srvcTag"
	interval := time.Duration(time.Second * 1)

	mdns := discovery.NewMdnsPeerDiscoverer(interval, serviceTag)
	err := mdns.Bootstrap()

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestMdnsPeerDiscoverer_BootstrapCalledOnceShouldWork(t *testing.T) {
	serviceTag := "srvcTag"
	interval := time.Duration(time.Second * 1)

	h := createDummyHost()
	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), h)

	mdns := discovery.NewMdnsPeerDiscoverer(interval, serviceTag)
	defer func() {
		h.Close()
	}()

	mdns.ApplyContext(ctx)
	err := mdns.Bootstrap()

	assert.Nil(t, err)
}

func TestMdnsPeerDiscoverer_BootstrapCalledTwiceShouldErr(t *testing.T) {
	serviceTag := "srvcTag"
	interval := time.Duration(time.Second * 1)

	h := createDummyHost()
	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), h)

	mdns := discovery.NewMdnsPeerDiscoverer(interval, serviceTag)

	defer func() {
		h.Close()
	}()

	mdns.ApplyContext(ctx)
	_ = mdns.Bootstrap()
	err := mdns.Bootstrap()

	assert.Equal(t, p2p.ErrPeerDiscoveryProcessAlreadyStarted, err)
}

//------- ApplyContext

func TestMdnsPeerDiscoverer_ApplyContextNilProviderShouldErr(t *testing.T) {
	mdns := discovery.NewMdnsPeerDiscoverer(time.Duration(time.Second*1), "srvcTag")

	err := mdns.ApplyContext(nil)

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestMdnsPeerDiscoverer_ApplyContextWrongProviderShouldErr(t *testing.T) {
	mdns := discovery.NewMdnsPeerDiscoverer(time.Duration(time.Second*1), "srvcTag")

	err := mdns.ApplyContext(&mock.ContextProviderMock{})

	assert.Equal(t, p2p.ErrWrongContextApplier, err)
}

func TestMdnsPeerDiscoverer_ApplyContextShouldWork(t *testing.T) {
	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	mdns := discovery.NewMdnsPeerDiscoverer(time.Duration(time.Second*1), "srvcTag")

	err := mdns.ApplyContext(ctx)

	assert.Nil(t, err)
	assert.True(t, ctx == mdns.ContextProvider())
}

//------- HandlePeerFoundPeer

func TestNetworkMessenger_HandlePeerFoundNotFoundShouldTryToConnect(t *testing.T) {
	newPeerInfo := peerstore.PeerInfo{
		ID: peer.ID("new found peerID"),
	}
	testAddress := "/ip4/127.0.0.1/tcp/23000/p2p/16Uiu2HAkyqtHSEJDkYhVWTtm9j58Mq5xQJgrApBYXMwS6sdamXuE"
	address, _ := multiaddr.NewMultiaddr(testAddress)
	newPeerInfo.Addrs = []multiaddr.Multiaddr{address}

	chanConnected := make(chan struct{})

	mockHost := &mock.ConnectableHostStub{
		PeerstoreCalled: func() peerstore.Peerstore {
			return peerstore.NewPeerstore(
				pstoremem.NewKeyBook(),
				pstoremem.NewAddrBook(),
				pstoremem.NewPeerMetadata())
		},
		ConnectCalled: func(ctx context.Context, pi peerstore.PeerInfo) error {
			if newPeerInfo.ID == pi.ID {
				chanConnected <- struct{}{}
			}

			return nil
		},
	}

	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), mockHost)
	mdns := discovery.NewMdnsPeerDiscoverer(time.Duration(time.Second*1), "srvcTag")
	_ = mdns.ApplyContext(ctx)

	mdns.HandlePeerFound(newPeerInfo)

	select {
	case <-chanConnected:
		return
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout while waiting to call host.Connect")
	}
}

func TestNetworkMessenger_HandlePeerFoundPeerFoundShouldNotTryToConnect(t *testing.T) {
	newPeerInfo := peerstore.PeerInfo{
		ID: peer.ID("new found peerID"),
	}
	testAddress := "/ip4/127.0.0.1/tcp/23000/p2p/16Uiu2HAkyqtHSEJDkYhVWTtm9j58Mq5xQJgrApBYXMwS6sdamXuE"
	address, _ := multiaddr.NewMultiaddr(testAddress)
	newPeerInfo.Addrs = []multiaddr.Multiaddr{address}

	chanConnected := make(chan struct{})

	mockHost := &mock.ConnectableHostStub{
		PeerstoreCalled: func() peerstore.Peerstore {
			ps := peerstore.NewPeerstore(
				pstoremem.NewKeyBook(),
				pstoremem.NewAddrBook(),
				pstoremem.NewPeerMetadata())
			ps.AddAddrs(newPeerInfo.ID, newPeerInfo.Addrs, peerstore.PermanentAddrTTL)

			return ps
		},
		ConnectCalled: func(ctx context.Context, pi peerstore.PeerInfo) error {
			if newPeerInfo.ID == pi.ID {
				chanConnected <- struct{}{}
			}

			return nil
		},
	}

	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), mockHost)
	mdns := discovery.NewMdnsPeerDiscoverer(time.Duration(time.Second*1), "srvcTag")
	_ = mdns.ApplyContext(ctx)

	mdns.HandlePeerFound(newPeerInfo)

	select {
	case <-chanConnected:
		assert.Fail(t, "should have not called host.Connect")
	case <-time.After(timeoutWaitResponses):
		return
	}
}
