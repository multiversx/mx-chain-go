package discovery_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
)

func createTestArgument() discovery.ArgKadDht {
	return discovery.ArgKadDht{
		PeersRefreshInterval: time.Second,
		RandezVous:           "randez vous",
		InitialPeersList:     []string{"peer1", "peer2"},
		BucketSize:           100,
		RoutingTableRefresh:  5 * time.Second,
	}
}

func TestNewContinuousKadDhtDiscoverer_InvalidPeersRefreshIntervalShouldErr(t *testing.T) {
	arg := createTestArgument()
	arg.PeersRefreshInterval = time.Second - time.Microsecond

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.Nil(t, kdd)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewContinuousKadDhtDiscoverer_InvalidRoutingTableRefreshIntervalShouldErr(t *testing.T) {
	arg := createTestArgument()
	arg.RoutingTableRefresh = time.Second - time.Microsecond

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.Nil(t, kdd)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

//------- Bootstrap

func TestContinuousKadDhtDiscoverer_BootstrapCalledWithoutContextAppliedShouldErr(t *testing.T) {
	arg := createTestArgument()

	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	err := ckdd.Bootstrap()

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestContinuousKadDhtDiscoverer_BootstrapCalledOnceShouldWork(t *testing.T) {
	arg := createTestArgument()
	h := createDummyHost()
	ctx, _ := libp2p.NewLibp2pContext(context.Background(), h)

	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	defer func() {
		_ = h.Close()
	}()

	_ = ckdd.ApplyContext(ctx)
	err := ckdd.Bootstrap()

	assert.Nil(t, err)

	if !testing.Short() {
		time.Sleep(arg.PeersRefreshInterval * 2)
	}
}

func TestContinuousKadDhtDiscoverer_BootstrapCalledTwiceShouldErr(t *testing.T) {
	arg := createTestArgument()
	h := createDummyHost()
	ctx, _ := libp2p.NewLibp2pContext(context.Background(), h)

	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	defer func() {
		_ = h.Close()
	}()

	_ = ckdd.ApplyContext(ctx)
	_ = ckdd.Bootstrap()
	err := ckdd.Bootstrap()

	assert.Equal(t, p2p.ErrPeerDiscoveryProcessAlreadyStarted, err)
}

//------- connectToOnePeerFromInitialPeersList

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersListNilListShouldRetWithChanFull(t *testing.T) {
	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	_ = ckdd.ApplyContext(lctx)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, nil)

	assert.Equal(t, 1, len(chanDone))
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersListEmptyListShouldRetWithChanFull(t *testing.T) {
	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	_ = ckdd.ApplyContext(lctx)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, make([]string, 0))

	assert.Equal(t, 1, len(chanDone))
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnect(t *testing.T) {
	arg := createTestArgument()
	peerID := "peer"
	wasConnectCalled := int32(0)

	uhs := &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			if peerID == address {
				atomic.AddInt32(&wasConnectCalled, 1)
			}

			return nil
		},
	}

	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), uhs)
	_ = ckdd.ApplyContext(lctx)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(1), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnectContinously(t *testing.T) {
	arg := createTestArgument()
	peerID := "peer"
	wasConnectCalled := int32(0)

	errDidNotConnect := errors.New("did not connect")
	noOfTimesToRefuseConnection := 5

	uhs := &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			if peerID != address {
				assert.Fail(t, "should have tried to connect to the same ID")
			}

			atomic.AddInt32(&wasConnectCalled, 1)

			if atomic.LoadInt32(&wasConnectCalled) < int32(noOfTimesToRefuseConnection) {
				return errDidNotConnect
			}

			return nil
		},
	}

	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), uhs)
	_ = ckdd.ApplyContext(lctx)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(noOfTimesToRefuseConnection), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersTwoPeersShouldAlternate(t *testing.T) {
	arg := createTestArgument()
	peerID1 := "peer1"
	peerID2 := "peer2"

	wasConnectCalled := int32(0)

	errDidNotConnect := errors.New("did not connect")
	noOfTimesToRefuseConnection := 5

	uhs := &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			connCalled := atomic.LoadInt32(&wasConnectCalled)

			atomic.AddInt32(&wasConnectCalled, 1)

			if connCalled >= int32(noOfTimesToRefuseConnection) {
				return nil
			}

			connCalled = connCalled % 2
			if connCalled == 0 {
				if peerID1 != address {
					assert.Fail(t, "should have tried to connect to "+peerID1)
				}
			}

			if connCalled == 1 {
				if peerID2 != address {
					assert.Fail(t, "should have tried to connect to "+peerID2)
				}
			}

			return errDidNotConnect
		},
	}

	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), uhs)
	_ = ckdd.ApplyContext(lctx)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID1, peerID2})

	select {
	case <-chanDone:
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

//------- ApplyContext

func TestContinuousKadDhtPeerDiscoverer_ApplyContextNilProviderShouldErr(t *testing.T) {
	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	err := ckdd.ApplyContext(nil)

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestContinuousKadDhtDiscoverer_ApplyContextWrongProviderShouldErr(t *testing.T) {
	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	err := ckdd.ApplyContext(&mock.ContextProviderMock{})

	assert.Equal(t, p2p.ErrWrongContextProvider, err)
}

func TestContinuousKadDhtDiscoverer_ApplyContextShouldWork(t *testing.T) {
	ctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	err := ckdd.ApplyContext(ctx)

	assert.Nil(t, err)
	assert.True(t, ctx == ckdd.ContextProvider())
}

//---------- Protocols

func TestContinuousKadDhtDiscoverer_Protocols(t *testing.T) {
	streams := make(map[protocol.ID]network.StreamHandler)
	notifeesCnt := 0
	net := &mock.NetworkStub{
		ConnectednessCalled: func(p peer.ID) network.Connectedness {
			t.Logf("Conn to %s", p.Pretty())
			return network.CannotConnect
		},

		NotifyCalled:     func(nn network.Notifiee) { notifeesCnt++ },
		StopNotifyCalled: func(nn network.Notifiee) { notifeesCnt-- },
	}

	host := &mock.ConnectableHostStub{
		IDCalled: func() peer.ID {
			return "local peer"
		},
		PeerstoreCalled: func() peerstore.Peerstore {
			return nil
		},
		NetworkCalled: func() network.Network {
			return net
		},
		SetStreamHandlerCalled: func(proto protocol.ID, sh network.StreamHandler) {
			t.Logf("Set strem hndl %v", proto)
			streams[proto] = sh
		},
		RemoveStreamHandlerCalled: func(proto protocol.ID) {
			t.Logf("Remove stream %v", proto)
			streams[proto] = nil
		},
		ConnManagerCalled: func() connmgr.ConnManager {
			return nil
		},
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			return nil
		},
	}

	ctx, _ := libp2p.NewLibp2pContext(context.Background(), host)
	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	err := ckdd.ApplyContext(ctx)
	assert.Nil(t, err)

	err = ckdd.Bootstrap()
	assert.Nil(t, err)

	assert.Equal(t, notifeesCnt, 1)
	err = ckdd.UpdateRandezVous("r2")
	assert.Nil(t, err)
	assert.Equal(t, notifeesCnt, 1)
	err = ckdd.StopDHT()
	assert.Nil(t, err)

	assert.Equal(t, notifeesCnt, 0)
	for p, cb := range streams {
		assert.Nil(t, cb, p, "should have no callback")
	}
}
