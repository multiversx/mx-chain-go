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
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
)

type testNetStringNewer struct {
	Data string
}

type structNetTest1 struct {
	Nonce int
	Data  float64
}

type structNetTest2 struct {
	Nonce string
	Data  []byte
}

//------- testNetStringNewer

// New will return a new instance of string. Dummy, just to implement Cloner interface as strings are immutable
func (sc *testNetStringNewer) New() p2p.Newer {
	return &testNetStringNewer{}
}

// ID will return the same string as ID
func (sc *testNetStringNewer) ID() string {
	return sc.Data
}

//------- structNetTest1

func (s1 *structNetTest1) New() p2p.Newer {
	return &structNetTest1{}
}

func (s1 *structNetTest1) ID() string {
	return strconv.Itoa(s1.Nonce)
}

//------- structNetTest2

func (s2 *structNetTest2) New() p2p.Newer {
	return &structNetTest2{}
}

func (s2 *structNetTest2) ID() string {
	return s2.Nonce
}

var testNetMessengerMaxWaitResponse = time.Duration(time.Second * 5)
var testNetMessengerWaitResponseUnreceivedMsg = time.Duration(time.Second)

var startingPort = 4000

func createNetMessenger(t *testing.T, port int, nConns int) (*p2p.NetMessenger, error) {
	return createNetMessengerPubSub(t, port, nConns, p2p.FloodSub)
}

func createNetMessengerPubSub(t *testing.T, port int, nConns int, strategy p2p.PubSubStrategy) (*p2p.NetMessenger, error) {
	cp, err := p2p.NewConnectParamsFromPort(port)
	assert.Nil(t, err)

	return p2p.NewNetMessenger(context.Background(), &mock.MarshalizerMock{}, &mock.HasherMock{}, cp, nConns, strategy)
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

func waitForValue(value *int32, expected int32, chanDone chan bool) {
	for {
		if atomic.LoadInt32(value) == expected {
			break
		}

		time.Sleep(time.Nanosecond)
	}

	chanDone <- true
}

func closeAllNodes(nodes []p2p.Messenger) {
	fmt.Println("### Closing nodes... ###")
	for i := 0; i < len(nodes); i++ {
		err := nodes[i].Close()
		if err != nil {
			p2p.Log.Error(err.Error())
		}
	}
}

func TestNetMessengerRecreationSameNodeShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	fmt.Println()

	port := startingPort

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

func TestNetMessengerSendToSelfShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, startingPort, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	defer closeAllNodes(nodes)

	wg := sync.WaitGroup{}
	wg.Add(1)
	chanDone := make(chan bool)
	go waitForWaitGroup(&wg, chanDone)

	err = nodes[0].AddTopic(p2p.NewTopic("test topic", &testNetStringNewer{}, &mock.MarshalizerMock{}))
	nodes[0].GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testNetStringNewer)).Data

		fmt.Printf("Got message: %v\n", payload)

		if payload == "ABC" {
			wg.Done()
		}
	})
	assert.Nil(t, err)

	err = nodes[0].GetTopic("test topic").Broadcast(testNetStringNewer{Data: "ABC"})
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Should have been 1 (message received to self)")
	}
}

func TestNetMessengerNodesPingPongOn2TopicsShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, startingPort, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	node, err = createNetMessenger(t, startingPort+1, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1}
	connectGraph[1] = []int{0}

	defer closeAllNodes(nodes)

	nodes[0].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0]})

	wg := sync.WaitGroup{}
	chanDone := make(chan bool)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Could not make a connection between the 2 peers")
		return
	}

	fmt.Printf("Node 1 is %s\n", nodes[0].Addresses()[0])
	fmt.Printf("Node 2 is %s\n", nodes[1].Addresses()[0])

	fmt.Printf("Node 1 has the addresses: %v\n", nodes[0].Addresses())
	fmt.Printf("Node 2 has the addresses: %v\n", nodes[1].Addresses())

	//create 2 topics on each node
	err = nodes[0].AddTopic(p2p.NewTopic("ping", &testNetStringNewer{}, &mock.MarshalizerMock{}))
	assert.Nil(t, err)
	err = nodes[0].AddTopic(p2p.NewTopic("pong", &testNetStringNewer{}, &mock.MarshalizerMock{}))
	assert.Nil(t, err)

	err = nodes[1].AddTopic(p2p.NewTopic("ping", &testNetStringNewer{}, &mock.MarshalizerMock{}))
	assert.Nil(t, err)
	err = nodes[1].AddTopic(p2p.NewTopic("pong", &testNetStringNewer{}, &mock.MarshalizerMock{}))
	assert.Nil(t, err)

	wg.Add(2)
	go waitForWaitGroup(&wg, chanDone)

	//assign some event handlers on topics
	nodes[0].GetTopic("ping").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testNetStringNewer)).Data

		if payload == "ping string" {
			fmt.Println("Ping received, sending pong...")
			err = nodes[0].GetTopic("pong").Broadcast(testNetStringNewer{"pong string"})
			assert.Nil(t, err)
		}
	})

	nodes[0].GetTopic("pong").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testNetStringNewer)).Data

		fmt.Printf("node1 received: %v\n", payload)

		if payload == "pong string" {
			fmt.Println("Pong received!")
			wg.Done()
		}
	})

	//for node2 topic ping we do not need an event handler in this test
	nodes[1].GetTopic("pong").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		payload := (*data.(*testNetStringNewer)).Data

		fmt.Printf("node2 received: %v\n", payload)

		if payload == "pong string" {
			fmt.Println("Pong received!")
			wg.Done()
		}
	})

	err = nodes[1].GetTopic("ping").Broadcast(testNetStringNewer{"ping string"})
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Should have been 2 (pong from node1: self and node2: received from node1)")
	}
}

func TestNetMessengerSimpleBroadcast5nodesInlineShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, startingPort+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addresses()[0])
	}

	defer closeAllNodes(nodes)

	//connect one with each other daisy-chain
	for i := 1; i < 5; i++ {
		node := nodes[i]
		node.ConnectToAddresses(context.Background(), []string{nodes[i-1].Addresses()[0]})
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
	case <-time.After(testNetMessengerMaxWaitResponse):
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

		err := node.AddTopic(p2p.NewTopic("test", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		node.GetTopic("test").AddDataReceived(
			func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
				fmt.Printf("%v received from %v: %v\n", node.ID(), msgInfo.Peer, data.(*testNetStringNewer).Data)
				wg.Done()
			})
		assert.Nil(t, err)
	}

	fmt.Println()
	fmt.Println()

	fmt.Println("Broadcasting...")
	err := nodes[0].GetTopic("test").Broadcast(testNetStringNewer{Data: "Foo"})
	assert.Nil(t, err)

	select {
	case <-chanDone:
		fmt.Println("Got all messages!")
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "not all messages were received")
	}
}

func TestNetMessengerSimpleBroadcast5nodesBetterConnectedShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, startingPort+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addresses()[0])
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

	nodes[1].ConnectToAddresses(context.Background(), []string{nodes[0].Addresses()[0]})
	nodes[2].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0], nodes[0].Addresses()[0]})
	nodes[3].ConnectToAddresses(context.Background(), []string{nodes[2].Addresses()[0]})
	nodes[4].ConnectToAddresses(context.Background(), []string{nodes[3].Addresses()[0], nodes[0].Addresses()[0]})

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
	case <-time.After(testNetMessengerMaxWaitResponse):
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

		err := node.AddTopic(p2p.NewTopic("test", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		node.GetTopic("test").AddDataReceived(
			func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
				fmt.Printf("%v received from %v: %v\n", node.ID(), msgInfo.Peer, data.(*testNetStringNewer).Data)
				wg.Done()
			})
		assert.Nil(t, err)
	}

	fmt.Println()
	fmt.Println()

	fmt.Println("Broadcasting...")
	err := nodes[0].GetTopic("test").Broadcast(testNetStringNewer{Data: "Foo"})
	assert.Nil(t, err)

	select {
	case <-chanDone:
		fmt.Println("Got all messages!")
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "not all messages were received")
	}
}

func TestNetMessengerSendingNilShouldErr(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, startingPort, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	defer closeAllNodes(nodes)

	err = node.AddTopic(p2p.NewTopic("test", &testNetStringNewer{}, &mock.MarshalizerMock{}))
	assert.Nil(t, err)
	err = node.GetTopic("test").Broadcast(nil)
	assert.NotNil(t, err)
}

func TestNetMessengerCreateNodeWithNilMarshalizerShouldErr(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	cp, err := p2p.NewConnectParamsFromPort(startingPort)
	assert.Nil(t, err)

	_, err = p2p.NewNetMessenger(context.Background(), nil, &mock.HasherMock{}, cp, 10, p2p.FloodSub)
	assert.NotNil(t, err)
}

func TestNetMessengerCreateNodeWithNilHasherShouldErr(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	cp, err := p2p.NewConnectParamsFromPort(startingPort)
	assert.Nil(t, err)

	_, err = p2p.NewNetMessenger(context.Background(), &mock.MarshalizerMock{}, nil, cp, 10, p2p.FloodSub)
	assert.NotNil(t, err)
}

func TestNetMessengerSingleRoundBootstrapShouldNotProduceLonelyNodes(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	startPort := startingPort
	endPort := startingPort + 9
	nConns := 4

	nodes := make([]p2p.Messenger, 0)

	recv := make(map[string]*p2p.MessageInfo)
	mut := sync.RWMutex{}

	//prepare messengers
	for i := startPort; i <= endPort; i++ {
		node, err := createNetMessenger(t, i, nConns)

		err = node.AddTopic(p2p.NewTopic("test topic", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		assert.Nil(t, err)

		node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
			mut.Lock()
			recv[node.ID().Pretty()] = msgInfo

			fmt.Printf("%v got message: %v\n", node.ID().Pretty(), (*data.(*testNetStringNewer)).Data)

			mut.Unlock()

		})

		nodes = append(nodes, node)
	}

	defer closeAllNodes(nodes)

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
	err := nodes[0].GetTopic("test topic").Broadcast(testNetStringNewer{"a string to broadcast"})
	assert.Nil(t, err)

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
}

func TestNetMessengerBadObjectToUnmarshalShouldFilteredOut(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	//stress test to check if the node is able to cope
	//with unmarshaling a bad object
	//both structs have the same fields but incompatible types

	//node1 registers topic 'test' with struct1
	//node2 registers topic 'test' with struct2

	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, startingPort, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	node, err = createNetMessenger(t, startingPort+1, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	defer closeAllNodes(nodes)

	//connect nodes
	nodes[0].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0]})

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1}
	connectGraph[1] = []int{0}

	chanDone := make(chan bool)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Could not make a connection between the 2 peers")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go waitForWaitGroup(&wg, chanDone)

	//create topics for each node
	err = nodes[0].AddTopic(p2p.NewTopic("test", &structNetTest1{}, &mock.MarshalizerMock{}))
	assert.Nil(t, err)
	err = nodes[1].AddTopic(p2p.NewTopic("test", &structNetTest2{}, &mock.MarshalizerMock{}))
	assert.Nil(t, err)

	//node 1 sends, node 2 receives
	nodes[1].GetTopic("test").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		fmt.Printf("received: %v", data)
		wg.Done()
	})

	err = nodes[0].GetTopic("test").Broadcast(&structNetTest1{Nonce: 4, Data: 4.5})
	assert.Nil(t, err)

	select {
	case <-chanDone:
		assert.Fail(t, "Should have not received the message")
	case <-time.After(testNetMessengerWaitResponseUnreceivedMsg):
	}
}

func TestNetMessengerBroadcastOnInexistentTopicShouldFilteredOut(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	//stress test to check if the node is able to cope
	//with receiving on an inexistent topic

	nodes := make([]p2p.Messenger, 0)

	node, err := createNetMessenger(t, startingPort, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	node, err = createNetMessenger(t, startingPort+1, 10)
	assert.Nil(t, err)
	nodes = append(nodes, node)

	defer closeAllNodes(nodes)

	//connect nodes
	nodes[0].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0]})

	connectGraph := make(map[int][]int)
	connectGraph[0] = []int{1}
	connectGraph[1] = []int{0}

	chanDone := make(chan bool)
	go waitForConnectionsToBeMade(nodes, connectGraph, chanDone)
	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Could not make a connection between the 2 peers")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go waitForWaitGroup(&wg, chanDone)

	//create topics for each node
	err = nodes[0].AddTopic(p2p.NewTopic("test1", &testNetStringNewer{}, &mock.MarshalizerMock{}))
	assert.Nil(t, err)
	err = nodes[1].AddTopic(p2p.NewTopic("test2", &testNetStringNewer{}, &mock.MarshalizerMock{}))
	assert.Nil(t, err)

	//node 1 sends, node 2 receives
	nodes[1].GetTopic("test2").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		fmt.Printf("received: %v", data)
		wg.Done()
	})

	err = nodes[0].GetTopic("test1").Broadcast(testNetStringNewer{"Foo"})
	assert.Nil(t, err)

	select {
	case <-chanDone:
		assert.Fail(t, "Should have not received the message")
	case <-time.After(testNetMessengerWaitResponseUnreceivedMsg):
	}
}

func TestNetMessengerMultipleRoundBootstrapShouldNotProduceLonelyNodes(t *testing.T) {
	//TODO refactor
	t.Skip("pubsub's implementation has bugs, skipping for now")

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	startPort := startingPort
	endPort := startingPort + 9
	nConns := 4

	nodes := make([]p2p.Messenger, 0)

	recv := make(map[string]*p2p.MessageInfo)
	mut := sync.RWMutex{}

	//prepare messengers
	for i := startPort; i <= endPort; i++ {
		node, err := createNetMessenger(t, i, nConns)

		err = node.AddTopic(p2p.NewTopic("test topic", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		assert.Nil(t, err)

		node.GetTopic("test topic").AddDataReceived(func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
			mut.Lock()
			recv[node.ID().Pretty()] = msgInfo

			fmt.Printf("%v got message: %v\n", node.ID().Pretty(), (*data.(*testNetStringNewer)).Data)

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
	err := nodes[0].GetTopic("test topic").Broadcast(testNetStringNewer{"a string to broadcast"})
	assert.Nil(t, err)

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

	//TODO remove the comment when pubsub will have its bug fixed
	//assert.Equal(t, 0, notRecv)
}

func TestNetMessengerBroadcastWithValidatorsShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, startingPort+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addresses()[0])
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

	nodes[1].ConnectToAddresses(context.Background(), []string{nodes[0].Addresses()[0]})
	nodes[2].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0], nodes[0].Addresses()[0]})
	nodes[3].ConnectToAddresses(context.Background(), []string{nodes[2].Addresses()[0]})
	nodes[4].ConnectToAddresses(context.Background(), []string{nodes[3].Addresses()[0], nodes[0].Addresses()[0]})

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
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	wg := sync.WaitGroup{}

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		fmt.Printf("%v got from %v the message: %v\n", msgInfo.CurrentPeer, msgInfo.Peer, data)
		wg.Done()
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		err := node.AddTopic(p2p.NewTopic("test", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		assert.Nil(t, err)
		node.GetTopic("test").AddDataReceived(recv)
	}

	// dummy validator that prevents propagation of "AAA" message
	v := func(ctx context.Context, mes *pubsub.Message) bool {
		obj := &testNetStringNewer{}

		marsh := mock.MarshalizerMock{}
		err := marsh.Unmarshal(obj, mes.GetData())
		assert.Nil(t, err)

		return obj.Data != "AAA"
	}

	//node 2 has validator in place
	err := nodes[2].GetTopic("test").RegisterValidator(v)
	assert.Nil(t, err)

	fmt.Println()
	fmt.Println()

	//send AAA, wait, check that 4 peers got the message
	fmt.Println("Broadcasting AAA...")
	wg.Add(4)
	go waitForWaitGroup(&wg, chanDone)
	err = nodes[0].GetTopic("test").Broadcast(testNetStringNewer{Data: "AAA"})
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "not all 4 peers got AAA message")
		return
	}

	//send BBB, wait, check that all peers got the message
	fmt.Println("Broadcasting BBB...")
	wg.Add(5)
	go waitForWaitGroup(&wg, chanDone)

	err = nodes[0].GetTopic("test").Broadcast(testNetStringNewer{Data: "BBB"})
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "not all 5 peers got BBB message")
		return
	}

	//add the validator on node 4
	err = nodes[4].GetTopic("test").RegisterValidator(v)
	assert.Nil(t, err)

	fmt.Println("Waiting for cooldown period (timecache should empty map)")
	time.Sleep(p2p.DurTimeCache + time.Millisecond*100)

	//send AAA, wait, check that 2 peers got the message
	fmt.Println("Resending AAA...")
	wg.Add(2)
	go waitForWaitGroup(&wg, chanDone)

	err = nodes[0].GetTopic("test").Broadcast(testNetStringNewer{Data: "AAA"})
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "not all 2 peers got AAA message")
	}
}

func TestNetMessengerBroadcastToGossipSubShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	fmt.Println()

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessengerPubSub(t, startingPort+i, 10, p2p.GossipSub)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addresses()[0])
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

	nodes[1].ConnectToAddresses(context.Background(), []string{nodes[0].Addresses()[0]})
	nodes[2].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0], nodes[0].Addresses()[0]})
	nodes[3].ConnectToAddresses(context.Background(), []string{nodes[2].Addresses()[0]})
	nodes[4].ConnectToAddresses(context.Background(), []string{nodes[3].Addresses()[0], nodes[0].Addresses()[0]})

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
	case <-time.After(testNetMessengerMaxWaitResponse):
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

		err := node.AddTopic(p2p.NewTopic("test", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		assert.Nil(t, err)
		node.GetTopic("test").AddDataReceived(recv1)
	}

	//send a piggyback message, wait 1 sec
	fmt.Println("Broadcasting piggyback message...")
	err := nodes[0].GetTopic("test").Broadcast(testNetStringNewer{Data: "piggyback"})
	assert.Nil(t, err)
	time.Sleep(time.Second)
	fmt.Printf("%d peers got the message!\n", atomic.LoadInt32(&counter))

	atomic.StoreInt32(&counter, 0)

	fmt.Println("Broadcasting AAA...")
	doWaitGroup = true
	wg.Add(5)
	go waitForWaitGroup(&wg, chanDone)
	err = nodes[0].GetTopic("test").Broadcast(testNetStringNewer{Data: "AAA"})
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "not all 5 peers got AAA message")
	}
}

func TestNetMessengerBroadcastToUnknownSubShouldErr(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	fmt.Println()

	_, err := createNetMessengerPubSub(t, startingPort, 10, 500)
	assert.NotNil(t, err)
}

func TestNetMessengerRequestResolveTestCfg1ShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, startingPort+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addresses()[0])
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

	nodes[1].ConnectToAddresses(context.Background(), []string{nodes[0].Addresses()[0]})
	nodes[2].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0], nodes[0].Addresses()[0]})
	nodes[3].ConnectToAddresses(context.Background(), []string{nodes[2].Addresses()[0]})
	nodes[4].ConnectToAddresses(context.Background(), []string{nodes[3].Addresses()[0], nodes[0].Addresses()[0]})

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
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		if data.(*testNetStringNewer).Data == "Real object1" {
			chanDone <- true
		}

		fmt.Printf("Received: %v\n", data.(*testNetStringNewer).Data)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		err := node.AddTopic(p2p.NewTopic("test", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		assert.Nil(t, err)
	}

	//to simplify, only node 0 should have a recv event handler
	nodes[0].GetTopic("test").AddDataReceived(recv)

	//setup a resolver func for node 3
	nodes[3].GetTopic("test").ResolveRequest = func(hash []byte) p2p.Newer {
		if bytes.Equal(hash, []byte("A000")) {
			return &testNetStringNewer{Data: "Real object1"}
		}

		return nil
	}

	//node0 requests an unavailable data
	err := nodes[0].GetTopic("test").SendRequest([]byte("B000"))
	assert.Nil(t, err)
	fmt.Println("Sent request B000")
	select {
	case <-chanDone:
		assert.Fail(t, "Should have not sent object")
	case <-time.After(testNetMessengerWaitResponseUnreceivedMsg):
	}

	//node0 requests an available data on node 3
	err = nodes[0].GetTopic("test").SendRequest([]byte("A000"))
	assert.Nil(t, err)
	fmt.Println("Sent request A000")

	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Should have sent object")
		return
	}
}

func TestNetMessengerRequestResolveTestCfg2ShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, startingPort+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addresses()[0])
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

	nodes[1].ConnectToAddresses(context.Background(), []string{nodes[0].Addresses()[0]})
	nodes[2].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0], nodes[0].Addresses()[0]})
	nodes[3].ConnectToAddresses(context.Background(), []string{nodes[2].Addresses()[0]})
	nodes[4].ConnectToAddresses(context.Background(), []string{nodes[3].Addresses()[0], nodes[0].Addresses()[0]})

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
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		if data.(*testNetStringNewer).Data == "Real object1" {
			chanDone <- true
		}

		fmt.Printf("Received: %v from %v\n", data.(*testNetStringNewer).Data, msgInfo.Peer)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		err := node.AddTopic(p2p.NewTopic("test", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		assert.Nil(t, err)
	}

	//to simplify, only node 1 should have a recv event handler
	nodes[1].GetTopic("test").AddDataReceived(recv)

	//resolver func for node 0 and 2
	resolverOK := func(hash []byte) p2p.Newer {
		if bytes.Equal(hash, []byte("A000")) {
			return &testNetStringNewer{Data: "Real object1"}
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
	err := nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	assert.Nil(t, err)
	fmt.Println("Sent request A000")

	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Should have sent object")
	}

}

func TestNetMessengerRequestResolveTestSelfShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, startingPort+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addresses()[0])
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

	nodes[1].ConnectToAddresses(context.Background(), []string{nodes[0].Addresses()[0]})
	nodes[2].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0], nodes[0].Addresses()[0]})
	nodes[3].ConnectToAddresses(context.Background(), []string{nodes[2].Addresses()[0]})
	nodes[4].ConnectToAddresses(context.Background(), []string{nodes[3].Addresses()[0], nodes[0].Addresses()[0]})

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
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		if data.(*testNetStringNewer).Data == "Real object1" {
			chanDone <- true
		}

		fmt.Printf("Received: %v from %v\n", data.(*testNetStringNewer).Data, msgInfo.Peer)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		err := node.AddTopic(p2p.NewTopic("test", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		assert.Nil(t, err)
	}

	//to simplify, only node 1 should have a recv event handler
	nodes[1].GetTopic("test").AddDataReceived(recv)

	//resolver func for node 1
	resolverOK := func(hash []byte) p2p.Newer {
		if bytes.Equal(hash, []byte("A000")) {
			return &testNetStringNewer{Data: "Real object1"}
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
	err := nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	assert.Nil(t, err)
	fmt.Println("Sent request A000")

	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Should have self-sent object")
	}

}

func TestNetMessengerRequestResolveResendingShouldWork(t *testing.T) {
	if skipP2PMessengerTests {
		t.Skip("test skipped for P2PMessenger struct")
	}

	nodes := make([]p2p.Messenger, 0)

	//create 5 nodes
	for i := 0; i < 5; i++ {
		node, err := createNetMessenger(t, startingPort+i, 10)
		assert.Nil(t, err)

		nodes = append(nodes, node)

		fmt.Printf("Node %v is %s\n", i+1, node.Addresses()[0])
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

	nodes[1].ConnectToAddresses(context.Background(), []string{nodes[0].Addresses()[0]})
	nodes[2].ConnectToAddresses(context.Background(), []string{nodes[1].Addresses()[0], nodes[0].Addresses()[0]})
	nodes[3].ConnectToAddresses(context.Background(), []string{nodes[2].Addresses()[0]})
	nodes[4].ConnectToAddresses(context.Background(), []string{nodes[3].Addresses()[0], nodes[0].Addresses()[0]})

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
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Could not make connections")
		return
	}

	counter := int32(0)

	recv := func(name string, data interface{}, msgInfo *p2p.MessageInfo) {
		atomic.AddInt32(&counter, 1)

		fmt.Printf("Received: %v from %v\n", data.(*testNetStringNewer).Data, msgInfo.Peer)
	}

	//print connected and create topics
	for i := 0; i < 5; i++ {
		node := nodes[i]
		node.PrintConnected()

		err := node.AddTopic(p2p.NewTopic("test", &testNetStringNewer{}, &mock.MarshalizerMock{}))
		assert.Nil(t, err)
	}

	//to simplify, only node 1 should have a recv event handler
	nodes[1].GetTopic("test").AddDataReceived(recv)

	//resolver func for node 0 and 2
	resolverOK := func(hash []byte) p2p.Newer {
		if bytes.Equal(hash, []byte("A000")) {
			return &testNetStringNewer{Data: "Real object0"}
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
	go waitForValue(&counter, 1, chanDone)
	err := nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	assert.Nil(t, err)
	fmt.Println("Sent request A000")
	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Should have received 1 object")
		return
	}

	//resending request. This should be filtered out
	atomic.StoreInt32(&counter, 0)
	go waitForValue(&counter, 1, chanDone)
	err = nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	assert.Nil(t, err)
	fmt.Println("Re-sent request A000")
	select {
	case <-chanDone:
		assert.Fail(t, "Should have not received")
		return
	case <-time.After(testNetMessengerWaitResponseUnreceivedMsg):
	}

	fmt.Println("delaying as to clear timecache buffer")
	time.Sleep(p2p.DurTimeCache + time.Millisecond*100)

	//resending
	atomic.StoreInt32(&counter, 0)
	go waitForValue(&counter, 1, chanDone)
	err = nodes[1].GetTopic("test").SendRequest([]byte("A000"))
	assert.Nil(t, err)
	fmt.Println("Re-sent request A000")
	select {
	case <-chanDone:
	case <-time.After(testNetMessengerMaxWaitResponse):
		assert.Fail(t, "Should have received 2 objects")
		return
	}

}
