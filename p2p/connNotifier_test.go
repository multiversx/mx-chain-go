package p2p_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/stretchr/testify/assert"
)

var testConnNotifierMaxWaitTime = time.Second * 5
var testConnNotifierWaitTimeForNoResponse = time.Second

func TestConnNotifierTaskResolveConnectionsNotAllowedToConnect(t *testing.T) {
	cn := p2p.NewConnNotifier(0)

	assert.Equal(t, p2p.WontConnect, cn.TaskResolveConnections())
}

func TestConnNotifierTaskResolveConnectionsOnlyInboundConnections(t *testing.T) {
	cn := p2p.NewConnNotifier(1)

	//will return 2 inbound connections
	cn.GetConnections = func(sender *p2p.ConnNotifier) []net.Conn {
		return []net.Conn{
			&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}},
			&mock.ConnMock{Status: net.Stat{Direction: net.DirInbound}}}
	}

	assert.Equal(t, p2p.OnlyInboundConnections, cn.TaskResolveConnections())
}

func TestConnNotifierTaskResolveConnectionsNothingDone(t *testing.T) {
	cn := p2p.NewConnNotifier(1)

	assert.Equal(t, p2p.NothingDone, cn.TaskResolveConnections())
}

func TestConnNotifierComputeInboundOutboundConns(t *testing.T) {
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

func TestConnNotifierTryToConnectWithSuccessRunning(t *testing.T) {
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

func TestConnNotifierTryToConnectWithSuccessOn2PeersRunning(t *testing.T) {
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

func TestConnNotifierTryToConnectNoOtherPeersRunning(t *testing.T) {
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

func TestConnNotifierConnectedGetConnNilShouldCloseConn(t *testing.T) {
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

func TestConnNotifierConnectedMaxConnReachedShouldCloseConn(t *testing.T) {
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

func TestConnNotifierConnectedCanAcceptShouldNotCloseConn(t *testing.T) {
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
