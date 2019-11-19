package libp2p_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
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
			ProcessMessageCalled: func(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
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

func createMockNetworkOf2() (mocknet.Mocknet, p2p.Messenger, p2p.Messenger) {
	netw := mocknet.New(context.Background())

	mes1, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
	mes2, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())

	_ = netw.LinkAll()

	return netw, mes1, mes2
}

func createMockNetwork(numOfPeers int) (mocknet.Mocknet, []p2p.Messenger) {
	netw := mocknet.New(context.Background())
	peers := make([]p2p.Messenger, numOfPeers)

	for i := 0; i < numOfPeers; i++ {
		peers[i], _ = libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
	}

	_ = netw.LinkAll()

	return netw, peers
}

func connectPeersFullMesh(peers []p2p.Messenger) {
	for i := 0; i < len(peers); i++ {
		for j := i + 1; j < len(peers); j++ {
			err := peers[i].ConnectToPeer(peers[j].Addresses()[0])
			if err != nil {
				fmt.Printf("error connecting: %s\n", err.Error())
			}
		}
	}
}

func createMockMessenger() p2p.Messenger {
	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())

	return mes
}

func createLibP2PCredentialsMessenger() (peer.ID, libp2pCrypto.PrivKey) {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	return id, sk
}

func containsPeerID(list []p2p.PeerID, searchFor p2p.PeerID) bool {
	for _, pid := range list {
		if bytes.Equal(pid.Bytes(), searchFor.Bytes()) {
			return true
		}
	}
	return false
}

//------- NewMemoryLibp2pMessenger

func TestNewMemoryLibp2pMessenger_NilContextShouldErr(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, err := libp2p.NewMemoryMessenger(nil, netw, discovery.NewNullDiscoverer())

	assert.Nil(t, mes)
	assert.Equal(t, p2p.ErrNilContext, err)
}

func TestNewMemoryLibp2pMessenger_NilMocknetShouldErr(t *testing.T) {
	mes, err := libp2p.NewMemoryMessenger(context.Background(), nil, discovery.NewNullDiscoverer())

	assert.Nil(t, mes)
	assert.Equal(t, p2p.ErrNilMockNet, err)
}

func TestNewMemoryLibp2pMessenger_NilPeerDiscovererShouldErr(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, err := libp2p.NewMemoryMessenger(context.Background(), netw, nil)

	assert.Nil(t, mes)
	assert.Equal(t, p2p.ErrNilPeerDiscoverer, err)
}

func TestNewMemoryLibp2pMessenger_PeerDiscovererFailsWhenApplyingContextShouldErr(t *testing.T) {
	netw := mocknet.New(context.Background())

	errExpected := errors.New("expected err")

	mes, err := libp2p.NewMemoryMessenger(
		context.Background(),
		netw,
		&mock.PeerDiscovererStub{
			ApplyContextCalled: func(ctxProvider p2p.ContextProvider) error {
				return errExpected
			},
		},
	)

	assert.Nil(t, mes)
	assert.Equal(t, errExpected, err)
}

func TestNewMemoryLibp2pMessenger_OkValsShouldWork(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, err := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())

	assert.Nil(t, err)
	assert.False(t, check.IfNil(mes))

	_ = mes.Close()
}

//------- NewNetworkMessenger

func TestNewNetworkMessenger_NilContextShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		nil,
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		&mock.ChannelLoadBalancerStub{},
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilContext)
}

func TestNewNetworkMessenger_InvalidPortShouldErr(t *testing.T) {
	port := -1

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		&mock.ChannelLoadBalancerStub{},
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrInvalidPort)
}

func TestNewNetworkMessenger_NilP2PprivateKeyShouldErr(t *testing.T) {
	port := 4000

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		nil,
		&mock.ConnManagerNotifieeStub{},
		&mock.ChannelLoadBalancerStub{},
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilP2PprivateKey)
}

func TestNewNetworkMessenger_NilPipeLoadBalancerShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		nil,
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilChannelLoadBalancer)
}

func TestNewNetworkMessenger_NoConnMgrShouldWork(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.ChannelLoadBalancerStub{
			CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
		},
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)

	_ = mes.Close()
}

func TestNewNetworkMessenger_WithConnMgrShouldWork(t *testing.T) {
	//TODO remove skip when external library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	cns := &mock.ConnManagerNotifieeStub{
		ListenCalled:      func(netw network.Network, ma multiaddr.Multiaddr) {},
		ListenCloseCalled: func(netw network.Network, ma multiaddr.Multiaddr) {},
		CloseCalled:       func() error { return nil },
	}

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		cns,
		&mock.ChannelLoadBalancerStub{
			CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
		},
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)
	assert.True(t, cns == mes.ConnManager())

	_ = mes.Close()
}

func TestNewNetworkMessenger_WithNullPeerDiscoveryShouldWork(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.ChannelLoadBalancerStub{
			CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
		},
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)

	_ = mes.Close()
}

func TestNewNetworkMessenger_NilPeerDiscoveryShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.ChannelLoadBalancerStub{
			CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
		},
		nil,
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	assert.Nil(t, mes)
	assert.Equal(t, p2p.ErrNilPeerDiscoverer, err)
}

func TestNewNetworkMessenger_PeerDiscovererFailsWhenApplyingContextShouldErr(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	errExpected := errors.New("expected err")

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.ChannelLoadBalancerStub{
			CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
		},
		&mock.PeerDiscovererStub{
			ApplyContextCalled: func(ctxProvider p2p.ContextProvider) error {
				return errExpected
			},
		},
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	assert.Nil(t, mes)
	assert.Equal(t, errExpected, err)
}

func TestNewNetworkMessengerWithPortSweep_ShouldFindFreePort(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessengerOnFreePort(
		context.Background(),
		sk,
		nil,
		&mock.ChannelLoadBalancerStub{
			CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
		},
		discovery.NewNullDiscoverer(),
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)

	_ = mes.Close()
}

//------- Messenger functionality

func TestLibp2pMessenger_ConnectToPeerShouldCallUpgradedHost(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
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

func TestLibp2pMessenger_CreateTopicTwiceShouldErr(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)
	err := mes.CreateTopic("test", false)
	assert.Equal(t, p2p.ErrTopicAlreadyExists, err)

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

func TestLibp2pMessenger_RegisterTopicValidatorOnInexistentTopicShouldErr(t *testing.T) {
	mes := createMockMessenger()

	err := mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	assert.Equal(t, p2p.ErrNilTopic, err)

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

	assert.Equal(t, p2p.ErrTopicValidatorOperationNotSupported, err)

	_ = mes.Close()
}

func TestLibp2pMessenger_UnegisterTopicValidatorOnInexistentTopicShouldErr(t *testing.T) {
	mes := createMockMessenger()

	err := mes.UnregisterMessageProcessor("test")

	assert.Equal(t, p2p.ErrNilTopic, err)

	_ = mes.Close()
}

func TestLibp2pMessenger_UnegisterTopicValidatorOnANotRegisteredTopicShouldErr(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	err := mes.UnregisterMessageProcessor("test")

	assert.Equal(t, p2p.ErrTopicValidatorOperationNotSupported, err)

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

func TestLibp2pMessenger_BroadcastDataLargeMessageShouldNotCallSend(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

	msg := make([]byte, libp2p.MaxSendBuffSize+1)

	_, sk := createLibP2PCredentialsMessenger()
	mes, err := libp2p.NewNetworkMessengerOnFreePort(
		context.Background(),
		sk,
		nil,
		&mock.ChannelLoadBalancerStub{
			GetChannelOrDefaultCalled: func(pipe string) chan *p2p.SendableData {
				assert.Fail(t, "should have not got to this line")

				return make(chan *p2p.SendableData, 1)
			},
			CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
				return nil
			},
		},
		&discovery.NullDiscoverer{},
	)
	assert.Nil(t, err)

	mes.Broadcast("topic", msg)

	_ = mes.Close()
}

func TestLibp2pMessenger_BroadcastDataBetween2PeersShouldWork(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

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
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

	port := 4000
	msg := []byte("test message")
	numBroadcasts := 10000

	_, sk := createLibP2PCredentialsMessenger()
	ch := make(chan *p2p.SendableData)

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.ChannelLoadBalancerStub{
			CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
			GetChannelOrDefaultCalled: func(pipe string) chan *p2p.SendableData {
				return ch
			},
		},
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	wg := sync.WaitGroup{}
	wg.Add(numBroadcasts - libp2p.BroadcastGoRoutines)
	for i := 0; i < numBroadcasts; i++ {
		go func() {
			err := mes.BroadcastOnChannelBlocking("test", "test", msg)
			if err != nil {
				wg.Done()
			}
		}()
	}

	wg.Wait()

	assert.True(t, libp2p.BroadcastGoRoutines >= emptyChannel(ch))
}

func emptyChannel(ch chan *p2p.SendableData) int {
	readsCnt := 0
	for {
		select {
		case <-ch:
			readsCnt++
		default:
			return readsCnt
		}
	}
}

func TestLibp2pMessenger_BroadcastDataBetween2PeersWithLargeMsgShouldWork(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

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

func TestLibp2pMessenger_BroadcastDataOnTopicPipeBetween2PeersShouldWork(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

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
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())

	_ = netw.LinkAll()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)

	//connected peers:  1 ----- 2 ----- 3

	assert.Equal(t, []p2p.PeerID{mes2.ID()}, mes1.ConnectedPeers())
	assert.Equal(t, []p2p.PeerID{mes2.ID()}, mes3.ConnectedPeers())
	assert.Equal(t, 2, len(mes2.ConnectedPeers()))
	//no need to further test that mes2 is connected to mes1 and mes3 s this was tested in first 2 asserts

	_ = mes1.Close()
	_ = mes2.Close()
	_ = mes3.Close()
}

func TestLibp2pMessenger_ConnectedAddresses(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())

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
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())

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

	adr1Recov := mes2.PeerAddress(mes1.ID())
	for _, addr := range mes1.Addresses() {
		if strings.Contains(addr, adr1Recov) {
			//address returned is valid, test is successful
			return
		}
	}

	assert.Fail(t, "Returned address is not valid!")
}

func TestLibp2pMessenger_PeerAddressDisconnectedPeerShouldWork(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())

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

	adr1Recov := mes2.PeerAddress(mes1.ID())
	for _, addr := range mes1.Addresses() {
		if strings.Contains(addr, adr1Recov) {
			//address returned is valid, test is successful
			return
		}
	}

	assert.Fail(t, "Returned address is not valid!")
}

func TestLibp2pMessenger_PeerAddressUnknownPeerShouldReturnEmpty(t *testing.T) {
	_, mes1, _ := createMockNetworkOf2()

	defer func() {
		_ = mes1.Close()
	}()

	adr1Recov := mes1.PeerAddress("unknown peer")
	assert.Equal(t, "", adr1Recov)
}

//------- ConnectedPeersOnTopic

func TestLibp2pMessenger_ConnectedPeersOnTopicInvalidTopicShouldRetEmptyList(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
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
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
	mes4, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
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
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
	mes4, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
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
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
	mes4, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
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
	pid1 := p2p.PeerID("pid1")
	pid2 := p2p.PeerID("pid2")
	pid3 := p2p.PeerID("pid3")
	pid4 := p2p.PeerID("pid4")

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
	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
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

func existInList(list []p2p.PeerID, pid p2p.PeerID) bool {
	for _, p := range list {
		if bytes.Equal(p.Bytes(), pid.Bytes()) {
			return true
		}
	}

	return false
}

func generateConnWithRemotePeer(pid p2p.PeerID) network.Conn {
	return &mock.ConnStub{
		RemotePeerCalled: func() peer.ID {
			return peer.ID(pid)
		},
	}
}

func TestLibp2pMessenger_TrimConnectionsCallsConnManagerTrimConnections(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	wasCalled := false

	cns := &mock.ConnManagerNotifieeStub{
		ListenCalled:      func(netw network.Network, ma multiaddr.Multiaddr) {},
		ListenCloseCalled: func(netw network.Network, ma multiaddr.Multiaddr) {},
		TrimOpenConnsCalled: func(ctx context.Context) {
			wasCalled = true
		},
		CloseCalled: func() error {
			return nil
		},
	}

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		cns,
		&mock.ChannelLoadBalancerStub{
			CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
		},
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	mes.TrimConnections()

	assert.True(t, wasCalled)

	_ = mes.Close()
}

func TestLibp2pMessenger_SendDataThrottlerShouldReturnCorrectObject(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	sdt := &mock.ChannelLoadBalancerStub{
		AddChannelCalled: func(pipe string) error {
			return nil
		},
		CollectOneElementFromChannelsCalled: func() *p2p.SendableData {
			time.Sleep(time.Millisecond * 100)
			return nil
		},
	}

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		sdt,
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	sdtReturned := mes.OutgoingChannelLoadBalancer()

	assert.True(t, sdt == sdtReturned)

	_ = mes.Close()
}

func TestLibp2pMessenger_SendDirectShouldNotBroadcastIfMessageIsPartiallyInvalid(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

	numOfPeers := 4
	_, peers := createMockNetwork(numOfPeers)
	connectPeersFullMesh(peers)

	broadcastMsgResolver := []byte("broadcast resolver msg")
	directMsgResolver := []byte("resolver msg")
	msgRequester := []byte("resolver msg is partially valid, mine is ok")
	numResolverMessagesReceived := int32(0)
	mesProcessorRequester := &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
			if !bytes.Equal(message.Data(), directMsgResolver) {
				// pass through all other messages
				return nil
			}

			atomic.AddInt32(&numResolverMessagesReceived, 1)
			if broadcastHandler != nil {
				broadcastHandler(msgRequester)
			}

			return errors.New("resolver msg is partially valid")
		},
	}

	mesProcessorResolverAndOtherPeers := &mock.MessageProcessorStub{
		ProcessMessageCalled: func(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
			if bytes.Equal(message.Data(), msgRequester) {
				assert.Fail(t, "other peers should have not received filtered out requester's message")
			}
			return nil
		},
	}

	idxRequester := 0

	topic := "testTopic"
	for i := 0; i < numOfPeers; i++ {
		_ = peers[i].CreateTopic(topic, true)
		if i == idxRequester {
			_ = peers[i].RegisterMessageProcessor(topic, mesProcessorRequester)
		} else {
			_ = peers[i].RegisterMessageProcessor(topic, mesProcessorResolverAndOtherPeers)
		}
	}

	fmt.Println("Delaying for peer connections and topic broadcast...")
	time.Sleep(time.Second * 5)

	idxResolver := 1
	fmt.Println("broadcasting a message")
	peers[idxResolver].Broadcast(topic, broadcastMsgResolver)

	time.Sleep(time.Second)

	fmt.Println("sending a direct message")
	_ = peers[idxResolver].SendToConnectedPeer(topic, directMsgResolver, peers[idxRequester].ID())

	time.Sleep(time.Second * 2)
	assert.Equal(t, int32(1), atomic.LoadInt32(&numResolverMessagesReceived))
}

func TestLibp2pMessenger_SendDirectWithMockNetToConnectedPeerShouldWork(t *testing.T) {
	//TODO remove skip when github.com/koron/go-ssdp library is concurrent safe
	if testing.Short() {
		t.Skip("this test fails with race detector on because of the github.com/koron/go-ssdp lib")
	}

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

	_, sk1 := createLibP2PCredentialsMessenger()
	_, sk2 := createLibP2PCredentialsMessenger()

	fmt.Println("Messenger 1:")
	mes1, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		4000,
		sk1,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

	fmt.Println("Messenger 2:")
	mes2, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		4001,
		sk2,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewNullDiscoverer(),
		libp2p.ListenLocalhostAddrWithIp4AndTcp,
		0,
	)

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

//------- Bootstrap

func TestNetworkMessenger_BootstrapPeerDiscoveryShouldCallPeerBootstrapper(t *testing.T) {
	wasCalled := false

	netw := mocknet.New(context.Background())

	pdm := &mock.PeerDiscovererStub{
		BootstrapCalled: func() error {
			wasCalled = true
			return nil
		},
		ApplyContextCalled: func(ctxProvider p2p.ContextProvider) error {
			return nil
		},
		CloseCalled: func() error {
			return nil
		},
	}

	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, pdm)

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
