package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-conn-security-multistream"
	"github.com/libp2p/go-floodsub"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-metrics"
	libP2PNet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-secio"
	"github.com/libp2p/go-libp2p-swarm"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-tcp-transport"
	"github.com/multiformats/go-multiaddr"
	msmux "github.com/whyrusleeping/go-smux-multistream"
	yamux "github.com/whyrusleeping/go-smux-yamux"
)

// NetMessenger implements a libP2P node with added functionality
type NetMessenger struct {
	protocol protocol.ID

	context context.Context

	p2pNode host.Host

	mutChansSend sync.RWMutex
	chansSend    map[string]chan []byte

	mutBootstrap sync.Mutex

	queue *MessageQueue
	marsh marshal.Marshalizer

	rt *RoutingTable
	cn *ConnNotifier

	mdns discovery.Service
	dn   *DiscoveryNotifier

	//onMsgRecv func(caller Messenger, peerID string, m *Message)

	mutClosed sync.RWMutex
	closed    bool

	pubsub      *floodsub.PubSub
	topicHolder *TopicHolder
}

func genUpgrader(n *swarm.Swarm) *tptu.Upgrader {
	id := n.LocalPeer()
	pk := n.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(secio.ID, &secio.Transport{
		LocalID:    id,
		PrivateKey: pk,
	})

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)

	return &tptu.Upgrader{
		Secure:  secMuxer,
		Muxer:   stMuxer,
		Filters: n.Filters,
	}
}

func genSwarm(ctx context.Context, cp ConnectParams) (*swarm.Swarm, error) {
	ps := pstore.NewPeerstore(pstoremem.NewKeyBook(), pstoremem.NewAddrBook(), pstoremem.NewPeerMetadata())
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

	s.Peerstore().AddAddrs(cp.ID, s.ListenAddresses(), pstore.PermanentAddrTTL)

	return s, nil
}

// NewNetMessenger creates a new instance of NetMessenger.
func NewNetMessenger(ctx context.Context, mrsh marshal.Marshalizer, cp ConnectParams, addresses []string, maxAllowedPeers int) (*NetMessenger, error) {
	if mrsh == nil {
		return nil, errors.New("marshalizer is nil! Can't create node")
	}

	var node NetMessenger
	node.marsh = mrsh
	node.context = ctx
	node.cn = NewConnNotifier(&node)
	node.cn.MaxAllowedPeers = maxAllowedPeers

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
	pubsub, err := floodsub.NewFloodSub(ctx, node.p2pNode) //floodsub.NewGossipSub(ctx, node.p2pNode)
	if err != nil {
		return nil, err
	}
	node.pubsub = pubsub

	node.topicHolder = NewTopicHolder()
	node.topicHolder.OnNeedToRegisterTopic = node.registerTopicHolder
	node.topicHolder.OnNeedToSendDataOnTopic = node.onNeedToSendDataOnTopic

	node.chansSend = make(map[string]chan []byte)
	node.queue = NewMessageQueue(50000)
	node.protocol = protocol.ID("elrondnetwork/1.0.0.0")

	node.ConnectToAddresses(ctx, addresses)

	node.rt = NewRoutingTable(h.ID())
	//register the notifier
	node.p2pNode.Network().Notify(node.cn)
	node.cn.OnDoSimpleTask = func(this interface{}) {
		TaskMonitorConnections(node.cn)
	}
	node.cn.OnGetKnownPeers = func(cn *ConnNotifier) []peer.ID {
		return cn.Msgr.RouteTable().NearestPeersAll()
	}
	node.cn.OnNeedToConn = func(cn *ConnNotifier, pid peer.ID) error {
		pinfo := node.p2pNode.Peerstore().PeerInfo(pid)

		if err := node.p2pNode.Connect(ctx, pinfo); err != nil {
			return err
		}

		return nil
	}

	node.mutClosed = sync.RWMutex{}
	node.closed = false

	return &node, nil
}

func (nm *NetMessenger) registerTopicHolder(topicName string) error {
	subscriber, err := nm.pubsub.Subscribe(topicName)

	if err != nil {
		return err
	}

	go func(topicName string) {
		for {
			msg, _ := subscriber.Next(nm.context)

			nm.topicHolder.GotNewMessage(topicName, msg.GetData())
		}
	}(topicName)

	return nil
}

func (nm *NetMessenger) onNeedToSendDataOnTopic(topicName string, buff []byte) error {
	return nm.pubsub.Publish(topicName, buff)
}

// Closes a NetMessenger
func (nm *NetMessenger) Close() error {
	nm.mutClosed.Lock()
	nm.closed = true
	nm.mutClosed.Unlock()

	nm.p2pNode.Close()

	return nil
}

// ID returns the current id
func (nm *NetMessenger) ID() peer.ID {
	return nm.p2pNode.ID()
}

// Peers returns the connected peers list
func (nm *NetMessenger) Peers() []peer.ID {
	return nm.p2pNode.Peerstore().Peers()
}

// Conns return the connections made by this memory messenger
func (nm *NetMessenger) Conns() []libP2PNet.Conn {
	return nm.p2pNode.Network().Conns()
}

// Marshalizer returns the used marshalizer object
func (nm *NetMessenger) Marshalizer() marshal.Marshalizer {
	return nm.marsh
}

// RouteTable will return the RoutingTable object
func (nm *NetMessenger) RouteTable() *RoutingTable {
	return nm.rt
}

// Addrs will return all addresses bind to current messenger
func (nm *NetMessenger) Addrs() []string {
	addrs := make([]string, 0)

	for _, adrs := range nm.p2pNode.Addrs() {
		addrs = append(addrs, adrs.String()+"/ipfs/"+nm.ID().Pretty())
	}

	return addrs
}

// ConnectToAddresses is used to explicitly connect to a well known set of addresses
func (nm *NetMessenger) ConnectToAddresses(ctx context.Context, addresses []string) {
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

// Bootstrap will try to connect to as many peers as possible
func (nm *NetMessenger) Bootstrap(ctx context.Context) {
	nm.mutClosed.RLock()
	if nm.closed {
		nm.mutClosed.RUnlock()
		return
	}
	nm.mutClosed.RUnlock()

	nm.mutBootstrap.Lock()
	defer nm.mutBootstrap.Unlock()

	if nm.mdns != nil {
		return
	}

	nm.dn = NewDiscoveryNotifier(nm)

	mdns, err := discovery.NewMdnsService(context.Background(), nm.p2pNode, time.Second, "discovery")

	if err != nil {
		panic(err)
	}

	mdns.RegisterNotifee(nm.dn)
	nm.mdns = mdns

	wait := time.Second * 10
	fmt.Printf("\n**** Waiting %v to bootstrap...****\n\n", wait)

	nm.cn.Start()

	time.Sleep(wait)
}

// PrintConnected displays the connected peers
func (nm *NetMessenger) PrintConnected() {
	conns := nm.Conns()

	fmt.Printf("Node %s is connected to: \n", nm.ID().Pretty())

	for i := 0; i < len(conns); i++ {
		fmt.Printf("\t- %s with distance %d\n", conns[i].RemotePeer().Pretty(),
			ComputeDistanceAD(nm.ID(), conns[i].RemotePeer()))
	}
}

// AddAddr adds a new address to peer store
func (nm *NetMessenger) AddAddr(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	nm.p2pNode.Network().Peerstore().AddAddr(p, addr, ttl)
}

// Connectedness tests for a connection between self and another peer
func (nm *NetMessenger) Connectedness(pid peer.ID) libP2PNet.Connectedness {
	return nm.p2pNode.Network().Connectedness(pid)
}

// ParseAddressIpfs translates the string containing the address of the node to a PeerInfo object
func (nm *NetMessenger) ParseAddressIpfs(address string) (*pstore.PeerInfo, error) {
	addr, err := ipfsaddr.ParseString(address)
	if err != nil {
		return nil, err
	}

	pinfo, err := pstore.InfoFromP2pAddr(addr.Multiaddr())
	if err != nil {
		return nil, err
	}

	return pinfo, nil
}

func (nm *NetMessenger) TopicHolder() *TopicHolder {
	return nm.topicHolder
}
