package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"time"

	pubsub "github.com/ElrondNetwork/go-libp2p-pubsub"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

type messenger struct {
	host   host.Host
	pb     *pubsub.PubSub
	topic  *pubsub.Topic
	subscr *pubsub.Subscription
}

func newMessenger() *messenger {
	address := "/ip4/0.0.0.0/tcp/0"
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(address),
		libp2p.Identity(createP2PPrivKey()),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.DefaultTransports,
		// we need the disable relay option in order to save the node's bandwidth as much as possible
		libp2p.DisableRelay(),
		libp2p.NATPortMap(),
	}

	h, _ := libp2p.New(opts...)
	optsPS := make([]pubsub.Option, 0)
	pb, _ := pubsub.NewGossipSub(context.Background(), h, optsPS...)

	return &messenger{
		host: h,
		pb:   pb,
	}
}

func createP2PPrivKey() *libp2pCrypto.Secp256k1PrivateKey {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	return (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
}

func (m *messenger) connectTo(target *messenger) {
	addr := peer.AddrInfo{
		ID:    target.host.ID(),
		Addrs: target.host.Addrs(),
	}

	err := m.host.Connect(context.Background(), addr)
	if err != nil {
		fmt.Println("error connecting to peer: " + err.Error())
	}
}

func (m *messenger) joinTopic(topic string) {
	m.topic, _ = m.pb.Join(topic)
	m.subscr, _ = m.topic.Subscribe()

	go func() {
		for {
			msg, err := m.subscr.Next(context.Background())
			if err != nil {
				return
			}

			fmt.Printf("%s: got message %s\n", m.host.ID().Pretty(), string(msg.Data))
		}
	}()

}

func main() {
	fmt.Println("creating 8 host connected statically...")
	peers := create8ConnectedPeers()

	defer func() {
		for _, p := range peers {
			_ = p.host.Close()
		}
	}()

	fmt.Println()

	for _, p := range peers {
		p.joinTopic("test")
	}

	go func() {
		time.Sleep(time.Second * 2)
		// TODO uncomment these 2 lines to make the pubsub create connections
		// peers[3].subscr.Cancel()
		// _ = peers[3].topic.Close()
	}()

	for i := 0; i < 10; i++ {
		printConnections(peers)
		fmt.Println()
		time.Sleep(time.Second)
	}
}

func printConnections(peers []*messenger) {
	for _, p := range peers {
		fmt.Printf(" %s is connected to %d peers\n", p.host.ID().Pretty(), len(p.host.Network().Peers()))
	}
}

// create8ConnectedPeers assembles a network as following:
//
//                             0------------------- 1
//                             |                    |
//        2 ------------------ 3 ------------------ 4
//        |                    |                    |
//        5                    6                    7
func create8ConnectedPeers() []*messenger {
	peers := make([]*messenger, 0)
	for i := 0; i < 8; i++ {
		p := newMessenger()
		fmt.Printf("%d - created peer %s\n", i, p.host.ID().Pretty())

		peers = append(peers, p)
	}

	connections := map[int][]int{
		0: {1, 3},
		1: {4},
		2: {5, 3},
		3: {4, 6},
		4: {7},
	}

	createConnections(peers, connections)

	return peers
}

func createConnections(peers []*messenger, connections map[int][]int) {
	for pid, connectTo := range connections {
		connectPeerToOthers(peers, pid, connectTo)
	}
}

func connectPeerToOthers(peers []*messenger, idx int, connectToIdxes []int) {
	for _, connectToIdx := range connectToIdxes {
		peers[idx].connectTo(peers[connectToIdx])
	}
}
