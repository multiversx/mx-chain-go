package p2p

import (
	"context"
	"github.com/ElrondNetwork/elrond-go-sandbox/service"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestShouldPanicOnNilNode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	NewConnNotifier(nil)
}

func TestStartingStoppingWorkingRoutine(t *testing.T) {
	counterCN01 := int32(0)

	fnc := func(cn *ConnNotifier) {
		atomic.AddInt32(&counterCN01, 1)

		time.Sleep(time.Second)
	}

	cn := NewConnNotifier(&Node{})

	cn.OnDoSimpleTask = fnc

	cn.Start()

	assert.Equal(t, STARTED, cn.Stat())

	//wait 0.5 sec
	time.Sleep(time.Millisecond * 500)

	//counter CN01 should have been 1 by now, closing
	assert.Equal(t, int32(1), atomic.LoadInt32(&counterCN01))

	cn.Stop()
	//since go routine is still waiting, status should be CLOSING
	assert.Equal(t, CLOSING, cn.Stat())
	//starting should not produce effects here
	cn.Start()
	assert.Equal(t, CLOSING, cn.Stat())

	time.Sleep(time.Second)

	//it should have stopped
	assert.Equal(t, CLOSED, cn.Stat())
}

func TestTaskNotDoingStuffOn0MaxPeers(t *testing.T) {
	cn := NewConnNotifier(&Node{})

	cn.MaxPeersAllowed = 0

	result := TaskMonitorConnections(cn)

	assert.Equal(t, 1, result)
}

func TestTryToConnect(t *testing.T) {
	mut := sync.Mutex{}
	lastString := ""

	node, err := NewNode(context.Background(), 4000, []string{}, service.GetMarshalizerService(), 10)
	assert.Nil(t, err)

	cn := NewConnNotifier(node)

	cn.MaxPeersAllowed = 10
	cn.OnGetKnownPeers = func(cn *ConnNotifier) []peer.ID {
		return []peer.ID{"aaa", "bbb"}
	}

	cn.OnNeedToConnectToOtherPeer = func(cn *ConnNotifier, id peer.ID) error {
		mut.Lock()
		lastString = string(id)
		mut.Unlock()

		return nil
	}

	result := TaskMonitorConnections(cn)

	assert.Equal(t, 0, result)
	mut.Lock()
	assert.Equal(t, "aaa", lastString)
	mut.Unlock()

	result = TaskMonitorConnections(cn)

	assert.Equal(t, 0, result)
	mut.Lock()
	assert.Equal(t, "bbb", lastString)
	mut.Unlock()
}

func TestRemoveInboundPeers(t *testing.T) {
	node1, err := NewNode(context.Background(), 5000, []string{}, service.GetMarshalizerService(), 10)
	assert.Nil(t, err)

	node2, err := NewNode(context.Background(), 5001, []string{}, service.GetMarshalizerService(), 10)
	assert.Nil(t, err)

	node3, err := NewNode(context.Background(), 5002, []string{}, service.GetMarshalizerService(), 10)
	assert.Nil(t, err)

	cn := NewConnNotifier(node1)
	cn.MaxPeersAllowed = 2

	time.Sleep(time.Second)

	strNode1 := node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()

	node2.ConnectToAddresses(context.Background(), []string{strNode1})
	node3.ConnectToAddresses(context.Background(), []string{strNode1})

	time.Sleep(time.Second)

	result := TaskMonitorConnections(cn)
	assert.Equal(t, 2, result)

	time.Sleep(time.Second)

	assert.Equal(t, 1, len(cn.node.P2pNode.Network().Conns()))
}

func TestTryToConnect2(t *testing.T) {
	node1, err := NewNode(context.Background(), 6000, []string{}, service.GetMarshalizerService(), 10)
	assert.Nil(t, err)

	node2, err := NewNode(context.Background(), 6001, []string{}, service.GetMarshalizerService(), 10)
	assert.Nil(t, err)

	node3, err := NewNode(context.Background(), 6002, []string{}, service.GetMarshalizerService(), 10)
	assert.Nil(t, err)

	cn := NewConnNotifier(node1)
	cn.MaxPeersAllowed = 4

	addresses := []string{node2.P2pNode.Addrs()[0].String() + "/ipfs/" + node2.P2pNode.ID().Pretty(),
		node3.P2pNode.Addrs()[0].String() + "/ipfs/" + node3.P2pNode.ID().Pretty()}

	for _, str := range addresses {
		addr, err := ipfsaddr.ParseString(str)

		if err != nil {
			panic(err)
		}

		pinfo, err := pstore.InfoFromP2pAddr(addr.Multiaddr())

		if err != nil {
			panic(err)
		}

		ma := pinfo.Addrs[0]

		node1.P2pNode.Peerstore().AddAddr(addr.ID(), ma, pstore.PermanentAddrTTL)
	}

	cn.OnNeedToConnectToOtherPeer = func(cn *ConnNotifier, pid peer.ID) error {
		pinfo := cn.node.P2pNode.Peerstore().PeerInfo(pid)

		if err := cn.node.P2pNode.Connect(context.Background(), pinfo); err != nil {
			return err
		} else {
			stream, err := cn.node.P2pNode.NewStream(context.Background(), pinfo.ID, "benchmark/nolimit/1.0.0.0")
			if err != nil {
				return err
			} else {
				cn.node.streamHandler(stream)
			}
		}

		return nil
	}

	cn.OnGetKnownPeers = func(sender *ConnNotifier) []peer.ID {
		return []peer.ID{node2.P2pNode.ID(), node3.P2pNode.ID()}
	}

	time.Sleep(time.Second)

	result := TaskMonitorConnections(cn)
	assert.Equal(t, 0, result)
	assert.Equal(t, 1, len(cn.node.P2pNode.Network().Conns()))

	result = TaskMonitorConnections(cn)
	assert.Equal(t, 0, result)
	assert.Equal(t, 2, len(cn.node.P2pNode.Network().Conns()))

	result = TaskMonitorConnections(cn)
	assert.Equal(t, 3, result)
	assert.Equal(t, 2, len(cn.node.P2pNode.Network().Conns()))

}
