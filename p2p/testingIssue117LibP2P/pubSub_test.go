package testingIssue117LibP2P

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-conn-security-multistream"
	"github.com/libp2p/go-floodsub"
	"github.com/libp2p/go-libp2p-crypto"
	cr "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-metrics"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/libp2p/go-libp2p-secio"
	"github.com/libp2p/go-libp2p-swarm"
	"github.com/libp2p/go-libp2p-transport-upgrader"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-tcp-transport"
	"github.com/multiformats/go-multiaddr"
	"github.com/whyrusleeping/go-smux-multistream"
	"github.com/whyrusleeping/go-smux-yamux"
)

type notifier struct {
}

// Listen is called when network starts listening on an addr
func (n *notifier) Listen(netw net.Network, ma multiaddr.Multiaddr) {}

// ListenClose is called when network starts listening on an addr
func (n *notifier) ListenClose(netw net.Network, ma multiaddr.Multiaddr) {}

// Connected is called when a connection opened
func (n *notifier) Connected(netw net.Network, conn net.Conn) {
	//retain max 4 connections
	if len(netw.Conns()) > 4 {
		conn.Close()
		fmt.Printf("Connection refused for peer: %v!\n", conn.RemotePeer().Pretty())
	}
}

// Disconnected is called when a connection closed
func (n *notifier) Disconnected(netw net.Network, conn net.Conn) {}

// OpenedStream is called when a stream opened
func (n *notifier) OpenedStream(netw net.Network, stream net.Stream) {}

// ClosedStream is called when a stream was closed
func (n *notifier) ClosedStream(netw net.Network, stream net.Stream) {}

type Messenger struct {
	ctx     context.Context
	p2pNode host.Host
	pubsub  *floodsub.PubSub
}

//helper func
func genUpgrader(n *swarm.Swarm) *stream.Upgrader {
	id := n.LocalPeer()
	pk := n.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(secio.ID, &secio.Transport{
		LocalID:    id,
		PrivateKey: pk,
	})

	stMuxer := multistream.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", sm_yamux.DefaultTransport)

	return &stream.Upgrader{
		Secure:  secMuxer,
		Muxer:   stMuxer,
		Filters: n.Filters,
	}
}

//helper func
func genSwarm(ctx context.Context, sKey crypto.PrivKey, pKey crypto.PubKey, id peer.ID, addr multiaddr.Multiaddr) (*swarm.Swarm, error) {
	ps := peerstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewPeerMetadata())
	ps.AddPubKey(id, pKey)
	ps.AddPrivKey(id, sKey)
	s := swarm.NewSwarm(ctx, id, ps, metrics.NewBandwidthCounter())

	tcpTransport := tcp.NewTCPTransport(genUpgrader(s))
	tcpTransport.DisableReuseport = false

	err := s.AddTransport(tcpTransport)
	if err != nil {
		return nil, err
	}

	err = s.Listen(addr)
	if err != nil {
		return nil, err
	}

	s.Peerstore().AddAddrs(id, s.ListenAddresses(), peerstore.PermanentAddrTTL)

	return s, nil
}

// NewMessenger creates a new messenger having as seed only the port
func NewMessenger(port int, refuseConns bool) (*Messenger, error) {
	mesgr := Messenger{ctx: context.Background()}

	//create private, public keys + id
	r := rand.New(rand.NewSource(int64(port)))

	prvKey, err := ecdsa.GenerateKey(btcec.S256(), r)

	if err != nil {
		panic(err)
	}

	k := (*cr.Secp256k1PrivateKey)(prvKey)

	sKey := k
	pKey := k.GetPublic()
	id, err := peer.IDFromPublicKey(pKey)
	if err != nil {
		return nil, err
	}

	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	//create host
	netw, err := genSwarm(mesgr.ctx, sKey, pKey, id, addr)
	if err != nil {
		return nil, err
	}
	mesgr.p2pNode = basichost.New(netw)

	//create pubsub
	pubsub, err := floodsub.NewFloodSub(mesgr.ctx, mesgr.p2pNode) //floodsub.NewGossipSub(ctx, node.p2pNode)
	if err != nil {
		return nil, err
	}
	mesgr.pubsub = pubsub

	if refuseConns {
		mesgr.p2pNode.Network().Notify(&notifier{})
	}

	return &mesgr, nil
}

func (mesgr *Messenger) ConnectToAddr(addr string) error {
	pinfo, err := mesgr.parseAddressIpfs(addr)
	if err != nil {
		return err
	}

	if err := mesgr.p2pNode.Connect(context.Background(), *pinfo); err != nil {
		return err
	}

	return nil
}

// ParseAddressIpfs translates the string containing the address of the node to a PeerInfo object
func (mesgr *Messenger) parseAddressIpfs(address string) (*peerstore.PeerInfo, error) {
	addr, err := ipfsaddr.ParseString(address)
	if err != nil {
		return nil, err
	}

	pinfo, err := peerstore.InfoFromP2pAddr(addr.Multiaddr())
	if err != nil {
		return nil, err
	}

	return pinfo, nil
}

func (mesgr *Messenger) FirstAddress() string {
	return mesgr.p2pNode.Addrs()[0].String() + "/ipfs/" + mesgr.p2pNode.ID().Pretty()
}

func TestPubSub(t *testing.T) {
	t.Skip("libP2P go-floodsub issue #117")

	//create node1 that refuses connections
	nodeMain, err := NewMessenger(4000, true)
	if err != nil {
		t.Fail()
	}

	noOfNodes := 1

	subscriberMain, err := nodeMain.pubsub.Subscribe("test")
	if err != nil {
		t.Fail()
	}

	go func() {
		for {
			msg, _ := subscriberMain.Next(nodeMain.ctx)

			fmt.Printf("Got message: %v\n", msg.GetData())
		}
	}()

	//create a data race condition that asynchronously close connection
	go func() {
		for {
			if len(nodeMain.p2pNode.Network().Conns()) > 0 {
				nodeMain.p2pNode.Network().Conns()[0].Close()
			}

			time.Sleep(time.Millisecond * 100)

		}
	}()

	nodes := make([]*Messenger, 0)

	for i := 0; i < noOfNodes; i++ {
		node, err := NewMessenger(4001+i, false)
		if err != nil {
			t.Fail()
		}

		subscriber, err := node.pubsub.Subscribe("test")
		if err != nil {
			t.Fail()
		}

		go func() {
			for {
				msg, _ := subscriber.Next(node.ctx)

				fmt.Printf("Got message: %v\n", msg.GetData())
			}
		}()

		nodes = append(nodes, node)
	}

	//try ten times
	for i := 0; i < 10; i++ {
		for j := 0; j < len(nodes); j++ {
			nodes[j].ConnectToAddr(nodeMain.FirstAddress())
		}

		time.Sleep(time.Second)
	}
}
