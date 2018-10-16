package p2p

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-conn-security-multistream"
	"github.com/libp2p/go-floodsub"
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
	"github.com/whyrusleeping/go-smux-multistream"
	"github.com/whyrusleeping/go-smux-yamux"
)

// NewMessenger implements a libP2P node with added functionality
type NewMessenger struct {
	p2pNode host.Host
	pubsub  *floodsub.PubSub
	ctx     context.Context

	marsh marshal.Marshalizer
}

func _genUpgrader(n *swarm.Swarm) *stream.Upgrader {
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

func _genSwarm(ctx context.Context, cp ConnectParams) (*swarm.Swarm, error) {
	ps := peerstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewPeerMetadata())
	ps.AddPubKey(cp.ID, cp.PubKey)
	ps.AddPrivKey(cp.ID, cp.PrivKey)
	s := swarm.NewSwarm(ctx, cp.ID, ps, metrics.NewBandwidthCounter())

	tcpTransport := tcp.NewTCPTransport(genUpgrader(s))
	tcpTransport.DisableReuseport = false

	err := s.AddTransport(tcpTransport)
	if err != nil {
		return nil, err
	}

	err = s.Listen(cp.Addr)
	if err != nil {
		return nil, err
	}

	s.Peerstore().AddAddrs(cp.ID, s.ListenAddresses(), peerstore.PermanentAddrTTL)

	return s, nil
}

// NewNewMessenger creates a new instance of NetMessenger.
func NewNewMessenger(ctx context.Context, mrsh marshal.Marshalizer, cp ConnectParams, addresses []string, maxAllowedPeers int) (*NewMessenger, error) {
	if mrsh == nil {
		return nil, errors.New("marshalizer is nil! Can't create node")
	}

	var node NewMessenger
	node.marsh = mrsh
	node.ctx = ctx

	timeStart := time.Now()

	netw, err := genSwarm(ctx, cp)
	if err != nil {
		return nil, err
	}
	h := basichost.New(netw)

	fmt.Printf("Node: %v has the following addr table: \n", h.ID().Pretty())

	for i, addr := range h.Addrs() {
		fmt.Printf("%d: %s/ipfs/%s\n", i, addr, h.ID().Pretty())
	}
	fmt.Println()

	fmt.Printf("Created node in %v\n", time.Now().Sub(timeStart))

	node.p2pNode = h
	pubsub, err := floodsub.NewFloodSub(ctx, node.p2pNode)
	if err != nil {
		return nil, err
	}
	node.pubsub = pubsub

	return &node, nil
}

func (nm *NewMessenger) GetIpfsAddress() string {
	return fmt.Sprintf("%s/ipfs/%s", nm.p2pNode.Addrs()[0], nm.p2pNode.ID().Pretty())
}

func (nm *NewMessenger) RegisterTopic(topicName string) error {
	subscriber, err := nm.pubsub.Subscribe(topicName)

	if err != nil {
		return err
	}

	go func(topicName string, id peer.ID) {
		for {
			msg, _ := subscriber.Next(nm.ctx)

			fmt.Printf("%v got the message: %v\n", id.Pretty(), msg.GetData())
		}
	}(topicName, nm.p2pNode.ID())

	return nil
}

func (nm *NewMessenger) PublishOnTopic(topicName string, data []byte) error {
	return nm.pubsub.Publish(topicName, data)
}

// Conns return the connections made by this memory messenger
func (nm *NewMessenger) Conns() []net.Conn {
	return nm.p2pNode.Network().Conns()
}

// ConnectToAddresses is used to explicitly connect to a well known set of addresses
func (nm *NewMessenger) ConnectToAddresses(ctx context.Context, addresses []string) {
	peers := 0

	timeStart := time.Now()

	for i := 0; i < len(addresses); i++ {
		pinfo, err := nm.ParseAddressIpfs(addresses[i])

		if err != nil {
			fmt.Printf("Bootstrapping the peer '%v' failed with error %v\n", addresses[i], err)
			continue
		}

		if err := nm.p2pNode.Connect(ctx, *pinfo); err != nil {
			fmt.Printf("Bootstrapping the peer '%v' failed with error %v\n", addresses[i], err)
			continue
		}

		peers++
	}

	fmt.Printf("Connected to %d peers in %v\n", peers, time.Now().Sub(timeStart))
}

// ParseAddressIpfs translates the string containing the address of the node to a PeerInfo object
func (nm *NewMessenger) ParseAddressIpfs(address string) (*peerstore.PeerInfo, error) {
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

// PrintConnected displays the connected peers
func (nm *NewMessenger) PrintConnected() {
	conns := nm.Conns()

	fmt.Printf("Node %s is connected to: \n", nm.p2pNode.ID().Pretty())

	for i := 0; i < len(conns); i++ {
		fmt.Printf("\t- %s with distance %d\n", conns[i].RemotePeer().Pretty(),
			ComputeDistanceAD(nm.p2pNode.ID(), conns[i].RemotePeer()))
	}
}
