package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	libp2p2 "github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
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
	interval := time.Second * 4

	mdns := discovery.NewMdnsPeerDiscoverer(interval, serviceTag)

	assert.Equal(t, serviceTag, mdns.ServiceTag())
	assert.Equal(t, interval, mdns.RefreshInterval())
}

//------- Bootstrap

func TestMdnsPeerDiscoverer_BootstrapCalledWithoutContextAppliedShouldErr(t *testing.T) {
	serviceTag := "srvcTag"
	interval := time.Second

	mdns := discovery.NewMdnsPeerDiscoverer(interval, serviceTag)
	err := mdns.Bootstrap()

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestMdnsPeerDiscoverer_BootstrapCalledOnceShouldWork(t *testing.T) {
	//TODO delete skip when mdns library is concurrent safe
	t.Skip("mdns library is not concurrent safe (yet)")
	serviceTag := "srvcTag"
	interval := time.Second

	h := createDummyHost()
	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), h)

	mdns := discovery.NewMdnsPeerDiscoverer(interval, serviceTag)
	defer func() {
		_ = h.Close()
	}()

	_ = mdns.ApplyContext(ctx)
	err := mdns.Bootstrap()

	assert.Nil(t, err)
}

func TestMdnsPeerDiscoverer_BootstrapCalledTwiceShouldErr(t *testing.T) {
	//TODO delete skip when mdns library is concurrent safe
	t.Skip("mdns library is not concurrent safe (yet)")
	serviceTag := "srvcTag"
	interval := time.Second

	h := createDummyHost()
	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), h)

	mdns := discovery.NewMdnsPeerDiscoverer(interval, serviceTag)

	defer func() {
		_ = h.Close()
	}()

	_ = mdns.ApplyContext(ctx)
	_ = mdns.Bootstrap()
	err := mdns.Bootstrap()

	assert.Equal(t, p2p.ErrPeerDiscoveryProcessAlreadyStarted, err)
}

//------- ApplyContext

func TestMdnsPeerDiscoverer_ApplyContextNilProviderShouldErr(t *testing.T) {
	mdns := discovery.NewMdnsPeerDiscoverer(time.Second, "srvcTag")

	err := mdns.ApplyContext(nil)

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestMdnsPeerDiscoverer_ApplyContextWrongProviderShouldErr(t *testing.T) {
	//TODO delete skip when mdns library is concurrent safe
	t.Skip("mdns library is not concurrent safe (yet)")
	mdns := discovery.NewMdnsPeerDiscoverer(time.Second, "srvcTag")

	err := mdns.ApplyContext(&mock.ContextProviderMock{})

	assert.Equal(t, p2p.ErrWrongContextApplier, err)
}

func TestMdnsPeerDiscoverer_ApplyContextShouldWork(t *testing.T) {
	//TODO delete skip when mdns library is concurrent safe
	t.Skip("mdns library is not concurrent safe (yet)")
	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	mdns := discovery.NewMdnsPeerDiscoverer(time.Second, "srvcTag")

	err := mdns.ApplyContext(ctx)

	assert.Nil(t, err)
	assert.True(t, ctx == mdns.ContextProvider())
}

//------- HandlePeerFoundPeer

func TestNetworkMessenger_HandlePeerFoundNotFoundShouldTryToConnect(t *testing.T) {
	newPeerInfo := peer.AddrInfo{
		ID: peer.ID("new found peerID"),
	}
	testAddress := "/ip4/127.0.0.1/tcp/23000/p2p/16Uiu2HAkyqtHSEJDkYhVWTtm9j58Mq5xQJgrApBYXMwS6sdamXuE"
	address, _ := multiaddr.NewMultiaddr(testAddress)
	newPeerInfo.Addrs = []multiaddr.Multiaddr{address}

	chanConnected := make(chan struct{})

	mockHost := &mock.ConnectableHostStub{
		PeerstoreCalled: func() peerstore.Peerstore {
			return pstoremem.NewPeerstore()
		},
		ConnectCalled: func(ctx context.Context, pi peer.AddrInfo) error {
			if newPeerInfo.ID == pi.ID {
				chanConnected <- struct{}{}
			}

			return nil
		},
	}

	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), mockHost)
	mdns := discovery.NewMdnsPeerDiscoverer(time.Second, "srvcTag")
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
	newPeerInfo := peer.AddrInfo{
		ID: peer.ID("new found peerID"),
	}
	testAddress := "/ip4/127.0.0.1/tcp/23000/p2p/16Uiu2HAkyqtHSEJDkYhVWTtm9j58Mq5xQJgrApBYXMwS6sdamXuE"
	address, _ := multiaddr.NewMultiaddr(testAddress)
	newPeerInfo.Addrs = []multiaddr.Multiaddr{address}

	chanConnected := make(chan struct{})

	mockHost := &mock.ConnectableHostStub{
		PeerstoreCalled: func() peerstore.Peerstore {
			ps := pstoremem.NewPeerstore()
			ps.AddAddrs(newPeerInfo.ID, newPeerInfo.Addrs, peerstore.PermanentAddrTTL)

			return ps
		},
		ConnectCalled: func(ctx context.Context, pi peer.AddrInfo) error {
			if newPeerInfo.ID == pi.ID {
				chanConnected <- struct{}{}
			}

			return nil
		},
	}

	ctx, _ := libp2p2.NewLibp2pContext(context.Background(), mockHost)
	mdns := discovery.NewMdnsPeerDiscoverer(time.Second, "srvcTag")
	_ = mdns.ApplyContext(ctx)

	mdns.HandlePeerFound(newPeerInfo)

	select {
	case <-chanConnected:
		assert.Fail(t, "should have not called host.Connect")
	case <-time.After(timeoutWaitResponses):
		return
	}
}
