package p2p_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestShouldPanicOnNilNode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	p2p.NewConnNotifier(nil)
}

func TestStartingStoppingWorkingRoutine(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	counterCN01 := int32(0)

	cn := p2p.NewConnNotifier(&p2p.MemoryMessenger{})
	cn.DurCalls = 0
	cn.OnDoSimpleTask = func(caller interface{}) {
		atomic.AddInt32(&counterCN01, 1)

		time.Sleep(time.Second)
	}

	cn.Start()

	assert.Equal(t, execution.Started, cn.Stat())

	//wait 0.5 sec
	time.Sleep(time.Millisecond * 500)

	//counter CN01 should have been 1 by now, closing
	assert.Equal(t, int32(1), atomic.LoadInt32(&counterCN01))

	cn.Stop()
	//since go routine is still waiting, status should be CLOSING
	assert.Equal(t, execution.Closing, cn.Stat())
	//starting should not produce effects here
	cn.Start()
	assert.Equal(t, execution.Closing, cn.Stat())

	time.Sleep(time.Second)

	//it should have stopped
	assert.Equal(t, execution.Closed, cn.Stat())
}

func TestTaskNotDoingStuffOn0MaxPeers(t *testing.T) {
	cn := p2p.NewConnNotifier(&p2p.MemoryMessenger{})

	cn.MaxAllowedPeers = 0

	result := p2p.TaskResolveConnections(cn)

	assert.Equal(t, p2p.WontConnect, result)
}

func TestTryToConnectWithSuccess(t *testing.T) {
	mut := sync.Mutex{}
	lastString := ""

	cp, err := p2p.NewConnectParamsFromPort(4000)
	assert.Nil(t, err)

	node, err := p2p.NewNetMessenger(
		context.Background(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		cp,
		10)
	assert.Nil(t, err)

	cn := p2p.NewConnNotifier(node)

	cn.MaxAllowedPeers = 10
	cn.GetKnownPeers = func(cn *p2p.ConnNotifier) []peer.ID {
		return []peer.ID{"aaa", "bbb"}
	}

	cn.ConnectToPeer = func(cn *p2p.ConnNotifier, id peer.ID) error {
		mut.Lock()
		lastString = string(id)
		mut.Unlock()

		return nil
	}

	result := p2p.TaskResolveConnections(cn)

	assert.Equal(t, p2p.SuccessfullyConnected, result)
	mut.Lock()
	assert.Equal(t, "aaa", lastString)
	mut.Unlock()

	result = p2p.TaskResolveConnections(cn)

	assert.Equal(t, p2p.SuccessfullyConnected, result)
	mut.Lock()
	assert.Equal(t, "bbb", lastString)
	mut.Unlock()
}

func TestRemoveInboundPeers(t *testing.T) {
	//steps:
	// - create 3 nodes
	// - node 1 has a limit of 2 peers to connect to
	// - node 2 and node 3 connects to node 1
	// - run the function that checks the connections and the function will determine that there are too many inbound
	//   connections, will close the first connection and will return the status OnlyInboundConnections
	cp1, err := p2p.NewConnectParamsFromPort(5000)
	assert.Nil(t, err)

	cp2, err := p2p.NewConnectParamsFromPort(5001)
	assert.Nil(t, err)

	cp3, err := p2p.NewConnectParamsFromPort(5002)
	assert.Nil(t, err)

	node1, err := p2p.NewNetMessenger(
		context.Background(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		cp1,
		10)
	assert.Nil(t, err)

	node2, err := p2p.NewNetMessenger(
		context.Background(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		cp2,
		10)
	assert.Nil(t, err)

	node3, err := p2p.NewNetMessenger(
		context.Background(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		cp3,
		10)
	assert.Nil(t, err)

	cn := p2p.NewConnNotifier(node1)
	cn.MaxAllowedPeers = 2

	time.Sleep(time.Second)

	node2.ConnectToAddresses(context.Background(), []string{node1.Addrs()[0]})
	node3.ConnectToAddresses(context.Background(), []string{node1.Addrs()[0]})

	time.Sleep(time.Second)

	result := p2p.TaskResolveConnections(cn)
	assert.Equal(t, p2p.OnlyInboundConnections, result)

	time.Sleep(time.Second)

	assert.Equal(t, 1, len(cn.Msgr.Conns()))
}

func TestTryToConnect3PeersWithSuccess(t *testing.T) {
	cp1, err := p2p.NewConnectParamsFromPort(6000)
	assert.Nil(t, err)

	cp2, err := p2p.NewConnectParamsFromPort(6001)
	assert.Nil(t, err)

	cp3, err := p2p.NewConnectParamsFromPort(6002)
	assert.Nil(t, err)

	node1, err := p2p.NewNetMessenger(
		context.Background(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		cp1,
		10)
	assert.Nil(t, err)

	node2, err := p2p.NewNetMessenger(
		context.Background(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		cp2,
		10)
	assert.Nil(t, err)

	node3, err := p2p.NewNetMessenger(
		context.Background(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		cp3,
		10)
	assert.Nil(t, err)

	cn := p2p.NewConnNotifier(node1)
	cn.MaxAllowedPeers = 4

	addresses := []string{node2.Addrs()[0], node3.Addrs()[0]}

	mutMapAddr := sync.Mutex{}
	mapAddr := make(map[peer.ID]multiaddr.Multiaddr, 0)

	for i := 0; i < len(addresses); i++ {
		addr, err := ipfsaddr.ParseString(addresses[i])

		if err != nil {
			panic(err)
		}

		pinfo, err := peerstore.InfoFromP2pAddr(addr.Multiaddr())

		if err != nil {
			panic(err)
		}

		ma := pinfo.Addrs[0]

		mapAddr[addr.ID()] = ma
	}

	cn.ConnectToPeer = func(cn *p2p.ConnNotifier, pid peer.ID) error {
		mutMapAddr.Lock()
		defer mutMapAddr.Unlock()

		addr, ok := mapAddr[pid]

		if !ok {
			return errors.New("should have not entered here")
		}

		cn.Msgr.ConnectToAddresses(context.Background(), []string{addr.String() + "/ipfs/" + pid.Pretty()})

		if err != nil {
			return err
		}

		return nil
	}

	cn.GetKnownPeers = func(sender *p2p.ConnNotifier) []peer.ID {
		return []peer.ID{node2.ID(), node3.ID()}
	}

	time.Sleep(time.Second)

	result := p2p.TaskResolveConnections(cn)
	assert.Equal(t, p2p.SuccessfullyConnected, result)
	assert.Equal(t, 1, len(cn.Msgr.Conns()))

	result = p2p.TaskResolveConnections(cn)
	assert.Equal(t, p2p.SuccessfullyConnected, result)
	assert.Equal(t, 2, len(cn.Msgr.Conns()))

	result = p2p.TaskResolveConnections(cn)
	assert.Equal(t, p2p.NothingDone, result)
	assert.Equal(t, 2, len(cn.Msgr.Conns()))

}
