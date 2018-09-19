package ex06Floodsub

import (
	"context"
	"fmt"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-floodsub"
	bhost "github.com/libp2p/go-libp2p-blankhost"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-metrics"
	"github.com/libp2p/go-libp2p-peerstore"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	tu2 "github.com/libp2p/go-libp2p-swarm/testing"
	"github.com/libp2p/go-tcp-transport"
	tu "github.com/libp2p/go-testutil"
	"sync"
)

type Node struct {
	p2pNode host.Host
	pubsub  *floodsub.PubSub

	DataRcv map[string][]byte

	mut sync.Mutex
}

type config struct {
	disableReuseport bool
	dialOnly         bool
}

type Option func(*config)

var OptDisableReuseport Option = func(c *config) {
	c.disableReuseport = true
}

// OptDialOnly prevents the test swarm from listening.
var OptDialOnly Option = func(c *config) {
	c.dialOnly = true
}

func RandPeerNetParamsOrFatal() tu.PeerNetParams {
	p, err := tu.RandPeerNetParams()
	if err != nil {
		panic(err)
		return tu.PeerNetParams{} // TODO return nil
	}
	return *p
}

func GenSwarm(ctx context.Context, opts ...Option) *swarm.Swarm {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}

	p := RandPeerNetParamsOrFatal()

	ps := pstore.NewPeerstore()
	ps.AddPubKey(p.ID, p.PubKey)
	ps.AddPrivKey(p.ID, p.PrivKey)
	s := swarm.NewSwarm(ctx, p.ID, ps, metrics.NewBandwidthCounter())

	tcpTransport := tcp.NewTCPTransport(tu2.GenUpgrader(s))
	tcpTransport.DisableReuseport = cfg.disableReuseport

	if err := s.AddTransport(tcpTransport); err != nil {
		panic(err)
	}

	if !cfg.dialOnly {
		if err := s.Listen(p.Addr); err != nil {
			panic(err)
		}

		s.Peerstore().AddAddrs(p.ID, s.ListenAddresses(), pstore.PermanentAddrTTL)
	}

	return s
}

func CreateNewNode(ctx context.Context, addrToConnect string, listenOnSecond bool) *Node {
	var node Node

	netw := GenSwarm(ctx)
	h := bhost.NewBlankHost(netw)

	pubsub, err := floodsub.NewFloodSub(ctx, h)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Node: %v has the following addr table: \n", h.ID().Pretty())

	for i, addr := range h.Addrs() {
		fmt.Printf("%d: %s/ipfs/%s\n", i, addr, h.ID().Pretty())
	}
	fmt.Println()

	if addrToConnect != "" {
		addr, err := ipfsaddr.ParseString(addrToConnect)
		if err != nil {
			panic(err)
		}
		pinfo, _ := peerstore.InfoFromP2pAddr(addr.Multiaddr())

		if err := h.Connect(ctx, *pinfo); err != nil {
			fmt.Println("bootstrapping a peer failed", err)
		}
	}

	node.p2pNode = h
	node.pubsub = pubsub
	node.DataRcv = make(map[string][]byte)
	node.mut = sync.Mutex{}

	node.ListenStreams(ctx, "stream1")
	if listenOnSecond {
		node.ListenStreams(ctx, "stream2")
	}

	return &node
}

func (node *Node) ListenStreams(ctx context.Context, topic string) {
	sub, err := node.pubsub.Subscribe(topic)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				panic(err)
			}

			node.mut.Lock()
			data := node.DataRcv[topic]
			node.DataRcv[topic] = append(data, msg.GetData()[0])
			node.mut.Unlock()

			fmt.Printf("%v got on topic %v, message: %v\n", node.p2pNode.ID().Pretty(), topic, msg.GetData())
		}
	}()
}
