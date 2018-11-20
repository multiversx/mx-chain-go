package p2p_test

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

var testMemMarshalizer = &mock.MarshalizerMock{}
var testMemHasher = &mock.HasherMock{}

type testMemStringCloner struct {
	Data string
}

// Clone will return a new instance of string. Dummy, just to implement Newer interface as strings are immutable
func (sc *testMemStringCloner) New() p2p.Newer {
	return &testMemStringCloner{}
}

// ID will return the same string as ID
func (sc *testMemStringCloner) ID() string {
	return sc.Data
}

type structMemTest1 struct {
	Nonce int
	Data  float64
}

func (s1 *structMemTest1) New() p2p.Newer {
	return &structMemTest1{}
}

func (s1 *structMemTest1) ID() string {
	return strconv.Itoa(s1.Nonce)
}

type structMemTest2 struct {
	Nonce string
	Data  []byte
}

func (s2 *structMemTest2) New() p2p.Newer {
	return &structMemTest2{}
}

func (s2 *structMemTest2) ID() string {
	return s2.Nonce
}

func createMemMessenger(t *testing.T, port int) (*p2p.MemMessenger, error) {
	cp, err := p2p.NewConnectParamsFromPort(port)
	assert.Nil(t, err)

	mm, err := p2p.NewMemMessenger(testMemMarshalizer, testMemHasher, cp)
	if err != nil {
		return nil, err
	}

	return mm, nil
}

func resetMemMessengers() {
	(*p2p.MutGloballyRegPeers).Lock()
	defer (*p2p.MutGloballyRegPeers).Unlock()

	p2p.GloballyRegisteredPeers = make(map[peer.ID]*p2p.MemMessenger)
}

func TestMemMessenger_RecreationSameNode_ShouldWork(t *testing.T) {
	fmt.Println()

	resetMemMessengers()

	port := 4000

	node1, err := createMemMessenger(t, port)
	assert.Nil(t, err)

	node2, err := createMemMessenger(t, port)
	assert.Nil(t, err)

	if node1.ID().Pretty() != node2.ID().Pretty() {
		t.Fatal("ID mismatch")
	}
}

func TestMemMessenger_SendToSelf_ShouldWork(t *testing.T) {
	resetMemMessengers()

	node, err := createMemMessenger(t, 4500)
	assert.Nil(t, err)

	var counter int32

	node.AddTopic(p2p.NewTopic("test topic", &testMemStringCloner{}, testMemMarshalizer))
	node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testMemStringCloner)).Data

		fmt.Printf("Got message: %v\n", payload)

		if payload == "ABC" {
			atomic.AddInt32(&counter, 1)
		}
	})

	node.GetTopic("test topic").Broadcast(testMemStringCloner{Data: "ABC"})

	time.Sleep(time.Second)

	if atomic.LoadInt32(&counter) != int32(1) {
		assert.Fail(t, "Should have been 1 (message received to self)")
	}

}

func TestMemMessenger_NodesPingPongOn2Topics_ShouldWork(t *testing.T) {
	fmt.Println()

	resetMemMessengers()

	node1, err := createMemMessenger(t, 5100)
	assert.Nil(t, err)

	node2, err := createMemMessenger(t, 5101)
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
	node1.AddTopic(p2p.NewTopic("ping", &testMemStringCloner{}, testMemMarshalizer))
	node1.AddTopic(p2p.NewTopic("pong", &testMemStringCloner{}, testMemMarshalizer))

	node2.AddTopic(p2p.NewTopic("ping", &testMemStringCloner{}, testMemMarshalizer))
	node2.AddTopic(p2p.NewTopic("pong", &testMemStringCloner{}, testMemMarshalizer))

	//assign some event handlers on topics
	node1.GetTopic("ping").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testMemStringCloner)).Data

		if payload == "ping string" {
			node1.GetTopic("pong").Broadcast(testMemStringCloner{"pong string"})
		}
	})

	node1.GetTopic("pong").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testMemStringCloner)).Data

		fmt.Printf("node1 received: %v\n", payload)

		if payload == "pong string" {
			atomic.AddInt32(&val, 1)
		}
	})

	//for node2 topic ping we do not need an event handler in this test
	node2.GetTopic("pong").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testMemStringCloner)).Data

		fmt.Printf("node2 received: %v\n", payload)

		if payload == "pong string" {
			atomic.AddInt32(&val, 1)
		}
	})

	node2.GetTopic("ping").Broadcast(testMemStringCloner{"ping string"})

	assert.Nil(t, err)

	time.Sleep(time.Second)

	if atomic.LoadInt32(&val) != 2 {
		t.Fatal("Should have been 2 (pong from node1: self and node2: received from node1)")
	}

	node1.Close()
	node2.Close()
}

func TestMemMessenger_SimpleBroadcast5nodesInline_ShouldWork(t *testing.T) {
	fmt.Println()

	resetMemMessengers()

	nodes := make([]*p2p.MemMessenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createMemMessenger(t, 6100+i)
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

		node.AddTopic(p2p.NewTopic("test", &testMemStringCloner{}, testMemMarshalizer))
		node.GetTopic("test").AddDataReceived(recv)
	}

	fmt.Println()
	fmt.Println()

	fmt.Println("Broadcasting...")
	nodes[0].GetTopic("test").Broadcast(testMemStringCloner{Data: "Foo"})

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

func TestMemMessenger_SimpleBroadcast5nodesBetterConnected_ShouldWork(t *testing.T) {
	fmt.Println()

	resetMemMessengers()

	nodes := make([]*p2p.MemMessenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createMemMessenger(t, 7000+i)
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

		node.AddTopic(p2p.NewTopic("test", &testMemStringCloner{}, testMemMarshalizer))
		node.GetTopic("test").AddDataReceived(recv)
	}

	fmt.Println()
	fmt.Println()

	fmt.Println("Broadcasting...")
	nodes[0].GetTopic("test").Broadcast(testMemStringCloner{Data: "Foo"})

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

func TestMemMessenger_SendingNil_ShouldErr(t *testing.T) {
	resetMemMessengers()

	node1, err := createMemMessenger(t, 9000)
	assert.Nil(t, err)

	node1.AddTopic(p2p.NewTopic("test", &testMemStringCloner{}, testMemMarshalizer))
	err = node1.GetTopic("test").Broadcast(nil)
	assert.NotNil(t, err)
}

func TestMemMessenger_CreateNodeWithNilMarshalizer_ShouldErr(t *testing.T) {
	resetMemMessengers()

	cp, err := p2p.NewConnectParamsFromPort(11000)
	assert.Nil(t, err)

	_, err = p2p.NewMemMessenger(nil, testMemHasher, cp)

	assert.NotNil(t, err)
}

func TestMemMessenger_CreateNodeWithNilHasher_ShouldErr(t *testing.T) {
	resetMemMessengers()

	cp, err := p2p.NewConnectParamsFromPort(12000)
	assert.Nil(t, err)

	_, err = p2p.NewMemMessenger(testMemMarshalizer, nil, cp)

	assert.NotNil(t, err)
}

func TestMemMessenger_SingleRoundBootstrap_ShouldNotProduceLonelyNodes(t *testing.T) {
	resetMemMessengers()

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	startPort := 12000
	endPort := 12009

	nodes := make([]p2p.Messenger, 0)

	recv := make(map[string]*p2p.MessageInfo)
	mut := sync.RWMutex{}

	//prepare messengers
	for i := startPort; i <= endPort; i++ {
		node, err := createMemMessenger(t, i)

		err = node.AddTopic(p2p.NewTopic("test topic", &testMemStringCloner{}, testMemMarshalizer))
		assert.Nil(t, err)

		node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
			mut.Lock()
			recv[node.ID().Pretty()] = msgInfo

			fmt.Printf("%v got message: %v\n", node.ID().Pretty(), (*data.(*testMemStringCloner)).Data)

			mut.Unlock()

		})

		nodes = append(nodes, node)
	}

	time.Sleep(time.Second)

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

	//broadcasting something
	fmt.Println("Broadcasting a message...")
	nodes[0].GetTopic("test topic").Broadcast(testMemStringCloner{"a string to broadcast"})

	fmt.Println("Waiting...")

	//waiting 2 seconds
	time.Sleep(time.Second * 2)

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

	assert.Equal(t, 0, notRecv)
}

func TestMemMessenger_BadObjectToUnmarshal_ShouldFilteredOut(t *testing.T) {
	//stress test to check if the node is able to cope
	//with unmarshaling a bad object
	//both structs have the same fields but incompatible types

	//node1 registers topic 'test' with struct1
	//node2 registers topic 'test' with struct2

	resetMemMessengers()

	node1, err := createMemMessenger(t, 13000)
	assert.Nil(t, err)

	node2, err := createMemMessenger(t, 13001)
	assert.Nil(t, err)

	//connect nodes
	node1.ConnectToAddresses(context.Background(), []string{node2.Addrs()[0]})

	//wait a bit
	time.Sleep(time.Second)

	//create topics for each node
	node1.AddTopic(p2p.NewTopic("test", &structMemTest1{}, testMemMarshalizer))
	node2.AddTopic(p2p.NewTopic("test", &structMemTest2{}, testMemMarshalizer))

	counter := int32(0)

	//node 1 sends, node 2 receives
	node2.GetTopic("test").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		fmt.Printf("received: %v", data)
		atomic.AddInt32(&counter, 1)
	})

	node1.GetTopic("test").Broadcast(&structMemTest1{Nonce: 4, Data: 4.5})

	//wait a bit
	time.Sleep(time.Second)

	//check that the message was filtered out
	assert.Equal(t, int32(0), atomic.LoadInt32(&counter))
}

func TestMemMessenger_BroadcastOnInexistentTopic_ShouldFilteredOut(t *testing.T) {
	//stress test to check if the node is able to cope
	//with receiving on an inexistent topic
	resetMemMessengers()

	node1, err := createMemMessenger(t, 14000)
	assert.Nil(t, err)

	node2, err := createMemMessenger(t, 14001)
	assert.Nil(t, err)

	//connect nodes
	node1.ConnectToAddresses(context.Background(), []string{node2.Addrs()[0]})

	//wait a bit
	time.Sleep(time.Second)

	//create topics for each node
	node1.AddTopic(p2p.NewTopic("test1", &testMemStringCloner{}, testMemMarshalizer))
	node2.AddTopic(p2p.NewTopic("test2", &testMemStringCloner{}, testMemMarshalizer))

	counter := int32(0)

	//node 1 sends, node 2 receives
	node2.GetTopic("test2").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		fmt.Printf("received: %v", data)
		atomic.AddInt32(&counter, 1)
	})

	node1.GetTopic("test1").Broadcast(testMemStringCloner{"Foo"})

	//wait a bit
	time.Sleep(time.Second)

	//check that the message was filtered out
	assert.Equal(t, int32(0), atomic.LoadInt32(&counter))
}

func TestMemMessenger_MultipleRoundBootstrap_ShouldNotProduceLonelyNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	resetMemMessengers()

	startPort := 12000
	endPort := 12009

	nodes := make([]p2p.Messenger, 0)

	recv := make(map[string]*p2p.MessageInfo)
	mut := sync.RWMutex{}

	//prepare messengers
	for i := startPort; i <= endPort; i++ {
		node, err := createMemMessenger(t, i)

		err = node.AddTopic(p2p.NewTopic("test topic", &testMemStringCloner{}, testMemMarshalizer))
		assert.Nil(t, err)

		node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
			mut.Lock()
			recv[node.ID().Pretty()] = msgInfo

			fmt.Printf("%v got message: %v\n", node.ID().Pretty(), (*data.(*testMemStringCloner)).Data)

			mut.Unlock()

		})

		nodes = append(nodes, node)
	}

	time.Sleep(time.Second)

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

	//broadcasting something
	fmt.Println("Broadcasting a message...")
	nodes[0].GetTopic("test topic").Broadcast(testMemStringCloner{"a string to broadcast"})

	fmt.Println("Waiting...")

	//waiting 2 seconds
	time.Sleep(time.Second * 2)

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

	assert.Equal(t, 0, notRecv)
}

func TestMemMessenger_BroadcastWithValidators_ShouldWork(t *testing.T) {
	resetMemMessengers()

	fmt.Println()

	nodes := make([]*p2p.MemMessenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createMemMessenger(t, 13150+i)
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

	counter := int32(0)

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		atomic.AddInt32(&counter, 1)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		node.AddTopic(p2p.NewTopic("test", &testMemStringCloner{}, testMemMarshalizer))
		node.GetTopic("test").AddDataReceived(recv)
	}

	// dummy validator that prevents propagation of "AAA" message
	v := func(ctx context.Context, mes *pubsub.Message) bool {
		marsh := mock.MarshalizerMock{}
		obj := &testMemStringCloner{}
		marsh.Unmarshal(obj, mes.GetData())

		return obj.Data != "AAA"
	}

	//node 2 has validator in place
	nodes[2].GetTopic("test").RegisterValidator(v)

	fmt.Println()
	fmt.Println()

	//send AAA, wait 1 sec, check that 4 peers got the message
	atomic.StoreInt32(&counter, 0)
	fmt.Println("Broadcasting AAA...")
	nodes[0].GetTopic("test").Broadcast(testMemStringCloner{Data: "AAA"})
	time.Sleep(time.Second)
	assert.Equal(t, int32(4), atomic.LoadInt32(&counter))
	fmt.Printf("%d peers got the message!\n", atomic.LoadInt32(&counter))

	//send BBB, wait 1 sec, check that all peers got the message
	atomic.StoreInt32(&counter, 0)
	fmt.Println("Broadcasting BBB...")
	nodes[0].GetTopic("test").Broadcast(testMemStringCloner{Data: "BBB"})
	time.Sleep(time.Second)
	assert.Equal(t, int32(5), atomic.LoadInt32(&counter))
	fmt.Printf("%d peers got the message!\n", atomic.LoadInt32(&counter))

	//add the validator on node 4
	nodes[4].GetTopic("test").RegisterValidator(v)

	//send AAA, wait 1 sec, check that no peers got the message as the filtering should work
	atomic.StoreInt32(&counter, 0)
	fmt.Println("Broadcasting AAA...")
	nodes[0].GetTopic("test").Broadcast(testMemStringCloner{Data: "AAA"})
	time.Sleep(time.Second)
	assert.Equal(t, int32(0), atomic.LoadInt32(&counter))
	fmt.Printf("%d peers got the message!\n", atomic.LoadInt32(&counter))

	//closing
	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}
}

func TestMemMessenger_RequestResolveTestCfg1_ShouldWork(t *testing.T) {
	resetMemMessengers()

	nodes := make([]*p2p.MemMessenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createMemMessenger(t, 15000+i)
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

	counter1 := int32(0)

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		if data.(*testMemStringCloner).Data == "Real object1" {
			atomic.AddInt32(&counter1, 1)
		}

		fmt.Printf("Received: %v\n", data.(*testMemStringCloner).Data)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		node.AddTopic(p2p.NewTopic("test", &testMemStringCloner{}, testMemMarshalizer))
	}

	//to simplify, only node 0 should have a recv event handler
	nodes[0].GetTopic("test").AddDataReceived(recv)

	//setup a resolver func for node 3
	nodes[3].GetTopic("test").ResolveRequest = func(hash []byte) p2p.Newer {
		if bytes.Equal(hash, []byte("A000")) {
			return &testMemStringCloner{Data: "Real object1"}
		}

		return nil
	}

	//node0 requests an unavailable data
	nodes[0].GetTopic("test").SendRequest([]byte("B000"))
	fmt.Println("Sent request B000")
	time.Sleep(time.Second * 2)
	assert.Equal(t, int32(0), atomic.LoadInt32(&counter1))

	//node0 requests an available data on node 3
	nodes[0].GetTopic("test").SendRequest([]byte("A000"))
	fmt.Println("Sent request A000")
	time.Sleep(time.Second * 2)
	assert.Equal(t, int32(1), atomic.LoadInt32(&counter1))
}

func TestMemMessenger_RequestResolveTestCfg2_ShouldWork(t *testing.T) {
	resetMemMessengers()

	nodes := make([]*p2p.MemMessenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createMemMessenger(t, 15100+i)
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

	counter1 := int32(0)

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		if data.(*testMemStringCloner).Data == "Real object1" {
			atomic.AddInt32(&counter1, 1)
		}

		fmt.Printf("Received: %v from %v\n", data.(*testMemStringCloner).Data, msgInfo.Peer)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		node.AddTopic(p2p.NewTopic("test", &testMemStringCloner{}, testMemMarshalizer))
	}

	//to simplify, only node 1 should have a recv event handler
	nodes[1].GetTopic("test").AddDataReceived(recv)

	//resolver func for node 0 and 2
	resolverOK := func(hash []byte) p2p.Newer {
		if bytes.Equal(hash, []byte("A000")) {
			return &testMemStringCloner{Data: "Real object1"}
		}

		return nil
	}

	//resolver func for other nodes
	resolverNOK := func(hash []byte) p2p.Newer {
		panic("Should have not reached this point")

		return nil
	}

	nodes[0].GetTopic("test").ResolveRequest = resolverOK
	nodes[2].GetTopic("test").ResolveRequest = resolverOK

	nodes[3].GetTopic("test").ResolveRequest = resolverNOK
	nodes[4].GetTopic("test").ResolveRequest = resolverNOK

	//node1 requests an available data
	nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	fmt.Println("Sent request A000")
	time.Sleep(time.Second * 2)
	assert.True(t, atomic.LoadInt32(&counter1) == int32(1) || atomic.LoadInt32(&counter1) == int32(2))

}

func TestMemMessenger_RequestResolveTestSelf_ShouldWork(t *testing.T) {
	resetMemMessengers()

	nodes := make([]*p2p.MemMessenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createMemMessenger(t, 15200+i)
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

	counter1 := int32(0)

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		if data.(*testMemStringCloner).Data == "Real object1" {
			atomic.AddInt32(&counter1, 1)
		}

		fmt.Printf("Received: %v from %v\n", data.(*testMemStringCloner).Data, msgInfo.Peer)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		node.AddTopic(p2p.NewTopic("test", &testMemStringCloner{}, testMemMarshalizer))
	}

	//to simplify, only node 1 should have a recv event handler
	nodes[1].GetTopic("test").AddDataReceived(recv)

	//resolver func for node 1
	resolverOK := func(hash []byte) p2p.Newer {
		if bytes.Equal(hash, []byte("A000")) {
			return &testMemStringCloner{Data: "Real object1"}
		}

		return nil
	}

	//resolver func for other nodes
	resolverNOK := func(hash []byte) p2p.Newer {
		panic("Should have not reached this point")

		return nil
	}

	nodes[1].GetTopic("test").ResolveRequest = resolverOK

	nodes[0].GetTopic("test").ResolveRequest = resolverNOK
	nodes[2].GetTopic("test").ResolveRequest = resolverNOK
	nodes[3].GetTopic("test").ResolveRequest = resolverNOK
	nodes[4].GetTopic("test").ResolveRequest = resolverNOK

	//node1 requests an available data
	nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	fmt.Println("Sent request A000")
	time.Sleep(time.Second * 2)
	assert.Equal(t, int32(1), atomic.LoadInt32(&counter1))

}

func TestMemMessenger_RequestResolve_Resending_ShouldWork(t *testing.T) {
	resetMemMessengers()

	nodes := make([]*p2p.MemMessenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createMemMessenger(t, 15300+i)
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

	counter1 := int32(0)

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		atomic.AddInt32(&counter1, 1)

		fmt.Printf("Received: %v from %v\n", data.(*testMemStringCloner).Data, msgInfo.Peer)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		node.AddTopic(p2p.NewTopic("test", &testMemStringCloner{}, testMemMarshalizer))
	}

	//to simplify, only node 1 should have a recv event handler
	nodes[1].GetTopic("test").AddDataReceived(recv)

	//resolver func for node 0 and 2
	resolverOK := func(hash []byte) p2p.Newer {
		if bytes.Equal(hash, []byte("A000")) {
			return &testMemStringCloner{Data: "Real object0"}
		}

		return nil
	}

	//resolver func for other nodes
	resolverNOK := func(hash []byte) p2p.Newer {
		panic("Should have not reached this point")

		return nil
	}

	nodes[0].GetTopic("test").ResolveRequest = resolverOK
	nodes[2].GetTopic("test").ResolveRequest = resolverOK

	nodes[3].GetTopic("test").ResolveRequest = resolverNOK
	nodes[4].GetTopic("test").ResolveRequest = resolverNOK

	//node1 requests an available data
	nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	fmt.Println("Sent request A000")
	time.Sleep(time.Second * 2)
	assert.True(t, atomic.LoadInt32(&counter1) == int32(1))

	//resending
	nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	fmt.Println("Re-sent request A000")
	time.Sleep(time.Second * 2)
	assert.True(t, atomic.LoadInt32(&counter1) == int32(1))

	//delaying
	time.Sleep(p2p.DurTimeCache - time.Second*3)

	//resending
	nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	fmt.Println("Re-sent request A000")
	time.Sleep(time.Second * 2)
	assert.True(t, atomic.LoadInt32(&counter1) == int32(2))

}
