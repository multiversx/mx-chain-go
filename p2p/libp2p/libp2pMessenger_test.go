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
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/dataThrottle"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/mock"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
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

	_ = mes.SetTopicValidator("test", func(message p2p.MessageP2P) error {
		if bytes.Equal(matchData, message.Data()) {
			fmt.Printf("%s got the message\n", mes.ID().Pretty())
			wg.Done()
		}

		return nil
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

	mes1, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)
	mes2, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)

	return netw, mes1, mes2
}

func createMockMessenger() p2p.Messenger {
	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)

	return mes
}

//------- NewMockLibp2pMessenger

func TestNewMockLibp2pMessenger_NilContextShouldErr(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, err := libp2p.NewMockLibp2pMessenger(nil, netw)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilContext)
}

func TestNewMockLibp2pMessenger_NilMocknetShouldErr(t *testing.T) {
	mes, err := libp2p.NewMockLibp2pMessenger(context.Background(), nil)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilMockNet)
}

func TestNewMockLibp2pMessenger_OkValsShouldWork(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes, err := libp2p.NewMockLibp2pMessenger(context.Background(), netw)

	assert.Nil(t, err)
	assert.NotNil(t, mes)
}

func createLibP2PCredentialsMessenger() (peer.ID, crypto.PrivKey) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	return id, sk
}

//------- NewSocketLibp2pMessenger

func TestNewSocketLibp2pMessenger_NilContextShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewSocketLibp2pMessenger(
		nil,
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		&mock.DataThrottleStub{},
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilContext)
}

func TestNewSocketLibp2pMessenger_InvalidPortShouldErr(t *testing.T) {
	port := 0

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewSocketLibp2pMessenger(
		context.Background(),
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		&mock.DataThrottleStub{},
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrInvalidPort)
}

func TestNewSocketLibp2pMessenger_NilP2PprivateKeyShouldErr(t *testing.T) {
	port := 4000

	mes, err := libp2p.NewSocketLibp2pMessenger(
		context.Background(),
		port,
		nil,
		&mock.ConnManagerNotifieeStub{},
		&mock.DataThrottleStub{},
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilP2PprivateKey)
}

func TestNewSocketLibp2pMessenger_NilSendDataThrottlerShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewSocketLibp2pMessenger(
		context.Background(),
		port,
		sk,
		&mock.ConnManagerNotifieeStub{},
		nil,
	)

	assert.Nil(t, mes)
	assert.Equal(t, err, p2p.ErrNilDataThrottler)
}

func TestNewSocketLibp2pMessenger_NoConnMgrShouldWork(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, err := libp2p.NewSocketLibp2pMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.DataThrottleStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
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

	mes, err := libp2p.NewSocketLibp2pMessenger(
		context.Background(),
		port,
		sk,
		cns,
		&mock.DataThrottleStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
	)

	assert.NotNil(t, mes)
	assert.Nil(t, err)
	assert.True(t, cns == mes.ConnManager())

	mes.Close()
}

//------- Messenger functionality

func TestLibp2pMessenger_ConnectToPeerWrongAddressShouldErr(t *testing.T) {
	mes1 := createMockMessenger()

	adr2 := "invalid_address"

	fmt.Printf("Connecting to %s...\n", adr2)

	err := mes1.ConnectToPeer(adr2)
	assert.NotNil(t, err)
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
}

func TestLibp2pMessenger_CreateTopicOkValsShouldWork(t *testing.T) {
	mes1 := createMockMessenger()

	err := mes1.CreateTopic("test", false)
	assert.Nil(t, err)
}

func TestLibp2pMessenger_CreateTopicTwiceShouldErr(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)
	err := mes1.CreateTopic("test", false)
	assert.Equal(t, p2p.ErrTopicAlreadyExists, err)
}

func TestLibp2pMessenger_HasTopicIfHaveTopicShouldReturnTrue(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)

	assert.True(t, mes1.HasTopic("test"))
}

func TestLibp2pMessenger_HasTopicIfDoNotHaveTopicShouldReturnFalse(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)

	assert.False(t, mes1.HasTopic("one topic"))
}

func TestLibp2pMessenger_HasTopicValidatorDoNotHaveTopicShouldReturnFalse(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)

	assert.False(t, mes1.HasTopicValidator("one topic"))
}

func TestLibp2pMessenger_HasTopicValidatorHaveTopicDoNotHaveValidatorShouldReturnFalse(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)

	assert.False(t, mes1.HasTopicValidator("test"))
}

func TestLibp2pMessenger_HasTopicValidatorHaveTopicHaveValidatorShouldReturnTrue(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)
	_ = mes1.SetTopicValidator("test", func(message p2p.MessageP2P) error {
		return nil
	})

	assert.True(t, mes1.HasTopicValidator("test"))
}

func TestLibp2pMessenger_SetTopicValidatorOnInexistentTopicShouldErr(t *testing.T) {
	mes1 := createMockMessenger()

	err := mes1.SetTopicValidator("test", func(message p2p.MessageP2P) error {
		return nil
	})

	assert.Equal(t, p2p.ErrNilTopic, err)
}

func TestLibp2pMessenger_SetTopicValidatorUnregisterInexistentValidatorShouldErr(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)

	err := mes1.SetTopicValidator("test", nil)

	assert.Equal(t, p2p.ErrTopicValidatorOperationNotSupported, err)
}

func TestLibp2pMessenger_SetTopicValidatorOkValsShouldWork(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)

	err := mes1.SetTopicValidator("test", func(message p2p.MessageP2P) error {
		return nil
	})

	assert.Nil(t, err)
}

func TestLibp2pMessenger_SetTopicValidatorReregistrationShouldErr(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)

	//registration
	_ = mes1.SetTopicValidator("test", func(message p2p.MessageP2P) error {
		return nil
	})

	//re-registration
	err := mes1.SetTopicValidator("test", func(message p2p.MessageP2P) error {
		return nil
	})

	assert.Equal(t, p2p.ErrTopicValidatorOperationNotSupported, err)
}

func TestLibp2pMessenger_SetTopicValidatorUnReregistrationShouldWork(t *testing.T) {
	mes1 := createMockMessenger()

	_ = mes1.CreateTopic("test", false)

	//registration
	_ = mes1.SetTopicValidator("test", func(message p2p.MessageP2P) error {
		return nil
	})

	//unregistration
	err := mes1.SetTopicValidator("test", nil)

	assert.Nil(t, err)
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

	mes1.BroadcastData("test", "test", msg)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)
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
}

func TestLibp2pMessenger_ConnectedPeers(t *testing.T) {
	netw, mes1, mes2 := createMockNetworkOf2()
	mes3, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)

	adr2 := mes2.Addresses()[0]

	fmt.Printf("Connecting to %s...\n", adr2)

	_ = mes1.ConnectToPeer(adr2)
	_ = mes3.ConnectToPeer(adr2)

	//connected peers:  1 ----- 2 ----- 3

	assert.Equal(t, []p2p.PeerID{mes2.ID()}, mes1.ConnectedPeers())
	assert.Equal(t, []p2p.PeerID{mes2.ID()}, mes3.ConnectedPeers())
	assert.Equal(t, 2, len(mes2.ConnectedPeers()))
	//no need to further test taht mes2 is connected to mes1 and mes3 s this was tested in first 2 asserts
}

func TestLibp2pMessenger_DiscoverNewPeersNilDiscovererShouldErr(t *testing.T) {
	netw := mocknet.New(context.Background())

	mes1, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)
	mes1.SetDiscoverer(nil)

	err := mes1.DiscoverNewPeers()
	assert.Equal(t, p2p.ErrNilDiscoverer, err)
}

func TestLibp2pMessenger_DiscoverNewPeersDiscovererErrsShouldErr(t *testing.T) {
	ds := &mock.DiscovererStub{}
	ds.FindPeersCalled = func(ctx context.Context, ns string, opts ...discovery.Option) (infos <-chan peerstore.PeerInfo, e error) {
		return nil, errors.New("error")
	}

	netw := mocknet.New(context.Background())

	mes, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)
	mes.SetDiscoverer(ds)

	err := mes.DiscoverNewPeers()
	assert.Equal(t, "error", err.Error())
}

func TestLibp2pMessenger_DiscoverNewPeersShouldWork(t *testing.T) {
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

	mes, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)
	mes.SetDiscoverer(ds)

	foundPeers := make([]peerstore.PeerInfo, 0)

	mes.PeerDiscoveredHandler = func(pInfo peerstore.PeerInfo) {
		foundPeers = append(foundPeers, pInfo)
	}

	err := mes.DiscoverNewPeers()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundPeers))
	assert.Equal(t, foundPeers[0], pInfo1)
	assert.Equal(t, foundPeers[1], pInfo2)
}

func TestLibp2pMessenger_DiscoverNewPeersWithRealDiscovererShouldWork(t *testing.T) {
	netw := mocknet.New(context.Background())

	advertiser, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)
	mes1, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)
	mes2, _ := libp2p.NewMockLibp2pMessenger(context.Background(), netw)

	adrAdvertiser := advertiser.Addresses()[0]

	mutAdvertiser := sync.Mutex{}
	peersFoundByAdvertiser := make(map[peer.ID]peerstore.PeerInfo)
	advertiser.PeerDiscoveredHandler = func(pInfo peerstore.PeerInfo) {
		mutAdvertiser.Lock()
		peersFoundByAdvertiser[pInfo.ID] = pInfo
		mutAdvertiser.Unlock()
	}

	mutMes1 := sync.Mutex{}
	peersFoundByMes1 := make(map[peer.ID]peerstore.PeerInfo)
	mes1.PeerDiscoveredHandler = func(pInfo peerstore.PeerInfo) {
		mutMes1.Lock()
		peersFoundByMes1[pInfo.ID] = pInfo
		mutMes1.Unlock()
	}

	mutMes2 := sync.Mutex{}
	peersFoundByMes2 := make(map[peer.ID]peerstore.PeerInfo)
	mes2.PeerDiscoveredHandler = func(pInfo peerstore.PeerInfo) {
		mutMes2.Lock()
		peersFoundByMes2[pInfo.ID] = pInfo
		mutMes2.Unlock()
	}

	mes1.ConnectToPeer(adrAdvertiser)
	mes2.ConnectToPeer(adrAdvertiser)

	time.Sleep(time.Second)

	err := advertiser.DiscoverNewPeers()
	assert.Nil(t, err)

	err = mes1.DiscoverNewPeers()
	assert.Nil(t, err)

	err = mes2.DiscoverNewPeers()
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

	mes, _ := libp2p.NewSocketLibp2pMessenger(
		context.Background(),
		port,
		sk,
		cns,
		&mock.DataThrottleStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
	)

	err := mes.TrimConnections()

	assert.True(t, wasCalled)
	assert.Nil(t, err)

	mes.Close()
}

func TestLibp2pMessenger_TrimConnectionsCallsOnNilConnManagerShouldErr(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	mes, _ := libp2p.NewSocketLibp2pMessenger(
		context.Background(),
		port,
		sk,
		nil,
		&mock.DataThrottleStub{
			CollectFromPipesCalled: func() []*p2p.SendableData {
				return make([]*p2p.SendableData, 0)
			},
		},
	)

	err := mes.TrimConnections()

	assert.Equal(t, p2p.ErrNilConnManager, err)

	mes.Close()
}

func TestLibp2pMessenger_SendDataThrottlerShouldReturnCorrectObject(t *testing.T) {
	port := 4000

	_, sk := createLibP2PCredentialsMessenger()

	sdt := &mock.DataThrottleStub{
		AddPipeCalled: func(pipe string) error {
			return nil
		},
		CollectFromPipesCalled: func() []*p2p.SendableData {
			return make([]*p2p.SendableData, 0)
		},
	}

	mes, _ := libp2p.NewSocketLibp2pMessenger(
		context.Background(),
		port,
		sk,
		nil,
		sdt,
	)

	sdtReturned := mes.SendDataThrottler()

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

	err := mes1.SendDirectToConnectedPeer("test", msg, mes2.ID())

	assert.Nil(t, err)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)
}

func TestLibp2pMessenger_SendDirectWithRealNetToConnectedPeerShouldWork(t *testing.T) {
	msg := []byte("test message")

	_, sk1 := createLibP2PCredentialsMessenger()
	_, sk2 := createLibP2PCredentialsMessenger()

	fmt.Println("Messenger 1:")
	mes1, _ := libp2p.NewSocketLibp2pMessenger(context.Background(), 4000, sk1, nil, dataThrottle.NewSendDataThrottle())
	fmt.Println("Messenger 2:")
	mes2, _ := libp2p.NewSocketLibp2pMessenger(context.Background(), 4001, sk2, nil, dataThrottle.NewSendDataThrottle())

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
	err = mes1.SendDirectToConnectedPeer("test", msg, mes2.ID())
	assert.Nil(t, err)

	time.Sleep(time.Second)
	fmt.Printf("Messenger 2 is sending message from %s...\n", mes2.ID().Pretty())
	err = mes2.SendDirectToConnectedPeer("test", msg, mes1.ID())
	assert.Nil(t, err)

	waitDoneWithTimeout(t, chanDone, timeoutWaitResponses)
}
