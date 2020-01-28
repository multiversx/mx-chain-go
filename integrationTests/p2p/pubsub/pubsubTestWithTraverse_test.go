package peerDisconnecting

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type printInterceptor struct {
	idx int
	pid p2p.PeerID
}

func (pi *printInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	str := fmt.Sprintf("Peer index %d - pid: %s\n%s", pi.idx, pi.pid.Pretty(), message.TraverseInfoTable())
	log := logger.GetOrCreate("p2pmes")
	log.Info(str)

	return nil
}

func (pi *printInterceptor) IsInterfaceNil() bool {
	return pi == nil
}

func TestPubsubMessageTraverse(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	peers := createNetwork()

	defer func() {
		for _, p := range peers {
			_ = p.Close()
		}
	}()

	time.Sleep(time.Second * 2)

	peers[0].Broadcast("topic", []byte("a message"))

	time.Sleep(time.Second * 2)
}

func createNetwork() []p2p.Messenger {
	numPeers := 8
	peers := createPeers(numPeers)

	mapConnection := map[int][]int{
		0: {1},
		1: {2, 5},
		2: {3, 6},
		4: {5},
		6: {7},
	}

	fmt.Println(
		`network config: 
0 ------- 1 ------- 2 --------- 3
          |         |
4 ------- 5         6 --------- 7`)
	fmt.Println()

	makeConnections(peers, mapConnection)
	assignInterceptors(peers)

	return peers
}

func createPeers(numPeers int) []p2p.Messenger {
	peers := make([]p2p.Messenger, numPeers)
	for i := 0; i < numPeers; i++ {
		peers[i] = integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	}

	return peers
}

func makeConnections(peers []p2p.Messenger, connections map[int][]int) {
	for idxConnector, connectTo := range connections {
		peerConnector := peers[idxConnector]
		for _, idx := range connectTo {
			peer := peers[idx]

			err := peerConnector.ConnectToPeer(peer.Addresses()[0])
			if err != nil {
				panic(err)
			}
		}
	}
}

func assignInterceptors(peers []p2p.Messenger) {
	topic := "topic"
	for i, p := range peers {
		_ = p.CreateTopic(topic, true)
		_ = p.RegisterMessageProcessor(
			topic,
			&printInterceptor{
				idx: i,
				pid: p.ID(),
			},
		)
	}
}
