package discovery_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
)

var timeoutWaitResponses = 2 * time.Second

func createTestArgument() discovery.ArgKadDht {
	return discovery.ArgKadDht{
		Context:              context.Background(),
		Host:                 &mock.ConnectableHostStub{},
		KddSharder:           &mock.SharderStub{},
		PeersRefreshInterval: time.Second,
		RandezVous:           "",
		InitialPeersList:     []string{"peer1", "peer2"},
		BucketSize:           100,
		RoutingTableRefresh:  5 * time.Second,
	}
}

//------- NewContinuousKadDhtDiscoverer

func TestNewContinuousKadDhtDiscoverer_NilContextShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.Context = nil

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.True(t, check.IfNil(kdd))
	assert.True(t, errors.Is(err, p2p.ErrNilContext))
}

func TestNewContinuousKadDhtDiscoverer_NilHostShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.Host = nil

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.True(t, check.IfNil(kdd))
	assert.True(t, errors.Is(err, p2p.ErrNilHost))
}

func TestNewContinuousKadDhtDiscoverer_NilSharderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.KddSharder = nil

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.True(t, check.IfNil(kdd))
	assert.True(t, errors.Is(err, p2p.ErrNilSharder))
}

func TestNewContinuousKadDhtDiscoverer_WrongSharderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.KddSharder = &mock.CommonSharder{}

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.True(t, check.IfNil(kdd))
	assert.True(t, errors.Is(err, p2p.ErrWrongTypeAssertion))
}

func TestNewContinuousKadDhtDiscoverer_InvalidPeersRefreshIntervalShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.PeersRefreshInterval = time.Second - time.Microsecond

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.Nil(t, kdd)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewContinuousKadDhtDiscoverer_InvalidRoutingTableRefreshIntervalShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.RoutingTableRefresh = time.Second - time.Microsecond

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.Nil(t, kdd)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewContinuousKadDhtDiscoverer_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.False(t, check.IfNil(kdd))
	assert.Nil(t, err)
}

func TestNewContinuousKadDhtDiscoverer_EmptyInitialPeersShouldWork(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.InitialPeersList = nil

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.False(t, check.IfNil(kdd))
	assert.Nil(t, err)
}

//------- Bootstrap

func TestContinuousKadDhtDiscoverer_BootstrapCalledOnceShouldWork(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	err := ckdd.Bootstrap()

	assert.Nil(t, err)
	time.Sleep(arg.PeersRefreshInterval * 2)
}

func TestContinuousKadDhtDiscoverer_BootstrapCalledTwiceShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	_ = ckdd.Bootstrap()
	err := ckdd.Bootstrap()

	assert.Equal(t, p2p.ErrPeerDiscoveryProcessAlreadyStarted, err)
}

//------- connectToOnePeerFromInitialPeersList

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersListNilListShouldRetWithChanFull(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, nil)

	assert.Equal(t, 1, len(chanDone))
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersListEmptyListShouldRetWithChanFull(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, make([]string, 0))

	assert.Equal(t, 1, len(chanDone))
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnect(t *testing.T) {
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
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(1), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnectContinously(t *testing.T) {
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
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(noOfTimesToRefuseConnection), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersTwoPeersShouldAlternate(t *testing.T) {
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

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID1, peerID2})

	select {
	case <-chanDone:
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

//---------- Protocols

func TestContinuousKadDhtDiscoverer_Protocols(t *testing.T) {
	streams := make(map[protocol.ID]network.StreamHandler)
	notifeesCnt := 0
	net := &mock.NetworkStub{
		ConnectednessCalled: func(p peer.ID) network.Connectedness {
			fmt.Printf("Conn to %s\n", p.Pretty())
			return network.CannotConnect
		},

		NotifyCalled:     func(nn network.Notifiee) { notifeesCnt++ },
		StopNotifyCalled: func(nn network.Notifiee) { notifeesCnt-- },
	}
	arg := createTestArgument()
	arg.Host = &mock.ConnectableHostStub{
		IDCalled: func() peer.ID {
			return "local peer"
		},
		NetworkCalled: func() network.Network {
			return net
		},
		SetStreamHandlerCalled: func(proto protocol.ID, sh network.StreamHandler) {
			fmt.Printf("Set strem hndl %v\n", proto)
			streams[proto] = sh
		},
		RemoveStreamHandlerCalled: func(proto protocol.ID) {
			fmt.Printf("Remove stream %v\n", proto)
			streams[proto] = nil
		},
	}
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	err := ckdd.Bootstrap()
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

func TestContinuousKadDhtDiscoverer_Name(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	kdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.Equal(t, discovery.KadDhtName, kdd.Name())
}
