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
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
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

	mes1, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryOff)
	mes2, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryOff)

	return netw, mes1, mes2
}

func createMockMessenger() p2p.Messenger {
	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryOff)

	return mes
}

func createLibP2PCredentialsMessenger() (peer.ID, crypto.PrivKey) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	return id, sk
}

//------- NewMockLibp2pMessenger

func TestNewMockLibp2pMessenger_NilContextShouldErr(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, err := libp2p.NewMemoryMessenger(nil, netw, p2p.PeerDiscoveryOff)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilContext)
}

func TestNewMockLibp2pMessenger_NilMocknetShouldErr(t *testing.T) {
	mes, err := libp2p.NewMemoryMessenger(context.Background(), nil, p2p.PeerDiscoveryOff)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilMockNet)
}

func TestNewMockLibp2pMessenger_OkValsShouldWork(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, err := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryOff)

	assert.Nil(t, err)
	assert.NotNil(t, mes)

	mes.Close()
}

//------- NewSocketLibp2pMessenger

func TestNewSocketLibp2pMessenger_NilContextShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		nil,
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		&mock.PipeLoadBalancerStub{},
		p2p.PeerDiscoveryOff,
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilContext)
}

func TestNewSocketLibp2pMessenger_InvalidPortShouldErr(t *testing.T) {
	port := 0

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		&mock.PipeLoadBalancerStub{},
		p2p.PeerDiscoveryOff,
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrInvalidPort)
}

func TestNewSocketLibp2pMessenger_NilP2PprivateKeyShouldErr(t *testing.T) {
	port := 4000

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		nil,
		&mock.ConnManagerNotifieeStub{},
		&mock.PipeLoadBalancerStub{},
		p2p.PeerDiscoveryOff,
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilP2PprivateKey)
}

func TestNewSocketLibp2pMessenger_NilPipeLoadBalancerShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewNetworkMessenger(
		context.Background(),
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		nil,
		p2p.PeerDiscoveryOff,
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilPipeLoadBalancer)
}

func TestNewSocketLibp2pMessenger_NoConnMgrShouldWork(t *testing.T) {
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
		p2p.PeerDiscoveryOff,
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)

	mes.Close()
}

func TestNewSocketLibp2pMessenger_WithConnMgrShouldWork(t *testing.T) {
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
		p2p.PeerDiscoveryOff,
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)
	assert.True(t, cns == mes.ConnManager())

	mes.Close()
}

func TestNewSocketLibp2pMessenger_WithMdnsPeerDiscoveryShouldWork(t *testing.T) {
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
		p2p.PeerDiscoveryMdns,
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)

	mes.Close()
}

func TestNewSocketLibp2pMessenger_NoPeerDiscoveryImplementationShouldError(t *testing.T) {
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
		10000,
	)

	assert.Nil(t, mes)
	assert.Equal(t, p2p.ErrPeerDiscoveryNotImplemented, err)
}

//------- Messenger functionality

func TestLibp2pMessenger_ConnectToPeerWrongAddressShouldErr(t *testing.T) {
	mes1 := createMockMessenger()

	adr2 := "invalid_address"

	fmt.Printf("Connecting to %s...\n", adr2)

	err := mes1.ConnectToPeer(adr2)
	assert.NotNil(t, err)

	mes1.Close()
}

func TestLibp2pMessenger_ConnectToPeerAndClose2PeersShouldWork(t *testing.T) {
	_, mes1, mes2 := createMockNetworkOf2()

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	err := mes1.ConnectToPeer(adr2)
	assert.Nil(t, err)

	err = mes1.Close()
	assert.Nil(t, err)

	err = mes2.Close()
	assert.Nil(t, err)
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
	mes3, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryOff)

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
}

func TestLibp2pMessenger_ConnectedPeersShouldReturnUniquePeers(t *testing.T) {
	pid1 := p2p.PeerID("pid1")
	pid2 := p2p.PeerID("pid2")
	pid3 := p2p.PeerID("pid3")
	pid4 := p2p.PeerID("pid4")

	hs := &mock.HostStub{
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
	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryOff)
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

func TestLibp2pMessenger_KadDhtDiscoverNewPeersNilDiscovererShouldErr(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryKadDht)
	mes.SetDiscoverer(nil)

	err := mes.KadDhtDiscoverNewPeers()
	assert.Equal(t, p2p.ErrNilDiscoverer, err)

	mes.Close()
}

func TestLibp2pMessenger_KadDhtDiscoverNewPeersDiscovererErrsShouldErr(t *testing.T) {
	ds := &mock.DiscovererStub{}
	ds.FindPeersCalled = func(ctx context.Context, ns string, opts ...discovery.Option) (infos <-chan peerstore.PeerInfo, e error) {
		return nil, errors.New("error")
	}

	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryKadDht)
	mes.SetDiscoverer(ds)

	err := mes.KadDhtDiscoverNewPeers()
	assert.Equal(t, "error", err.Error())

	mes.Close()
}

func TestLibp2pMessenger_KadDhtDiscoverNewPeersShouldWork(t *testing.T) {
	pInfo1 := peerstore.PeerInfo{ID: peer.ID("peer1")}
	pInfo2 := peerstore.PeerInfo{ID: peer.ID("peer2")}

	ds := &mock.DiscovererStub{}
	ds.FindPeersCalled = func(ctx context.Context, ns string, opts ...discovery.Option) (infos <-chan peerstore.PeerInfo, e error) {
		ch := make(chan peerstore.PeerInfo)

		go func(ch chan peerstore.PeerInfo) {
			//emulating find peers taking some time
			time.Sleep(time.Millisecond * 100)

			ch <- pInfo1

			time.Sleep(time.Millisecond * 100)

			ch <- pInfo2

			time.Sleep(time.Millisecond * 100)

			close(ch)
		}(ch)

		return ch, nil
	}

	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryKadDht)
	mes.SetDiscoverer(ds)

	foundPeers := make([]peerstore.PeerInfo, 0)

	mes.SetPeerDiscoveredHandler(func(pInfo peerstore.PeerInfo) {
		foundPeers = append(foundPeers, pInfo)
	})

	err := mes.KadDhtDiscoverNewPeers()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundPeers))
	assert.Equal(t, foundPeers[0], pInfo1)
	assert.Equal(t, foundPeers[1], pInfo2)

	mes.Close()
}

func TestLibp2pMessenger_KadDhtDiscoverNewPeersWithRealDiscovererShouldWork(t *testing.T) {
	netw := mocknet.New(context.Background())

	advertiser, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryKadDht)
	mes1, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryKadDht)
	mes2, _ := libp2p.NewMemoryMessenger(context.Background(), netw, p2p.PeerDiscoveryKadDht)

	adrAdvertiser := advertiser.Addresses()[0]

	mutAdvertiser := sync.Mutex{}
	peersFoundByAdvertiser := make(map[peer.ID]peerstore.PeerInfo)
	advertiser.SetPeerDiscoveredHandler(func(pInfo peerstore.PeerInfo) {
		mutAdvertiser.Lock()
		peersFoundByAdvertiser[pInfo.ID] = pInfo
		mutAdvertiser.Unlock()
	})

	mutMes1 := sync.Mutex{}
	peersFoundByMes1 := make(map[peer.ID]peerstore.PeerInfo)
	mes1.SetPeerDiscoveredHandler(func(pInfo peerstore.PeerInfo) {
		mutMes1.Lock()
		peersFoundByMes1[pInfo.ID] = pInfo
		mutMes1.Unlock()
	})

	mutMes2 := sync.Mutex{}
	peersFoundByMes2 := make(map[peer.ID]peerstore.PeerInfo)
	mes2.SetPeerDiscoveredHandler(func(pInfo peerstore.PeerInfo) {
		mutMes2.Lock()
		peersFoundByMes2[pInfo.ID] = pInfo
		mutMes2.Unlock()
	})

	mes1.ConnectToPeer(adrAdvertiser)
	mes2.ConnectToPeer(adrAdvertiser)

	time.Sleep(time.Second)

	err := advertiser.KadDhtDiscoverNewPeers()
	assert.Nil(t, err)

	err = mes1.KadDhtDiscoverNewPeers()
	assert.Nil(t, err)

	err = mes2.KadDhtDiscoverNewPeers()
	assert.Nil(t, err)

	//we can not make an assertion for len to be equal to 3 because there is simple no guarantee
	//that a peer always fetch entire networks known by its peers
	assert.True(t, len(peersFoundByAdvertiser) >= 2)
	assert.True(t, len(peersFoundByMes1) >= 2)
	assert.True(t, len(peersFoundByMes2) >= 2)

	mes1.Close()
	mes2.Close()
	advertiser.Close()
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
		p2p.PeerDiscoveryOff,
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
		p2p.PeerDiscoveryOff,
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
		p2p.PeerDiscoveryOff,
	)

	fmt.Println("Messenger 2:")
	mes2, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		4001,
		sk2,
		nil,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		p2p.PeerDiscoveryOff,
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

func TestNetworkMessenger_BootstrapPeerDiscoveryOffShouldReturnNil(t *testing.T) {
	_, sk := createLibP2PCredentialsMessenger()

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		4000,
		sk,
		nil,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		p2p.PeerDiscoveryOff,
	)

	err := mes.Bootstrap()

	assert.Nil(t, err)

	mes.Close()
}

func TestNetworkMessenger_BootstrapMdnsPeerDiscoveryShouldReturnNil(t *testing.T) {
	_, sk := createLibP2PCredentialsMessenger()

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		23000,
		sk,
		nil,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		p2p.PeerDiscoveryMdns,
	)

	err := mes.Bootstrap()
	assert.Nil(t, err)

	mes.Close()
}

func TestNetworkMessenger_BootstrapMdnsPeerDiscoveryCalledTwiceShouldErr(t *testing.T) {
	_, sk := createLibP2PCredentialsMessenger()

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		23000,
		sk,
		nil,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		p2p.PeerDiscoveryMdns,
	)

	_ = mes.Bootstrap()
	err := mes.Bootstrap()
	assert.Equal(t, p2p.ErrPeerDiscoveryProcessAlreadyStarted, err)

	mes.Close()
}

func TestNetworkMessenger_HandlePeerFoundNotFoundShouldTryToConnect(t *testing.T) {
	_, sk := createLibP2PCredentialsMessenger()

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		23000,
		sk,
		nil,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		p2p.PeerDiscoveryOff,
	)
	//closing "real" host as to check with a mock host
	mes.Close()

	newPeerInfo := peerstore.PeerInfo{
		ID: peer.ID("new found peerID"),
	}
	testAddress := "/ip4/127.0.0.1/tcp/23000/p2p/16Uiu2HAkyqtHSEJDkYhVWTtm9j58Mq5xQJgrApBYXMwS6sdamXuE"
	address, _ := multiaddr.NewMultiaddr(testAddress)
	newPeerInfo.Addrs = []multiaddr.Multiaddr{address}

	chanConnected := make(chan struct{})

	mockHost := &mock.HostStub{
		PeerstoreCalled: func() peerstore.Peerstore {
			return peerstore.NewPeerstore(
				pstoremem.NewKeyBook(),
				pstoremem.NewAddrBook(),
				pstoremem.NewPeerMetadata())
		},
		ConnectCalled: func(ctx context.Context, pi peerstore.PeerInfo) error {
			if newPeerInfo.ID == pi.ID {
				chanConnected <- struct{}{}
			}

			return nil
		},
	}

	mes.SetHost(mockHost)

	mes.HandlePeerFound(newPeerInfo)

	select {
	case <-chanConnected:
		return
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout while waiting to call host.Connect")
	}
}

func TestNetworkMessenger_HandlePeerFoundPeerFoundShouldNotTryToConnect(t *testing.T) {
	_, sk := createLibP2PCredentialsMessenger()

	mes, _ := libp2p.NewNetworkMessenger(
		context.Background(),
		23000,
		sk,
		nil,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		p2p.PeerDiscoveryOff,
	)
	//closing "real" host as to check with a mock host
	mes.Close()

	newPeerInfo := peerstore.PeerInfo{
		ID: peer.ID("new found peerID"),
	}
	testAddress := "/ip4/127.0.0.1/tcp/23000/p2p/16Uiu2HAkyqtHSEJDkYhVWTtm9j58Mq5xQJgrApBYXMwS6sdamXuE"
	address, _ := multiaddr.NewMultiaddr(testAddress)
	newPeerInfo.Addrs = []multiaddr.Multiaddr{address}

	chanConnected := make(chan struct{})

	mockHost := &mock.HostStub{
		PeerstoreCalled: func() peerstore.Peerstore {
			ps := peerstore.NewPeerstore(
				pstoremem.NewKeyBook(),
				pstoremem.NewAddrBook(),
				pstoremem.NewPeerMetadata())
			ps.AddAddrs(newPeerInfo.ID, newPeerInfo.Addrs, peerstore.PermanentAddrTTL)

			return ps
		},
		ConnectCalled: func(ctx context.Context, pi peerstore.PeerInfo) error {
			if newPeerInfo.ID == pi.ID {
				chanConnected <- struct{}{}
			}

			return nil
		},
	}

	mes.SetHost(mockHost)

	mes.HandlePeerFound(newPeerInfo)

	select {
	case <-chanConnected:
		assert.Fail(t, "should have not called host.Connect")
	case <-time.After(timeoutWaitResponses):
		return
	}
}
