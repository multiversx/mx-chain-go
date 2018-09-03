package ex04TestDiscovery

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	mdns "github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/multiformats/go-multiaddr"
	mrand "math/rand"
	"time"
)

func genNodes(max int, randomSource *mrand.Rand) []host.Host {

	i := 0

	nodes := []host.Host{}

	for i < max {
		//generating a host
		prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomSource)

		if err != nil {
			panic(err)
		}

		sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 4000+i))

		host, err := libp2p.New(
			context.Background(),
			libp2p.ListenAddrs(sourceMultiAddr),
			libp2p.Identity(prvKey),
		)

		if err != nil {
			panic(err)
		}

		nodes = append(nodes, host)

		i++
	}

	return nodes
}

func genServs(nodes []host.Host) (servs []mdns.Service) {
	servs = []mdns.Service{}

	for _, node := range nodes {
		s, err := mdns.NewMdnsService(context.Background(), node, time.Second, "discovery")

		if err != nil {
			panic(err)
		}

		servs = append(servs, s)
	}

	return servs
}

type DiscoveryNotifee struct {
	h host.Host
}

func (n *DiscoveryNotifee) HandlePeerFound(pi pstore.PeerInfo) {
	n.h.Connect(context.Background(), pi)
	fmt.Printf("Host: %v connected to %v\n", n.h.ID().Pretty(), pi.ID.Pretty())
}

func Main() {
	randSource := mrand.New(mrand.NewSource(100))

	fmt.Println("Generating 10 nodes...")

	//slice of nodes
	nodes := genNodes(10, randSource)

	for _, node := range nodes {
		fmt.Printf("Generated node %v %v\n", node.ID().Pretty(), node.Addrs())
	}

	fmt.Println("Adding service discovery on each node...")

	servs := genServs(nodes)

	for idx, serv := range servs {
		n := &DiscoveryNotifee{nodes[idx]}

		serv.RegisterNotifee(n)
	}

	fmt.Println("Added service discovery! Connecting nodes...")

	nodes[0].Connect(context.Background(), pstore.PeerInfo{ID: nodes[1].ID()})

	time.Sleep(time.Duration(time.Second * 10))

}
