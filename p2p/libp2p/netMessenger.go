package libp2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
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
	"github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
)

const elrondRandezVousString = "ElrondNetworkRandezVous"
const durationBetweenSends = time.Duration(time.Microsecond * 10)

// DirectSendID represents the protocol ID for sending and receiving direct P2P messages
const DirectSendID = protocol.ID("/directsend/1.0.0")

var log = logger.NewDefaultLogger()

// PeerInfoHandler is the signature of the handler that gets called whenever an action for a peerInfo is triggered
type PeerInfoHandler func(pInfo peerstore.PeerInfo)

type networkMessenger struct {
	ctx        context.Context
	hostP2P    host.Host
	pb         *pubsub.PubSub
	ds         p2p.DirectSender
	kadDHT     *dht.IpfsDHT
	discoverer discovery.Discoverer

	mutTopics sync.RWMutex
	topics    map[string]p2p.TopicValidator
	// preconnectPeerHandler is used for notifying that a peer wants to connect to another peer so
	// in the case of mocknet use, mocknet should first link the peers
	preconnectPeerHandler PeerInfoHandler
	peerDiscoveredHandler PeerInfoHandler

	outgoingPLB p2p.PipeLoadBalancer
}

// NewMockNetworkMessenger creates a new sandbox testable instance of libP2P messenger
// It should not open ports on current machine
// Should be used only in testing!
func NewMockNetworkMessenger(ctx context.Context, mockNet mocknet.Mocknet) (*networkMessenger, error) {
	if ctx == nil {
		return nil, p2p.ErrNilContext
	}

	if mockNet == nil {
		return nil, p2p.ErrNilMockNet
	}

	h, err := mockNet.GenPeer()
	if err != nil {
		return nil, err
	}

	mes, err := createMessenger(ctx, h, false, loadBalancer.NewOutgoingPipeLoadBalancer())
	if err != nil {
		return nil, err
	}

	mes.preconnectPeerHandler = func(pInfo peerstore.PeerInfo) {
		_ = mockNet.LinkAll()
	}

	return mes, err
}

// NewNetworkMessenger creates a libP2P messenger by opening a port on the current machine
// Should be used in production!
func NewNetworkMessenger(
	ctx context.Context,
	port int,
	p2pPrivKey crypto.PrivKey,
	conMgr ifconnmgr.ConnManager,
	outgoingPLB p2p.PipeLoadBalancer,
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

	p2pNode, err := createMessenger(ctx, h, true, outgoingPLB)
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
) (*networkMessenger, error) {

	pb, err := createPubSub(ctx, h, withSigning)
	if err != nil {
		return nil, err
	}

	kad, discoverer, err := createKadDHT(ctx, h, elrondRandezVousString)
	if err != nil {
		return nil, err
	}

	netMes := networkMessenger{
		hostP2P:     h,
		pb:          pb,
		topics:      make(map[string]p2p.TopicValidator),
		ctx:         ctx,
		kadDHT:      kad,
		discoverer:  discoverer,
		outgoingPLB: outgoingPLB,
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

func createKadDHT(ctx context.Context, h host.Host, randezvous string) (*dht.IpfsDHT, *discovery.RoutingDiscovery, error) {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		return nil, nil, err
	}

	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, nil, err
	}

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	log.Debug("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, randezvous)
	log.Debug("Successfully announced!")

	return kademliaDHT, routingDiscovery, nil
}

// Close closes the host, connections and streams
func (netMes *networkMessenger) Close() error {
	err := netMes.kadDHT.Close()
	log.LogIfError(err)

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

// DiscoverNewPeers starts a blocking function that searches for all known peers querying all connected peers
// The default libp2p implementation tries to connect to all of them
func (netMes *networkMessenger) DiscoverNewPeers() error {
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

// BroadcastData tries to send a byte buffer onto a topic
func (netMes *networkMessenger) Broadcast(pipe string, topic string, buff []byte) {
	go func() {
		sendable := &p2p.SendableData{
			Buff:  buff,
			Topic: topic,
		}
		netMes.outgoingPLB.GetChannelOrDefault(pipe) <- sendable
	}()
}

// RegisterTopicValidator registers a validator on a topic
func (netMes *networkMessenger) RegisterTopicValidator(topic string, handler p2p.TopicValidator) error {
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
		err := handler.Validate(NewMessage(message))

		return err == nil
	})
	if err != nil {
		return err
	}

	netMes.topics[topic] = handler
	return nil
}

// UnregisterTopicValidator registers a validator on a topic
func (netMes *networkMessenger) UnregisterTopicValidator(topic string) error {
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
	var validator p2p.TopicValidator

	netMes.mutTopics.RLock()
	validator = netMes.topics[message.TopicIDs()[0]]
	netMes.mutTopics.RUnlock()

	if validator == nil {
		return p2p.ErrNilValidator
	}

	go func(msg p2p.MessageP2P) {
		log.LogIfError(validator.Validate(msg))
	}(message)

	return nil
}
