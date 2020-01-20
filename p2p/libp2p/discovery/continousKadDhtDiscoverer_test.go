package discovery

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
)

//------- Bootstrap

func TestContinousKadDhtPeerDiscoverer2_BootstrapCalledWithoutContextAppliedShouldErr(t *testing.T) {
	interval := time.Second

	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)
	err := ckdd.Bootstrap()

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestContinousKadDhtPeerDiscoverer2_BootstrapCalledOnceShouldWork(t *testing.T) {
	interval := time.Microsecond * 50

	h := createDummyHost()
	ctx, _ := libp2p.NewLibp2pContext(context.Background(), h)

	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)
	defer func() {
		_ = h.Close()
	}()

	_ = ckdd.ApplyContext(ctx)
	err := ckdd.Bootstrap()

	assert.Nil(t, err)

	if !testing.Short() {
		time.Sleep(interval * 20)
	}
}

func TestContinousKadDhtPeerDiscoverer2_BootstrapCalledTwiceShouldErr(t *testing.T) {
	interval := time.Second

	h := createDummyHost()
	ctx, _ := libp2p.NewLibp2pContext(context.Background(), h)

	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)

	defer func() {
		_ = h.Close()
	}()

	_ = ckdd.ApplyContext(ctx)
	_ = ckdd.Bootstrap()
	err := ckdd.Bootstrap()

	assert.Equal(t, p2p.ErrPeerDiscoveryProcessAlreadyStarted, err)
}

//------- connectToOnePeerFromInitialPeersList

func TestContinousKadDhtPeerDiscoverer2_ConnectToOnePeerFromInitialPeersListNilListShouldRetWithChanFull(t *testing.T) {
	interval := time.Second

	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	_ = ckdd.ApplyContext(lctx)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, nil)

	assert.Equal(t, 1, len(chanDone))
}

func TestContinousKadDhtPeerDiscoverer2_ConnectToOnePeerFromInitialPeersListEmptyListShouldRetWithChanFull(t *testing.T) {
	interval := time.Second

	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	_ = ckdd.ApplyContext(lctx)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, make([]string, 0))

	assert.Equal(t, 1, len(chanDone))
}

func TestContinousKadDhtPeerDiscoverer2_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnect(t *testing.T) {
	interval := time.Second

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

	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)
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

func TestContinousKadDhtPeerDiscoverer2_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnectContinously(t *testing.T) {
	interval := time.Second

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

	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)
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

func TestContinousKadDhtPeerDiscoverer2_ConnectToOnePeerFromInitialPeersTwoPeersShouldAlternate(t *testing.T) {
	interval := time.Second

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

	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)
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

func TestContinousKadDhtPeerDiscoverer2_ApplyContextNilProviderShouldErr(t *testing.T) {
	interval := time.Second
	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)

	err := ckdd.ApplyContext(nil)

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestContinousKadDhtPeerDiscoverer2_ApplyContextWrongProviderShouldErr(t *testing.T) {
	interval := time.Second
	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)

	err := ckdd.ApplyContext(&mock.ContextProviderMock{})

	assert.Equal(t, p2p.ErrWrongContextProvider, err)
}

func TestContinousKadDhtPeerDiscoverer2_ApplyContextShouldWork(t *testing.T) {
	ctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	interval := time.Second
	ckdd := NewKadDhtPeerDiscoverer(interval, "", nil)

	err := ckdd.ApplyContext(ctx)

	assert.Nil(t, err)
	assert.True(t, ctx == ckdd.ContextProvider())
}

//---------- Protocols
func TestContinousKadDhtPeerDiscoverer2_Protocols(t *testing.T) {
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
	}

	ctx, _ := libp2p.NewLibp2pContext(context.Background(), host)
	interval := time.Second
	ckdd := NewKadDhtPeerDiscoverer(interval, "r1", nil)

	err := ckdd.ApplyContext(ctx)
	assert.Nil(t, err)

	err = ckdd.Bootstrap()
	assert.Nil(t, err)

	assert.Equal(t, notifeesCnt, 1)
	err = ckdd.UpdateRandezVous("r2")
	assert.Nil(t, err)
	assert.Equal(t, notifeesCnt, 1)
	// lock the mutex
	ckdd.mutKadDht.Lock()
	err = ckdd.stopDHT()
	ckdd.mutKadDht.Unlock()
	assert.Nil(t, err)

	assert.Equal(t, notifeesCnt, 0)
	for p, cb := range streams {
		assert.Nil(t, cb, p, "should have no callback")
	}
}
