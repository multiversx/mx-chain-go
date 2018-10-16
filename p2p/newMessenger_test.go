package p2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/stretchr/testify/assert"
)

var testMarshalizer = &mock.MarshalizerMock{}
var testHasher = &mock.HasherMock{}

func createNewMessenger(port int, maxAllowedPeers int) (*NewMessenger, error) {
	cp := NewConnectParamsFromPort(port)

	return NewNewMessenger(context.Background(), testMarshalizer, *cp, []string{}, maxAllowedPeers)
}

func TestNewMessenger_Bootstrap(t *testing.T) {

	startPort := 12000
	endPort := 12009
	nConns := 4

	nodes := make([]*NewMessenger, 0)

	//recv := make(map[string]*string)
	//mut := sync.RWMutex{}

	mapHops := make(map[int]int)

	for i := startPort; i <= endPort; i++ {
		node, err := createNewMessenger(i, nConns)

		node.RegisterTopic("AAA")

		assert.Nil(t, err)

		nodes = append(nodes, node)
	}

	for i := 1; i < len(nodes); i++ {
		nodes[i].ConnectToAddresses(context.Background(), []string{nodes[i-1].GetIpfsAddress()})
	}

	time.Sleep(time.Second * 10)

	for i := 0; i < len(nodes); i++ {
		nodes[i].PrintConnected()
		fmt.Println()
	}

	time.Sleep(time.Second)

	for i := 0; i < len(nodes); i++ {
		peers := nodes[i].pubsub.ListPeers("AAA")

		fmt.Printf("%v is connected to: \n", nodes[i].p2pNode.ID().Pretty())

		for j := 0; j < len(peers); j++ {
			fmt.Printf(" - %v\n", peers[j].Pretty())
		}

		fmt.Println()
	}

	//broadcastind something
	fmt.Println("Broadcasting a message...")
	err := nodes[0].PublishOnTopic("AAA", []byte{65, 66, 67})
	assert.Nil(t, err)

	fmt.Println("Waiting...")
	time.Sleep(time.Second * 2)

	maxHops := 0

	notRecv := 0
	didRecv := 0

	//for i := 0; i < len(nodes); i++ {
	//
	//	mut.RLock()
	//	v, found := recv[(*nodes[i]).ID().Pretty()]
	//	mut.RUnlock()
	//
	//	if !found {
	//		fmt.Printf("Peer %s didn't got the message!\n", (*nodes[i]).ID().Pretty())
	//		notRecv++
	//	} else {
	//		fmt.Printf("Peer %s got the message in %d hops!\n", (*nodes[i]).ID().Pretty(), v.Hops)
	//		didRecv++
	//
	//		val, found := mapHops[v.Hops]
	//		if !found {
	//			mapHops[v.Hops] = 1
	//		} else {
	//			mapHops[v.Hops] = val + 1
	//		}
	//
	//		if maxHops < v.Hops {
	//			maxHops = v.Hops
	//		}
	//	}
	//}

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
