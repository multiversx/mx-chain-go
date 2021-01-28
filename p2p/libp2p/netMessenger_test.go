package libp2p_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/data"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
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

func prepareMessengerForMatchDataReceive(mes p2p.Messenger, matchData []byte, wg *sync.WaitGroup) {
	_ = mes.CreateTopic("test", false)

	_ = mes.RegisterMessageProcessor("test",
		&mock.MessageProcessorStub{
			ProcessMessageCalled: func(message p2p.MessageP2P, _ core.PeerID) error {
				if bytes.Equal(matchData, message.Data()) {
					fmt.Printf("%s got the message\n", mes.ID().Pretty())
					wg.Done()
				}

				return nil
			},
		})
}

func getConnectableAddress(mes p2p.Messenger) string {
	for _, addr := range mes.Addresses() {
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
				Port: "0",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer: &libp2p.LocalSyncTimer{},
	}
}

func createMockNetworkOf2() (mocknet.Mocknet, p2p.Messenger, p2p.Messenger) {
	netw := mocknet.New(context.Background())

	mes1, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	mes2, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	return netw, mes1, mes2
}

func createMockMessenger() p2p.Messenger {
	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	return mes
}

func containsPeerID(list []core.PeerID, searchFor core.PeerID) bool {
	for _, pid := range list {
		if bytes.Equal(pid.Bytes(), searchFor.Bytes()) {
			return true
		}
	}
	return false
}

//------- NewMemoryLibp2pMessenger

func TestNewMemoryLibp2pMessenger_NilMockNetShouldErr(t *testing.T) {
	args := createMockNetworkArgs()
	mes, err := libp2p.NewMockMessenger(args, nil)

	assert.Nil(t, mes)
	assert.Equal(t, p2p.ErrNilMockNet, err)
}

func TestNewMemoryLibp2pMessenger_OkValsWithoutDiscoveryShouldWork(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, err := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(mes))

	_ = mes.Close()
}

//------- NewNetworkMessenger

func TestNewNetworkMessenger_NilMessengerShouldErr(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.Marshalizer = nil
	mes, err := libp2p.NewNetworkMessenger(arg)

	assert.True(t, check.IfNil(mes))
	assert.True(t, errors.Is(err, p2p.ErrNilMarshalizer))
}

func TestNewNetworkMessenger_NilSyncTimerShouldErr(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.SyncTimer = nil
	mes, err := libp2p.NewNetworkMessenger(arg)

	assert.True(t, check.IfNil(mes))
	assert.True(t, errors.Is(err, p2p.ErrNilSyncTimer))
}

func TestNewNetworkMessenger_WithDeactivatedKadDiscovererShouldWork(t *testing.T) {
	arg := createMockNetworkArgs()
	mes, err := libp2p.NewNetworkMessenger(arg)

	assert.NotNil(t, mes)
	assert.Nil(t, err)

	_ = mes.Close()
}

func TestNewNetworkMessenger_WithKadDiscovererListsSharderInvalidTargetConnShouldErr(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.P2pConfig.KadDhtPeerDiscovery = config.KadDhtPeerDiscoveryConfig{
		Enabled:                          true,
		RefreshIntervalInSec:             10,
		ProtocolID:                       "/erd/kad/1.0.0",
		InitialPeerList:                  nil,
		BucketSize:                       100,
		RoutingTableRefreshIntervalInSec: 10,
	}
	arg.P2pConfig.Sharding.Type = p2p.ListsSharder
	mes, err := libp2p.NewNetworkMessenger(arg)

	assert.True(t, check.IfNil(mes))
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestNewNetworkMessenger_WithKadDiscovererListSharderShouldWork(t *testing.T) {
	arg := createMockNetworkArgs()
	arg.P2pConfig.KadDhtPeerDiscovery = config.KadDhtPeerDiscoveryConfig{
		Enabled:                          true,
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
	mes, err := libp2p.NewNetworkMessenger(arg)

	assert.False(t, check.IfNil(mes))
	assert.Nil(t, err)

	_ = mes.Close()
}

//------- Messenger functionality

func TestLibp2pMessenger_ConnectToPeerShouldCallUpgradedHost(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = mes.Close()

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

	mes.SetHost(uhs)
	_ = mes.ConnectToPeer(p)
	assert.True(t, wasCalled)
}

func TestLibp2pMessenger_IsConnectedShouldWork(t *testing.T) {
	_, mes1, mes2 := createMockNetworkOf2()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)

	assert.True(t, mes1.IsConnected(mes2.ID()))
	assert.True(t, mes2.IsConnected(mes1.ID()))

	_ = mes1.Close()
	_ = mes2.Close()
}

func TestLibp2pMessenger_CreateTopicOkValsShouldWork(t *testing.T) {
	mes := createMockMessenger()

	err := mes.CreateTopic("test", true)
	assert.Nil(t, err)

	_ = mes.Close()
}

func TestLibp2pMessenger_CreateTopicTwiceShouldNotErr(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)
	err := mes.CreateTopic("test", false)
	assert.Nil(t, err)

	_ = mes.Close()
}

func TestLibp2pMessenger_HasTopicIfHaveTopicShouldReturnTrue(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	assert.True(t, mes.HasTopic("test"))

	_ = mes.Close()
}

func TestLibp2pMessenger_HasTopicIfDoNotHaveTopicShouldReturnFalse(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	assert.False(t, mes.HasTopic("one topic"))

	_ = mes.Close()
}

func TestLibp2pMessenger_HasTopicValidatorDoNotHaveTopicShouldReturnFalse(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	assert.False(t, mes.HasTopicValidator("one topic"))

	_ = mes.Close()
}

func TestLibp2pMessenger_HasTopicValidatorHaveTopicDoNotHaveValidatorShouldReturnFalse(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	assert.False(t, mes.HasTopicValidator("test"))

	_ = mes.Close()
}

func TestLibp2pMessenger_HasTopicValidatorHaveTopicHaveValidatorShouldReturnTrue(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)
	_ = mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	assert.True(t, mes.HasTopicValidator("test"))

	_ = mes.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorOnInexistentTopicShouldWork(t *testing.T) {
	mes := createMockMessenger()

	err := mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	assert.Nil(t, err)

	_ = mes.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorWithNilHandlerShouldErr(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	err := mes.RegisterMessageProcessor("test", nil)

	assert.Equal(t, p2p.ErrNilValidator, err)

	_ = mes.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorOkValsShouldWork(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	err := mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	assert.Nil(t, err)

	_ = mes.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorReregistrationShouldErr(t *testing.T) {
	mes := createMockMessenger()
	_ = mes.CreateTopic("test", false)
	//registration
	_ = mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})
	//re-registration
	err := mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	assert.True(t, errors.Is(err, p2p.ErrTopicValidatorOperationNotSupported))

	_ = mes.Close()
}

func TestLibp2pMessenger_UnegisterTopicValidatorOnANotRegisteredTopicShouldNotErr(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)
	err := mes.UnregisterMessageProcessor("test")

	assert.Nil(t, err)

	_ = mes.Close()
}

func TestLibp2pMessenger_UnregisterTopicValidatorShouldWork(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	//registration
	_ = mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	//unregistration
	err := mes.UnregisterMessageProcessor("test")

	assert.Nil(t, err)

	_ = mes.Close()
}

func TestLibp2pMessenger_UnregisterAllTopicValidatorShouldWork(t *testing.T) {
	mes := createMockMessenger()
	_ = mes.CreateTopic("test", false)
	//registration
	_ = mes.CreateTopic("test1", false)
	_ = mes.RegisterMessageProcessor("test1", &mock.MessageProcessorStub{})
	_ = mes.CreateTopic("test2", false)
	_ = mes.RegisterMessageProcessor("test2", &mock.MessageProcessorStub{})
	//unregistration
	err := mes.UnregisterAllMessageProcessors()
	assert.Nil(t, err)
	err = mes.RegisterMessageProcessor("test1", &mock.MessageProcessorStub{})
	assert.Nil(t, err)
	err = mes.RegisterMessageProcessor("test2", &mock.MessageProcessorStub{})
	assert.Nil(t, err)
	_ = mes.Close()
}

func TestLibp2pMessenger_BroadcastDataLargeMessageShouldNotCallSend(t *testing.T) {
	msg := make([]byte, libp2p.MaxSendBuffSize+1)
	mes, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())
	mes.SetLoadBalancer(&mock.ChannelLoadBalancerStub{
		GetChannelOrDefaultCalled: func(pipe string) chan *p2p.SendableData {
			assert.Fail(t, "should have not got to this line")

			return make(chan *p2p.SendableData, 1)
		},
		CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
			return nil
		},
	})

	mes.Broadcast("topic", msg)

	_ = mes.Close()
}

func TestLibp2pMessenger_BroadcastDataBetween2PeersShouldWork(t *testing.T) {
	msg := []byte("test message")

	_, mes1, mes2 := createMockNetworkOf2()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(2)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	prepareMessengerForMatchDataReceive(mes1, msg, wg)
	prepareMessengerForMatchDataReceive(mes2, msg, wg)

	fmt.Println("Delaying as to allow peers to announce themselves on the opened topic...")
	time.Sleep(time.Second)

	fmt.Printf("sending message from %s...\n", mes1.ID().Pretty())

	mes1.Broadcast("test", msg)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)

	_ = mes1.Close()
	_ = mes2.Close()
}

func TestLibp2pMessenger_BroadcastOnChannelBlockingShouldLimitNumberOfGoRoutines(t *testing.T) {
	if testing.Short() {
		t.Skip("this test does not perform well in TC with race detector on")
	}

	msg := []byte("test message")
	numBroadcasts := libp2p.BroadcastGoRoutines + 1

	ch := make(chan *p2p.SendableData)

	mes, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())
	mes.SetLoadBalancer(&mock.ChannelLoadBalancerStub{
		CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
			return nil
		},
		GetChannelOrDefaultCalled: func(pipe string) chan *p2p.SendableData {
			return ch
		},
	})

	numErrors := uint32(0)

	wg := sync.WaitGroup{}
	wg.Add(numBroadcasts - libp2p.BroadcastGoRoutines)
	for i := 0; i < numBroadcasts; i++ {
		go func() {
			err := mes.BroadcastOnChannelBlocking("test", "test", msg)
			if err == p2p.ErrTooManyGoroutines {
				atomic.AddUint32(&numErrors, 1)
				wg.Done()
			}
		}()
	}

	wg.Wait()

	// cleanup stuck go routines that are trying to write on the ch channel
	for i := 0; i < libp2p.BroadcastGoRoutines; i++ {
		<-ch
	}

	assert.Equal(t, atomic.LoadUint32(&numErrors), uint32(numBroadcasts-libp2p.BroadcastGoRoutines))

	_ = mes.Close()
}

func TestLibp2pMessenger_BroadcastDataBetween2PeersWithLargeMsgShouldWork(t *testing.T) {
	msg := make([]byte, libp2p.MaxSendBuffSize)

	_, mes1, mes2 := createMockNetworkOf2()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(2)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	prepareMessengerForMatchDataReceive(mes1, msg, wg)
	prepareMessengerForMatchDataReceive(mes2, msg, wg)

	fmt.Println("Delaying as to allow peers to announce themselves on the opened topic...")
	time.Sleep(time.Second)

	fmt.Printf("sending message from %s...\n", mes1.ID().Pretty())

	mes1.Broadcast("test", msg)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)

	_ = mes1.Close()
	_ = mes2.Close()
}

func TestLibp2pMessenger_Peers(t *testing.T) {
	_, mes1, mes2 := createMockNetworkOf2()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)

	//should know both peers
	foundCurrent := false
	foundConnected := false

	for _, p := range mes1.Peers() {
		fmt.Println(p.Pretty())

		if p.Pretty() == mes1.ID().Pretty() {
			foundCurrent = true
		}
		if p.Pretty() == mes2.ID().Pretty() {
			foundConnected = true
		}
	}

	assert.True(t, foundCurrent && foundConnected)

	_ = mes1.Close()
	_ = mes2.Close()
}

func TestLibp2pMessenger_ConnectedPeers(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)

	//connected peers:  1 ----- 2 ----- 3

	assert.Equal(t, []core.PeerID{mes2.ID()}, mes1.ConnectedPeers())
	assert.Equal(t, []core.PeerID{mes2.ID()}, mes3.ConnectedPeers())
	assert.Equal(t, 2, len(mes2.ConnectedPeers()))
	//no need to further test that mes2 is connected to mes1 and mes3 s this was tested in first 2 asserts

	_ = mes1.Close()
	_ = mes2.Close()
	_ = mes3.Close()
}

func TestLibp2pMessenger_ConnectedAddresses(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)

	//connected peers:  1 ----- 2 ----- 3

	foundAddr1 := false
	foundAddr3 := false

	for _, addr := range mes2.ConnectedAddresses() {
		for _, addrMes1 := range mes1.Addresses() {
			if addr == addrMes1 {
				foundAddr1 = true
			}
		}

		for _, addrMes3 := range mes3.Addresses() {
			if addr == addrMes3 {
				foundAddr3 = true
			}
		}
	}

	assert.True(t, foundAddr1)
	assert.True(t, foundAddr3)
	assert.Equal(t, 2, len(mes2.ConnectedAddresses()))
	//no need to further test that mes2 is connected to mes1 and mes3 s this was tested in first 2 asserts

	_ = mes1.Close()
	_ = mes2.Close()
	_ = mes3.Close()
}

func TestLibp2pMessenger_PeerAddressConnectedPeerShouldWork(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)

	//connected peers:  1 ----- 2 ----- 3

	defer func() {
		_ = mes1.Close()
		_ = mes2.Close()
		_ = mes3.Close()
	}()

	addressesRecov := mes2.PeerAddresses(mes1.ID())
	for _, addr := range mes1.Addresses() {
		for _, addrRecov := range addressesRecov {
			if strings.Contains(addr, addrRecov) {
				//address returned is valid, test is successful
				return
			}
		}
	}

	assert.Fail(t, "Returned address is not valid!")
}

func TestLibp2pMessenger_PeerAddressDisconnectedPeerShouldWork(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)

	_ = netw.LinkAll()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)

	defer func() {
		_ = mes1.Close()
		_ = mes2.Close()
		_ = mes3.Close()
	}()

	_ = netw.UnlinkPeers(peer.ID(mes1.ID().Bytes()), peer.ID(mes2.ID().Bytes()))
	_ = netw.DisconnectPeers(peer.ID(mes1.ID().Bytes()), peer.ID(mes2.ID().Bytes()))
	_ = netw.DisconnectPeers(peer.ID(mes2.ID().Bytes()), peer.ID(mes1.ID().Bytes()))

	//connected peers:  1 --x-- 2 ----- 3

	assert.False(t, mes2.IsConnected(mes1.ID()))
}

func TestLibp2pMessenger_PeerAddressUnknownPeerShouldReturnEmpty(t *testing.T) {
	_, mes1, _ := createMockNetworkOf2()

	defer func() {
		_ = mes1.Close()
	}()

	adr1Recov := mes1.PeerAddresses("unknown peer")
	assert.Equal(t, 0, len(adr1Recov))
}

//------- ConnectedPeersOnTopic

func TestLibp2pMessenger_ConnectedPeersOnTopicInvalidTopicShouldRetEmptyList(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = netw.LinkAll()

	adr2 := mes2.Addresses()[0]
	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)
	//connected peers:  1 ----- 2 ----- 3
	connPeers := mes1.ConnectedPeersOnTopic("non-existent topic")
	assert.Equal(t, 0, len(connPeers))

	_ = mes1.Close()
	_ = mes2.Close()
	_ = mes3.Close()
}

func TestLibp2pMessenger_ConnectedPeersOnTopicOneTopicShouldWork(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	mes4, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = netw.LinkAll()

	adr2 := mes2.Addresses()[0]
	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)
	_ = mes4.ConnectToPeer(adr2)
	//connected peers:  1 ----- 2 ----- 3
	//                          |
	//                          4
	//1, 2, 3 should be on topic "topic123"
	_ = mes1.CreateTopic("topic123", false)
	_ = mes2.CreateTopic("topic123", false)
	_ = mes3.CreateTopic("topic123", false)

	//wait a bit for topic announcements
	time.Sleep(time.Second)

	peersOnTopic123 := mes2.ConnectedPeersOnTopic("topic123")

	assert.Equal(t, 2, len(peersOnTopic123))
	assert.True(t, containsPeerID(peersOnTopic123, mes1.ID()))
	assert.True(t, containsPeerID(peersOnTopic123, mes3.ID()))

	_ = mes1.Close()
	_ = mes2.Close()
	_ = mes3.Close()
	_ = mes4.Close()
}

func TestLibp2pMessenger_ConnectedPeersOnTopicOneTopicDifferentViewsShouldWork(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	mes4, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = netw.LinkAll()

	adr2 := mes2.Addresses()[0]
	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)
	_ = mes4.ConnectToPeer(adr2)
	//connected peers:  1 ----- 2 ----- 3
	//                          |
	//                          4
	//1, 2, 3 should be on topic "topic123"
	_ = mes1.CreateTopic("topic123", false)
	_ = mes2.CreateTopic("topic123", false)
	_ = mes3.CreateTopic("topic123", false)

	//wait a bit for topic announcements
	time.Sleep(time.Second)

	peersOnTopic123FromMes2 := mes2.ConnectedPeersOnTopic("topic123")
	peersOnTopic123FromMes4 := mes4.ConnectedPeersOnTopic("topic123")

	//keep the same checks as the test above as to be 100% that the returned list are correct
	assert.Equal(t, 2, len(peersOnTopic123FromMes2))
	assert.True(t, containsPeerID(peersOnTopic123FromMes2, mes1.ID()))
	assert.True(t, containsPeerID(peersOnTopic123FromMes2, mes3.ID()))

	assert.Equal(t, 1, len(peersOnTopic123FromMes4))
	assert.True(t, containsPeerID(peersOnTopic123FromMes4, mes2.ID()))

	_ = mes1.Close()
	_ = mes2.Close()
	_ = mes3.Close()
	_ = mes4.Close()
}

func TestLibp2pMessenger_ConnectedPeersOnTopicTwoTopicsShouldWork(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	mes4, _ := libp2p.NewMockMessenger(createMockNetworkArgs(), netw)
	_ = netw.LinkAll()

	adr2 := mes2.Addresses()[0]
	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)
	_ = mes4.ConnectToPeer(adr2)
	//connected peers:  1 ----- 2 ----- 3
	//                          |
	//                          4
	//1, 2, 3 should be on topic "topic123"
	//2, 4 should be on topic "topic24"
	_ = mes1.CreateTopic("topic123", false)
	_ = mes2.CreateTopic("topic123", false)
	_ = mes2.CreateTopic("topic24", false)
	_ = mes3.CreateTopic("topic123", false)
	_ = mes4.CreateTopic("topic24", false)

	//wait a bit for topic announcements
	time.Sleep(time.Second)

	peersOnTopic123 := mes2.ConnectedPeersOnTopic("topic123")
	peersOnTopic24 := mes2.ConnectedPeersOnTopic("topic24")

	//keep the same checks as the test above as to be 100% that the returned list are correct
	assert.Equal(t, 2, len(peersOnTopic123))
	assert.True(t, containsPeerID(peersOnTopic123, mes1.ID()))
	assert.True(t, containsPeerID(peersOnTopic123, mes3.ID()))

	assert.Equal(t, 1, len(peersOnTopic24))
	assert.True(t, containsPeerID(peersOnTopic24, mes4.ID()))

	_ = mes1.Close()
	_ = mes2.Close()
	_ = mes3.Close()
	_ = mes4.Close()
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
					//generate a mock list that contain duplicates
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
	//we can safely close the host as the next operations will be done on a mock
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

	_, mes1, mes2 := createMockNetworkOf2()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(1)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	prepareMessengerForMatchDataReceive(mes2, msg, wg)

	fmt.Println("Delaying as to allow peers to announce themselves on the opened topic...")
	time.Sleep(time.Second)

	fmt.Printf("sending message from %s...\n", mes1.ID().Pretty())

	err := mes1.SendToConnectedPeer("test", msg, mes2.ID())

	assert.Nil(t, err)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)

	_ = mes1.Close()
	_ = mes2.Close()
}

func TestLibp2pMessenger_SendDirectWithRealNetToConnectedPeerShouldWork(t *testing.T) {
	msg := []byte("test message")

	fmt.Println("Messenger 1:")
	mes1, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

	fmt.Println("Messenger 2:")
	mes2, _ := libp2p.NewNetworkMessenger(createMockNetworkArgs())

	err := mes1.ConnectToPeer(getConnectableAddress(mes2))
	assert.Nil(t, err)

	wg := &sync.WaitGroup{}
	chanDone := make(chan bool)
	wg.Add(2)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	prepareMessengerForMatchDataReceive(mes1, msg, wg)
	prepareMessengerForMatchDataReceive(mes2, msg, wg)

	fmt.Println("Delaying as to allow peers to announce themselves on the opened topic...")
	time.Sleep(time.Second)

	fmt.Printf("Messenger 1 is sending message from %s...\n", mes1.ID().Pretty())
	err = mes1.SendToConnectedPeer("test", msg, mes2.ID())
	assert.Nil(t, err)

	time.Sleep(time.Second)
	fmt.Printf("Messenger 2 is sending message from %s...\n", mes2.ID().Pretty())
	err = mes2.SendToConnectedPeer("test", msg, mes1.ID())
	assert.Nil(t, err)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)

	_ = mes1.Close()
	_ = mes2.Close()
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

//------- Bootstrap

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

//------- SetThresholdMinConnectedPeers

func TestNetworkMessenger_SetThresholdMinConnectedPeersInvalidValueShouldErr(t *testing.T) {
	mes := createMockMessenger()
	defer func() {
		_ = mes.Close()
	}()

	err := mes.SetThresholdMinConnectedPeers(-1)

	assert.Equal(t, p2p.ErrInvalidValue, err)
}

func TestNetworkMessenger_SetThresholdMinConnectedPeersShouldWork(t *testing.T) {
	mes := createMockMessenger()
	defer func() {
		_ = mes.Close()
	}()

	minConnectedPeers := 56
	err := mes.SetThresholdMinConnectedPeers(minConnectedPeers)

	assert.Nil(t, err)
	assert.Equal(t, minConnectedPeers, mes.ThresholdMinConnectedPeers())
}

//------- IsConnectedToTheNetwork

func TestNetworkMessenger_IsConnectedToTheNetworkRetFalse(t *testing.T) {
	mes := createMockMessenger()
	defer func() {
		_ = mes.Close()
	}()

	minConnectedPeers := 56
	_ = mes.SetThresholdMinConnectedPeers(minConnectedPeers)

	assert.False(t, mes.IsConnectedToTheNetwork())
}

func TestNetworkMessenger_IsConnectedToTheNetworkWithZeroRetTrue(t *testing.T) {
	mes := createMockMessenger()
	defer func() {
		_ = mes.Close()
	}()

	minConnectedPeers := 0
	_ = mes.SetThresholdMinConnectedPeers(minConnectedPeers)

	assert.True(t, mes.IsConnectedToTheNetwork())
}

//------- SetPeerShardResolver

func TestNetworkMessenger_SetPeerShardResolverNilShouldErr(t *testing.T) {
	mes := createMockMessenger()
	defer func() {
		_ = mes.Close()
	}()

	err := mes.SetPeerShardResolver(nil)

	assert.Equal(t, p2p.ErrNilPeerShardResolver, err)
}

func TestNetworkMessenger_SetPeerShardResolver(t *testing.T) {
	mes := createMockMessenger()
	defer func() {
		_ = mes.Close()
	}()

	err := mes.SetPeerShardResolver(&mock.PeerShardResolverStub{})

	assert.Nil(t, err)
}

func TestNetworkMessenger_DoubleCloseShouldWork(t *testing.T) {
	mes := createMessenger()

	time.Sleep(time.Second)

	err := mes.Close()
	assert.Nil(t, err)

	err = mes.Close()
	assert.Nil(t, err)
}

func TestNetworkMessenger_PreventReprocessingShouldWork(t *testing.T) {
	args := libp2p.ArgsNetworkMessenger{
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		Marshalizer:   &testscommon.ProtoMarshalizerMock{},
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port: "0",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer: &libp2p.LocalSyncTimer{},
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
		Message: &pubsub_pb.Message{
			From:                 []byte(pid),
			Data:                 buff,
			Seqno:                []byte{0, 0, 0, 1},
			TopicIDs:             nil,
			Signature:            nil,
			Key:                  nil,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		},
		ReceivedFrom:  "",
		ValidatorData: nil,
	}

	assert.False(t, callBackFunc(ctx, pid, msg)) //this will not call
	assert.False(t, callBackFunc(ctx, pid, msg)) //this will not call
	assert.Equal(t, uint32(0), atomic.LoadUint32(&numCalled))

	_ = mes.Close()
}

func TestNetworkMessenger_PubsubCallbackNotMessageNotValidShouldNotCallHandler(t *testing.T) {
	args := libp2p.ArgsNetworkMessenger{
		Marshalizer:   &testscommon.ProtoMarshalizerMock{},
		ListenAddress: libp2p.ListenLocalhostAddrWithIp4AndTcp,
		P2pConfig: config.P2PConfig{
			Node: config.NodeConfig{
				Port: "0",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer: &libp2p.LocalSyncTimer{},
	}

	mes, _ := libp2p.NewNetworkMessenger(args)
	numUpserts := int32(0)
	_ = mes.SetPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{
		UpsertPeerIDCalled: func(pid core.PeerID, duration time.Duration) error {
			atomic.AddInt32(&numUpserts, 1)
			//any error thrown here should not impact the execution
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
		Message: &pubsub_pb.Message{
			From:                 []byte("not a valid pid"),
			Data:                 buff,
			Seqno:                []byte{0, 0, 0, 1},
			TopicIDs:             nil,
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
				Port: "0",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer: &libp2p.LocalSyncTimer{},
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
	msg := &pubsub.Message{
		Message: &pubsub_pb.Message{
			From:                 []byte(mes.ID()),
			Data:                 buff,
			Seqno:                []byte{0, 0, 0, 1},
			TopicIDs:             nil,
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
				Port: "0",
			},
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
			Sharding: config.ShardingConfig{
				Type: p2p.NilListSharder,
			},
		},
		SyncTimer: &libp2p.LocalSyncTimer{},
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
		TimestampField: now.Unix() - int64(libp2p.PubsubTimeCacheDuration.Seconds()) +
			int64(libp2p.AcceptMessagesInAdvanceDuration.Seconds()) - 1,
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
