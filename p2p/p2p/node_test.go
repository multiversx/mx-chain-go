package p2p

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

var counter1 int32
var counter2 int32

func TestRecreationSameNode(t *testing.T) {

	port := 4000

	node1 := CreateNewNode(context.Background(), port, []string{})

	node2 := CreateNewNode(context.Background(), port, []string{})

	if node1.P2pNode.ID().Pretty() != node2.P2pNode.ID().Pretty() {
		t.Fatal("ID mismatch")
	}
}

func TestSimpleSend2NodesPingPong(t *testing.T) {
	node1 := CreateNewNode(context.Background(), 5000, []string{})

	fmt.Println(node1.P2pNode.Addrs()[0].String())

	node2 := CreateNewNode(context.Background(), 5001, []string{node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()})

	time.Sleep(time.Second)

	var val int32 = 0

	node1.OnMsgRecv = func(sender *Node, peerID string, message string) {
		fmt.Printf("Got message from peerID %v: %v\n", peerID, message)

		if message == "Ping" {
			atomic.AddInt32(&val, 1)
			sender.SendDirect(peerID, "Pong")
		}
	}

	node2.OnMsgRecv = func(sender *Node, peerID string, message string) {
		fmt.Printf("Got message from peerID %v: %v\n", peerID, message)

		if message == "Ping" {
			sender.SendDirect(peerID, "Pong")
		}

		if message == "Pong" {
			atomic.AddInt32(&val, 1)
		}

	}

	node2.SendDirect(node1.P2pNode.ID().Pretty(), "Ping")

	time.Sleep(time.Second)

	if val != 2 {
		t.Fatal("Should have been 2 (ping-pong pair)")
	}

	node1.P2pNode.Close()
	node2.P2pNode.Close()

}

func recv1(sender *Node, peerID string, message string) {
	atomic.AddInt32(&counter1, 1)
	fmt.Printf("%v > %v: Got message from peerID %v: %v\n", time.Now(), sender.P2pNode.ID().Pretty(), peerID, message)
	sender.Broadcast(message, []string{peerID})
}

func recv2(sender *Node, peerID string, message string) {
	atomic.AddInt32(&counter2, 1)
	fmt.Printf("%v > %v: Got message from peerID %v: %v\n", time.Now(), sender.P2pNode.ID().Pretty(), peerID, message)
	sender.Broadcast(message, []string{peerID})
}

func TestSimpleBroadcast5nodesInline(t *testing.T) {
	counter1 = 0

	node1 := CreateNewNode(context.Background(), 6000, []string{})
	node2 := CreateNewNode(context.Background(), 6001, []string{node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()})
	node3 := CreateNewNode(context.Background(), 6002, []string{node2.P2pNode.Addrs()[0].String() + "/ipfs/" + node2.P2pNode.ID().Pretty()})
	node4 := CreateNewNode(context.Background(), 6003, []string{node3.P2pNode.Addrs()[0].String() + "/ipfs/" + node3.P2pNode.ID().Pretty()})
	node5 := CreateNewNode(context.Background(), 6004, []string{node4.P2pNode.Addrs()[0].String() + "/ipfs/" + node4.P2pNode.ID().Pretty()})

	time.Sleep(time.Second)

	node1.OnMsgRecv = recv1
	node2.OnMsgRecv = recv1
	node3.OnMsgRecv = recv1
	node4.OnMsgRecv = recv1
	node5.OnMsgRecv = recv1

	fmt.Println()
	fmt.Println()

	node1.Broadcast("Boo", []string{})

	time.Sleep(time.Second)

	if counter1 != 4 {
		t.Fatal("Should have been 4 (traversed 4 peers)")
	}

	node1.P2pNode.Close()
	node2.P2pNode.Close()
	node3.P2pNode.Close()
	node4.P2pNode.Close()
	node5.P2pNode.Close()

}

func TestSimpleBroadcast5nodesBeterConnected(t *testing.T) {
	counter2 = 0

	node1 := CreateNewNode(context.Background(), 7000, []string{})
	node2 := CreateNewNode(context.Background(), 7001, []string{node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()})
	node3 := CreateNewNode(context.Background(), 7002, []string{node2.P2pNode.Addrs()[0].String() + "/ipfs/" + node2.P2pNode.ID().Pretty(),
		node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()})
	node4 := CreateNewNode(context.Background(), 7003, []string{node3.P2pNode.Addrs()[0].String() + "/ipfs/" + node3.P2pNode.ID().Pretty()})
	node5 := CreateNewNode(context.Background(), 7004, []string{node4.P2pNode.Addrs()[0].String() + "/ipfs/" + node4.P2pNode.ID().Pretty(),
		node1.P2pNode.Addrs()[0].String() + "/ipfs/" + node1.P2pNode.ID().Pretty()})

	time.Sleep(time.Second)

	node1.OnMsgRecv = recv2
	node2.OnMsgRecv = recv2
	node3.OnMsgRecv = recv2
	node4.OnMsgRecv = recv2
	node5.OnMsgRecv = recv2

	fmt.Println()
	fmt.Println()

	node1.Broadcast("Boo", []string{})

	time.Sleep(time.Second)

	if counter1 != 4 {
		t.Fatal("Should have been 4 (traversed 4 peers)")
	}

	node1.P2pNode.Close()
	node2.P2pNode.Close()
	node3.P2pNode.Close()
	node4.P2pNode.Close()
	node5.P2pNode.Close()
}
