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
	"github.com/libp2p/go-libp2p-interface-connmgr"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-pubsub"
)

const durationBetweenSends = time.Duration(time.Microsecond * 10)

// DirectSendID represents the protocol ID for sending and receiving direct P2P messages
const DirectSendID = protocol.ID("/directsend/1.0.0")

var log = logger.NewDefaultLogger()

type networkMessenger struct {
	ctxProvider *Libp2pContext

	pb *pubsub.PubSub
	ds p2p.DirectSender

	peerDiscoverer p2p.PeerDiscoverer

	mutTopics sync.RWMutex
	topics    map[string]p2p.MessageProcessor

	outgoingPLB p2p.ChannelLoadBalancer
}

// NewNetworkMessenger creates a libP2P messenger by opening a port on the current machine
// Should be used in production!
func NewNetworkMessenger(
	ctx context.Context,
	port int,
	p2pPrivKey crypto.PrivKey,
	conMgr ifconnmgr.ConnManager,
	outgoingPLB p2p.ChannelLoadBalancer,
	peerDiscoverer p2p.PeerDiscoverer,
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
		return nil, p2p.ErrNilChannelLoadBalancer
	}

	if peerDiscoverer == nil {
		return nil, p2p.ErrNilPeerDiscoverer
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.Identity(p2pPrivKey),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.ConnectionManager(conMgr),
	}

	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	lctx, err := NewLibp2pContext(ctx, NewConnectableHost(h))
	if err != nil {
		log.LogIfError(h.Close())
		return nil, err
	}

	p2pNode, err := createMessenger(lctx, true, outgoingPLB, peerDiscoverer)
	if err != nil {
		log.LogIfError(h.Close())
		return nil, err
	}

	return p2pNode, nil
}

func createMessenger(
	lctx *Libp2pContext,
	withSigning bool,
	outgoingPLB p2p.ChannelLoadBalancer,
	peerDiscoverer p2p.PeerDiscoverer,
) (*networkMessenger, error) {

	pb, err := createPubSub(lctx, withSigning)
	if err != nil {
		return nil, err
	}

	err = peerDiscoverer.ApplyContext(lctx)
	if err != nil {
		return nil, err
	}

	netMes := networkMessenger{
		ctxProvider:    lctx,
		pb:             pb,
		topics:         make(map[string]p2p.MessageProcessor),
		outgoingPLB:    outgoingPLB,
		peerDiscoverer: peerDiscoverer,
	}

	netMes.ds, err = NewDirectSender(lctx.Context(), lctx.Host(), netMes.directMessageHandler)
	if err != nil {
		return nil, err
	}

	go func(pubsub *pubsub.PubSub, plb p2p.ChannelLoadBalancer) {
		for {
			sendableData := plb.CollectOneElementFromChannels()

			if sendableData == nil {
				continue
			}

			_ = pb.Publish(sendableData.Topic, sendableData.Buff)
			time.Sleep(durationBetweenSends)
		}
	}(pb, netMes.outgoingPLB)

	for _, address := range netMes.ctxProvider.Host().Addrs() {
		fmt.Println(address.String() + "/p2p/" + netMes.ID().Pretty())
	}

	return &netMes, nil
}

func createPubSub(ctxProvider *Libp2pContext, withSigning bool) (*pubsub.PubSub, error) {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(withSigning),
	}

	ps, err := pubsub.NewGossipSub(ctxProvider.Context(), ctxProvider.Host(), optsPS...)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

// Close closes the host, connections and streams
func (netMes *networkMessenger) Close() error {
	return netMes.ctxProvider.Host().Close()
}

// ID returns the messenger's ID
func (netMes *networkMessenger) ID() p2p.PeerID {
	h := netMes.ctxProvider.Host()

	return p2p.PeerID(h.ID())
}

// Peers returns the list of all known peers ID (including self)
func (netMes *networkMessenger) Peers() []p2p.PeerID {
	h := netMes.ctxProvider.Host()
	peers := make([]p2p.PeerID, 0)

	for _, p := range h.Peerstore().Peers() {
		peers = append(peers, p2p.PeerID(p))
	}
	return peers
}

// Addresses returns all addresses found in peerstore
func (netMes *networkMessenger) Addresses() []string {
	h := netMes.ctxProvider.Host()
	addrs := make([]string, 0)

	for _, address := range h.Addrs() {
		addrs = append(addrs, address.String()+"/p2p/"+netMes.ID().Pretty())
	}

	return addrs
}

// ConnectToPeer tries to open a new connection to a peer
func (netMes *networkMessenger) ConnectToPeer(address string) error {
	h := netMes.ctxProvider.Host()
	ctx := netMes.ctxProvider.ctx

	return h.ConnectToPeer(ctx, address)
}

// TrimConnections will trigger a manual sweep onto current connection set reducing the
// number of connections if needed
func (netMes *networkMessenger) TrimConnections() {
	h := netMes.ctxProvider.Host()
	ctx := netMes.ctxProvider.Context()

	h.ConnManager().TrimOpenConns(ctx)
}

// Bootstrap will start the peer discovery mechanism
func (netMes *networkMessenger) Bootstrap() error {
	return netMes.peerDiscoverer.Bootstrap()
}

// IsConnected returns true if current node is connected to provided peer
func (netMes *networkMessenger) IsConnected(peerID p2p.PeerID) bool {
	h := netMes.ctxProvider.Host()

	connectedness := h.Network().Connectedness(peer.ID(peerID))

	return connectedness == net.Connected
}

// ConnectedPeers returns the current connected peers list
func (netMes *networkMessenger) ConnectedPeers() []p2p.PeerID {
	h := netMes.ctxProvider.Host()

	connectedPeers := make(map[p2p.PeerID]struct{})

	for _, conn := range h.Network().Conns() {
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

// ConnectedAddresses returns all connected peer's addresses
func (netMes *networkMessenger) ConnectedAddresses() []string {
	h := netMes.ctxProvider.Host()
	conns := make([]string, 0)

	for _, c := range h.Network().Conns() {
		conns = append(conns, c.RemoteMultiaddr().String()+"/p2p/"+c.RemotePeer().Pretty())
	}
	return conns
}

// ConnectedPeersOnTopic returns the connected peers on a provided topic
func (netMes *networkMessenger) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
	//as the peers in pubsub impl are held inside a map where the key is the peerID,
	//the returned list will hold distinct values
	list := netMes.pb.ListPeers(topic)
	connectedPeers := make([]p2p.PeerID, len(list))

	for idx, pid := range list {
		connectedPeers[idx] = p2p.PeerID(pid)
	}

	return connectedPeers
}

// CreateTopic opens a new topic using pubsub infrastructure
func (netMes *networkMessenger) CreateTopic(name string, createChannelForTopic bool) error {
	ctx := netMes.ctxProvider.Context()

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

	if createChannelForTopic {
		err = netMes.outgoingPLB.AddChannel(name)
	}

	//just a dummy func to consume messages received by the newly created topic
	go func() {
		for {
			_, _ = subscrRequest.Next(ctx)
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

// OutgoingChannelLoadBalancer returns the channel load balancer object used by the messenger to send data
func (netMes *networkMessenger) OutgoingChannelLoadBalancer() p2p.ChannelLoadBalancer {
	return netMes.outgoingPLB
}

// BroadcastOnChannel tries to send a byte buffer onto a topic using provided channel
func (netMes *networkMessenger) BroadcastOnChannel(channel string, topic string, buff []byte) {
	go func() {
		sendable := &p2p.SendableData{
			Buff:  buff,
			Topic: topic,
		}
		netMes.outgoingPLB.GetChannelOrDefault(channel) <- sendable
	}()
}

// Broadcast tries to send a byte buffer onto a topic using the topic name as channel
func (netMes *networkMessenger) Broadcast(topic string, buff []byte) {
	netMes.BroadcastOnChannel(topic, topic, buff)
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

	err := netMes.pb.RegisterTopicValidator(topic, func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
		err := handler.ProcessReceivedMessage(NewMessage(message))

		if err != nil {
			log.Debug(err.Error())
		}

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
		err := processor.ProcessReceivedMessage(msg)

		if err != nil {
			log.Debug(err.Error())
		}
	}(message)

	return nil
}
