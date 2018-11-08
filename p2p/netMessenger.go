package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/multiformats/go-multiaddr"
)

// maxMessageQueueNetMessenger is used to control the maximum message queue in a network messenger
const maxMessageQueueNetMessenger = 50000

// durMdnsCalls is used to define the duration used by mdns service when polling peers
const durMdnsCalls = time.Second

// PubSubStrategy defines the strategy for broadcasting messages in the network
type PubSubStrategy int

const (
	// FloodSub strategy to use when broadcasting messages
	FloodSub = iota
	// GossipSub strategy to use when broadcasting messages
	GossipSub
	// RandomSub strategy to use when broadcasting messages
	RandomSub
)

// NetMessenger implements a libP2P node with added functionality
type NetMessenger struct {
	context  context.Context
	protocol protocol.ID
	p2pNode  host.Host
	ps       *pubsub.PubSub
	mdns     discovery.Service

	mutChansSend sync.RWMutex
	chansSend    map[string]chan []byte

	mutBootstrap sync.Mutex

	queue  *MessageQueue
	marsh  marshal.Marshalizer
	hasher hashing.Hasher
	rt     *RoutingTable
	cn     *ConnNotifier
	dn     *DiscoveryNotifier

	onMsgRecv func(caller Messenger, peerID string, data []byte)

	mutClosed sync.RWMutex
	closed    bool

	mutTopics sync.RWMutex
	topics    map[string]*Topic
}

// NewNetMessenger creates a new instance of NetMessenger.
func NewNetMessenger(ctx context.Context, marsh marshal.Marshalizer, hasher hashing.Hasher,
	cp *ConnectParams, maxAllowedPeers int, pubsubStrategy PubSubStrategy) (*NetMessenger, error) {

	if marsh == nil {
		return nil, errors.New("marshalizer is nil! Can't create node")
	}

	if hasher == nil {
		return nil, errors.New("hasher is nil! Can't create node")
	}

	node := NetMessenger{
		context: ctx,
		marsh:   marsh,
		hasher:  hasher,
		topics:  make(map[string]*Topic, 0),
	}

	node.cn = NewConnNotifier(&node)
	node.cn.MaxAllowedPeers = maxAllowedPeers

	timeStart := time.Now()

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cp.Port)),
		libp2p.Identity(cp.PrivKey),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	}

	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Node: %v has the following addr table: \n", h.ID().Pretty())
	for i, addr := range h.Addrs() {
		fmt.Printf("%d: %s/ipfs/%s\n", i, addr, h.ID().Pretty())
	}
	fmt.Println()

	fmt.Printf("Created node in %v\n", time.Now().Sub(timeStart))

	node.p2pNode = h
	node.chansSend = make(map[string]chan []byte)
	node.queue = NewMessageQueue(maxMessageQueueNetMessenger)

	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(true),
	}

	switch pubsubStrategy {
	case FloodSub:
		{
			ps, err := pubsub.NewFloodSub(ctx, h, optsPS...)
			if err != nil {
				return nil, err
			}
			node.ps = ps
		}
	case GossipSub:
		{
			ps, err := pubsub.NewGossipSub(ctx, h, optsPS...)
			if err != nil {
				return nil, err
			}
			node.ps = ps
		}
	case RandomSub:
		{
			ps, err := pubsub.NewRandomSub(ctx, h, optsPS...)
			if err != nil {
				return nil, err
			}
			node.ps = ps
		}
	default:
		return nil, errors.New("unknown pubsub strategy")
	}

	node.rt = NewRoutingTable(h.ID())
	//register the notifier
	node.p2pNode.Network().Notify(node.cn)
	node.cn.OnDoSimpleTask = func(this interface{}) {
		TaskResolveConnections(node.cn)
	}
	node.cn.GetKnownPeers = func(cn *ConnNotifier) []peer.ID {
		return cn.Msgr.RouteTable().NearestPeersAll()
	}
	node.cn.ConnectToPeer = func(cn *ConnNotifier, pid peer.ID) error {
		pinfo := node.p2pNode.Peerstore().PeerInfo(pid)

		if err := node.p2pNode.Connect(ctx, pinfo); err != nil {
			return err
		}

		return nil
	}

	return &node, nil
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
func (nm *NetMessenger) Conns() []net.Conn {
	return nm.p2pNode.Network().Conns()
}

// Marshalizer returns the used marshalizer object
func (nm *NetMessenger) Marshalizer() marshal.Marshalizer {
	return nm.marsh
}

// Hasher returns the used object for hashing data
func (nm *NetMessenger) Hasher() hashing.Hasher {
	return nm.hasher
}

// RouteTable will return the RoutingTable object
func (nm *NetMessenger) RouteTable() *RoutingTable {
	return nm.rt
}

// Addrs will return all addresses bound to current messenger
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
	if nm.mdns != nil {
		//already started the bootstrap process, return
		nm.mutBootstrap.Unlock()
		return
	}

	nm.dn = NewDiscoveryNotifier(nm)

	mdns, err := discovery.NewMdnsService(context.Background(), nm.p2pNode, durMdnsCalls, "discovery")

	if err != nil {
		panic(err)
	}

	mdns.RegisterNotifee(nm.dn)
	nm.mdns = mdns

	nm.mutBootstrap.Unlock()

	nm.cn.Start()
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
func (nm *NetMessenger) Connectedness(pid peer.ID) net.Connectedness {
	return nm.p2pNode.Network().Connectedness(pid)
}

// ParseAddressIpfs translates the string containing the address of the node to a PeerInfo object
func (nm *NetMessenger) ParseAddressIpfs(address string) (*peerstore.PeerInfo, error) {
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

// AddTopic registers a new topic to this messenger
func (nm *NetMessenger) AddTopic(t *Topic) error {
	if t == nil {
		return errors.New("topic can not be nil")
	}

	subscr, err := nm.ps.Subscribe(t.Name)
	if err != nil {
		return err
	}

	nm.mutTopics.Lock()
	_, ok := nm.topics[t.Name]

	if ok {
		nm.mutTopics.Unlock()
		return errors.New("topic already exists")
	}

	nm.topics[t.Name] = t
	nm.mutTopics.Unlock()

	go func() {
		for {
			msg, err := subscr.Next(nm.context)

			if err != nil {
				continue
			}

			t.NewDataReceived(msg.GetData(), msg.GetFrom().Pretty())
		}
	}()

	t.SendData = func(data []byte) error {
		return nm.ps.Publish(t.Name, data)
	}

	t.registerTopicValidator = func(v pubsub.Validator) error {
		return nm.ps.RegisterTopicValidator(t.Name, v)
	}

	t.unregisterTopicValidator = func() error {
		return nm.ps.UnregisterTopicValidator(t.Name)
	}

	return nil
}

// GetTopic returns the topic from its name or nil if no topic with that name
// was ever registered
func (nm *NetMessenger) GetTopic(topicName string) *Topic {
	nm.mutTopics.RLock()
	defer nm.mutTopics.RUnlock()

	t, ok := nm.topics[topicName]

	if !ok {
		return nil
	}

	return t
}
