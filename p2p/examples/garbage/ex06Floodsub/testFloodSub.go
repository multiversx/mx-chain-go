package ex06Floodsub

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-peer"
	"time"
)

func Main() {
	nodes := []Node{}

	i := 0

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeBootStrap := CreateNewNode(ctx, "", true)

	addr := nodeBootStrap.p2pNode.Addrs()[0].String() + "/ipfs/" + nodeBootStrap.p2pNode.ID().Pretty()
	fmt.Printf("Bootstrap node: %v\n", addr)

	//var ctxs []context.Context
	//var cancels []func()

	for i < 100 {
		//ctx1, cancel1 := context.WithCancel(context.Background())
		//defer cancel1()
		//
		//ctxs = append(ctxs, ctx1)
		//cancels = append(cancels, cancel1)

		nodes = append(nodes, *CreateNewNode(ctx, addr, i < 50))

		i++
	}

	time.Sleep(time.Duration(time.Second))

	i = 0

	for i < 32 {
		nodes[i].pubsub.Publish("stream1", []byte{byte(i), 65, 78})
		//time.Sleep(time.Duration(time.Millisecond))
		//nodes[i].pubsub.Publish("stream2", []byte{byte(i), 66, 89})
		//time.Sleep(time.Duration(time.Millisecond))

		i++
	}

	time.Sleep(time.Duration(time.Second * 5))

	i = 0

	for i < 100 {
		fmt.Printf("%v, topic stream1 found %v\n", nodes[i].p2pNode.ID().Pretty(), len(nodes[i].DataRcv["stream1"]))
		fmt.Printf("%v, topic stream2 found %v\n", nodes[i].p2pNode.ID().Pretty(), len(nodes[i].DataRcv["stream2"]))

		for _, peerid := range nodes[i].p2pNode.Peerstore().Peers() {
			fmt.Printf("%v is connected to %v\n", nodes[i].p2pNode.ID().Pretty(), peer.ID(peerid).Pretty())
		}
		fmt.Println()

		i++
	}

	time.Sleep(time.Duration(time.Second * 10))

}
