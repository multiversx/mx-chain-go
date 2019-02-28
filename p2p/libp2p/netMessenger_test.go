package libp2p_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
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
			ProcessMessageCalled: func(message p2p.MessageP2P) error {
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
		if strings.Contains(addr, "circuit") {
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

	netw.LinkAll()

	return netw, mes1, mes2
}

func createMockMessenger() p2p.Messenger {
	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())

	return mes
}

func createLibP2PCredentialsMessenger() (peer.ID, crypto.PrivKey) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	return id, sk
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
	assert.NotNil(t, mes)

	mes.Close()
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
		&mock.PipeLoadBalancerStub{},
		discovery.NewNullDiscoverer(),
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilContext)
}

func TestNewNetworkMessenger_InvalidPortShouldErr(t *testing.T) {
	port := 0

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		&mock.PipeLoadBalancerStub{},
		discovery.NewNullDiscoverer(),
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
		&mock.PipeLoadBalancerStub{},
		discovery.NewNullDiscoverer(),
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
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilPipeLoadBalancer)
}

func TestNewNetworkMessenger_NoConnMgrShouldWork(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.PipeLoadBalancerStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
		discovery.NewNullDiscoverer(),
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)

	mes.Close()
}

func TestNewNetworkMessenger_WithConnMgrShouldWork(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	cns := &mock.ConnManagerNotifieeStub{
		ListenCalled:      func(netw net.Network, ma multiaddr.Multiaddr) {},
		ListenCloseCalled: func(netw net.Network, ma multiaddr.Multiaddr) {},
	}

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		cns,
		&mock.PipeLoadBalancerStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
		discovery.NewNullDiscoverer(),
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)
	assert.True(t, cns == mes.ConnManager())

	mes.Close()
}

func TestNewNetworkMessenger_WithNullPeerDiscoveryShouldWork(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.PipeLoadBalancerStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
		discovery.NewNullDiscoverer(),
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)

	mes.Close()
}

func TestNewNetworkMessenger_NilPeerDiscoveryShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.PipeLoadBalancerStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
		nil,
	)

	assert.Nil(t, mes)
	assert.Equal(t, p2p.ErrNilPeerDiscoverer, err)
}

func TestNewNetworkMessenger_PeerDiscovererFailsWhenApplyingContextShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	errExpected := errors.New("expected err")

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.PipeLoadBalancerStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
		&mock.PeerDiscovererStub{
			ApplyContextCalled: func(ctxProvider p2p.ContextProvider) error {
				return errExpected
			},
		},
	)

	assert.Nil(t, mes)
	assert.Equal(t, errExpected, err)
}

//------- Messenger functionality

func TestLibp2pMessenger_ConnectToPeerShoulsCallUpgradedHost(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
	mes.Close()

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
	mes.ConnectToPeer(p)
	assert.True(t, wasCalled)
}

func TestLibp2pMessenger_IsConnectedShouldWork(t *testing.T) {
	_, mes1, mes2 := createMockNetworkOf2()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)

	assert.True(t, mes1.IsConnected(mes2.ID()))
	assert.True(t, mes2.IsConnected(mes1.ID()))

	mes1.Close()
	mes2.Close()
}

func TestLibp2pMessenger_CreateTopicOkValsShouldWork(t *testing.T) {
	mes := createMockMessenger()

	err := mes.CreateTopic("test", true)
	assert.Nil(t, err)

	mes.Close()
}

func TestLibp2pMessenger_CreateTopicTwiceShouldErr(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)
	err := mes.CreateTopic("test", false)
	assert.Equal(t, p2p.ErrTopicAlreadyExists, err)

	mes.Close()
}

func TestLibp2pMessenger_HasTopicIfHaveTopicShouldReturnTrue(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	assert.True(t, mes.HasTopic("test"))

	mes.Close()
}

func TestLibp2pMessenger_HasTopicIfDoNotHaveTopicShouldReturnFalse(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	assert.False(t, mes.HasTopic("one topic"))

	mes.Close()
}

func TestLibp2pMessenger_HasTopicValidatorDoNotHaveTopicShouldReturnFalse(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	assert.False(t, mes.HasTopicValidator("one topic"))

	mes.Close()
}

func TestLibp2pMessenger_HasTopicValidatorHaveTopicDoNotHaveValidatorShouldReturnFalse(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	assert.False(t, mes.HasTopicValidator("test"))

	mes.Close()
}

func TestLibp2pMessenger_HasTopicValidatorHaveTopicHaveValidatorShouldReturnTrue(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)
	_ = mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	assert.True(t, mes.HasTopicValidator("test"))

	mes.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorOnInexistentTopicShouldErr(t *testing.T) {
	mes := createMockMessenger()

	err := mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	assert.Equal(t, p2p.ErrNilTopic, err)

	mes.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorWithNilHandlerShouldErr(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	err := mes.RegisterMessageProcessor("test", nil)

	assert.Equal(t, p2p.ErrNilValidator, err)

	mes.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorOkValsShouldWork(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	err := mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	assert.Nil(t, err)

	mes.Close()
}

func TestLibp2pMessenger_RegisterTopicValidatorReregistrationShouldErr(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	//registration
	_ = mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	//re-registration
	err := mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	assert.Equal(t, p2p.ErrTopicValidatorOperationNotSupported, err)

	mes.Close()
}

func TestLibp2pMessenger_UnegisterTopicValidatorOnInexistentTopicShouldErr(t *testing.T) {
	mes := createMockMessenger()

	err := mes.UnregisterMessageProcessor("test")

	assert.Equal(t, p2p.ErrNilTopic, err)

	mes.Close()
}

func TestLibp2pMessenger_UnegisterTopicValidatorOnANotRegisteredTopicShouldErr(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	err := mes.UnregisterMessageProcessor("test")

	assert.Equal(t, p2p.ErrTopicValidatorOperationNotSupported, err)

	mes.Close()
}

func TestLibp2pMessenger_UnregisterTopicValidatorShouldWork(t *testing.T) {
	mes := createMockMessenger()

	_ = mes.CreateTopic("test", false)

	//registration
	_ = mes.RegisterMessageProcessor("test", &mock.MessageProcessorStub{})

	//unregistration
	err := mes.UnregisterMessageProcessor("test")

	assert.Nil(t, err)

	mes.Close()
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

	mes1.Close()
	mes2.Close()
}

func TestLibp2pMessenger_BroadcastDataOnTopicPipeBetween2PeersShouldWork(t *testing.T) {
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

	mes1.Close()
	mes2.Close()
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

	mes1.Close()
	mes2.Close()
}

func TestLibp2pMessenger_ConnectedPeers(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())

	netw.LinkAll()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)

	//connected peers:  1 ----- 2 ----- 3

	assert.Equal(t, []p2p.PeerID{mes2.ID()}, mes1.ConnectedPeers())
	assert.Equal(t, []p2p.PeerID{mes2.ID()}, mes3.ConnectedPeers())
	assert.Equal(t, 2, len(mes2.ConnectedPeers()))
	//no need to further test that mes2 is connected to mes1 and mes3 s this was tested in first 2 asserts

	mes1.Close()
	mes2.Close()
	mes3.Close()
}

func TestLibp2pMessenger_ConnectedPeersShouldReturnUniquePeers(t *testing.T) {
	pid1 := p2p.PeerID("pid1")
	pid2 := p2p.PeerID("pid2")
	pid3 := p2p.PeerID("pid3")
	pid4 := p2p.PeerID("pid4")

	hs := &mock.ConnectableHostStub{
		NetworkCalled: func() net.Network {
			return &mock.NetworkStub{
				ConnsCalled: func() []net.Conn {
					//generate a mock list that contain duplicates
					return []net.Conn{
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
				ConnectednessCalled: func(id peer.ID) net.Connectedness {
					return net.Connected
				},
			}
		},
	}

	netw := mocknet.New(context.Background())
	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, discovery.NewNullDiscoverer())
	//we can safely close the host as the next operations will be done on a mock
	mes.Close()

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

func generateConnWithRemotePeer(pid p2p.PeerID) net.Conn {
	return &mock.ConnStub{
		RemotePeerCalled: func() peer.ID {
			return peer.ID(pid)
		},
	}
}

func TestLibp2pMessenger_TrimConnectionsCallsConnManagerTrimConnections(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	wasCalled := false

	cns := &mock.ConnManagerNotifieeStub{
		ListenCalled:      func(netw net.Network, ma multiaddr.Multiaddr) {},
		ListenCloseCalled: func(netw net.Network, ma multiaddr.Multiaddr) {},
		TrimOpenConnsCalled: func(ctx context.Context) {
			wasCalled = true
		},
	}

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		cns,
		&mock.PipeLoadBalancerStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
		discovery.NewNullDiscoverer(),
	)

	mes.TrimConnections()

	assert.True(t, wasCalled)

	mes.Close()
}

func TestLibp2pMessenger_SendDataThrottlerShouldReturnCorrectObject(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	sdt := &mock.PipeLoadBalancerStub{
		AddPipeCalled: func(pipe string) error {
			return nil
		},
		CollectFromPipesCalled: func() []*p2p.SendableData {
			return make([]*p2p.SendableData, 0)
		},
	}

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		nil,
		sdt,
		discovery.NewNullDiscoverer(),
	)

	sdtReturned := mes.OutgoingPipeLoadBalancer()

	assert.True(t, sdt == sdtReturned)

	mes.Close()
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

	mes1.Close()
	mes2.Close()
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
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		discovery.NewNullDiscoverer(),
	)

	fmt.Println("Messenger 2:")
	mes2, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		4001,
		sk2,
		nil,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		discovery.NewNullDiscoverer(),
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

	mes1.Close()
	mes2.Close()
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

	mes.Close()
}
