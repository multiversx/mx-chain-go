package discovery_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewKadDhtPeerDiscoverer_InvalidPeersRefreshIntervalShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.PeersRefreshInterval = time.Second - time.Microsecond

	kdd, err := discovery.NewKadDhtPeerDiscoverer(arg)

	assert.True(t, check.IfNil(kdd))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewKadDhtPeerDiscoverer_InvalidRoutingTableRefreshIntervalShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.RoutingTableRefresh = time.Second - time.Microsecond

	kdd, err := discovery.NewKadDhtPeerDiscoverer(arg)

	assert.Nil(t, kdd)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewKadDhtPeerDiscoverer_ShouldSetValues(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	kdd, err := discovery.NewKadDhtPeerDiscoverer(arg)

	assert.Nil(t, err)
	assert.Equal(t, arg.PeersRefreshInterval, kdd.PeersRefreshInterval())
	assert.Equal(t, arg.RandezVous, kdd.RandezVous())
	assert.Equal(t, arg.InitialPeersList, kdd.InitialPeersList())
	assert.Equal(t, arg.RoutingTableRefresh, kdd.RoutingTableRefresh())
	assert.Equal(t, arg.BucketSize, kdd.BucketSize())

	assert.False(t, kdd.IsDiscoveryPaused())
	kdd.Pause()
	assert.True(t, kdd.IsDiscoveryPaused())
	kdd.Resume()
	assert.False(t, kdd.IsDiscoveryPaused())
}

//------- Bootstrap

func TestKadDhtPeerDiscoverer_BootstrapCalledOnceShouldWork(t *testing.T) {
	t.Parallel()

	interval := time.Second

	arg := createTestArgument()
	kdd, _ := discovery.NewKadDhtPeerDiscoverer(arg)

	err := kdd.Bootstrap()
	assert.Nil(t, err)

	time.Sleep(interval * 1)
	kdd.Pause()
	time.Sleep(interval * 2)
}

func TestKadDhtPeerDiscoverer_BootstrapCalledTwiceShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	kdd, _ := discovery.NewKadDhtPeerDiscoverer(arg)

	_ = kdd.Bootstrap()
	err := kdd.Bootstrap()

	assert.Equal(t, p2p.ErrPeerDiscoveryProcessAlreadyStarted, err)
}

//------- connectToOnePeerFromInitialPeersList

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersListNilListShouldRetWithChanFull(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	kdd, _ := discovery.NewKadDhtPeerDiscoverer(arg)

	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Second, nil)

	assert.Equal(t, 1, len(chanDone))
}

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersListEmptyListShouldRetWithChanFull(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	kdd, _ := discovery.NewKadDhtPeerDiscoverer(arg)
	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Second, make([]string, 0))

	assert.Equal(t, 1, len(chanDone))
}

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnect(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	peerID := "peer"
	wasConnectCalled := int32(0)
	arg.Host = &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			if peerID == address {
				atomic.AddInt32(&wasConnectCalled, 1)
			}

			return nil
		},
	}
	kdd, _ := discovery.NewKadDhtPeerDiscoverer(arg)

	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Second, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(1), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnectContinously(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	peerID := "peer"
	wasConnectCalled := int32(0)
	errDidNotConnect := errors.New("did not connect")
	noOfTimesToRefuseConnection := 5
	arg.Host = &mock.ConnectableHostStub{
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
	kdd, _ := discovery.NewKadDhtPeerDiscoverer(arg)

	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(noOfTimesToRefuseConnection), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestKadDhtPeerDiscoverer_ConnectToOnePeerFromInitialPeersTwoPeersShouldAlternate(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	peerID1 := "peer1"
	peerID2 := "peer2"
	wasConnectCalled := int32(0)
	errDidNotConnect := errors.New("did not connect")
	noOfTimesToRefuseConnection := 5
	arg.Host = &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			connCalled := atomic.LoadInt32(&wasConnectCalled)

			atomic.AddInt32(&wasConnectCalled, 1)

			if connCalled >= int32(noOfTimesToRefuseConnection) {
				return nil
			}

			connCalled %= 2
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
	kdd, _ := discovery.NewKadDhtPeerDiscoverer(arg)

	chanDone := kdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID1, peerID2})

	select {
	case <-chanDone:
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

//----------- Watchdog

func TestKadDhtPeerDiscoverer_Watchdog(t *testing.T) {
	t.Parallel()

	interval := time.Second
	arg := createTestArgument()
	arg.InitialPeersList = nil
	arg.PeersRefreshInterval = interval
	kdd, _ := discovery.NewKadDhtPeerDiscoverer(arg)

	err := kdd.StopWatchdog()
	assert.NotNil(t, err)

	err = kdd.KickWatchdog()
	assert.NotNil(t, err)

	err = kdd.StartWatchdog(interval / 2)
	assert.NotNil(t, err)

	s1 := kdd.StartWatchdog(interval)
	s2 := kdd.StartWatchdog(interval)

	assert.Nil(t, s1)
	assert.Equal(t, s2, p2p.ErrWatchdogAlreadyStarted)

	if !testing.Short() {
		//kick
		_ = kdd.KickWatchdog()
		kdd.Pause()
		assert.True(t, kdd.IsDiscoveryPaused())
		time.Sleep(interval / 2)
		assert.True(t, kdd.IsDiscoveryPaused())
		time.Sleep(interval)
		assert.False(t, kdd.IsDiscoveryPaused())
	}

	_ = kdd.StopWatchdog()
}
