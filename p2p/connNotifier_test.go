package p2p_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
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
	counterCN01 := int32(0)

	fnc := func(caller interface{}) {
		atomic.AddInt32(&counterCN01, 1)

		time.Sleep(time.Second)
	}

	cn := p2p.NewConnNotifier(&p2p.MemoryMessenger{})

	cn.OnDoSimpleTask = fnc

	cn.Start()

	assert.Equal(t, execution.STARTED, cn.Stat())

	//wait 0.5 sec
	time.Sleep(time.Millisecond * 500)

	//counter CN01 should have been 1 by now, closing
	assert.Equal(t, int32(1), atomic.LoadInt32(&counterCN01))

	cn.Stop()
	//since go routine is still waiting, status should be CLOSING
	assert.Equal(t, execution.CLOSING, cn.Stat())
	//starting should not produce effects here
	cn.Start()
	assert.Equal(t, execution.CLOSING, cn.Stat())

	time.Sleep(time.Second)

	//it should have stopped
	assert.Equal(t, execution.CLOSED, cn.Stat())
}

func TestTaskNotDoingStuffOn0MaxPeers(t *testing.T) {
	cn := p2p.NewConnNotifier(&p2p.MemoryMessenger{})

	cn.MaxPeersAllowed = 0

	result := p2p.TaskMonitorConnections(cn)

	assert.Equal(t, 1, result)
}

func TestTryToConnect(t *testing.T) {
	mut := sync.Mutex{}
	lastString := ""

	node, err := p2p.NewNetMessenger(
		context.Background(),
		marshal.DefaultMarshalizer(),
		*p2p.NewConnectParamsFromPort(4000),
		[]string{},
		10)
	assert.Nil(t, err)

	cn := p2p.NewConnNotifier(node)

	cn.MaxPeersAllowed = 10
	cn.OnGetKnownPeers = func(cn *p2p.ConnNotifier) []peer.ID {
		return []peer.ID{"aaa", "bbb"}
	}

	cn.OnNeedToConnectToOtherPeer = func(cn *p2p.ConnNotifier, id peer.ID) error {
		mut.Lock()
		lastString = string(id)
		mut.Unlock()

		return nil
	}

	result := p2p.TaskMonitorConnections(cn)

	assert.Equal(t, 0, result)
	mut.Lock()
	assert.Equal(t, "aaa", lastString)
	mut.Unlock()

	result = p2p.TaskMonitorConnections(cn)

	assert.Equal(t, 0, result)
	mut.Lock()
	assert.Equal(t, "bbb", lastString)
	mut.Unlock()
}

func TestRemoveInboundPeers(t *testing.T) {
	node1, err := p2p.NewNetMessenger(
		context.Background(),
		marshal.DefaultMarshalizer(),
		*p2p.NewConnectParamsFromPort(5000),
		[]string{},
		10)
	assert.Nil(t, err)

	node2, err := p2p.NewNetMessenger(
		context.Background(),
		marshal.DefaultMarshalizer(),
		*p2p.NewConnectParamsFromPort(5001),
		[]string{},
		10)
	assert.Nil(t, err)

	node3, err := p2p.NewNetMessenger(
		context.Background(),
		marshal.DefaultMarshalizer(),
		*p2p.NewConnectParamsFromPort(5002),
		[]string{},
		10)
	assert.Nil(t, err)

	cn := p2p.NewConnNotifier(node1)
	cn.MaxPeersAllowed = 2

	time.Sleep(time.Second)

	node2.ConnectToAddresses(context.Background(), []string{node1.Addrs()[0]})
	node3.ConnectToAddresses(context.Background(), []string{node1.Addrs()[0]})

	time.Sleep(time.Second)

	result := p2p.TaskMonitorConnections(cn)
	assert.Equal(t, 2, result)

	time.Sleep(time.Second)

	assert.Equal(t, 1, len(cn.Mes.Conns()))
}

func TestTryToConnect2(t *testing.T) {
	node1, err := p2p.NewNetMessenger(
		context.Background(),
		marshal.DefaultMarshalizer(),
		*p2p.NewConnectParamsFromPort(6000),
		[]string{},
		10)
	assert.Nil(t, err)

	node2, err := p2p.NewNetMessenger(
		context.Background(),
		marshal.DefaultMarshalizer(),
		*p2p.NewConnectParamsFromPort(6001),
		[]string{},
		10)
	assert.Nil(t, err)

	node3, err := p2p.NewNetMessenger(
		context.Background(),
		marshal.DefaultMarshalizer(),
		*p2p.NewConnectParamsFromPort(6002),
		[]string{},
		10)
	assert.Nil(t, err)

	cn := p2p.NewConnNotifier(node1)
	cn.MaxPeersAllowed = 4

	addresses := []string{node2.Addrs()[0], node3.Addrs()[0]}

	mutMapAddr := sync.Mutex{}
	mapAddr := make(map[peer.ID]multiaddr.Multiaddr, 0)

	for i := 0; i < len(addresses); i++ {
		addr, err := ipfsaddr.ParseString(addresses[i])

		if err != nil {
			panic(err)
		}

		pinfo, err := pstore.InfoFromP2pAddr(addr.Multiaddr())

		if err != nil {
			panic(err)
		}

		ma := pinfo.Addrs[0]

		mapAddr[addr.ID()] = ma
	}

	cn.OnNeedToConnectToOtherPeer = func(cn *p2p.ConnNotifier, pid peer.ID) error {
		mutMapAddr.Lock()
		defer mutMapAddr.Unlock()

		addr, ok := mapAddr[pid]

		if !ok {
			return errors.New("Should not entered here!")
		}

		cn.Mes.ConnectToAddresses(context.Background(), []string{addr.String() + "/ipfs/" + pid.Pretty()})

		if err != nil {
			return err
		}

		return nil
	}

	cn.OnGetKnownPeers = func(sender *p2p.ConnNotifier) []peer.ID {
		return []peer.ID{node2.ID(), node3.ID()}
	}

	time.Sleep(time.Second)

	result := p2p.TaskMonitorConnections(cn)
	assert.Equal(t, 0, result)
	assert.Equal(t, 1, len(cn.Mes.Conns()))

	result = p2p.TaskMonitorConnections(cn)
	assert.Equal(t, 0, result)
	assert.Equal(t, 2, len(cn.Mes.Conns()))

	result = p2p.TaskMonitorConnections(cn)
	assert.Equal(t, 3, result)
	assert.Equal(t, 2, len(cn.Mes.Conns()))

}
