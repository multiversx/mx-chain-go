package libp2p_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/data"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	pubsub "github.com/ElrondNetwork/go-libp2p-pubsub"
	pubsubPb "github.com/ElrondNetwork/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var timeoutWaitResponses = time.Second * 2

func waitDoneWithTimeout(t *testing.T, chanDone chan bool, timeout time.Duration) {
	select {
	case <-chanDone:
		return
	case <-time.After(timeout):
		assert.Fail(t, "timeout reached")
	}
}

func prepareMessengerForMatchDataReceive(messenger p2p.Messenger, matchData []byte, wg *sync.WaitGroup) {
	_ = messenger.CreateTopic("test", false)

	_ = messenger.RegisterMessageProcessor("test", "identifier",
		&mock.MessageProcessorStub{
			ProcessMessageCalled: func(message p2p.MessageP2P, _ core.PeerID) error {
				if bytes.Equal(matchData, message.Data()) {
					fmt.Printf("%s got the message\n", messenger.ID().Pretty())
					wg.Done()
				}

				return nil
			},
		})
}

func getConnectableAddress(messenger p2p.Messenger) string {
	for _, addr := range messenger.Addresses() {
		if strings.Contains(addr, "circuit") || strings.Contains(addr, "169.254") {
			continue
		}

		return addr
	}

	return ""
}

func createMockNetworkArgs() libp2p.ArgsNetworkMessenger {
	return libp2p.ArgsNetworkMessenger{
		Marshalizer:   &testscommon.ProtoMarshalizerMock{},
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port:                  "0",
				ConnectionWatcherType: "print",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer:            &libp2p.LocalSyncTimer{},
		PreferredPeersHolder: &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:   &p2pmocks.PeersRatingHandlerStub{},
	}
}

func createMockNetworkOf2() (mocknet.Mocknet, p2p.Messenger, p2p.Messenger) {
	netw := mocknet.New(context.Background())

	messenger1, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	messenger2, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	return netw, messenger1, messenger2
}

func createMockNetworkOf3() (p2p.Messenger, p2p.Messenger, p2p.Messenger) {
	netw := mocknet.New(context.Background())

	messenger1, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	messenger2, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	messenger3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	nscm1 := mock.NewNetworkShardingCollectorMock()
	nscm1.PutPeerIdSubType(messenger1.ID(), core.FullHistoryObserver)
	nscm1.PutPeerIdSubType(messenger2.ID(), core.FullHistoryObserver)
	nscm1.PutPeerIdSubType(messenger3.ID(), core.RegularPeer)
	_ = messenger1.SetPeerShardResolver(nscm1)

	nscm2 := mock.NewNetworkShardingCollectorMock()
	nscm2.PutPeerIdSubType(messenger1.ID(), core.FullHistoryObserver)
	nscm2.PutPeerIdSubType(messenger2.ID(), core.FullHistoryObserver)
	nscm2.PutPeerIdSubType(messenger3.ID(), core.RegularPeer)
	_ = messenger2.SetPeerShardResolver(nscm2)

	nscm3 := mock.NewNetworkShardingCollectorMock()
	nscm3.PutPeerIdSubType(messenger1.ID(), core.FullHistoryObserver)
	nscm3.PutPeerIdSubType(messenger2.ID(), core.FullHistoryObserver)
	nscm3.PutPeerIdSubType(messenger3.ID(), core.RegularPeer)
	_ = messenger3.SetPeerShardResolver(nscm3)

	return messenger1, messenger2, messenger3
}

func createMockMessenger() p2p.Messenger {
	netw := mocknet.New(context.Background())

	messenger, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	return messenger
}

func containsPeerID(list []core.PeerID, searchFor core.PeerID) bool {
	for _, pid := range list {
		if bytes.Equal(pid.Bytes(), searchFor.Bytes()) {
			return true
		}
	}
	return false
}

// ------- NewMemoryLibp2pMessenger

func TestNewMemoryLibp2pMessenger_NilMockNetShouldErr(t *testing.T) {
	args := createMockNetworkArgs()
	messenger, err := libp2p.NewMockMessenger(args, nil)

	assert.Nil(t, messenger)
	assert.Equal(t, p2p.ErrNilMockNet, err)
}

func TestNewMemoryLibp2pMessenger_OkValsWithoutDiscoveryShouldWork(t *testing.T) {
	netw := mocknet.New(context.Background())

	messenger, err := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(messenger))

	_ = messenger.Close()
}

// ------- NewNetworkMessenger

func TestNewNetworkMessenger_NilMessengerShouldErr(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.Marshalizer = nil
	messenger, err := libp2p.NewNetworkMessenger(arg)

	assert.True(t, check.IfNil(messenger))
	assert.True(t, errors.Is(err, p2p.ErrNilMarshalizer))
}

func TestNewNetworkMessenger_NilPreferredPeersHolderShouldErr(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.PreferredPeersHolder = nil
	messenger, err := libp2p.NewNetworkMessenger(arg)

	assert.True(t, check.IfNil(messenger))
	assert.True(t, errors.Is(err, p2p.ErrNilPreferredPeersHolder))
}

func TestNewNetworkMessenger_NilPeersRatingHandlerShouldErr(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.PeersRatingHandler = nil
	mes, err := libp2p.NewNetworkMessenger(arg)

	assert.True(t, check.IfNil(mes))
	assert.True(t, errors.Is(err, p2p.ErrNilPeersRatingHandler))
}

func TestNewNetworkMessenger_NilSyncTimerShouldErr(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.SyncTimer = nil
	messenger, err := libp2p.NewNetworkMessenger(arg)

	assert.True(t, check.IfNil(messenger))
	assert.True(t, errors.Is(err, p2p.ErrNilSyncTimer))
}

func TestNewNetworkMessenger_WithDeactivatedKadDiscovererShouldWork(t *testing.T) {
	arg := createMockNetworkArgs()
	messenger, err := libp2p.NewNetworkMessenger(arg)

	assert.NotNil(t, messenger)
	assert.Nil(t, err)

	_ = messenger.Close()
}

func TestNewNetworkMessenger_WithKadDiscovererListsSharderInvalidTargetConnShouldErr(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.P2pConfig.KadDhtPeerDiscovery = config.KadDhtPeerDiscoveryConfig{
		Enabled:                          true,
		Type:                             "optimized",
		RefreshIntervalInSec:             10,
		ProtocolID:                       "/erd/kad/1.0.0",
		InitialPeerList:                  nil,
		BucketSize:                       100,
		RoutingTableRefreshIntervalInSec: 10,
	}
	arg.P2pConfig.Sharding.Type = p2p.ListsSharder
	messenger, err := libp2p.NewNetworkMessenger(arg)

	assert.True(t, check.IfNil(messenger))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewNetworkMessenger_WithKadDiscovererListSharderShouldWork(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.P2pConfig.KadDhtPeerDiscovery = config.KadDhtPeerDiscoveryConfig{
		Enabled:                          true,
		Type:                             "optimized",
		RefreshIntervalInSec:             10,
		ProtocolID:                       "/erd/kad/1.0.0",
		InitialPeerList:                  nil,
		BucketSize:                       100,
		RoutingTableRefreshIntervalInSec: 10,
	}
	arg.P2pConfig.Sharding = config.ShardingConfig{
		Type:            p2p.NilListSharder,
		TargetPeerCount: 10,
	}
	messenger, err := libp2p.NewNetworkMessenger(arg)

	assert.False(t, check.IfNil(messenger))
	assert.Nil(t, err)

	_ = messenger.Close()
}

// ------- Messenger functionality

func TestLibp2pMessenger_ConnectToPeerShouldCallUpgradedHost(t *testing.T) {
	netw := mocknet.New(context.Background())

	messenger, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = messenger.Close()

	wasCalled := false

	p := "peer"

	uhs := &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			if p == address {
				wasCalled = true
			}
			return nil
		},
	}

	messenger.SetHost(uhs)
	_ = messenger.ConnectToPeer(p)
	assert.True(t, wasCalled)
}

func TestLibp2pMessenger_IsConnectedShouldWork(t *testing.T) {
	_, messenger1, messenger2 := createMockNetworkOf2()

	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)

	assert.True(t, messenger1.IsConnected(messenger2.ID()))
	assert.True(t, messenger2.IsConnected(messenger1.ID()))

	_ = messenger1.Close()
	_ = messenger2.Close()
}

func TestLibp2pMessenger_CreateTopicOkValsShouldWork(t *testing.T) {
	messenger := createMockMessenger()

	err := messenger.CreateTopic("test", true)
	assert.Nil(t, err)

	_ = messenger.Close()
}

func TestLibp2pMessenger_CreateTopicTwiceShouldNotErr(t *testing.T) {
	messenger := createMockMessenger()

	_ = messenger.CreateTopic("test", false)
	err := messenger.CreateTopic("test", false)
	assert.Nil(t, err)

	_ = messenger.Close()
}

func TestLibp2pMessenger_HasTopicIfHaveTopicShouldReturnTrue(t *testing.T) {
	messenger := createMockMessenger()

	_ = messenger.CreateTopic("test", false)

	assert.True(t, messenger.HasTopic("test"))

	_ = messenger.Close()
}

func TestLibp2pMessenger_HasTopicIfDoNotHaveTopicShouldReturnFalse(t *testing.T) {
	messenger := createMockMessenger()

	_ = messenger.CreateTopic("test", false)

	assert.False(t, messenger.HasTopic("one topic"))

	_ = messenger.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorOnInexistentTopicShouldWork(t *testing.T) {
	messenger := createMockMessenger()

	err := messenger.RegisterMessageProcessor("test", "identifier", &mock.MessageProcessorStub{})

	assert.Nil(t, err)

	_ = messenger.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorWithNilHandlerShouldErr(t *testing.T) {
	messenger := createMockMessenger()

	_ = messenger.CreateTopic("test", false)

	err := messenger.RegisterMessageProcessor("test", "identifier", nil)

	assert.True(t, errors.Is(err, p2p.ErrNilValidator))

	_ = messenger.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorOkValsShouldWork(t *testing.T) {
	messenger := createMockMessenger()

	_ = messenger.CreateTopic("test", false)

	err := messenger.RegisterMessageProcessor("test", "identifier", &mock.MessageProcessorStub{})

	assert.Nil(t, err)

	_ = messenger.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorReregistrationShouldErr(t *testing.T) {
	messenger := createMockMessenger()
	_ = messenger.CreateTopic("test", false)
	// registration
	_ = messenger.RegisterMessageProcessor("test", "identifier", &mock.MessageProcessorStub{})
	// re-registration
	err := messenger.RegisterMessageProcessor("test", "identifier", &mock.MessageProcessorStub{})

	assert.True(t, errors.Is(err, p2p.ErrMessageProcessorAlreadyDefined))

	_ = messenger.Close()
}

func TestLibp2pMessenger_UnegisterTopicValidatorOnANotRegisteredTopicShouldNotErr(t *testing.T) {
	messenger := createMockMessenger()

	_ = messenger.CreateTopic("test", false)
	err := messenger.UnregisterMessageProcessor("test", "identifier")

	assert.Nil(t, err)

	_ = messenger.Close()
}

func TestLibp2pMessenger_UnregisterTopicValidatorShouldWork(t *testing.T) {
	messenger := createMockMessenger()

	_ = messenger.CreateTopic("test", false)

	// registration
	_ = messenger.RegisterMessageProcessor("test", "identifier", &mock.MessageProcessorStub{})

	// unregistration
	err := messenger.UnregisterMessageProcessor("test", "identifier")

	assert.Nil(t, err)

	_ = messenger.Close()
}

func TestLibp2pMessenger_UnregisterAllTopicValidatorShouldWork(t *testing.T) {
	messenger := createMockMessenger()
	_ = messenger.CreateTopic("test", false)
	// registration
	_ = messenger.CreateTopic("test1", false)
	_ = messenger.RegisterMessageProcessor("test1", "identifier", &mock.MessageProcessorStub{})
	_ = messenger.CreateTopic("test2", false)
	_ = messenger.RegisterMessageProcessor("test2", "identifier", &mock.MessageProcessorStub{})
	// unregistration
	err := messenger.UnregisterAllMessageProcessors()
	assert.Nil(t, err)
	err = messenger.RegisterMessageProcessor("test1", "identifier", &mock.MessageProcessorStub{})
	assert.Nil(t, err)
	err = messenger.RegisterMessageProcessor("test2", "identifier", &mock.MessageProcessorStub{})
	assert.Nil(t, err)
	_ = messenger.Close()
}

func TestLibp2pMessenger_RegisterUnregisterConcurrentlyShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	messenger := createMockMessenger()
	topic := "test topic"
	_ = messenger.CreateTopic(topic, false)

	numIdentifiers := 100
	identifiers := make([]string, 0, numIdentifiers)
	for i := 0; i < numIdentifiers; i++ {
		identifiers = append(identifiers, fmt.Sprintf("identifier%d", i))
	}

	wg := sync.WaitGroup{}
	wg.Add(numIdentifiers * 3)
	for i := 0; i < numIdentifiers; i++ {
		go func(index int) {
			_ = messenger.RegisterMessageProcessor(topic, identifiers[index], &mock.MessageProcessorStub{})
			wg.Done()
		}(i)

		go func(index int) {
			_ = messenger.UnregisterMessageProcessor(topic, identifiers[index])
			wg.Done()
		}(i)

		go func() {
			messenger.Broadcast(topic, []byte("buff"))
			wg.Done()
		}()
	}

	wg.Wait()
	_ = messenger.Close()
}

func TestLibp2pMessenger_BroadcastDataLargeMessageShouldNotCallSend(t *testing.T) {
	msg := make([]byte, libp2p.MaxSendBuffSize+1)
	messenger, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())
	messenger.SetLoadBalancer(&mock.ChannelLoadBalancerStub{
		GetChannelOrDefaultCalled: func(pipe string) chan *p2p.SendableData {
			assert.Fail(t, "should have not got to this line")

			return make(chan *p2p.SendableData, 1)
		},
		CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
			return nil
		},
	})

	messenger.Broadcast("topic", msg)

	_ = messenger.Close()
}

func TestLibp2pMessenger_BroadcastDataBetween2PeersShouldWork(t *testing.T) {
	msg := []byte("test message")

	_, messenger1, messenger2 := createMockNetworkOf2()

	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(2)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	prepareMessengerForMatchDataReceive(messenger1, msg, wg)
	prepareMessengerForMatchDataReceive(messenger2, msg, wg)

	fmt.Println("Delaying as to allow peers to announce themselves on the opened topic...")
	time.Sleep(time.Second)

	fmt.Printf("sending message from %s...\n", messenger1.ID().Pretty())

	messenger1.Broadcast("test", msg)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)

	_ = messenger1.Close()
	_ = messenger2.Close()
}

func TestLibp2pMessenger_BroadcastOnChannelBlockingShouldLimitNumberOfGoRoutines(t *testing.T) {
	if testing.Short() {
		t.Skip("this test does not perform well in TC with race detector on")
	}

	msg := []byte("test message")
	numBroadcasts := libp2p.BroadcastGoRoutines + 5

	ch := make(chan *p2p.SendableData)

	wg := sync.WaitGroup{}
	wg.Add(numBroadcasts)

	messenger, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())
	messenger.SetLoadBalancer(&mock.ChannelLoadBalancerStub{
		CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
			return nil
		},
		GetChannelOrDefaultCalled: func(pipe string) chan *p2p.SendableData {
			wg.Done()
			return ch
		},
	})

	numErrors := uint32(0)

	for i := 0; i < numBroadcasts; i++ {
		go func() {
			err := messenger.BroadcastOnChannelBlocking("test", "test", msg)
			if err == p2p.ErrTooManyGoroutines {
				atomic.AddUint32(&numErrors, 1)
				wg.Done()
			}
		}()
	}

	wg.Wait()

	// cleanup stuck go routines that are trying to write on the ch channel
	for i := 0; i < libp2p.BroadcastGoRoutines; i++ {
		select {
		case <-ch:
		default:
		}
	}

	assert.True(t, atomic.LoadUint32(&numErrors) > 0)

	_ = messenger.Close()
}

func TestLibp2pMessenger_BroadcastDataBetween2PeersWithLargeMsgShouldWork(t *testing.T) {
	msg := bytes.Repeat([]byte{'A'}, libp2p.MaxSendBuffSize)

	_, messenger1, messenger2 := createMockNetworkOf2()

	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(2)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	prepareMessengerForMatchDataReceive(messenger1, msg, wg)
	prepareMessengerForMatchDataReceive(messenger2, msg, wg)

	fmt.Println("Delaying as to allow peers to announce themselves on the opened topic...")
	time.Sleep(time.Second)

	fmt.Printf("sending message from %s...\n", messenger1.ID().Pretty())

	messenger1.Broadcast("test", msg)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)

	_ = messenger1.Close()
	_ = messenger2.Close()
}

func TestLibp2pMessenger_Peers(t *testing.T) {
	_, messenger1, messenger2 := createMockNetworkOf2()

	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)

	// should know both peers
	foundCurrent := false
	foundConnected := false

	for _, p := range messenger1.Peers() {
		fmt.Println(p.Pretty())

		if p.Pretty() == messenger1.ID().Pretty() {
			foundCurrent = true
		}
		if p.Pretty() == messenger2.ID().Pretty() {
			foundConnected = true
		}
	}

	assert.True(t, foundCurrent && foundConnected)

	_ = messenger1.Close()
	_ = messenger2.Close()
}

func TestLibp2pMessenger_ConnectedPeers(t *testing.T) {
	netw, messenger1, messenger2 := createMockNetworkOf2()
	messenger3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)
	_ = messenger3.ConnectToPeer(adr2)

	// connected peers:  1 ----- 2 ----- 3

	assert.Equal(t, []core.PeerID{messenger2.ID()}, messenger1.ConnectedPeers())
	assert.Equal(t, []core.PeerID{messenger2.ID()}, messenger3.ConnectedPeers())
	assert.Equal(t, 2, len(messenger2.ConnectedPeers()))
	// no need to further test that messenger2 is connected to messenger1 and messenger3 as this was tested in first 2 asserts

	_ = messenger1.Close()
	_ = messenger2.Close()
	_ = messenger3.Close()
}

func TestLibp2pMessenger_ConnectedAddresses(t *testing.T) {
	netw, messenger1, messenger2 := createMockNetworkOf2()
	messenger3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)
	_ = messenger3.ConnectToPeer(adr2)

	// connected peers:  1 ----- 2 ----- 3

	foundAddr1 := false
	foundAddr3 := false

	for _, addr := range messenger2.ConnectedAddresses() {
		for _, address := range messenger1.Addresses() {
			if addr == address {
				foundAddr1 = true
			}
		}

		for _, address := range messenger3.Addresses() {
			if addr == address {
				foundAddr3 = true
			}
		}
	}

	assert.True(t, foundAddr1)
	assert.True(t, foundAddr3)
	assert.Equal(t, 2, len(messenger2.ConnectedAddresses()))
	// no need to further test that messenger2 is connected to messenger1 and messenger3 as this was tested in first 2 asserts

	_ = messenger1.Close()
	_ = messenger2.Close()
	_ = messenger3.Close()
}

func TestLibp2pMessenger_PeerAddressConnectedPeerShouldWork(t *testing.T) {
	netw, messenger1, messenger2 := createMockNetworkOf2()
	messenger3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)
	_ = messenger3.ConnectToPeer(adr2)

	// connected peers:  1 ----- 2 ----- 3

	defer func() {
		_ = messenger1.Close()
		_ = messenger2.Close()
		_ = messenger3.Close()
	}()

	addressesRecov := messenger2.PeerAddresses(messenger1.ID())
	for _, addr := range messenger1.Addresses() {
		for _, addrRecov := range addressesRecov {
			if strings.Contains(addr, addrRecov) {
				// address returned is valid, test is successful
				return
			}
		}
	}

	assert.Fail(t, "Returned address is not valid!")
}

func TestLibp2pMessenger_PeerAddressNotConnectedShouldReturnFromPeerstore(t *testing.T) {
	netw := mocknet.New(context.Background())
	messenger, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	networkHandler := &mock.NetworkStub{
		ConnsCalled: func() []network.Conn {
			return nil
		},
	}

	peerstoreHandler := &mock.PeerstoreStub{
		AddrsCalled: func(p peer.ID) []multiaddr.Multiaddr {
			return []multiaddr.Multiaddr{
				&mock.MultiaddrStub{
					StringCalled: func() string {
						return "multiaddress 1"
					},
				},
				&mock.MultiaddrStub{
					StringCalled: func() string {
						return "multiaddress 2"
					},
				},
			}
		},
	}

	messenger.SetHost(&mock.ConnectableHostStub{
		NetworkCalled: func() network.Network {
			return networkHandler
		},
		PeerstoreCalled: func() peerstore.Peerstore {
			return peerstoreHandler
		},
	})

	addresses := messenger.PeerAddresses("pid")
	require.Equal(t, 2, len(addresses))
	assert.Equal(t, addresses[0], "multiaddress 1")
	assert.Equal(t, addresses[1], "multiaddress 2")
}

func TestLibp2pMessenger_PeerAddressDisconnectedPeerShouldWork(t *testing.T) {
	netw, messenger1, messenger2 := createMockNetworkOf2()
	messenger3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)
	_ = messenger3.ConnectToPeer(adr2)

	defer func() {
		_ = messenger1.Close()
		_ = messenger2.Close()
		_ = messenger3.Close()
	}()

	_ = netw.UnlinkPeers(peer.ID(messenger1.ID().Bytes()), peer.ID(messenger2.ID().Bytes()))
	_ = netw.DisconnectPeers(peer.ID(messenger1.ID().Bytes()), peer.ID(messenger2.ID().Bytes()))
	_ = netw.DisconnectPeers(peer.ID(messenger2.ID().Bytes()), peer.ID(messenger1.ID().Bytes()))

	// connected peers:  1 --x-- 2 ----- 3

	assert.False(t, messenger2.IsConnected(messenger1.ID()))
}

func TestLibp2pMessenger_PeerAddressUnknownPeerShouldReturnEmpty(t *testing.T) {
	_, messenger1, _ := createMockNetworkOf2()

	defer func() {
		_ = messenger1.Close()
	}()

	adr1Recov := messenger1.PeerAddresses("unknown peer")
	assert.Equal(t, 0, len(adr1Recov))
}

// ------- ConnectedPeersOnTopic

func TestLibp2pMessenger_ConnectedPeersOnTopicInvalidTopicShouldRetEmptyList(t *testing.T) {
	netw, messenger1, messenger2 := createMockNetworkOf2()
	messenger3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = netw.LinkAll()

	adr2 := messenger2.Addresses()[0]
	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)
	_ = messenger3.ConnectToPeer(adr2)
	// connected peers:  1 ----- 2 ----- 3
	connPeers := messenger1.ConnectedPeersOnTopic("non-existent topic")
	assert.Equal(t, 0, len(connPeers))

	_ = messenger1.Close()
	_ = messenger2.Close()
	_ = messenger3.Close()
}

func TestLibp2pMessenger_ConnectedPeersOnTopicOneTopicShouldWork(t *testing.T) {
	netw, messenger1, messenger2 := createMockNetworkOf2()
	messenger3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	messenger4, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = netw.LinkAll()

	adr2 := messenger2.Addresses()[0]
	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)
	_ = messenger3.ConnectToPeer(adr2)
	_ = messenger4.ConnectToPeer(adr2)
	// connected peers:  1 ----- 2 ----- 3
	//                          |
	//                          4
	// 1, 2, 3 should be on topic "topic123"
	_ = messenger1.CreateTopic("topic123", false)
	_ = messenger2.CreateTopic("topic123", false)
	_ = messenger3.CreateTopic("topic123", false)

	// wait a bit for topic announcements
	time.Sleep(time.Second)

	peersOnTopic123 := messenger2.ConnectedPeersOnTopic("topic123")

	assert.Equal(t, 2, len(peersOnTopic123))
	assert.True(t, containsPeerID(peersOnTopic123, messenger1.ID()))
	assert.True(t, containsPeerID(peersOnTopic123, messenger3.ID()))

	_ = messenger1.Close()
	_ = messenger2.Close()
	_ = messenger3.Close()
	_ = messenger4.Close()
}

func TestLibp2pMessenger_ConnectedPeersOnTopicOneTopicDifferentViewsShouldWork(t *testing.T) {
	netw, messenger1, messenger2 := createMockNetworkOf2()
	messenger3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	messenger4, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = netw.LinkAll()

	adr2 := messenger2.Addresses()[0]
	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)
	_ = messenger3.ConnectToPeer(adr2)
	_ = messenger4.ConnectToPeer(adr2)
	// connected peers:  1 ----- 2 ----- 3
	//                          |
	//                          4
	// 1, 2, 3 should be on topic "topic123"
	_ = messenger1.CreateTopic("topic123", false)
	_ = messenger2.CreateTopic("topic123", false)
	_ = messenger3.CreateTopic("topic123", false)

	// wait a bit for topic announcements
	time.Sleep(time.Second)

	peersOnTopic123FromMessenger2 := messenger2.ConnectedPeersOnTopic("topic123")
	peersOnTopic123FromMessenger4 := messenger4.ConnectedPeersOnTopic("topic123")

	// keep the same checks as the test above as to be 100% that the returned list are correct
	assert.Equal(t, 2, len(peersOnTopic123FromMessenger2))
	assert.True(t, containsPeerID(peersOnTopic123FromMessenger2, messenger1.ID()))
	assert.True(t, containsPeerID(peersOnTopic123FromMessenger2, messenger3.ID()))

	assert.Equal(t, 1, len(peersOnTopic123FromMessenger4))
	assert.True(t, containsPeerID(peersOnTopic123FromMessenger4, messenger2.ID()))

	_ = messenger1.Close()
	_ = messenger2.Close()
	_ = messenger3.Close()
	_ = messenger4.Close()
}

func TestLibp2pMessenger_ConnectedPeersOnTopicTwoTopicsShouldWork(t *testing.T) {
	netw, messenger1, messenger2 := createMockNetworkOf2()
	messenger3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	messenger4, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = netw.LinkAll()

	adr2 := messenger2.Addresses()[0]
	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)
	_ = messenger3.ConnectToPeer(adr2)
	_ = messenger4.ConnectToPeer(adr2)
	// connected peers:  1 ----- 2 ----- 3
	//                          |
	//                          4
	// 1, 2, 3 should be on topic "topic123"
	// 2, 4 should be on topic "topic24"
	_ = messenger1.CreateTopic("topic123", false)
	_ = messenger2.CreateTopic("topic123", false)
	_ = messenger2.CreateTopic("topic24", false)
	_ = messenger3.CreateTopic("topic123", false)
	_ = messenger4.CreateTopic("topic24", false)

	// wait a bit for topic announcements
	time.Sleep(time.Second)

	peersOnTopic123 := messenger2.ConnectedPeersOnTopic("topic123")
	peersOnTopic24 := messenger2.ConnectedPeersOnTopic("topic24")

	// keep the same checks as the test above as to be 100% that the returned list are correct
	assert.Equal(t, 2, len(peersOnTopic123))
	assert.True(t, containsPeerID(peersOnTopic123, messenger1.ID()))
	assert.True(t, containsPeerID(peersOnTopic123, messenger3.ID()))

	assert.Equal(t, 1, len(peersOnTopic24))
	assert.True(t, containsPeerID(peersOnTopic24, messenger4.ID()))

	_ = messenger1.Close()
	_ = messenger2.Close()
	_ = messenger3.Close()
	_ = messenger4.Close()
}

// ------- ConnectedFullHistoryPeersOnTopic

func TestLibp2pMessenger_ConnectedFullHistoryPeersOnTopicShouldWork(t *testing.T) {
	messenger1, messenger2, messenger3 := createMockNetworkOf3()

	adr2 := messenger2.Addresses()[0]
	adr3 := messenger3.Addresses()[0]
	fmt.Println("Connecting ...")

	_ = messenger1.ConnectToPeer(adr2)
	_ = messenger3.ConnectToPeer(adr2)
	_ = messenger1.ConnectToPeer(adr3)
	// connected peers:  1 ----- 2
	//                   |       |
	//                   3 ------+

	_ = messenger1.CreateTopic("topic123", false)
	_ = messenger2.CreateTopic("topic123", false)
	_ = messenger3.CreateTopic("topic123", false)

	// wait a bit for topic announcements
	time.Sleep(time.Second)

	assert.Equal(t, 2, len(messenger1.ConnectedPeersOnTopic("topic123")))
	assert.Equal(t, 1, len(messenger1.ConnectedFullHistoryPeersOnTopic("topic123")))

	assert.Equal(t, 2, len(messenger2.ConnectedPeersOnTopic("topic123")))
	assert.Equal(t, 1, len(messenger2.ConnectedFullHistoryPeersOnTopic("topic123")))

	assert.Equal(t, 2, len(messenger3.ConnectedPeersOnTopic("topic123")))
	assert.Equal(t, 2, len(messenger3.ConnectedFullHistoryPeersOnTopic("topic123")))

	_ = messenger1.Close()
	_ = messenger2.Close()
	_ = messenger3.Close()
}

func TestLibp2pMessenger_ConnectedPeersShouldReturnUniquePeers(t *testing.T) {
	pid1 := core.PeerID("pid1")
	pid2 := core.PeerID("pid2")
	pid3 := core.PeerID("pid3")
	pid4 := core.PeerID("pid4")

	hs := &mock.ConnectableHostStub{
		NetworkCalled: func() network.Network {
			return &mock.NetworkStub{
				ConnsCalled: func() []network.Conn {
					// generate a mock list that contain duplicates
					return []network.Conn{
						generateConnWithRemotePeer(pid1),
						generateConnWithRemotePeer(pid1),
						generateConnWithRemotePeer(pid2),
						generateConnWithRemotePeer(pid1),
						generateConnWithRemotePeer(pid4),
						generateConnWithRemotePeer(pid3),
						generateConnWithRemotePeer(pid1),
						generateConnWithRemotePeer(pid3),
						generateConnWithRemotePeer(pid4),
						generateConnWithRemotePeer(pid2),
						generateConnWithRemotePeer(pid1),
						generateConnWithRemotePeer(pid1),
					}
				},
				ConnectednessCalled: func(id peer.ID) network.Connectedness {
					return network.Connected
				},
			}
		},
	}

	netw := mocknet.New(context.Background())
	mes, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	// we can safely close the host as the next operations will be done on a mock
	_ = mes.Close()

	mes.SetHost(hs)

	peerList := mes.ConnectedPeers()

	assert.Equal(t, 4, len(peerList))
	assert.True(t, existInList(peerList, pid1))
	assert.True(t, existInList(peerList, pid2))
	assert.True(t, existInList(peerList, pid3))
	assert.True(t, existInList(peerList, pid4))
}

func existInList(list []core.PeerID, pid core.PeerID) bool {
	for _, p := range list {
		if bytes.Equal(p.Bytes(), pid.Bytes()) {
			return true
		}
	}

	return false
}

func generateConnWithRemotePeer(pid core.PeerID) network.Conn {
	return &mock.ConnStub{
		RemotePeerCalled: func() peer.ID {
			return peer.ID(pid)
		},
	}
}

func TestLibp2pMessenger_SendDirectWithMockNetToConnectedPeerShouldWork(t *testing.T) {
	msg := []byte("test message")

	_, messenger1, messenger2 := createMockNetworkOf2()

	adr2 := messenger2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = messenger1.ConnectToPeer(adr2)

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(1)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	prepareMessengerForMatchDataReceive(messenger2, msg, wg)

	fmt.Println("Delaying as to allow peers to announce themselves on the opened topic...")
	time.Sleep(time.Second)

	fmt.Printf("sending message from %s...\n", messenger1.ID().Pretty())

	err := messenger1.SendToConnectedPeer("test", msg, messenger2.ID())

	assert.Nil(t, err)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)

	_ = messenger1.Close()
	_ = messenger2.Close()
}

func TestLibp2pMessenger_SendDirectWithRealNetToConnectedPeerShouldWork(t *testing.T) {
	msg := []byte("test message")

	fmt.Println("Messenger 1:")
	messenger1, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

	fmt.Println("Messenger 2:")
	messenger2, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

	err := messenger1.ConnectToPeer(getConnectableAddress(messenger2))
	assert.Nil(t, err)

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(2)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	prepareMessengerForMatchDataReceive(messenger1, msg, wg)
	prepareMessengerForMatchDataReceive(messenger2, msg, wg)

	fmt.Println("Delaying as to allow peers to announce themselves on the opened topic...")
	time.Sleep(time.Second)

	fmt.Printf("Messenger 1 is sending message from %s...\n", messenger1.ID().Pretty())
	err = messenger1.SendToConnectedPeer("test", msg, messenger2.ID())
	assert.Nil(t, err)

	time.Sleep(time.Second)
	fmt.Printf("Messenger 2 is sending message from %s...\n", messenger2.ID().Pretty())
	err = messenger2.SendToConnectedPeer("test", msg, messenger1.ID())
	assert.Nil(t, err)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)

	_ = messenger1.Close()
	_ = messenger2.Close()
}

func TestLibp2pMessenger_SendDirectWithRealNetToSelfShouldWork(t *testing.T) {
	msg := []byte("test message")

	fmt.Println("Messenger 1:")
	mes, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(1)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	prepareMessengerForMatchDataReceive(mes, msg, wg)

	fmt.Printf("Messenger 1 is sending message from %s to self...\n", mes.ID().Pretty())
	err := mes.SendToConnectedPeer("test", msg, mes.ID())
	assert.Nil(t, err)

	time.Sleep(time.Second)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)

	_ = mes.Close()
}

// ------- Bootstrap

func TestNetworkMessenger_BootstrapPeerDiscoveryShouldCallPeerBootstrapper(t *testing.T) {
	wasCalled := false

	netw := mocknet.New(context.Background())
	pdm := &mock.PeerDiscovererStub{
		BootstrapCalled: func() error {
			wasCalled = true
			return nil
		},
		CloseCalled: func() error {
			return nil
		},
	}
	mes, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	mes.SetPeerDiscoverer(pdm)

	_ = mes.Bootstrap()

	assert.True(t, wasCalled)

	_ = mes.Close()
}

// ------- SetThresholdMinConnectedPeers

func TestNetworkMessenger_SetThresholdMinConnectedPeersInvalidValueShouldErr(t *testing.T) {
	messenger := createMockMessenger()
	defer func() {
		_ = messenger.Close()
	}()

	err := messenger.SetThresholdMinConnectedPeers(-1)

	assert.Equal(t, p2p.ErrInvalidValue, err)
}

func TestNetworkMessenger_SetThresholdMinConnectedPeersShouldWork(t *testing.T) {
	messenger := createMockMessenger()
	defer func() {
		_ = messenger.Close()
	}()

	minConnectedPeers := 56
	err := messenger.SetThresholdMinConnectedPeers(minConnectedPeers)

	assert.Nil(t, err)
	assert.Equal(t, minConnectedPeers, messenger.ThresholdMinConnectedPeers())
}

// ------- IsConnectedToTheNetwork

func TestNetworkMessenger_IsConnectedToTheNetworkRetFalse(t *testing.T) {
	messenger := createMockMessenger()
	defer func() {
		_ = messenger.Close()
	}()

	minConnectedPeers := 56
	_ = messenger.SetThresholdMinConnectedPeers(minConnectedPeers)

	assert.False(t, messenger.IsConnectedToTheNetwork())
}

func TestNetworkMessenger_IsConnectedToTheNetworkWithZeroRetTrue(t *testing.T) {
	messenger := createMockMessenger()
	defer func() {
		_ = messenger.Close()
	}()

	minConnectedPeers := 0
	_ = messenger.SetThresholdMinConnectedPeers(minConnectedPeers)

	assert.True(t, messenger.IsConnectedToTheNetwork())
}

// ------- SetPeerShardResolver

func TestNetworkMessenger_SetPeerShardResolverNilShouldErr(t *testing.T) {
	messenger := createMockMessenger()
	defer func() {
		_ = messenger.Close()
	}()

	err := messenger.SetPeerShardResolver(nil)

	assert.Equal(t, p2p.ErrNilPeerShardResolver, err)
}

func TestNetworkMessenger_SetPeerShardResolver(t *testing.T) {
	messenger := createMockMessenger()
	defer func() {
		_ = messenger.Close()
	}()

	err := messenger.SetPeerShardResolver(&mock.PeerShardResolverStub{})

	assert.Nil(t, err)
}

func TestNetworkMessenger_DoubleCloseShouldWork(t *testing.T) {
	messenger := createMessenger()

	time.Sleep(time.Second)

	err := messenger.Close()
	assert.Nil(t, err)

	err = messenger.Close()
	assert.Nil(t, err)
}

func TestNetworkMessenger_PreventReprocessingShouldWork(t *testing.T) {
	args := libp2p.ArgsNetworkMessenger{
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		Marshalizer:   &testscommon.ProtoMarshalizerMock{},
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port:                  "0",
				ConnectionWatcherType: "print",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer:            &libp2p.LocalSyncTimer{},
		PreferredPeersHolder: &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:   &p2pmocks.PeersRatingHandlerStub{},
	}

	mes, _ := libp2p.NewNetworkMessenger(args)

	numCalled := uint32(0)
	handler := &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			atomic.AddUint32(&numCalled, 1)
			return nil
		},
	}

	callBackFunc := mes.PubsubCallback(handler, "")
	ctx := context.Background()
	pid := peer.ID(mes.ID())
	timeStamp := time.Now().Unix() - 1
	timeStamp -= int64(libp2p.AcceptMessagesInAdvanceDuration.Seconds())
	timeStamp -= int64(libp2p.PubsubTimeCacheDuration.Seconds())

	innerMessage := &data.TopicMessage{
		Payload:   []byte("data"),
		Timestamp: timeStamp,
	}
	buff, _ := args.Marshalizer.Marshal(innerMessage)
	msg := &pubsub.Message{
		Message: &pubsubPb.Message{
			From:                 []byte(pid),
			Data:                 buff,
			Seqno:                []byte{0, 0, 0, 1},
			Topic:                nil,
			Signature:            nil,
			Key:                  nil,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		},
		ReceivedFrom:  "",
		ValidatorData: nil,
	}

	assert.False(t, callBackFunc(ctx, pid, msg)) // this will not call
	assert.False(t, callBackFunc(ctx, pid, msg)) // this will not call
	assert.Equal(t, uint32(0), atomic.LoadUint32(&numCalled))

	_ = mes.Close()
}

func TestNetworkMessenger_PubsubCallbackNotMessageNotValidShouldNotCallHandler(t *testing.T) {
	args := libp2p.ArgsNetworkMessenger{
		Marshalizer:   &testscommon.ProtoMarshalizerMock{},
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port:                  "0",
				ConnectionWatcherType: "print",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer:            &libp2p.LocalSyncTimer{},
		PreferredPeersHolder: &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:   &p2pmocks.PeersRatingHandlerStub{},
	}

	mes, _ := libp2p.NewNetworkMessenger(args)
	numUpserts := int32(0)
	_ = mes.SetPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{
		UpsertPeerIDCalled: func(pid core.PeerID, duration time.Duration) error {
			atomic.AddInt32(&numUpserts, 1)
			// any error thrown here should not impact the execution
			return fmt.Errorf("expected error")
		},
		IsDeniedCalled: func(pid core.PeerID) bool {
			return false
		},
	})

	numCalled := uint32(0)
	handler := &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			atomic.AddUint32(&numCalled, 1)
			return nil
		},
	}

	callBackFunc := mes.PubsubCallback(handler, "")
	ctx := context.Background()
	pid := peer.ID(mes.ID())
	innerMessage := &data.TopicMessage{
		Payload:   []byte("data"),
		Timestamp: time.Now().Unix(),
	}
	buff, _ := args.Marshalizer.Marshal(innerMessage)
	msg := &pubsub.Message{
		Message: &pubsubPb.Message{
			From:                 []byte("not a valid pid"),
			Data:                 buff,
			Seqno:                []byte{0, 0, 0, 1},
			Topic:                nil,
			Signature:            nil,
			Key:                  nil,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		},
		ReceivedFrom:  "",
		ValidatorData: nil,
	}

	assert.False(t, callBackFunc(ctx, pid, msg))
	assert.Equal(t, uint32(0), atomic.LoadUint32(&numCalled))
	assert.Equal(t, int32(2), atomic.LoadInt32(&numUpserts))

	_ = mes.Close()
}

func TestNetworkMessenger_PubsubCallbackReturnsFalseIfHandlerErrors(t *testing.T) {
	args := libp2p.ArgsNetworkMessenger{
		Marshalizer:   &testscommon.ProtoMarshalizerMock{},
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port:                  "0",
				ConnectionWatcherType: "print",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer:            &libp2p.LocalSyncTimer{},
		PreferredPeersHolder: &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:   &p2pmocks.PeersRatingHandlerStub{},
	}

	mes, _ := libp2p.NewNetworkMessenger(args)

	numCalled := uint32(0)
	expectedErr := errors.New("expected error")
	handler := &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			atomic.AddUint32(&numCalled, 1)
			return expectedErr
		},
	}

	callBackFunc := mes.PubsubCallback(handler, "")
	ctx := context.Background()
	pid := peer.ID(mes.ID())
	innerMessage := &data.TopicMessage{
		Payload:   []byte("data"),
		Timestamp: time.Now().Unix(),
		Version:   libp2p.CurrentTopicMessageVersion,
	}
	buff, _ := args.Marshalizer.Marshal(innerMessage)
	topic := "topic"
	msg := &pubsub.Message{
		Message: &pubsubPb.Message{
			From:                 []byte(mes.ID()),
			Data:                 buff,
			Seqno:                []byte{0, 0, 0, 1},
			Topic:                &topic,
			Signature:            nil,
			Key:                  nil,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		},
		ReceivedFrom:  "",
		ValidatorData: nil,
	}

	assert.False(t, callBackFunc(ctx, pid, msg))
	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalled))

	_ = mes.Close()
}

func TestNetworkMessenger_UnjoinAllTopicsShouldWork(t *testing.T) {
	args := libp2p.ArgsNetworkMessenger{
		Marshalizer:   &testscommon.ProtoMarshalizerMock{},
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port:                  "0",
				ConnectionWatcherType: "print",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer:            &libp2p.LocalSyncTimer{},
		PreferredPeersHolder: &p2pmocks.PeersHolderStub{},
		PeersRatingHandler:   &p2pmocks.PeersRatingHandlerStub{},
	}

	mes, _ := libp2p.NewNetworkMessenger(args)

	topic := "topic"
	_ = mes.CreateTopic(topic, true)
	assert.True(t, mes.HasTopic(topic))

	err := mes.UnjoinAllTopics()
	assert.Nil(t, err)

	assert.False(t, mes.HasTopic(topic))
}

func TestNetworkMessenger_ValidMessageByTimestampMessageTooOld(t *testing.T) {
	args := createMockNetworkArgs()
	now := time.Now()
	args.SyncTimer = &mock.SyncTimerStub{
		CurrentTimeCalled: func() time.Time {
			return now
		},
	}
	mes, _ := libp2p.NewNetworkMessenger(args)

	msg := &message.Message{
		TimestampField: now.Unix() - int64(libp2p.PubsubTimeCacheDuration.Seconds()) - 1,
	}
	err := mes.ValidMessageByTimestamp(msg)

	assert.True(t, errors.Is(err, p2p.ErrMessageTooOld))
}

func TestNetworkMessenger_ValidMessageByTimestampMessageAtLowerLimitShouldWork(t *testing.T) {
	mes, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

	now := time.Now()
	msg := &message.Message{
		TimestampField: now.Unix() - int64(libp2p.PubsubTimeCacheDuration.Seconds()) + int64(libp2p.AcceptMessagesInAdvanceDuration.Seconds()),
	}
	err := mes.ValidMessageByTimestamp(msg)

	assert.Nil(t, err)
}

func TestNetworkMessenger_ValidMessageByTimestampMessageTooNew(t *testing.T) {
	args := createMockNetworkArgs()
	now := time.Now()
	args.SyncTimer = &mock.SyncTimerStub{
		CurrentTimeCalled: func() time.Time {
			return now
		},
	}
	mes, _ := libp2p.NewNetworkMessenger(args)

	msg := &message.Message{
		TimestampField: now.Unix() + int64(libp2p.AcceptMessagesInAdvanceDuration.Seconds()) + 1,
	}
	err := mes.ValidMessageByTimestamp(msg)

	assert.True(t, errors.Is(err, p2p.ErrMessageTooNew))
}

func TestNetworkMessenger_ValidMessageByTimestampMessageAtUpperLimitShouldWork(t *testing.T) {
	args := createMockNetworkArgs()
	now := time.Now()
	args.SyncTimer = &mock.SyncTimerStub{
		CurrentTimeCalled: func() time.Time {
			return now
		},
	}
	mes, _ := libp2p.NewNetworkMessenger(args)

	msg := &message.Message{
		TimestampField: now.Unix() + int64(libp2p.AcceptMessagesInAdvanceDuration.Seconds()),
	}
	err := mes.ValidMessageByTimestamp(msg)

	assert.Nil(t, err)
}

func TestNetworkMessenger_GetConnectedPeersInfo(t *testing.T) {
	netw := mocknet.New(context.Background())

	peers := []peer.ID{
		"valI1",
		"valC1",
		"valC2",
		"obsI1",
		"obsI2",
		"obsI3",
		"obsC1",
		"obsC2",
		"obsC3",
		"obsC4",
		"unknown",
	}
	mes, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	mes.SetHost(&mock.ConnectableHostStub{
		NetworkCalled: func() network.Network {
			return &mock.NetworkStub{
				PeersCall: func() []peer.ID {
					return peers
				},
				ConnsToPeerCalled: func(p peer.ID) []network.Conn {
					return make([]network.Conn, 0)
				},
			}
		},
	})
	selfShardID := uint32(0)
	crossShardID := uint32(1)
	_ = mes.SetPeerShardResolver(&mock.PeerShardResolverStub{
		GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
			pinfo := core.P2PPeerInfo{
				PeerType: core.UnknownPeer,
			}
			if pid.Pretty() == mes.ID().Pretty() {
				pinfo.ShardID = selfShardID
				pinfo.PeerType = core.ObserverPeer
				return pinfo
			}

			strPid := string(pid)
			if strings.Contains(strPid, "I") {
				pinfo.ShardID = selfShardID
			}
			if strings.Contains(strPid, "C") {
				pinfo.ShardID = crossShardID
			}

			if strings.Contains(strPid, "val") {
				pinfo.PeerType = core.ValidatorPeer
			}

			if strings.Contains(strPid, "obs") {
				pinfo.PeerType = core.ObserverPeer
			}

			return pinfo
		},
	})

	cpi := mes.GetConnectedPeersInfo()

	assert.Equal(t, 4, cpi.NumCrossShardObservers)
	assert.Equal(t, 2, cpi.NumCrossShardValidators)
	assert.Equal(t, 3, cpi.NumIntraShardObservers)
	assert.Equal(t, 1, cpi.NumIntraShardValidators)
	assert.Equal(t, 3, cpi.NumObserversOnShard[selfShardID])
	assert.Equal(t, 4, cpi.NumObserversOnShard[crossShardID])
	assert.Equal(t, 1, cpi.NumValidatorsOnShard[selfShardID])
	assert.Equal(t, 2, cpi.NumValidatorsOnShard[crossShardID])
	assert.Equal(t, selfShardID, cpi.SelfShardID)
	assert.Equal(t, 1, len(cpi.UnknownPeers))
}

func TestNetworkMessenger_mapHistogram(t *testing.T) {
	t.Parallel()

	args := createMockNetworkArgs()
	netMes, _ := libp2p.NewNetworkMessenger(args)

	inp := map[uint32]int{
		0:                     5,
		1:                     7,
		2:                     9,
		core.MetachainShardId: 11,
	}
	output := `shard 0: 5, shard 1: 7, shard 2: 9, meta: 11`

	require.Equal(t, output, netMes.MapHistogram(inp))
}

func TestNetworkMessenger_ChooseAnotherPortIfBindFails(t *testing.T) {
	t.Skip("complex test used to debug port reuse mechanism on netMessenger")

	port := "37000-37010"
	mutMessengers := sync.Mutex{}
	messengers := make([]p2p.Messenger, 0)

	numMessengers := 10
	for i := 0; i < numMessengers; i++ {
		go func() {
			time.Sleep(time.Millisecond)
			args := createMockNetworkArgs()
			args.P2pConfig.Node.Port = port

			netMes, err := libp2p.NewNetworkMessengerWithoutPortReuse(args)
			assert.Nil(t, err)
			require.False(t, check.IfNil(netMes))

			mutMessengers.Lock()
			messengers = append(messengers, netMes)
			mutMessengers.Unlock()
		}()
	}

	time.Sleep(time.Second)

	mutMessengers.Lock()
	for index1, messenger1 := range messengers {
		for index2, messenger2 := range messengers {
			if index1 == index2 {
				continue
			}

			assert.NotEqual(t, messenger1.Port(), messenger2.Port())
		}
	}

	for _, messenger := range messengers {
		_ = messenger.Close()
	}
	mutMessengers.Unlock()
}

func TestNetworkMessenger_Bootstrap(t *testing.T) {
	t.Skip("long test used to debug go routines closing on the netMessenger")

	t.Parallel()

	_ = logger.SetLogLevel("*:DEBUG")
	log := logger.GetOrCreate("internal tests")

	args := libp2p.ArgsNetworkMessenger{
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		Marshalizer:   &marshal.GogoProtoMarshalizer{},
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port:                       "0",
				Seed:                       "",
				MaximumExpectedPeerCount:   1,
				ThresholdMinConnectedPeers: 1,
				ConnectionWatcherType:      "print",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled:                          true,
				Type:                             "optimized",
				RefreshIntervalInSec:             10,
				ProtocolID:                       "erd/kad/1.0.0",
				InitialPeerList:                  []string{"/ip4/35.214.140.83/tcp/10000/p2p/16Uiu2HAm6hPymvkZyFgbvWaVBKhEoPjmXhkV32r9JaFvQ7Rk8ynU"},
				BucketSize:                       10,
				RoutingTableRefreshIntervalInSec: 5,
			},
			Sharding: config.ShardingConfig{
				TargetPeerCount:         0,
				MaxIntraShardValidators: 0,
				MaxCrossShardValidators: 0,
				MaxIntraShardObservers:  0,
				MaxCrossShardObservers:  0,
				MaxSeeders:              0,
				Type:                    "NilListSharder",
			},
		},
		SyncTimer:          &mock.SyncTimerStub{},
		PeersRatingHandler: &p2pmocks.PeersRatingHandlerStub{},
		PreferredPeersHolder: &p2pmocks.PeersHolderStub{},
	}

	netMes, err := libp2p.NewNetworkMessenger(args)
	require.Nil(t, err)

	go func() {
		time.Sleep(time.Second * 1)
		goRoutinesNumberStart := runtime.NumGoroutine()
		log.Info("before closing", "num go routines", goRoutinesNumberStart)

		_ = netMes.Close()
	}()

	_ = netMes.Bootstrap()

	time.Sleep(time.Second * 5)

	goRoutinesNumberStart := runtime.NumGoroutine()
	core.DumpGoRoutinesToLog(goRoutinesNumberStart, log)
}

func TestNetworkMessenger_WaitForConnections(t *testing.T) {
	t.Parallel()

	t.Run("min num of peers is 0", func(t *testing.T) {
		t.Parallel()

		startTime := time.Now()
		_, mes1, mes2 := createMockNetworkOf2()
		_ = mes1.ConnectToPeer(mes2.Addresses()[0])

		defer func() {
			_ = mes1.Close()
			_ = mes2.Close()
		}()

		timeToWait := time.Second * 3
		mes1.WaitForConnections(timeToWait, 0)

		assert.True(t, timeToWait <= time.Since(startTime))
	})
	t.Run("min num of peers is 2", func(t *testing.T) {
		t.Parallel()

		startTime := time.Now()
		netw, mes1, mes2 := createMockNetworkOf2()
		mes3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
		_ = netw.LinkAll()

		_ = mes1.ConnectToPeer(mes2.Addresses()[0])
		go func() {
			time.Sleep(time.Second * 2)
			_ = mes1.ConnectToPeer(mes3.Addresses()[0])
		}()

		defer func() {
			_ = mes1.Close()
			_ = mes2.Close()
			_ = mes3.Close()
		}()

		timeToWait := time.Second * 10
		mes1.WaitForConnections(timeToWait, 2)

		assert.True(t, timeToWait > time.Since(startTime))
		assert.True(t, libp2p.PollWaitForConnectionsInterval <= time.Since(startTime))
	})
	t.Run("min num of peers is 2 but we only connected to 1 peer", func(t *testing.T) {
		t.Parallel()

		startTime := time.Now()
		_, mes1, mes2 := createMockNetworkOf2()

		_ = mes1.ConnectToPeer(mes2.Addresses()[0])

		defer func() {
			_ = mes1.Close()
			_ = mes2.Close()
		}()

		timeToWait := time.Second * 10
		mes1.WaitForConnections(timeToWait, 2)

		assert.True(t, timeToWait < time.Since(startTime))
	})
}

func TestLibp2pMessenger_SignVerifyPayloadShouldWork(t *testing.T) {
	fmt.Println("Messenger 1:")
	messenger1, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

	fmt.Println("Messenger 2:")
	messenger2, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

	err := messenger1.ConnectToPeer(getConnectableAddress(messenger2))
	assert.Nil(t, err)

	defer func() {
		_ = messenger1.Close()
		_ = messenger2.Close()
	}()

	payload := []byte("payload")
	sig, err := messenger1.Sign(payload)
	assert.Nil(t, err)

	err = messenger2.Verify(payload, messenger1.ID(), sig)
	assert.Nil(t, err)

	err = messenger1.Verify(payload, messenger1.ID(), sig)
	assert.Nil(t, err)
}

func TestLibp2pMessenger_ConnectionTopic(t *testing.T) {
	t.Parallel()

	t.Run("create topic should work", func(t *testing.T) {
		t.Parallel()

		netMes, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

		topic := common.ConnectionTopic
		err := netMes.CreateTopic(topic, true)
		assert.Nil(t, err)
		assert.False(t, netMes.HasTopic(topic))
		assert.False(t, netMes.PubsubHasTopic(topic))

		testTopic := "test topic"
		err = netMes.CreateTopic(testTopic, true)
		assert.Nil(t, err)
		assert.True(t, netMes.HasTopic(testTopic))
		assert.True(t, netMes.PubsubHasTopic(testTopic))

		err = netMes.UnjoinAllTopics()
		assert.Nil(t, err)
		assert.False(t, netMes.HasTopic(topic))
		assert.False(t, netMes.PubsubHasTopic(topic))
		assert.False(t, netMes.HasTopic(testTopic))
		assert.False(t, netMes.PubsubHasTopic(testTopic))

		_ = netMes.Close()
	})
	t.Run("register-unregister message processor should work", func(t *testing.T) {
		t.Parallel()

		netMes, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

		identifier := "identifier"
		topic := common.ConnectionTopic
		err := netMes.RegisterMessageProcessor(topic, identifier, &mock.MessageProcessorStub{})
		assert.Nil(t, err)
		assert.True(t, netMes.HasProcessorForTopic(topic))

		err = netMes.UnregisterMessageProcessor(topic, identifier)
		assert.Nil(t, err)
		assert.False(t, netMes.HasProcessorForTopic(topic))

		_ = netMes.Close()
	})
	t.Run("unregister all processors should work", func(t *testing.T) {
		t.Parallel()

		netMes, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

		topic := common.ConnectionTopic
		err := netMes.RegisterMessageProcessor(topic, "identifier", &mock.MessageProcessorStub{})
		assert.Nil(t, err)
		assert.True(t, netMes.HasProcessorForTopic(topic))

		testTopic := "test topic"
		err = netMes.RegisterMessageProcessor(testTopic, "identifier", &mock.MessageProcessorStub{})
		assert.Nil(t, err)
		assert.True(t, netMes.HasProcessorForTopic(testTopic))

		err = netMes.UnregisterAllMessageProcessors()
		assert.Nil(t, err)
		assert.False(t, netMes.HasProcessorForTopic(topic))
		assert.False(t, netMes.HasProcessorForTopic(testTopic))

		_ = netMes.Close()
	})
	t.Run("unregister all processors should work", func(t *testing.T) {
		t.Parallel()

		netMes, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

		topic := common.ConnectionTopic
		err := netMes.RegisterMessageProcessor(topic, "identifier", &mock.MessageProcessorStub{})
		assert.Nil(t, err)
		assert.True(t, netMes.HasProcessorForTopic(topic))

		testTopic := "test topic"
		err = netMes.RegisterMessageProcessor(testTopic, "identifier", &mock.MessageProcessorStub{})
		assert.Nil(t, err)
		assert.True(t, netMes.HasProcessorForTopic(testTopic))

		err = netMes.UnregisterAllMessageProcessors()
		assert.Nil(t, err)
		assert.False(t, netMes.HasProcessorForTopic(topic))
		assert.False(t, netMes.HasProcessorForTopic(testTopic))

		_ = netMes.Close()
	})
}
