package p2p

import (
	"context"
	"crypto"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/stretchr/testify/assert"
)

const (
	MEMORY int = iota
	NETWORK
)

var counter1 int32
var counter2 int32
var counter3 int32

var defaultMarshalizer marshal.Marshalizer

func getDefaultMarshlizer() marshal.Marshalizer {
	if defaultMarshalizer == nil {
		defaultMarshalizer = &marshal.JsonMarshalizer{}
	}

	return defaultMarshalizer
}

func TestSuite(t *testing.T) {
	Suite(t, MEMORY)
}

//func TestSimple(t *testing.T){
//	TestingBootstrap(t, MEMORY)
//}

func Suite(t *testing.T, mesType int) {
	TestingRecreationSameNode(t, mesType)
	TestingSimpleSend2NodesPingPong(t, mesType)
	TestingSimpleBroadcast5nodesInline(t, mesType)
	TestingSimpleBroadcast5nodesBeterConnected(t, mesType)
	TestingMessageHops(t, mesType)
	TestingSendingNilShouldReturnError(t, mesType)
	TestingMultipleErrorsOnBroadcasting(t, mesType)
	TestingCreateNodeWithNilMarshalizer(t, mesType)
	TestingBootstrap(t, mesType)
}

func createMessenger(mesType int, port int, maxAllowedPeers int, marsh marshal.Marshalizer) (Messenger, error) {
	switch mesType {
	case MEMORY:
		sha3 := crypto.SHA3_256.New()
		id := peer.ID(sha3.Sum([]byte("Name" + strconv.Itoa(port))))

		return NewMemoryMessenger(marsh, id, maxAllowedPeers)
	case NETWORK:
		return nil, nil
	default:
		panic("Type not defined!")
	}

}

func TestingRecreationSameNode(t *testing.T, mesType int) {
	fmt.Println()

	port := 4000

	node1, err := createMessenger(mesType, port, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node2, err := createMessenger(mesType, port, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	if node1.ID().Pretty() != node2.ID().Pretty() {
		t.Fatal("ID mismatch")
	}
}

func TestingSimpleSend2NodesPingPong(t *testing.T, mesType int) {
	fmt.Println()

	node1, err := createMessenger(mesType, 5000, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node2, err := createMessenger(mesType, 5001, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	time.Sleep(time.Second)

	node1.ConnectToAddresses(context.Background(), []string{node2.Addrs()[0].String()})

	fmt.Printf("Node 1 is %s\n", node1.Addrs()[0])
	fmt.Printf("Node 2 is %s\n", node2.Addrs()[0])

	var val int32 = 0

	node1.SetOnRecvMsg(func(caller Messenger, peerID string, m *Message) {
		fmt.Printf("%v: got message from peerID %v: %v\n", caller.ID().Pretty(), peerID, string(m.Payload))

		if string(m.Payload) == "Ping" {
			atomic.AddInt32(&val, 1)
			caller.SendDirectString(peerID, "Pong")
		}
	})

	node2.SetOnRecvMsg(func(caller Messenger, peerID string, m *Message) {
		fmt.Printf("%v: got message from peerID %v: %v\n", caller.ID().Pretty(), peerID, string(m.Payload))

		if string(m.Payload) == "Ping" {
			caller.SendDirectString(peerID, "Pong")
		}

		if string(m.Payload) == "Pong" {
			atomic.AddInt32(&val, 1)
		}
	})

	err = node2.SendDirectString(node1.ID().Pretty(), "Ping")

	assert.Nil(t, err)

	time.Sleep(time.Second)

	if atomic.LoadInt32(&val) != 2 {
		t.Fatal("Should have been 2 (ping-pong pair)")
	}

	node1.Close()
	node2.Close()
}

func recv1(caller Messenger, peerID string, m *Message) {
	atomic.AddInt32(&counter1, 1)
	fmt.Printf("%v > %v: Got message from peerID %v: %v\n", time.Now(), caller.ID().Pretty(), peerID, string(m.Payload))
	caller.BroadcastBuff(m.Payload, []string{peerID})
}

func recv2(caller Messenger, peerID string, m *Message) {
	atomic.AddInt32(&counter2, 1)
	fmt.Printf("%v > %v: Got message from peerID %v: %v\n", time.Now(), caller.ID().Pretty(), peerID, string(m.Payload))
	caller.BroadcastString(string(m.Payload), []string{peerID})
}

func TestingSimpleBroadcast5nodesInline(t *testing.T, mesType int) {
	fmt.Println()

	counter1 = 0

	node1, err := createMessenger(mesType, 6100, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node2, err := createMessenger(mesType, 6101, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node3, err := createMessenger(mesType, 6102, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node4, err := createMessenger(mesType, 6103, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node5, err := createMessenger(mesType, 6104, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	fmt.Printf("Node 1 is %s\n", node1.Addrs()[0])
	fmt.Printf("Node 2 is %s\n", node2.Addrs()[0])
	fmt.Printf("Node 3 is %s\n", node3.Addrs()[0])
	fmt.Printf("Node 4 is %s\n", node4.Addrs()[0])
	fmt.Printf("Node 5 is %s\n", node5.Addrs()[0])

	time.Sleep(time.Second)

	node2.ConnectToAddresses(context.Background(), []string{node1.Addrs()[0].String()})
	node3.ConnectToAddresses(context.Background(), []string{node2.Addrs()[0].String()})
	node4.ConnectToAddresses(context.Background(), []string{node3.Addrs()[0].String()})
	node5.ConnectToAddresses(context.Background(), []string{node4.Addrs()[0].String()})

	node1.SetOnRecvMsg(recv1)
	node2.SetOnRecvMsg(recv1)
	node3.SetOnRecvMsg(recv1)
	node4.SetOnRecvMsg(recv1)
	node5.SetOnRecvMsg(recv1)

	fmt.Println()
	fmt.Println()

	node1.BroadcastString("Boo", []string{})

	time.Sleep(time.Second)

	if atomic.LoadInt32(&counter1) != 4 {
		t.Fatal("Should have been 4 (traversed 4 peers)")
	}

	node1.Close()
	node2.Close()
	node3.Close()
	node4.Close()
	node5.Close()

}

func TestingSimpleBroadcast5nodesBeterConnected(t *testing.T, mesType int) {
	fmt.Println()

	counter2 = 0

	node1, err := createMessenger(mesType, 7000, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node2, err := createMessenger(mesType, 7001, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node3, err := createMessenger(mesType, 7002, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node4, err := createMessenger(mesType, 7003, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node5, err := createMessenger(mesType, 7004, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	fmt.Printf("Node 1 is %s\n", node1.Addrs()[0])
	fmt.Printf("Node 2 is %s\n", node2.Addrs()[0])
	fmt.Printf("Node 3 is %s\n", node3.Addrs()[0])
	fmt.Printf("Node 4 is %s\n", node4.Addrs()[0])
	fmt.Printf("Node 5 is %s\n", node5.Addrs()[0])

	time.Sleep(time.Second)

	node2.ConnectToAddresses(context.Background(), []string{node1.Addrs()[0].String()})
	node3.ConnectToAddresses(context.Background(), []string{node2.Addrs()[0].String(), node1.Addrs()[0].String()})
	node4.ConnectToAddresses(context.Background(), []string{node3.Addrs()[0].String()})
	node5.ConnectToAddresses(context.Background(), []string{node4.Addrs()[0].String(), node1.Addrs()[0].String()})

	time.Sleep(time.Second)

	node1.SetOnRecvMsg(recv2)
	node2.SetOnRecvMsg(recv2)
	node3.SetOnRecvMsg(recv2)
	node4.SetOnRecvMsg(recv2)
	node5.SetOnRecvMsg(recv2)

	fmt.Println()
	fmt.Println()

	node1.BroadcastString("Boo", []string{})

	time.Sleep(time.Second)

	if atomic.LoadInt32(&counter2) != 4 {
		t.Fatal("Should have been 4 (traversed 4 peers), got", counter2)
	}

	node1.Close()
	node2.Close()
	node3.Close()
	node4.Close()
	node5.Close()
}

func TestingMessageHops(t *testing.T, mesType int) {
	fmt.Println()

	marsh := &marshal.JsonMarshalizer{}

	node1, err := createMessenger(mesType, 8000, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	node2, err := createMessenger(mesType, 8001, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	fmt.Printf("Node 1 is %s\n", node1.Addrs()[0])
	fmt.Printf("Node 2 is %s\n", node2.Addrs()[0])

	time.Sleep(time.Second)

	node1.ConnectToAddresses(context.Background(), []string{node2.Addrs()[0].String()})

	m := NewMessage(node1.ID().Pretty(), []byte("A"), marsh)

	mut := sync.RWMutex{}
	var recv *Message = nil

	counter3 = 0

	node1.SetOnRecvMsg(func(caller Messenger, peerID string, m *Message) {

		if counter3 < 10 {
			atomic.AddInt32(&counter3, 1)

			fmt.Printf("Node 1, recv %v, resending...\n", *m)
			m.AddHop(caller.ID().Pretty())
			caller.BroadcastMessage(m, []string{})

			mut.Lock()
			recv = m
			mut.Unlock()
		}
	})

	node2.SetOnRecvMsg(func(caller Messenger, peerID string, m *Message) {

		if counter3 < 10 {
			atomic.AddInt32(&counter3, 1)

			fmt.Printf("Node 2, recv %v, resending...\n", *m)
			m.AddHop(caller.ID().Pretty())
			caller.BroadcastMessage(m, []string{})

			mut.Lock()
			recv = m
			mut.Unlock()
		}
	})

	node1.BroadcastMessage(m, []string{})

	time.Sleep(time.Second)

	if atomic.LoadInt32(&counter3) != 2 {
		t.Fatal(fmt.Sprintf("Shuld have been 2 iterations (messageQueue filtering), got %v!", counter3))
	}

	mut.RLock()
	if recv == nil {
		t.Fatal("Not broadcasted?")
	}

	if recv.Hops != 2 {
		t.Fatal("Hops should have been 2")
	}

	if recv.Peers[0] != node1.ID().Pretty() {
		t.Fatal("hop 1 should have been node's 1")
	}

	if recv.Peers[1] != node2.ID().Pretty() {
		t.Fatal("hop 2 should have been node's 2")
	}

	if recv.Peers[2] != node1.ID().Pretty() {
		t.Fatal("hop 3 should have been node's 1")
	}
	mut.RUnlock()

	node1.Close()
	node2.Close()

}

func TestingSendingNilShouldReturnError(t *testing.T, mesType int) {
	node1, err := createMessenger(mesType, 9000, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	err = node1.BroadcastMessage(nil, []string{})
	assert.NotNil(t, err)

	err = node1.SendDirectMessage("", nil)
	assert.NotNil(t, err)

}

func TestingMultipleErrorsOnBroadcasting(t *testing.T, mesType int) {
	node1, err := createMessenger(mesType, 10000, 10, getDefaultMarshlizer())
	assert.Nil(t, err)

	err = node1.BroadcastString("aaa", []string{})
	assert.NotNil(t, err)

	node1.AddAddr("A", node1.Addrs()[0], peerstore.PermanentAddrTTL)

	err = node1.BroadcastString("aaa", []string{})
	assert.NotNil(t, err)

	if len(err.(*NodeError).NestedErrors) != 0 {
		t.Fatal("Should have had 0 nested errs")
	}

	node1.AddAddr("B", node1.Addrs()[0], peerstore.PermanentAddrTTL)

	err = node1.BroadcastString("aaa", []string{"aaa", "bbbbb"})
	assert.NotNil(t, err)

	if len(err.(*NodeError).NestedErrors) != 2 {
		t.Fatal("Should have had 2 nested errs")
	}

}

func TestingCreateNodeWithNilMarshalizer(t *testing.T, mesType int) {
	_, err := createMessenger(mesType, 11000, 10, nil)

	assert.NotNil(t, err)
}

func TestingBootstrap(t *testing.T, mesType int) {
	startPort := 12000
	endPort := 12009
	nConns := 4

	nodes := make([]*Messenger, 0)

	recv := make(map[string]*Message)
	mut := sync.RWMutex{}

	mapHops := make(map[int]int)

	for i := startPort; i <= endPort; i++ {
		node, err := createMessenger(mesType, i, nConns, getDefaultMarshlizer())

		node.SetOnRecvMsg(func(caller Messenger, peerID string, m *Message) {

			m.AddHop(caller.ID().Pretty())

			mut.Lock()
			recv[caller.ID().Pretty()] = m
			mut.Unlock()

			caller.BroadcastMessage(m, []string{peerID})
		})

		assert.Nil(t, err)

		nodes = append(nodes, &node)
	}

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(len(nodes))

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]

		go func() {
			(*node).Bootstrap(context.Background())
			wg.Done()
		}()
	}

	wg.Wait()

	for i := 0; i < len(nodes); i++ {
		(*nodes[i]).PrintConnected()
		fmt.Println()
	}

	time.Sleep(time.Second)

	//broadcastind something
	fmt.Println("Broadcasting a message...")
	m := NewMessage((*nodes[0]).ID().Pretty(), []byte{65, 66, 67}, getDefaultMarshlizer())

	(*nodes[0]).BroadcastMessage(m, []string{})
	mut.Lock()
	recv[(*nodes[0]).ID().Pretty()] = m
	mut.Unlock()

	fmt.Println("Waiting...")
	time.Sleep(time.Second * 2)

	maxHops := 0

	notRecv := 0
	didRecv := 0

	for i := 0; i < len(nodes); i++ {

		mut.RLock()
		v, found := recv[(*nodes[i]).ID().Pretty()]
		mut.RUnlock()

		if !found {
			fmt.Printf("Peer %s didn't got the message!\n", (*nodes[i]).ID().Pretty())
			notRecv++
		} else {
			fmt.Printf("Peer %s got the message in %d hops!\n", (*nodes[i]).ID().Pretty(), v.Hops)
			didRecv++

			val, found := mapHops[v.Hops]
			if !found {
				mapHops[v.Hops] = 1
			} else {
				mapHops[v.Hops] = val + 1
			}

			if maxHops < v.Hops {
				maxHops = v.Hops
			}
		}
	}

	fmt.Println("Max hops:", maxHops)
	fmt.Print("Hops: ")

	for i := 0; i <= maxHops; i++ {
		fmt.Printf("\tH%d: %d", i, mapHops[i])
	}
	fmt.Println()

	fmt.Println("Did recv:", didRecv)
	fmt.Println("Did not recv:", notRecv)

	assert.Equal(t, notRecv, 0)
}
