package p2p_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"
)

var testMarshalizer = &mock.MarshalizerMock{}
var testHasher = &mock.HasherMock{}

func createMessenger(port int, maxAllowedPeers int) (p2p.Messenger, error) {
	cp := p2p.NewConnectParamsFromPort(port)

	return p2p.NewNetMessenger(context.Background(), testMarshalizer, *cp, []string{}, maxAllowedPeers)
}

func TestNetMessenger_Bootstrap(t *testing.T) {
	logging.SetLogLevel("pubsub", "CRITICAL")

	startPort := 12000
	endPort := 12009
	nConns := 4

	nodes := make([]p2p.Messenger, 0)

	mapHops := make(map[int]int)

	for i := startPort; i <= endPort; i++ {
		node, err := createMessenger(i, nConns)

		node.TopicHolder().AddTopic(p2p.NewTopic("string_topic", "", testMarshalizer, testHasher, 10000))
		node.TopicHolder().GetTopic("string_topic").AddEventHandler(func(name string, data interface{}) {
			fmt.Printf("%v got on topic %v message: %v\n", node.ID(), name, *data.(*string))
		})

		assert.Nil(t, err)

		nodes = append(nodes, node)
	}

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(len(nodes))

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]

		go func() {
			node.Bootstrap(context.Background())
			wg.Done()
		}()
	}

	wg.Wait()

	for i := 0; i < len(nodes); i++ {
		nodes[i].PrintConnected()
		fmt.Println()
	}

	time.Sleep(time.Second)

	//broadcastind something
	fmt.Println("Broadcasting a message...")
	nodes[0].TopicHolder().GetTopic("string_topic").Broadcast("string", nodes[0].ID(), nil)

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
