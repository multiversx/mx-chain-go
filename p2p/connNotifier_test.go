package p2p_test

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

var testConnNotifierMaxWaitTime = time.Second * 5
var testConnNotifierWaitTimeForNoResponse = time.Second

func TestConnNotifier_TaskResolveConnections_NotAllowedToConnect(t *testing.T) {
	cn := p2p.NewConnNotifier(0)

	assert.Equal(t, p2p.WontConnect, cn.TaskResolveConnections())
}

func TestConnNotifier_TaskResolveConnections_OnlyInboundConnections(t *testing.T) {
	cn := p2p.NewConnNotifier(1)

	//will return 2 inbound connections
	cn.GetConnections = func(sender *p2p.ConnNotifier) []net.Conn {
		return []net.Conn{
			&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}},
			&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}}}
	}

	assert.Equal(t, p2p.OnlyInboundConnections, cn.TaskResolveConnections())
}

func TestConnNotifier_TaskResolveConnections_NothingDone(t *testing.T) {
	cn := p2p.NewConnNotifier(1)

	assert.Equal(t, p2p.NothingDone, cn.TaskResolveConnections())
}

func TestConnNotifier_ComputeInboundOutboundConns(t *testing.T) {
	cn := p2p.NewConnNotifier(1)

	//2 inbound, 3 outbound
	conns := []net.Conn{
		&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}},
		&mock.ConnMock{Status: net.Stat{Direction: net.DirOutbound}},
		&mock.ConnMock{Status: net.Stat{Direction: net.DirOutbound}},
		&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}},
		&mock.ConnMock{Status: net.Stat{Direction: net.DirOutbound}}}

	inbound, outbound := cn.ComputeInboundOutboundConns(conns)

	assert.Equal(t, inbound, 2)
	assert.Equal(t, outbound, 3)

}

func TestConnNotifier_TryToConnectWithSuccess_Running(t *testing.T) {
	chanDone := make(chan bool, 0)

	cn := p2p.NewConnNotifier(1)

	cn.GetConnections = func(sender *p2p.ConnNotifier) []net.Conn {
		return make([]net.Conn, 0)
	}

	cn.IsConnected = func(sender *p2p.ConnNotifier, pid peer.ID) bool {
		return false
	}

	cn.GetKnownPeers = func(sender *p2p.ConnNotifier) []peer.ID {
		return []peer.ID{"aaa", "bbb"}
	}

	cn.ConnectToPeer = func(cn *p2p.ConnNotifier, id peer.ID) error {
		chanDone <- true

		return nil
	}

	cn.Start()
	defer cn.Stop()

	select {
	case <-chanDone:
		fmt.Println("ConnectToPeer called!")
	case <-time.After(testConnNotifierMaxWaitTime):
		assert.Fail(t, "Should have called to connect!")
		return
	}
}

func TestConnNotifier_TryToConnectWithSuccessOn2Peers_Running(t *testing.T) {
	chanDone := make(chan bool, 0)

	cn := p2p.NewConnNotifier(1)

	aaaTriedToConnect := int32(0)
	bbbTriedToConnect := int32(0)

	cn.GetConnections = func(sender *p2p.ConnNotifier) []net.Conn {
		return make([]net.Conn, 0)
	}

	cn.IsConnected = func(sender *p2p.ConnNotifier, pid peer.ID) bool {
		return false
	}

	cn.GetKnownPeers = func(sender *p2p.ConnNotifier) []peer.ID {
		return []peer.ID{"aaa", "bbb"}
	}

	cn.ConnectToPeer = func(cn *p2p.ConnNotifier, id peer.ID) error {
		if id == "aaa" {
			atomic.AddInt32(&aaaTriedToConnect, 1)
		}

		if id == "bbb" {
			atomic.AddInt32(&bbbTriedToConnect, 1)
		}

		return nil
	}

	cn.Start()
	defer cn.Stop()

	go func() {
		//function to check that it tried to connect 2 times to aaa and 2 times to bbb

		for {
			if atomic.LoadInt32(&aaaTriedToConnect) == 2 &&
				atomic.LoadInt32(&bbbTriedToConnect) == 2 {
				chanDone <- true
				return
			}
		}
	}()

	select {
	case <-chanDone:
		fmt.Println("ConnectToPeer called 2 times for aaa and 2 times for bbb!")
	case <-time.After(testConnNotifierMaxWaitTime):
		assert.Fail(t, "ConnectToPeer have called 2 times for aaa and 2 times for bbb!")
		return
	}
}

func TestConnNotifier_TryToConnectNoOtherPeers_Running(t *testing.T) {
	chanDone := make(chan bool, 0)

	cn := p2p.NewConnNotifier(1)

	cn.GetConnections = func(sender *p2p.ConnNotifier) []net.Conn {
		return make([]net.Conn, 0)
	}

	cn.IsConnected = func(sender *p2p.ConnNotifier, pid peer.ID) bool {
		return true
	}

	cn.GetKnownPeers = func(sender *p2p.ConnNotifier) []peer.ID {
		return []peer.ID{"aaa", "bbb"}
	}

	cn.ConnectToPeer = func(cn *p2p.ConnNotifier, id peer.ID) error {
		chanDone <- true

		return nil
	}

	cn.Start()
	defer cn.Stop()

	select {
	case <-chanDone:
		fmt.Println("Should have not called to connect!")
		return
	case <-time.After(testConnNotifierMaxWaitTime):

	}
}

func TestConnNotifier_Connected_GetConnNil_ShouldCloseConn(t *testing.T) {
	cn := p2p.NewConnNotifier(1)

	chanDone := make(chan bool, 0)

	connMonitored := mock.ConnMock{}
	connMonitored.CloseCalled = func(connMock *mock.ConnMock) error {
		chanDone <- true

		return nil
	}

	go cn.Connected(nil, &connMonitored)

	select {
	case <-chanDone:
		fmt.Println("Connection closed as expected!")
	case <-time.After(testConnNotifierMaxWaitTime):
		assert.Fail(t, "Should have called conn.Close()!")
		return
	}
}

func TestConnNotifier_Connected_MaxConnReached_ShouldCloseConn(t *testing.T) {
	cn := p2p.NewConnNotifier(2)

	cn.GetConnections = func(sender *p2p.ConnNotifier) []net.Conn {
		return []net.Conn{
			&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}},
			&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}},
			&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}}}
	}

	chanDone := make(chan bool, 0)

	connMonitored := mock.ConnMock{}
	connMonitored.CloseCalled = func(connMock *mock.ConnMock) error {
		chanDone <- true

		return nil
	}

	go cn.Connected(nil, &connMonitored)

	select {
	case <-chanDone:
		fmt.Println("Connection closed as expected!")
	case <-time.After(testConnNotifierMaxWaitTime):
		assert.Fail(t, "Should have called conn.Close()!")
		return
	}
}

func TestConnNotifier_Connected_CanAccept_ShouldNotCloseConn(t *testing.T) {
	cn := p2p.NewConnNotifier(2)

	cn.GetConnections = func(sender *p2p.ConnNotifier) []net.Conn {
		return []net.Conn{
			&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}}}
	}

	chanDone := make(chan bool, 0)

	connMonitored := mock.ConnMock{}
	connMonitored.CloseCalled = func(connMock *mock.ConnMock) error {
		chanDone <- true

		return nil
	}

	go cn.Connected(nil, &connMonitored)

	select {
	case <-chanDone:
		assert.Fail(t, "Should have not called conn.Close()!")
		return
	case <-time.After(testConnNotifierWaitTimeForNoResponse):
		fmt.Println("Connection not closed!")

	}
}

//
//func TestRemoveInboundPeers(t *testing.T) {
//	//steps:
//	// - create 3 nodes
//	// - node 1 has a limit of 2 peers to connect to
//	// - node 2 and node 3 connects to node 1
//	// - run the function that checks the connections and the function will determine that there are too many inbound
//	//   connections, will close the first connection and will return the status OnlyInboundConnections
//	cp1, err := p2p.NewConnectParamsFromPort(5000)
//	assert.Nil(t, err)
//
//	cp2, err := p2p.NewConnectParamsFromPort(5001)
//	assert.Nil(t, err)
//
//	cp3, err := p2p.NewConnectParamsFromPort(5002)
//	assert.Nil(t, err)
//
//	node1, err := p2p.NewNetMessenger(
//		context.Background(),
//		&mock.MarshalizerMock{},
//		&mock.HasherMock{},
//		cp1,
//		10,
//		p2p.FloodSub)
//	assert.Nil(t, err)
//
//	node2, err := p2p.NewNetMessenger(
//		context.Background(),
//		&mock.MarshalizerMock{},
//		&mock.HasherMock{},
//		cp2,
//		10,
//		p2p.FloodSub)
//	assert.Nil(t, err)
//
//	node3, err := p2p.NewNetMessenger(
//		context.Background(),
//		&mock.MarshalizerMock{},
//		&mock.HasherMock{},
//		cp3,
//		10,
//		p2p.FloodSub)
//	assert.Nil(t, err)
//
//	cn := p2p.NewConnNotifier(node1)
//	cn.MaxAllowedPeers = 2
//
//	time.Sleep(time.Second)
//
//	node2.ConnectToAddresses(context.Background(), []string{node1.Addresses()[0]})
//	node3.ConnectToAddresses(context.Background(), []string{node1.Addresses()[0]})
//
//	time.Sleep(time.Second)
//
//	result := p2p.TaskResolveConnections(cn)
//	assert.Equal(t, p2p.OnlyInboundConnections, result)
//
//	time.Sleep(time.Second)
//
//	assert.Equal(t, 1, len(cn.Msgr.Conns()))
//}
//
//func TestTryToConnect3PeersWithSuccess(t *testing.T) {
//	cp1, err := p2p.NewConnectParamsFromPort(6000)
//	assert.Nil(t, err)
//
//	cp2, err := p2p.NewConnectParamsFromPort(6001)
//	assert.Nil(t, err)
//
//	cp3, err := p2p.NewConnectParamsFromPort(6002)
//	assert.Nil(t, err)
//
//	node1, err := p2p.NewNetMessenger(
//		context.Background(),
//		&mock.MarshalizerMock{},
//		&mock.HasherMock{},
//		cp1,
//		10,
//		p2p.FloodSub)
//	assert.Nil(t, err)
//
//	node2, err := p2p.NewNetMessenger(
//		context.Background(),
//		&mock.MarshalizerMock{},
//		&mock.HasherMock{},
//		cp2,
//		10,
//		p2p.FloodSub)
//	assert.Nil(t, err)
//
//	node3, err := p2p.NewNetMessenger(
//		context.Background(),
//		&mock.MarshalizerMock{},
//		&mock.HasherMock{},
//		cp3,
//		10,
//		p2p.FloodSub)
//	assert.Nil(t, err)
//
//	cn := p2p.NewConnNotifier(node1)
//	cn.MaxAllowedPeers = 4
//
//	addresses := []string{node2.Addresses()[0], node3.Addresses()[0]}
//
//	mutMapAddr := sync.Mutex{}
//	mapAddr := make(map[peer.ID]multiaddr.Multiaddr, 0)
//
//	for i := 0; i < len(addresses); i++ {
//		addr, err := ipfsaddr.ParseString(addresses[i])
//
//		if err != nil {
//			panic(err)
//		}
//
//		pinfo, err := peerstore.InfoFromP2pAddr(addr.Multiaddr())
//
//		if err != nil {
//			panic(err)
//		}
//
//		ma := pinfo.Addrs[0]
//
//		mapAddr[addr.ID()] = ma
//	}
//
//	cn.ConnectToPeer = func(cn *p2p.ConnNotifier, pid peer.ID) error {
//		mutMapAddr.Lock()
//		defer mutMapAddr.Unlock()
//
//		addr, ok := mapAddr[pid]
//
//		if !ok {
//			return errors.New("should have not entered here")
//		}
//
//		cn.Msgr.ConnectToAddresses(context.Background(), []string{addr.String() + "/ipfs/" + pid.Pretty()})
//
//		if err != nil {
//			return err
//		}
//
//		return nil
//	}
//
//	cn.GetKnownPeers = func(sender *p2p.ConnNotifier) []peer.ID {
//		return []peer.ID{node2.ID(), node3.ID()}
//	}
//
//	time.Sleep(time.Second)
//
//	result := p2p.TaskResolveConnections(cn)
//	assert.Equal(t, p2p.SuccessfullyConnected, result)
//	assert.Equal(t, 1, len(cn.Msgr.Conns()))
//
//	result = p2p.TaskResolveConnections(cn)
//	assert.Equal(t, p2p.SuccessfullyConnected, result)
//	assert.Equal(t, 2, len(cn.Msgr.Conns()))
//
//	result = p2p.TaskResolveConnections(cn)
//	assert.Equal(t, p2p.NothingDone, result)
//	assert.Equal(t, 2, len(cn.Msgr.Conns()))
//
//}
