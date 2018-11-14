package p2p_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/libp2p/go-libp2p-net"
	"github.com/stretchr/testify/assert"
)

var testMemoryMarshalizer = &mock.MarshalizerMock{}
var testMemoryHasher = &mock.HasherMock{}

func createMemoryMessenger(t *testing.T, port int, nConns int) (*p2p.MemoryMessenger, error) {
	cp, err := p2p.NewConnectParamsFromPort(port)
	assert.Nil(t, err)

	mm, err := p2p.NewMemoryMessenger(testMemoryMarshalizer, testMemoryHasher, cp.ID, nConns)
	if err != nil {
		return nil, err
	}
	mm.PrivKey = cp.PrivKey
	return mm, nil
}

func TestMemoryMessenger_RecreationSameNode_ShouldWork(t *testing.T) {
	fmt.Println()

	port := 4000

	node1, err := createMemoryMessenger(t, port, 10)
	assert.Nil(t, err)

	node2, err := createMemoryMessenger(t, port, 10)
	assert.Nil(t, err)

	if node1.ID().Pretty() != node2.ID().Pretty() {
		t.Fatal("ID mismatch")
	}
}

func TestMemoryMessenger_SendToSelf_ShouldWork(t *testing.T) {
	node, err := createMemoryMessenger(t, 4500, 10)
	assert.Nil(t, err)

	var counter int32

	node.AddTopic(p2p.NewTopic("test topic", &testStringNewer{}, testMemoryMarshalizer))
	node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testStringNewer)).Data

		fmt.Printf("Got message: %v\n", payload)

		if payload == "ABC" {
			atomic.AddInt32(&counter, 1)
		}
	})

	node.GetTopic("test topic").Broadcast(&testStringNewer{Data: "ABC"})

	time.Sleep(time.Second)

	if atomic.LoadInt32(&counter) != int32(1) {
		assert.Fail(t, "Should have been 1 (message received to self)")
	}

}

func TestMemoryMessenger_NodesPingPongOn2Topics_ShouldWork(t *testing.T) {
	fmt.Println()

	node1, err := createMemoryMessenger(t, 5100, 10)
	assert.Nil(t, err)

	node2, err := createMemoryMessenger(t, 5101, 10)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	node1.ConnectToAddresses(context.Background(), []string{node2.Addrs()[0]})

	time.Sleep(time.Second)

	assert.Equal(t, net.Connected, node1.Connectedness(node2.ID()))
	assert.Equal(t, net.Connected, node2.Connectedness(node1.ID()))

	fmt.Printf("Node 1 is %s\n", node1.Addrs()[0])
	fmt.Printf("Node 2 is %s\n", node2.Addrs()[0])

	fmt.Printf("Node 1 has the addresses: %v\n", node1.Addrs())
	fmt.Printf("Node 2 has the addresses: %v\n", node2.Addrs())

	var val int32 = 0

	//create 2 topics on each node
	node1.AddTopic(p2p.NewTopic("ping", &testStringNewer{}, testMemoryMarshalizer))
	node1.AddTopic(p2p.NewTopic("pong", &testStringNewer{}, testMemoryMarshalizer))

	node2.AddTopic(p2p.NewTopic("ping", &testStringNewer{}, testMemoryMarshalizer))
	node2.AddTopic(p2p.NewTopic("pong", &testStringNewer{}, testMemoryMarshalizer))

	//assign some event handlers on topics
	node1.GetTopic("ping").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testStringNewer)).Data

		if payload == "ping string" {
			node1.GetTopic("pong").Broadcast(&testStringNewer{Data: "pong string"})
		}
	})

	node1.GetTopic("pong").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testStringNewer)).Data

		fmt.Printf("node1 received: %v\n", payload)

		if payload == "pong string" {
			atomic.AddInt32(&val, 1)
		}
	})

	//for node2 topic ping we do not need an event handler in this test
	node2.GetTopic("pong").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testStringNewer)).Data

		fmt.Printf("node2 received: %v\n", payload)

		if payload == "pong string" {
			atomic.AddInt32(&val, 1)
		}
	})

	node2.GetTopic("ping").Broadcast(&testStringNewer{Data: "ping string"})

	assert.Nil(t, err)

	time.Sleep(time.Second)

	if atomic.LoadInt32(&val) != 2 {
		t.Fatal("Should have been 2 (pong from node1: self and node2: received from node1)")
	}

	node1.Close()
	node2.Close()
}

func TestMemoryMessenger_SimpleBroadcast5nodesInline_ShouldWork(t *testing.T) {
	fmt.Println()

	nodes := make([]*p2p.MemoryMessenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createMemoryMessenger(t, 6100+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addrs()[0])
	}

	//connect one with each other daisy-chain
	for i := 1; i < 5; i++ {
		node := nodes[i]
		node.ConnectToAddresses(context.Background(), []string{nodes[i-1].Addrs()[0]})
	}

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(5)
	done := make(chan bool, 0)

	go func() {
		defer close(done)

		wg.Wait()
	}()

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		wg.Done()
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		node.AddTopic(p2p.NewTopic("test", &testStringNewer{}, testMemoryMarshalizer))
		node.GetTopic("test").AddDataReceived(recv)
	}

	fmt.Println()
	fmt.Println()

	fmt.Println("Broadcasting...")
	nodes[0].GetTopic("test").Broadcast(&testStringNewer{Data: "Foo"})

	select {
	case <-done:
		fmt.Println("Got all messages!")
	case <-time.After(time.Second):
		assert.Fail(t, "not all messages were received")
	}

	//closing
	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}
}

func TestMemoryMessenger_SimpleBroadcast5nodesBetterConnected_ShouldWork(t *testing.T) {
	fmt.Println()

	nodes := make([]*p2p.MemoryMessenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createMemoryMessenger(t, 7000+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addrs()[0])
	}

	//connect one with each other manually
	// node0 --------- node1
	//   |               |
	//   +------------ node2
	//   |               |
	//   |             node3
	//   |               |
	//   +------------ node4

	nodes[1].ConnectToAddresses(context.Background(), []string{nodes[0].Addrs()[0]})
	nodes[2].ConnectToAddresses(context.Background(), []string{nodes[1].Addrs()[0], nodes[0].Addrs()[0]})
	nodes[3].ConnectToAddresses(context.Background(), []string{nodes[2].Addrs()[0]})
	nodes[4].ConnectToAddresses(context.Background(), []string{nodes[3].Addrs()[0], nodes[0].Addrs()[0]})

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(5)
	done := make(chan bool, 0)

	go func() {
		defer close(done)

		wg.Wait()
	}()

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		wg.Done()
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		node.AddTopic(p2p.NewTopic("test", &testStringNewer{}, testMemoryMarshalizer))
		node.GetTopic("test").AddDataReceived(recv)
	}

	fmt.Println()
	fmt.Println()

	fmt.Println("Broadcasting...")
	nodes[0].GetTopic("test").Broadcast(&testStringNewer{Data: "Foo"})

	select {
	case <-done:
		fmt.Println("Got all messages!")
	case <-time.After(time.Second):
		assert.Fail(t, "not all messages were received")
	}

	//closing
	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}
}

func TestMemoryMessenger_SendingNil_ShouldErr(t *testing.T) {
	node1, err := createMemoryMessenger(t, 9000, 10)
	assert.Nil(t, err)

	node1.AddTopic(p2p.NewTopic("test", &testStringNewer{}, testMemoryMarshalizer))
	err = node1.GetTopic("test").Broadcast(nil)
	assert.NotNil(t, err)
}

func TestMemoryMessenger_CreateNodeWithNilMarshalizer_ShouldErr(t *testing.T) {
	cp, err := p2p.NewConnectParamsFromPort(11000)
	assert.Nil(t, err)

	_, err = p2p.NewMemoryMessenger(nil, testMemoryHasher, cp.ID, 10)

	assert.NotNil(t, err)
}

func TestMemoryMessenger_CreateNodeWithNilHasher_ShouldErr(t *testing.T) {
	cp, err := p2p.NewConnectParamsFromPort(12000)
	assert.Nil(t, err)

	_, err = p2p.NewMemoryMessenger(testMemoryMarshalizer, nil, cp.ID, 10)

	assert.NotNil(t, err)
}

func TestMemoryMessenger_SingleRoundBootstrap_ShouldNotProduceLonelyNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	startPort := 12000
	endPort := 12009
	nConns := 4

	nodes := make([]p2p.Messenger, 0)

	recv := make(map[string]*p2p.MessageInfo)
	mut := sync.RWMutex{}

	wgGot := sync.WaitGroup{}

	//prepare messengers
	for i := startPort; i <= endPort; i++ {
		node, err := createMemoryMessenger(t, i, nConns)

		err = node.AddTopic(p2p.NewTopic("test topic", &testStringNewer{}, testMemoryMarshalizer))
		assert.Nil(t, err)

		node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
			mut.Lock()
			recv[node.ID().Pretty()] = msgInfo

			fmt.Printf("%v got message: %v\n", node.ID().Pretty(), (*data.(*testStringNewer)).Data)

			wgGot.Done()

			mut.Unlock()

		})

		nodes = append(nodes, node)
	}

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(len(nodes))

	//call bootstrap to connect with each other
	for i := 0; i < len(nodes); i++ {
		node := nodes[i]

		node.Bootstrap(context.Background())
	}

	time.Sleep(time.Second * 10)

	for i := 0; i < len(nodes); i++ {
		nodes[i].PrintConnected()
		fmt.Println()
	}

	time.Sleep(time.Second)

	wgGot.Add(len(nodes))
	chanGot := make(chan bool)

	go func() {
		defer close(chanGot)

		wgGot.Wait()
	}()

	//broadcasting something
	fmt.Println("Broadcasting a message...")
	nodes[0].GetTopic("test topic").Broadcast(&testStringNewer{Data: "a string to broadcast"})

	fmt.Println("Waiting...")

	//waiting to complete or timeout
	select {
	case <-chanGot:
		break
	case <-time.After(time.Second * 2):
		assert.Fail(t, "Not all nodes received the message!")
		return
	}

	notRecv := 0
	didRecv := 0

	for i := 0; i < len(nodes); i++ {

		mut.RLock()
		_, found := recv[nodes[i].ID().Pretty()]
		mut.RUnlock()

		if !found {
			fmt.Printf("Peer %s didn't got the message!\n", nodes[i].ID().Pretty())
			notRecv++
		} else {
			didRecv++
		}
	}

	fmt.Println()
	fmt.Println("Did recv:", didRecv)
	fmt.Println("Did not recv:", notRecv)

	assert.Equal(t, notRecv, 0)
}

type structTest1 struct {
	Nonce int
	Data  float64
}

func (s1 *structTest1) New() p2p.Newer {
	return &structTest1{}
}

type structTest2 struct {
	Nonce string
	Data  []byte
}

func (s2 *structTest2) New() p2p.Newer {
	return &structTest2{}
}

func TestMemoryMessenger_BadObjectToUnmarshal_ShouldFilteredOut(t *testing.T) {
	//stress test to check if the node is able to cope
	//with unmarshaling a bad object
	//both structs have the same fields but incompatible types

	//node1 registers topic 'test' with struct1
	//node2 registers topic 'test' with struct2

	node1, err := createMemoryMessenger(t, 13000, 10)
	assert.Nil(t, err)

	node2, err := createMemoryMessenger(t, 13001, 10)
	assert.Nil(t, err)

	//connect nodes
	node1.ConnectToAddresses(context.Background(), []string{node2.Addrs()[0]})

	//wait a bit
	time.Sleep(time.Second)

	//create topics for each node
	node1.AddTopic(p2p.NewTopic("test", &structTest1{}, testMemoryMarshalizer))
	node2.AddTopic(p2p.NewTopic("test", &structTest2{}, testMemoryMarshalizer))

	counter := int32(0)

	//node 1 sends, node 2 receives
	node2.GetTopic("test").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		fmt.Printf("received: %v", data)
		atomic.AddInt32(&counter, 1)
	})

	node1.GetTopic("test").Broadcast(&structTest1{Nonce: 4, Data: 4.5})

	//wait a bit
	time.Sleep(time.Second)

	//check that the message was filtered out
	assert.Equal(t, int32(0), atomic.LoadInt32(&counter))
}

func TestMemoryMessenger_BroadcastOnInexistentTopic_ShouldFilteredOut(t *testing.T) {
	//stress test to check if the node is able to cope
	//with receiving on an inexistent topic

	node1, err := createMemoryMessenger(t, 14000, 10)
	assert.Nil(t, err)

	node2, err := createMemoryMessenger(t, 14001, 10)
	assert.Nil(t, err)

	//connect nodes
	node1.ConnectToAddresses(context.Background(), []string{node2.Addrs()[0]})

	//wait a bit
	time.Sleep(time.Second)

	//create topics for each node
	node1.AddTopic(p2p.NewTopic("test1", &testStringNewer{}, testMemoryMarshalizer))
	node2.AddTopic(p2p.NewTopic("test2", &testStringNewer{}, testMemoryMarshalizer))

	counter := int32(0)

	//node 1 sends, node 2 receives
	node2.GetTopic("test2").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		fmt.Printf("received: %v", data)
		atomic.AddInt32(&counter, 1)
	})

	node1.GetTopic("test1").Broadcast(&testStringNewer{Data: "Foo"})

	//wait a bit
	time.Sleep(time.Second)

	//check that the message was filtered out
	assert.Equal(t, int32(0), atomic.LoadInt32(&counter))
}

func TestMemoryMessenger_MultipleRoundBootstrap_ShouldNotProduceLonelyNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	startPort := 12000
	endPort := 12009
	nConns := 4

	nodes := make([]p2p.Messenger, 0)

	recv := make(map[string]*p2p.MessageInfo)
	mut := sync.RWMutex{}

	wgGot := sync.WaitGroup{}

	//prepare messengers
	for i := startPort; i <= endPort; i++ {
		node, err := createMemoryMessenger(t, i, nConns)

		err = node.AddTopic(p2p.NewTopic("test topic", &testStringNewer{}, testMemoryMarshalizer))
		assert.Nil(t, err)

		node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
			mut.Lock()
			recv[node.ID().Pretty()] = msgInfo

			fmt.Printf("%v got message: %v\n", node.ID().Pretty(), (*data.(*testStringNewer)).Data)

			wgGot.Done()

			mut.Unlock()

		})

		nodes = append(nodes, node)
	}

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(len(nodes))

	//call bootstrap to connect with each other only on n - 2 nodes
	for i := 0; i < len(nodes)-2; i++ {
		node := nodes[i]

		node.Bootstrap(context.Background())
	}

	fmt.Println("Bootstrapping round 1...")
	time.Sleep(time.Second * 10)

	for i := 0; i < len(nodes); i++ {
		nodes[i].PrintConnected()
		fmt.Println()
	}

	time.Sleep(time.Second)

	//second round bootstrap for the last 2 nodes
	for i := len(nodes) - 2; i < len(nodes); i++ {
		node := nodes[i]

		node.Bootstrap(context.Background())
	}

	fmt.Println("Bootstrapping round 2...")
	time.Sleep(time.Second * 10)

	for i := 0; i < len(nodes); i++ {
		nodes[i].PrintConnected()
		fmt.Println()
	}

	time.Sleep(time.Second)

	wgGot.Add(len(nodes))
	chanGot := make(chan bool)

	go func() {
		defer close(chanGot)

		wgGot.Wait()
	}()

	//broadcasting something
	fmt.Println("Broadcasting a message...")
	nodes[0].GetTopic("test topic").Broadcast(&testStringNewer{Data: "a string to broadcast"})

	fmt.Println("Waiting...")

	//waiting to complete or timeout
	select {
	case <-chanGot:
		break
	case <-time.After(time.Second * 2):
		assert.Fail(t, "Not all nodes received the message!")
		return
	}

	notRecv := 0
	didRecv := 0

	for i := 0; i < len(nodes); i++ {

		mut.RLock()
		_, found := recv[nodes[i].ID().Pretty()]
		mut.RUnlock()

		if !found {
			fmt.Printf("Peer %s didn't got the message!\n", nodes[i].ID().Pretty())
			notRecv++
		} else {
			didRecv++
		}
	}

	fmt.Println()
	fmt.Println("Did recv:", didRecv)
	fmt.Println("Did not recv:", notRecv)

	assert.Equal(t, notRecv, 0)
}
