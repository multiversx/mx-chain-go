package p2p

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// durMdnsCalls is used to define the duration used by mdns service when polling peers
const durMdnsCalls = time.Second

// durTimeCache represents the duration for gossip messages to be saved in cache
const durTimeCache = time.Second * 5

// requestTopicSuffix is added to a known topic to generate the topic's request counterpart
const requestTopicSuffix = "_REQUEST"

// PubSubStrategy defines the strategy for broadcasting messages in the network
type PubSubStrategy int

const (
	// FloodSub strategy to use when broadcasting messages
	FloodSub = iota
	// GossipSub strategy to use when broadcasting messages
	GossipSub
)

type message struct {
	buff  []byte
	topic string
}

// NetMessenger implements a libP2P node with added functionality
type NetMessenger struct {
	context        context.Context
	protocol       protocol.ID
	p2pNode        host.Host
	ps             *pubsub.PubSub
	mdns           discovery.Service
	mutChansSend   sync.RWMutex
	chansSend      map[string]chan []byte
	mutBootstrap   sync.Mutex
	marsh          marshal.Marshalizer
	hasher         hashing.Hasher
	rt             *RoutingTable
	cn             *ConnNotifier
	mutClosed      sync.RWMutex
	closed         bool
	mutTopics      sync.RWMutex
	topics         map[string]*Topic
	mutGossipCache sync.RWMutex
	gossipCache    *TimeCache

	chSendMessages chan *message
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
		context:        ctx,
		marsh:          marsh,
		hasher:         hasher,
		topics:         make(map[string]*Topic, 0),
		mutGossipCache: sync.RWMutex{},
		gossipCache:    NewTimeCache(durTimeCache),
		chSendMessages: make(chan *message),
	}

	node.cn = NewConnNotifier(maxAllowedPeers)

	timeStart := time.Now()

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cp.Port)),
		libp2p.Identity(cp.PrivKey),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	}

	hostP2P, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = hostP2P.Close()
		}
	}()

	adrTable := fmt.Sprintf("Node: %v has the following addr table: \n", hostP2P.ID().Pretty())
	for i, addr := range hostP2P.Addrs() {
		adrTable = adrTable + fmt.Sprintf("%d: %s/ipfs/%s\n", i, addr, hostP2P.ID().Pretty())
	}
	log.Debug(adrTable)

	node.p2pNode = hostP2P
	node.chansSend = make(map[string]chan []byte)

	err = node.createPubSub(hostP2P, pubsubStrategy, ctx)
	if err != nil {
		return nil, err
	}

	node.rt = NewRoutingTable(hostP2P.ID())

	node.connNotifierRegistration(ctx)

	log.Debug(fmt.Sprintf("Created node in %v\n", time.Now().Sub(timeStart)))

	return &node, nil
}

func (nm *NetMessenger) createPubSub(hostP2P host.Host, pubsubStrategy PubSubStrategy, ctx context.Context) error {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(true),
	}

	switch pubsubStrategy {
	case FloodSub:
		{
			ps, err := pubsub.NewFloodSub(ctx, hostP2P, optsPS...)
			if err != nil {
				return err
			}
			nm.ps = ps
		}
	case GossipSub:
		{
			ps, err := pubsub.NewGossipSub(ctx, hostP2P, optsPS...)
			if err != nil {
				return err
			}
			nm.ps = ps
		}
	default:
		return errors.New("unknown pubsub strategy")
	}

	go func(ps *pubsub.PubSub, ch chan *message) {
		for {
			select {
			case msg := <-ch:
				err := ps.Publish(msg.topic, msg.buff)

				log.LogIfError(err)
			}

			time.Sleep(time.Microsecond * 100)
		}
	}(nm.ps, nm.chSendMessages)

	return nil
}

func (nm *NetMessenger) connNotifierRegistration(ctx context.Context) {
	//register the notifier
	nm.p2pNode.Network().Notify(nm.cn)

	nm.cn.GetKnownPeers = func(cn *ConnNotifier) []peer.ID {
		return nm.RouteTable().NearestPeersAll()
	}

	nm.cn.ConnectToPeer = func(cn *ConnNotifier, pid peer.ID) error {
		pinfo := nm.p2pNode.Peerstore().PeerInfo(pid)

		if err := nm.p2pNode.Connect(ctx, pinfo); err != nil {
			return err
		}

		return nil
	}

	nm.cn.GetConnections = func(sender *ConnNotifier) []net.Conn {
		return nm.Conns()
	}

	nm.cn.IsConnected = func(sender *ConnNotifier, pid peer.ID) bool {
		return nm.Connectedness(pid) == net.Connected
	}

	nm.cn.Start()
}

// Closes a NetMessenger
func (nm *NetMessenger) Close() error {
	nm.mutClosed.Lock()
	nm.closed = true
	nm.mutClosed.Unlock()

	errorEncountered := error(nil)

	if nm.mdns != nil {
		//unregistration and closing
		nm.mdns.UnregisterNotifee(nm)
		err := nm.mdns.Close()
		if err != nil {
			errorEncountered = err
		}
	}

	nm.cn.Stop()
	err := nm.p2pNode.Network().Close()
	if err != nil {
		errorEncountered = err
	}
	err = nm.p2pNode.Close()
	if err != nil {
		errorEncountered = err
	}

	nm.mdns = nil
	nm.cn = nil
	nm.ps = nil

	return errorEncountered
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

// Addresses will return all addresses bound to current messenger
func (nm *NetMessenger) Addresses() []string {
	addrs := make([]string, 0)

	for _, address := range nm.p2pNode.Addrs() {
		addrs = append(addrs, address.String()+"/ipfs/"+nm.ID().Pretty())
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
			log.Error(fmt.Sprintf("Bootstrapping the peer '%v' failed with error %v\n", addresses[i], err))
			continue
		}

		if err := nm.p2pNode.Connect(ctx, *pinfo); err != nil {
			log.Error(fmt.Sprintf("Bootstrapping the peer '%v' failed with error %v\n", addresses[i], err))
			continue
		}

		peers++
	}

	log.Debug(fmt.Sprintf("Connected to %d peers in %v\n", peers, time.Now().Sub(timeStart)))
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

	mdns, err := discovery.NewMdnsService(context.Background(), nm.p2pNode, durMdnsCalls, "discovery")

	if err != nil {
		panic(err)
	}

	mdns.RegisterNotifee(nm)
	nm.mdns = mdns

	nm.mutBootstrap.Unlock()

	nm.cn.Start()
}

// HandlePeerFound updates the routing table with this new peer
func (nm *NetMessenger) HandlePeerFound(pi peerstore.PeerInfo) {
	if nm.Peers() == nil {
		return
	}

	peers := nm.Peers()
	found := false

	for i := 0; i < len(peers); i++ {
		if peers[i] == pi.ID {
			found = true
			break
		}
	}

	if found {
		return
	}

	for i := 0; i < len(pi.Addrs); i++ {
		nm.AddAddress(pi.ID, pi.Addrs[i], peerstore.PermanentAddrTTL)
	}

	nm.RouteTable().Update(pi.ID)
}

// PrintConnected displays the connected peers
func (nm *NetMessenger) PrintConnected() {
	conns := nm.Conns()

	connectedTo := fmt.Sprintf("Node %s is connected to: \n", nm.ID().Pretty())
	for i := 0; i < len(conns); i++ {
		connectedTo = connectedTo + fmt.Sprintf("\t- %s with distance %d\n", conns[i].RemotePeer().Pretty(),
			ComputeDistanceAD(nm.ID(), conns[i].RemotePeer()))
	}

	log.Debug(connectedTo)
}

// AddAddress adds a new address to peer store
func (nm *NetMessenger) AddAddress(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
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
	//sanity checks
	if t == nil {
		return errors.New("topic can not be nil")
	}

	if strings.Contains(t.Name(), requestTopicSuffix) {
		return errors.New("topic name contains request suffix")
	}

	nm.mutTopics.Lock()

	_, ok := nm.topics[t.Name()]
	if ok {
		nm.mutTopics.Unlock()
		return errors.New("topic already exists")
	}

	nm.topics[t.Name()] = t
	t.CurrentPeer = nm.ID()
	nm.mutTopics.Unlock()

	subscr, err := nm.ps.Subscribe(t.Name())
	if err != nil {
		return err
	}

	subscrRequest, err := nm.ps.Subscribe(t.Name() + requestTopicSuffix)
	if err != nil {
		return err
	}

	// async func for passing received data to Topic object
	go func() {
		for {
			msg, err := subscr.Next(nm.context)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			obj, err := t.CreateObject(msg.GetData())
			if err != nil {
				log.Error(err.Error())
				continue
			}

			nm.mutGossipCache.Lock()
			if nm.gossipCache.Has(obj.ID()) {
				//duplicate object, skip
				nm.mutGossipCache.Unlock()
				continue
			}

			nm.gossipCache.Add(obj.ID())
			nm.mutGossipCache.Unlock()

			err = t.NewObjReceived(obj, msg.GetFrom().Pretty())
			if err != nil {
				log.Error(err.Error())
				continue
			}
		}
	}()

	// func that publishes on network from Topic object
	t.SendData = func(data []byte) error {
		nm.mutClosed.RLock()
		if nm.closed {
			nm.mutClosed.RUnlock()
			return nil
		}
		nm.mutClosed.RUnlock()

		go func(topicName string, buffer []byte) {
			nm.chSendMessages <- &message{
				buff:  buffer,
				topic: topicName,
			}
		}(t.Name(), data)

		return nil
	}

	// validator registration func
	t.RegisterTopicValidator = func(v pubsub.Validator) error {
		return nm.ps.RegisterTopicValidator(t.Name(), v)
	}

	// validator unregistration func
	t.UnregisterTopicValidator = func() error {
		return nm.ps.UnregisterTopicValidator(t.Name())
	}

	nm.createRequestTopicAndBind(t, subscrRequest)

	return nil
}

// createRequestTopicAndBind is used to wire-up the func pointers to the request channel created automatically
// it also implements a validator function for not broadcast the request if it can resolve
func (nm *NetMessenger) createRequestTopicAndBind(t *Topic, subscriberRequest *pubsub.Subscription) {
	// there is no need to have a function on received data
	// the logic will be called inside validator func as this is the first func called
	// and only if the result was nil, the validator actually let the message pass through its peers
	v := func(ctx context.Context, mes *pubsub.Message) bool {
		//resolver has not been set up, let the message go to the other peers, maybe they can resolve the request
		if t.ResolveRequest == nil {
			return true
		}

		//resolved payload
		buff := t.ResolveRequest(mes.GetData())

		if buff == nil {
			//object not found
			return true
		}

		//found object, no need to resend the request message to peers
		//test whether we also should broadcast the message (others might have broadcast it just before us)
		has := false

		nm.mutGossipCache.RLock()
		has = nm.gossipCache.Has(string(buff))
		nm.mutGossipCache.RUnlock()

		if !has {
			//only if the current peer did not receive an equal object to cloner,
			//then it shall broadcast it
			err := t.BroadcastBuff(buff)
			if err != nil {
				log.Error(err.Error())
			}
		}
		return false
	}

	//wire-up a plain func for publishing on request channel
	t.Request = func(hash []byte) error {
		go func(topicName string, buffer []byte) {
			nm.chSendMessages <- &message{
				buff:  buffer,
				topic: topicName,
			}
		}(t.Name()+requestTopicSuffix, hash)

		return nil
	}

	//wire-up the validator
	err := nm.ps.RegisterTopicValidator(t.Name()+requestTopicSuffix, v)
	if err != nil {
		log.Error(err.Error())
	}
}

// GetTopic returns the topic from its name or nil if no topic with that name
// was ever registered
func (nm *NetMessenger) GetTopic(topicName string) *Topic {
	nm.mutTopics.RLock()
	defer nm.mutTopics.RUnlock()

	if t, ok := nm.topics[topicName]; ok {
		return t
	}

	return nil
}
