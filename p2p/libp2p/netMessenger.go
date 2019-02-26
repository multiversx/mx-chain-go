package libp2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-interface-connmgr"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-pubsub"
	discovery2 "github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/multiformats/go-multiaddr"
)

const elrondRandezVousString = "ElrondNetworkRandezVous"
const durationBetweenSends = time.Duration(time.Microsecond * 10)

// durMdnsCalls is used to define the duration used by mdns service when polling peers
const durationMdnsCalls = time.Second

// DirectSendID represents the protocol ID for sending and receiving direct P2P messages
const DirectSendID = protocol.ID("/directsend/1.0.0")

var log = logger.NewDefaultLogger()

// PeerInfoHandler is the signature of the handler that gets called whenever an action for a peerInfo is triggered
type PeerInfoHandler func(pInfo peerstore.PeerInfo)

type networkMessenger struct {
	ctx               context.Context
	hostP2P           host.Host
	pb                *pubsub.PubSub
	ds                p2p.DirectSender
	kadDHT            *dht.IpfsDHT
	discoverer        discovery.Discoverer
	peerDiscoveryType p2p.PeerDiscoveryType
	mutMdns           sync.Mutex
	mdns              discovery2.Service

	mutTopics sync.RWMutex
	topics    map[string]p2p.MessageProcessor
	// preconnectPeerHandler is used for notifying that a peer wants to connect to another peer so
	// in the case of mocknet use, mocknet should first link the peers
	preconnectPeerHandler PeerInfoHandler
	peerDiscoveredHandler PeerInfoHandler

	outgoingPLB p2p.PipeLoadBalancer
}

// NewNetworkMessenger creates a libP2P messenger by opening a port on the current machine
// Should be used in production!
func NewNetworkMessenger(
	ctx context.Context,
	port int,
	p2pPrivKey crypto.PrivKey,
	conMgr ifconnmgr.ConnManager,
	outgoingPLB p2p.PipeLoadBalancer,
	peerDiscoveryType p2p.PeerDiscoveryType,
) (*networkMessenger, error) {

	if ctx == nil {
		return nil, p2p.ErrNilContext
	}

	if port < 1 {
		return nil, p2p.ErrInvalidPort
	}

	if p2pPrivKey == nil {
		return nil, p2p.ErrNilP2PprivateKey
	}

	if outgoingPLB == nil {
		return nil, p2p.ErrNilPipeLoadBalancer
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.Identity(p2pPrivKey),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
		libp2p.ConnectionManager(conMgr),
	}

	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	p2pNode, err := createMessenger(ctx, h, true, outgoingPLB, peerDiscoveryType)
	if err != nil {
		log.LogIfError(h.Close())
		return nil, err
	}

	return p2pNode, nil
}

func createMessenger(
	ctx context.Context,
	h host.Host,
	withSigning bool,
	outgoingPLB p2p.PipeLoadBalancer,
	peerDiscoveryType p2p.PeerDiscoveryType,
) (*networkMessenger, error) {

	pb, err := createPubSub(ctx, h, withSigning)
	if err != nil {
		return nil, err
	}

	netMes := networkMessenger{
		hostP2P:           h,
		pb:                pb,
		topics:            make(map[string]p2p.MessageProcessor),
		ctx:               ctx,
		outgoingPLB:       outgoingPLB,
		peerDiscoveryType: peerDiscoveryType,
	}

	err = netMes.applyDiscoveryMechanism(peerDiscoveryType)
	if err != nil {
		return nil, err
	}

	netMes.ds, err = NewDirectSender(ctx, h, netMes.directMessageHandler)
	if err != nil {
		return nil, err
	}

	go func(pubsub *pubsub.PubSub, plb p2p.PipeLoadBalancer) {
		for {
			dataToBeSent := plb.CollectFromPipes()

			wasSent := false
			for i := 0; i < len(dataToBeSent); i++ {
				sendableData := dataToBeSent[i]

				if sendableData == nil {
					continue
				}

				_ = pb.Publish(sendableData.Topic, sendableData.Buff)
				wasSent = true

				time.Sleep(durationBetweenSends)
			}

			//if nothing was sent over the network, it makes sense to sleep for a bit
			//as to not make this for loop iterate at max CPU speed
			if !wasSent {
				time.Sleep(durationBetweenSends)
			}
		}
	}(pb, netMes.outgoingPLB)

	for _, address := range netMes.hostP2P.Addrs() {
		fmt.Println(address.String() + "/p2p/" + netMes.ID().Pretty())
	}

	return &netMes, nil
}

func createPubSub(ctx context.Context, host host.Host, withSigning bool) (*pubsub.PubSub, error) {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(withSigning),
	}

	ps, err := pubsub.NewGossipSub(ctx, host, optsPS...)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (netMes *networkMessenger) applyDiscoveryMechanism(peerDiscoveryType p2p.PeerDiscoveryType) error {
	switch peerDiscoveryType {
	case p2p.PeerDiscoveryKadDht:
		return netMes.createKadDHT(elrondRandezVousString)
	case p2p.PeerDiscoveryMdns:
		return nil
	case p2p.PeerDiscoveryOff:
		return nil
	default:
		return p2p.ErrPeerDiscoveryNotImplemented
	}
}

// HandlePeerFound updates the routing table with this new peer
func (netMes *networkMessenger) HandlePeerFound(pi peerstore.PeerInfo) {
	peers := netMes.hostP2P.Peerstore().Peers()
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
		netMes.hostP2P.Peerstore().AddAddr(pi.ID, pi.Addrs[i], peerstore.PermanentAddrTTL)
	}

	//will try to connect for now as the connections and peer filtering is not done yet
	//TODO design a connection manager component
	go func() {
		err := netMes.hostP2P.Connect(netMes.ctx, pi)

		if err != nil {
			log.Debug(err.Error())
		}
	}()
}

func (netMes *networkMessenger) createKadDHT(randezvous string) error {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(netMes.ctx, netMes.hostP2P)
	if err != nil {
		return err
	}

	if err = kademliaDHT.Bootstrap(netMes.ctx); err != nil {
		return err
	}

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	log.Debug("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(netMes.ctx, routingDiscovery, randezvous)
	log.Debug("Successfully announced!")

	netMes.kadDHT = kademliaDHT
	netMes.discoverer = routingDiscovery

	return nil
}

// Close closes the host, connections and streams
func (netMes *networkMessenger) Close() error {
	if netMes.kadDHT != nil {
		err := netMes.kadDHT.Close()
		log.LogIfError(err)
	}

	netMes.mutMdns.Lock()
	if netMes.mdns != nil {
		err := netMes.mdns.Close()
		log.LogIfError(err)
	}
	netMes.mutMdns.Unlock()

	return netMes.hostP2P.Close()
}

// ID returns the messenger's ID
func (netMes *networkMessenger) ID() p2p.PeerID {
	return p2p.PeerID(netMes.hostP2P.ID())
}

// Peers returns the list of all known peers ID (including self)
func (netMes *networkMessenger) Peers() []p2p.PeerID {
	peers := make([]p2p.PeerID, 0)

	for _, p := range netMes.hostP2P.Peerstore().Peers() {
		peers = append(peers, p2p.PeerID(p))
	}
	return peers
}

// Addresses returns all addresses found in peerstore
func (netMes *networkMessenger) Addresses() []string {
	addrs := make([]string, 0)

	for _, address := range netMes.hostP2P.Addrs() {
		addrs = append(addrs, address.String()+"/p2p/"+netMes.ID().Pretty())
	}

	return addrs
}

// ConnectToPeer tries to open a new connection to a peer
func (netMes *networkMessenger) ConnectToPeer(address string) error {
	multiAddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return err
	}

	pInfo, err := peerstore.InfoFromP2pAddr(multiAddr)
	if err != nil {
		return err
	}

	if netMes.preconnectPeerHandler != nil {
		netMes.preconnectPeerHandler(*pInfo)
	}

	return netMes.hostP2P.Connect(netMes.ctx, *pInfo)
}

// KadDhtDiscoverNewPeers starts a blocking function that searches for all known peers querying all connected peers
// The default libp2p kad-dht implementation tries to connect to all of them
func (netMes *networkMessenger) KadDhtDiscoverNewPeers() error {
	if netMes.discoverer == nil {
		return p2p.ErrNilDiscoverer
	}

	peerChan, err := netMes.discoverer.FindPeers(netMes.ctx, elrondRandezVousString)
	if err != nil {
		return err
	}

	for {
		pInfo, more := <-peerChan

		if !more {
			//discovered peers channel closed
			break
		}

		handler := netMes.peerDiscoveredHandler

		if handler != nil {
			handler(pInfo)
		}
	}

	return nil
}

// TrimConnections will trigger a manual sweep onto current connection set reducing the
// number of connections if needed
func (netMes *networkMessenger) TrimConnections() {
	netMes.hostP2P.ConnManager().TrimOpenConns(netMes.ctx)
}

// Bootstrap will start the peer discovery mechanism
func (netMes *networkMessenger) Bootstrap() error {
	if netMes.peerDiscoveryType == p2p.PeerDiscoveryMdns {
		netMes.mutMdns.Lock()
		defer netMes.mutMdns.Unlock()

		if netMes.mdns != nil {
			return p2p.ErrPeerDiscoveryProcessAlreadyStarted
		}

		mdns, err := discovery2.NewMdnsService(
			netMes.ctx,
			netMes.hostP2P,
			durationMdnsCalls,
			"discovery")

		if err != nil {
			return err
		}

		mdns.RegisterNotifee(netMes)
		netMes.mdns = mdns
		return nil
	}

	return nil
}

// IsConnected returns true if current node is connected to provided peer
func (netMes *networkMessenger) IsConnected(peerID p2p.PeerID) bool {
	connectedness := netMes.hostP2P.Network().Connectedness(peer.ID(peerID))

	return connectedness == net.Connected
}

// ConnectedPeers returns the current connected peers list
func (netMes *networkMessenger) ConnectedPeers() []p2p.PeerID {

	connectedPeers := make(map[p2p.PeerID]struct{})

	for _, conn := range netMes.hostP2P.Network().Conns() {
		p := p2p.PeerID(conn.RemotePeer())

		if netMes.IsConnected(p) {
			connectedPeers[p] = struct{}{}
		}
	}

	peerList := make([]p2p.PeerID, len(connectedPeers))

	index := 0
	for k := range connectedPeers {
		peerList[index] = k
		index++
	}

	return peerList
}

// CreateTopic opens a new topic using pubsub infrastructure
func (netMes *networkMessenger) CreateTopic(name string, createPipeForTopic bool) error {
	netMes.mutTopics.Lock()
	_, found := netMes.topics[name]
	if found {
		netMes.mutTopics.Unlock()
		return p2p.ErrTopicAlreadyExists
	}

	netMes.topics[name] = nil
	subscrRequest, err := netMes.pb.Subscribe(name)
	if err != nil {
		netMes.mutTopics.Unlock()
		return err
	}
	netMes.mutTopics.Unlock()

	if createPipeForTopic {
		err = netMes.outgoingPLB.AddPipe(name)
	}

	//just a dummy func to consume messages received by the newly created topic
	go func() {
		for {
			_, _ = subscrRequest.Next(netMes.ctx)
		}
	}()

	return err
}

// HasTopic returns true if the topic has been created
func (netMes *networkMessenger) HasTopic(name string) bool {
	netMes.mutTopics.RLock()
	_, found := netMes.topics[name]
	netMes.mutTopics.RUnlock()

	return found
}

// HasTopicValidator returns true if the topic has a validator set
func (netMes *networkMessenger) HasTopicValidator(name string) bool {
	netMes.mutTopics.RLock()
	validator, _ := netMes.topics[name]
	netMes.mutTopics.RUnlock()

	return validator != nil
}

// OutgoingPipeLoadBalancer returns the pipe load balancer object used by the messenger to send data
func (netMes *networkMessenger) OutgoingPipeLoadBalancer() p2p.PipeLoadBalancer {
	return netMes.outgoingPLB
}

// BroadcastOnPipe tries to send a byte buffer onto a topic using provided pipe
func (netMes *networkMessenger) BroadcastOnPipe(pipe string, topic string, buff []byte) {
	go func() {
		sendable := &p2p.SendableData{
			Buff:  buff,
			Topic: topic,
		}
		netMes.outgoingPLB.GetChannelOrDefault(pipe) <- sendable
	}()
}

// BroadcastOnTopicPipe tries to send a byte buffer onto a topic using the topic name as pipe
func (netMes *networkMessenger) Broadcast(topic string, buff []byte) {
	netMes.BroadcastOnPipe(topic, topic, buff)
}

// RegisterMessageProcessor registers a message process on a topic
func (netMes *networkMessenger) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	if handler == nil {
		return p2p.ErrNilValidator
	}

	netMes.mutTopics.Lock()
	defer netMes.mutTopics.Unlock()
	validator, found := netMes.topics[topic]

	if !found {
		return p2p.ErrNilTopic
	}

	if validator != nil {
		return p2p.ErrTopicValidatorOperationNotSupported
	}

	err := netMes.pb.RegisterTopicValidator(topic, func(i context.Context, message *pubsub.Message) bool {
		err := handler.ProcessReceivedMessage(NewMessage(message))

		return err == nil
	})
	if err != nil {
		return err
	}

	netMes.topics[topic] = handler
	return nil
}

// UnregisterMessageProcessor registers a message processes on a topic
func (netMes *networkMessenger) UnregisterMessageProcessor(topic string) error {
	netMes.mutTopics.Lock()
	defer netMes.mutTopics.Unlock()
	validator, found := netMes.topics[topic]

	if !found {
		return p2p.ErrNilTopic
	}

	if validator == nil {
		return p2p.ErrTopicValidatorOperationNotSupported
	}

	err := netMes.pb.UnregisterTopicValidator(topic)
	if err != nil {
		return err
	}

	netMes.topics[topic] = nil
	return nil
}

// SendToConnectedPeer sends a direct message to a connected peer
func (netMes *networkMessenger) SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	return netMes.ds.Send(topic, buff, peerID)
}

func (netMes *networkMessenger) directMessageHandler(message p2p.MessageP2P) error {
	var processor p2p.MessageProcessor

	netMes.mutTopics.RLock()
	processor = netMes.topics[message.TopicIDs()[0]]
	netMes.mutTopics.RUnlock()

	if processor == nil {
		return p2p.ErrNilValidator
	}

	go func(msg p2p.MessageP2P) {
		log.LogIfError(processor.ProcessReceivedMessage(msg))
	}(message)

	return nil
}
