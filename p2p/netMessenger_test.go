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
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

var testNetMarshalizer = &mock.MarshalizerMock{}
var testNetHasher = &mock.HasherMock{}
var maxDelayWaitResponse = time.Duration(time.Second * 5)
var delayWaitResponseUnreceivableMessages = time.Duration(time.Second)

func createNetMessenger(t *testing.T, port int, nConns int) (*p2p.NetMessenger, error) {
	return createNetMessengerPubSub(t, port, nConns, p2p.FloodSub)
}

func createNetMessengerPubSub(t *testing.T, port int, nConns int, strategy p2p.PubSubStrategy) (*p2p.NetMessenger, error) {
	cp, err := p2p.NewConnectParamsFromPort(port)
	assert.Nil(t, err)

	return p2p.NewNetMessenger(context.Background(), testNetMarshalizer, testNetHasher, cp, nConns, strategy)
}

func waitForConnectionsToBeMade(nodes []p2p.Messenger, connectGraph map[int][]int, chanDone chan bool) {
	for {
		fullyConnected := true

		//for each element in the connect graph, check that is really connected to other peers
		for k, v := range connectGraph {
			for _, peerIndex := range v {
				if nodes[k].Connectedness(nodes[peerIndex].ID()) != net.Connected {
					fullyConnected = false
					break
				}
			}
		}

		if fullyConnected {
			break
		}

		time.Sleep(time.Millisecond)
	}
	chanDone <- true
}

func waitForWaitGroup(wg *sync.WaitGroup, chanDone chan bool) {
	wg.Wait()
	chanDone <- true
}

func closeAllNodes(nodes []p2p.Messenger) {
	for i := 0; i < len(nodes); i++ {
		_ = nodes[i].Close()
	}
}

func TestNetMessenger_RecreationSameNode_ShouldWork(t *testing.T) {
	fmt.Println()

	port := 4000

	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, port, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	node, err = createNetMessenger(t, port, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	defer closeAllNodes(nodes)

	if nodes[0].ID().Pretty() != nodes[1].ID().Pretty() {
		t.Fatal("ID mismatch")
	}
}

func TestNetMessenger_SendToSelf_ShouldWork(t *testing.T) {
	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, 4500, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	defer closeAllNodes(nodes)

	wg := sync.WaitGroup{}
	wg.Add(1)
	chanDone := make(chan bool)
	go waitForWaitGroup(&wg, chanDone)

	_ = nodes[0].AddTopic(p2p.NewTopic("test topic", &testStringNewer{}, testNetMarshalizer))
	nodes[0].GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testStringNewer)).Data

		fmt.Printf("Got message: %v\n", payload)

		if payload == "ABC" {
			wg.Done()
		}
	})

	_ = nodes[0].GetTopic("test topic").Broadcast(testStringNewer{Data: "ABC"})

	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Should have been 1 (message received to self)")
	}
}

func TestNetMessenger_NodesPingPongOn2Topics_ShouldWork(t *testing.T) {
	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, 5110, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	node, err = createNetMessenger(t, 5111, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1}
	connectGraph[1] = []int{0}

	defer closeAllNodes(nodes)

	nodes[0].ConnectToAddresses(context.Background(), []string{nodes[1].Addrs()[0]})

	wg := sync.WaitGroup{}
	chanDone := make(chan bool)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Could not make a connection between the 2 peers")
		return
	}

	fmt.Printf("Node 1 is %s\n", nodes[0].Addrs()[0])
	fmt.Printf("Node 2 is %s\n", nodes[1].Addrs()[0])

	fmt.Printf("Node 1 has the addresses: %v\n", nodes[0].Addrs())
	fmt.Printf("Node 2 has the addresses: %v\n", nodes[1].Addrs())

	//create 2 topics on each node
	_ = nodes[0].AddTopic(p2p.NewTopic("ping", &testStringNewer{}, testNetMarshalizer))
	_ = nodes[0].AddTopic(p2p.NewTopic("pong", &testStringNewer{}, testNetMarshalizer))

	_ = nodes[1].AddTopic(p2p.NewTopic("ping", &testStringNewer{}, testNetMarshalizer))
	_ = nodes[1].AddTopic(p2p.NewTopic("pong", &testStringNewer{}, testNetMarshalizer))

	wg.Add(2)
	go waitForWaitGroup(&wg, chanDone)

	//assign some event handlers on topics
	nodes[0].GetTopic("ping").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testStringNewer)).Data

		if payload == "ping string" {
			fmt.Println("Ping received, sending pong...")
			_ = nodes[0].GetTopic("pong").Broadcast(testStringNewer{"pong string"})
		}
	})

	nodes[0].GetTopic("pong").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testStringNewer)).Data

		fmt.Printf("node1 received: %v\n", payload)

		if payload == "pong string" {
			fmt.Println("Pong received!")
			wg.Done()
		}
	})

	//for node2 topic ping we do not need an event handler in this test
	nodes[1].GetTopic("pong").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testStringNewer)).Data

		fmt.Printf("node2 received: %v\n", payload)

		if payload == "pong string" {
			fmt.Println("Pong received!")
			wg.Done()
		}
	})

	_ = nodes[1].GetTopic("ping").Broadcast(testStringNewer{"ping string"})

	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Should have been 2 (pong from node1: self and node2: received from node1)")
	}
}

func TestNetMessenger_SimpleBroadcast5nodesInline_ShouldWork(t *testing.T) {
	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, 6100+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addrs()[0])
	}

	defer closeAllNodes(nodes)

	//connect one with each other daisy-chain
	for i := 1; i < 5; i++ {
		node := nodes[i]
		node.ConnectToAddresses(context.Background(), []string{nodes[i-1].Addrs()[0]})
	}

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1}
	connectGraph[1] = []int{0, 2}
	connectGraph[2] = []int{1, 3}
	connectGraph[3] = []int{2, 4}
	connectGraph[4] = []int{3}

	chanDone := make(chan bool)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(5)
	go waitForWaitGroup(&wg, chanDone)

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		_ = node.AddTopic(p2p.NewTopic("test", &testStringNewer{}, testNetMarshalizer))
		node.GetTopic("test").AddDataReceived(
			func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
				fmt.Printf("%v received from %v: %v\n", node.ID(), msgInfo.Peer, data.(*testStringNewer).Data)
				wg.Done()
			})
	}

	fmt.Println()
	fmt.Println()

	fmt.Println("Broadcasting...")
	_ = nodes[0].GetTopic("test").Broadcast(testStringNewer{Data: "Foo"})

	select {
	case <-chanDone:
		fmt.Println("Got all messages!")
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "not all messages were received")
	}
}

func TestNetMessenger_SimpleBroadcast5nodesBetterConnected_ShouldWork(t *testing.T) {
	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, 7000+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addrs()[0])
	}

	defer closeAllNodes(nodes)

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

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1, 2, 4}
	connectGraph[1] = []int{0, 2}
	connectGraph[2] = []int{0, 1, 3}
	connectGraph[3] = []int{2, 4}
	connectGraph[4] = []int{0, 3}

	chanDone := make(chan bool)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(5)
	go waitForWaitGroup(&wg, chanDone)

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		_ = node.AddTopic(p2p.NewTopic("test", &testStringNewer{}, testNetMarshalizer))
		node.GetTopic("test").AddDataReceived(
			func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
				fmt.Printf("%v received from %v: %v\n", node.ID(), msgInfo.Peer, data.(*testStringNewer).Data)
				wg.Done()
			})
	}

	fmt.Println()
	fmt.Println()

	fmt.Println("Broadcasting...")
	_ = nodes[0].GetTopic("test").Broadcast(testStringNewer{Data: "Foo"})

	select {
	case <-chanDone:
		fmt.Println("Got all messages!")
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "not all messages were received")
	}
}

func TestNetMessenger_SendingNil_ShouldErr(t *testing.T) {
	node1, err := createNetMessenger(t, 9000, 10)
	assert.Nil(t, err)

	_ = node1.AddTopic(p2p.NewTopic("test", &testStringNewer{}, testNetMarshalizer))
	err = node1.GetTopic("test").Broadcast(nil)
	assert.NotNil(t, err)

	_ = node1.Close()
}

func TestNetMessenger_CreateNodeWithNilMarshalizer_ShouldErr(t *testing.T) {
	cp, err := p2p.NewConnectParamsFromPort(11000)
	assert.Nil(t, err)

	_, err = p2p.NewNetMessenger(context.Background(), nil, testNetHasher, cp, 10, p2p.FloodSub)

	assert.NotNil(t, err)
}

func TestNetMessenger_CreateNodeWithNilHasher_ShouldErr(t *testing.T) {
	cp, err := p2p.NewConnectParamsFromPort(12000)
	assert.Nil(t, err)

	_, err = p2p.NewNetMessenger(context.Background(), testNetMarshalizer, nil, cp, 10, p2p.FloodSub)

	assert.NotNil(t, err)
}

func TestNetMessenger_SingleRoundBootstrap_ShouldNotProduceLonelyNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	startPort := 12000
	endPort := 12009
	nConns := 4

	nodes := make([]p2p.Messenger, 0)

	recv := make(map[string]*p2p.MessageInfo)
	mut := sync.RWMutex{}

	//prepare messengers
	for i := startPort; i <= endPort; i++ {
		node, err := createNetMessenger(t, i, nConns)

		err = node.AddTopic(p2p.NewTopic("test topic", &testStringNewer{}, testNetMarshalizer))
		assert.Nil(t, err)

		node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
			mut.Lock()
			recv[node.ID().Pretty()] = msgInfo

			fmt.Printf("%v got message: %v\n", node.ID().Pretty(), (*data.(*testStringNewer)).Data)

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
	_ = nodes[0].GetTopic("test topic").Broadcast(testStringNewer{"a string to broadcast"})

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

	//TODO uncomment this when pubsub issue is done
	//assert.Equal(t, 0, notRecv)

	//closing
	for i := 0; i < len(nodes); i++ {
		_ = nodes[i].Close()
	}
}

func TestNetMessenger_BadObjectToUnmarshal_ShouldFilteredOut(t *testing.T) {
	//stress test to check if the node is able to cope
	//with unmarshaling a bad object
	//both structs have the same fields but incompatible types

	//node1 registers topic 'test' with struct1
	//node2 registers topic 'test' with struct2

	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, 13000, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	node, err = createNetMessenger(t, 13001, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	defer closeAllNodes(nodes)

	//connect nodes
	nodes[0].ConnectToAddresses(context.Background(), []string{nodes[1].Addrs()[0]})

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1}
	connectGraph[1] = []int{0}

	chanDone := make(chan bool)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Could not make a connection between the 2 peers")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go waitForWaitGroup(&wg, chanDone)

	//create topics for each node
	_ = nodes[0].AddTopic(p2p.NewTopic("test", &structTest1{}, testNetMarshalizer))
	_ = nodes[1].AddTopic(p2p.NewTopic("test", &structTest2{}, testNetMarshalizer))

	//node 1 sends, node 2 receives
	nodes[1].GetTopic("test").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		fmt.Printf("received: %v", data)
		wg.Done()
	})

	_ = nodes[0].GetTopic("test").Broadcast(&structTest1{Nonce: 4, Data: 4.5})

	select {
	case <-chanDone:
		assert.Fail(t, "Should have not received the message")
	case <-time.After(delayWaitResponseUnreceivableMessages):
	}
}

func TestNetMessenger_BroadcastOnInexistentTopic_ShouldFilteredOut(t *testing.T) {
	//stress test to check if the node is able to cope
	//with receiving on an inexistent topic

	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, 14000, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	node, err = createNetMessenger(t, 14001, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	defer closeAllNodes(nodes)

	//connect nodes
	nodes[0].ConnectToAddresses(context.Background(), []string{nodes[1].Addrs()[0]})

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1}
	connectGraph[1] = []int{0}

	chanDone := make(chan bool)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Could not make a connection between the 2 peers")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go waitForWaitGroup(&wg, chanDone)

	//create topics for each node
	_ = nodes[0].AddTopic(p2p.NewTopic("test1", &testStringNewer{}, testNetMarshalizer))
	_ = nodes[1].AddTopic(p2p.NewTopic("test2", &testStringNewer{}, testNetMarshalizer))

	//node 1 sends, node 2 receives
	nodes[1].GetTopic("test2").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		fmt.Printf("received: %v", data)
		wg.Done()
	})

	_ = nodes[0].GetTopic("test1").Broadcast(testStringNewer{"Foo"})

	select {
	case <-chanDone:
		assert.Fail(t, "Should have not received the message")
	case <-time.After(delayWaitResponseUnreceivableMessages):
	}
}

func TestNetMessenger_MultipleRoundBootstrap_ShouldNotProduceLonelyNodes(t *testing.T) {
	t.Skip("pubsub's implementation has bugs, skipping for now")

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	startPort := 12000
	endPort := 12009
	nConns := 4

	nodes := make([]p2p.Messenger, 0)

	recv := make(map[string]*p2p.MessageInfo)
	mut := sync.RWMutex{}

	//prepare messengers
	for i := startPort; i <= endPort; i++ {
		node, err := createNetMessenger(t, i, nConns)

		err = node.AddTopic(p2p.NewTopic("test topic", &testStringNewer{}, testNetMarshalizer))
		assert.Nil(t, err)

		node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
			mut.Lock()
			recv[node.ID().Pretty()] = msgInfo

			fmt.Printf("%v got message: %v\n", node.ID().Pretty(), (*data.(*testStringNewer)).Data)

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
	_ = nodes[0].GetTopic("test topic").Broadcast(testStringNewer{"a string to broadcast"})

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

	//closing
	for i := 0; i < len(nodes); i++ {
		_ = nodes[i].Close()
	}
}

func TestNetMessenger_BroadcastWithValidators_ShouldWork(t *testing.T) {
	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, 13120+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addrs()[0])
	}

	defer closeAllNodes(nodes)

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

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1, 2, 4}
	connectGraph[1] = []int{0, 2}
	connectGraph[2] = []int{0, 1, 3}
	connectGraph[3] = []int{2, 4}
	connectGraph[4] = []int{0, 3}

	chanDone := make(chan bool, 0)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	wg := sync.WaitGroup{}

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		wg.Done()
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		_ = node.AddTopic(p2p.NewTopic("test", &testStringNewer{}, testNetMarshalizer))
		node.GetTopic("test").AddDataReceived(recv)
	}

	// dummy validator that prevents propagation of "AAA" message
	v := func(ctx context.Context, mes *pubsub.Message) bool {
		obj := &testStringNewer{}

		marsh := mock.MarshalizerMock{}
		_ = marsh.Unmarshal(obj, mes.GetData())

		return obj.Data != "AAA"
	}

	//node 2 has validator in place
	_ = nodes[2].GetTopic("test").RegisterValidator(v)

	fmt.Println()
	fmt.Println()

	//send AAA, wait, check that 4 peers got the message
	fmt.Println("Broadcasting AAA...")
	wg.Add(4)
	go waitForWaitGroup(&wg, chanDone)
	_ = nodes[0].GetTopic("test").Broadcast(testStringNewer{Data: "AAA"})
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "not all 4 peers got AAA message")
		return
	}

	//send BBB, wait, check that all peers got the message
	fmt.Println("Broadcasting BBB...")
	wg.Add(5)
	go waitForWaitGroup(&wg, chanDone)

	_ = nodes[0].GetTopic("test").Broadcast(testStringNewer{Data: "BBB"})
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "not all 5 peers got BBB message")
		return
	}

	//add the validator on node 4
	_ = nodes[4].GetTopic("test").RegisterValidator(v)

	//send AAA, wait, check that 2 peers got the message
	fmt.Println("Resending AAA...")
	wg.Add(2)
	go waitForWaitGroup(&wg, chanDone)

	_ = nodes[0].GetTopic("test").Broadcast(testStringNewer{Data: "AAA"})
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "not all 2 peers got AAA message")
	}
}

func TestNetMessenger_BroadcastToGossipSub_ShouldWork(t *testing.T) {
	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessengerPubSub(t, 14000+i, 10, p2p.GossipSub)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addrs()[0])
	}

	defer closeAllNodes(nodes)

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

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1, 2, 4}
	connectGraph[1] = []int{0, 2}
	connectGraph[2] = []int{0, 1, 3}
	connectGraph[3] = []int{2, 4}
	connectGraph[4] = []int{0, 3}

	chanDone := make(chan bool, 0)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	wg := sync.WaitGroup{}
	doWaitGroup := false
	counter := int32(0)

	recv1 := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		if doWaitGroup {
			wg.Done()
		}

		atomic.AddInt32(&counter, 1)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		_ = node.AddTopic(p2p.NewTopic("test", &testStringNewer{}, testNetMarshalizer))
		node.GetTopic("test").AddDataReceived(recv1)
	}

	//send a piggyback message, wait 1 sec
	fmt.Println("Broadcasting piggyback message...")
	_ = nodes[0].GetTopic("test").Broadcast(testStringNewer{Data: "piggyback"})
	time.Sleep(time.Second)
	fmt.Printf("%d peers got the message!\n", atomic.LoadInt32(&counter))

	atomic.StoreInt32(&counter, 0)

	fmt.Println("Broadcasting AAA...")
	doWaitGroup = true
	wg.Add(5)
	go waitForWaitGroup(&wg, chanDone)
	_ = nodes[0].GetTopic("test").Broadcast(testStringNewer{Data: "AAA"})
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "not all 5 peers got AAA message")
	}
}

func TestNetMessenger_BroadcastToRandomSub_ShouldWork(t *testing.T) {
	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessengerPubSub(t, 14100+i, 10, p2p.RandomSub)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addrs()[0])
	}

	defer closeAllNodes(nodes)

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

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1, 2, 4}
	connectGraph[1] = []int{0, 2}
	connectGraph[2] = []int{0, 1, 3}
	connectGraph[3] = []int{2, 4}
	connectGraph[4] = []int{0, 3}

	chanDone := make(chan bool, 0)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(5)
	go waitForWaitGroup(&wg, chanDone)

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		wg.Done()
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		_ = node.AddTopic(p2p.NewTopic("test", &testStringNewer{}, testNetMarshalizer))
		node.GetTopic("test").AddDataReceived(recv)
	}

	//send AAA, wait, check that 5 peers got the message
	fmt.Println("Broadcasting AAA...")
	_ = nodes[0].GetTopic("test").Broadcast(testStringNewer{Data: "AAA"})
	select {
	case <-chanDone:
	case <-time.After(maxDelayWaitResponse):
		assert.Fail(t, "not all 5 peers got AAA message")
	}
}

func TestNetMessenger_BroadcastToUnknownSub_ShouldErr(t *testing.T) {
	fmt.Println()

	_, err := createNetMessengerPubSub(t, 14200, 10, 500)
	assert.NotNil(t, err)
}
