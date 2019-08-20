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
	"github.com/stretchr/testify/assert"
)

func TestNewKadDhtPeerDiscoverer_ShouldSetValues(t *testing.T) {
	initialPeersList := []string{"peer1", "peer2"}
	interval := time.Second * 4
	randezVous := "randez vous"

	kdd := discovery.NewKadDhtPeerDiscoverer(interval, randezVous, initialPeersList)

	assert.Equal(t, interval, kdd.RefreshInterval())
	assert.Equal(t, randezVous, kdd.RandezVous())
	assert.Equal(t, initialPeersList, kdd.InitialPeersList())
}

//------- Bootstrap

func TestKadDhtPeerDiscoverer_BootstrapCalledWithoutContextAppliedShouldErr(t *testing.T) {
	interval := time.Second * 1

	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)
	err := kdd.Bootstrap()

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestKadDhtPeerDiscoverer_BootstrapCalledOnceShouldWork(t *testing.T) {
	interval := time.Second * 1

	h := createDummyHost()
	ctx, _ := libp2p.NewLibp2pContext(context.Background(), h)

	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)
	defer func() {
		_ = h.Close()
	}()

	_ = kdd.ApplyContext(ctx)
	err := kdd.Bootstrap()

	assert.Nil(t, err)
}

func TestKadDhtPeerDiscoverer_BootstrapCalledTwiceShouldErr(t *testing.T) {
	interval := time.Second * 1

	h := createDummyHost()
	ctx, _ := libp2p.NewLibp2pContext(context.Background(), h)

	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)

	defer func() {
		_ = h.Close()
	}()

	_ = kdd.ApplyContext(ctx)
	_ = kdd.Bootstrap()
	err := kdd.Bootstrap()

	assert.Equal(t, p2p.ErrPeerDiscoveryProcessAlreadyStarted, err)
}

//------- connectToOnePeerFromInitialPeersList

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersListNilListShouldRetWithChanFull(t *testing.T) {
	interval := time.Second * 1

	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	_ = kdd.ApplyContext(lctx)

	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Second, nil)

	assert.Equal(t, 1, len(chanDone))
}

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersListEmptyListShouldRetWithChanFull(t *testing.T) {
	interval := time.Second * 1

	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	_ = kdd.ApplyContext(lctx)

	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Second, make([]string, 0))

	assert.Equal(t, 1, len(chanDone))
}

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnect(t *testing.T) {
	interval := time.Second * 1

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

	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), uhs)
	_ = kdd.ApplyContext(lctx)

	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Second, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(1), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnectContinously(t *testing.T) {
	interval := time.Second * 1

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

	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), uhs)
	_ = kdd.ApplyContext(lctx)

	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(noOfTimesToRefuseConnection), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersTwoPeersShouldAlternate(t *testing.T) {
	interval := time.Second * 1

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

	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)
	lctx, _ := libp2p.NewLibp2pContext(context.Background(), uhs)
	_ = kdd.ApplyContext(lctx)

	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID1, peerID2})

	select {
	case <-chanDone:
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

//------- ApplyContext

func TestKadDhtPeerDiscoverer_ApplyContextNilProviderShouldErr(t *testing.T) {
	interval := time.Second * 1
	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)

	err := kdd.ApplyContext(nil)

	assert.Equal(t, p2p.ErrNilContextProvider, err)
}

func TestKadDhtPeerDiscoverer_ApplyContextWrongProviderShouldErr(t *testing.T) {
	interval := time.Second * 1
	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)

	err := kdd.ApplyContext(&mock.ContextProviderMock{})

	assert.Equal(t, p2p.ErrWrongContextApplier, err)
}

func TestKadDhtPeerDiscoverer_ApplyContextShouldWork(t *testing.T) {
	ctx, _ := libp2p.NewLibp2pContext(context.Background(), &mock.ConnectableHostStub{})
	interval := time.Second * 1
	kdd := discovery.NewKadDhtPeerDiscoverer(interval, "", nil)

	err := kdd.ApplyContext(ctx)

	assert.Nil(t, err)
	assert.True(t, ctx == kdd.ContextProvider())
}
