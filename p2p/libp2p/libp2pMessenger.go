package libp2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/dataThrottle"
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

type libp2pMessenger struct {
	ctx         context.Context
	hostP2P     host.Host
	pb          *pubsub.PubSub
	ds          p2p.DirectSender
	kadDHT      *dht.IpfsDHT
	discoverer  discovery.Discoverer
	connManager ifconnmgr.ConnManager

	mutTopics sync.RWMutex
	topics    map[string]func(message p2p.MessageP2P) error
	// preconnectPeerHandler is used for notifying that a peer wants to connect to another peer so
	// in the case of mocknet use, mocknet should first link the peers
	preconnectPeerHandler func(pInfo peerstore.PeerInfo)
	PeerDiscoveredHandler func(pInfo peerstore.PeerInfo)

	sendDataThrottle p2p.DataThrottler
}

// NewMockLibp2pMessenger creates a new sandbox testable instance of libP2P messenger
// It should not open ports on current machine
// Should be used only in testing!
func NewMockLibp2pMessenger(ctx context.Context, mockNet mocknet.Mocknet) (*libp2pMessenger, error) {
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

	mes, err := createLibP2P(ctx, h, false, dataThrottle.NewSendDataThrottle())
	if err != nil {
		return nil, err
	}

	mes.preconnectPeerHandler = func(pInfo peerstore.PeerInfo) {
		_ = mockNet.LinkAll()
	}

	return mes, err
}

// NewSocketLibp2pMessenger creates a libP2P messenger by opening a port on the current machine
// Should be used in production!
func NewSocketLibp2pMessenger(
	ctx context.Context,
	port int,
	p2pPrivKey crypto.PrivKey,
	conMgr ifconnmgr.ConnManager,
	sendDataThrottler p2p.DataThrottler,
) (*libp2pMessenger, error) {

	if ctx == nil {
		return nil, p2p.ErrNilContext
	}

	if port < 1 {
		return nil, p2p.ErrInvalidPort
	}

	if p2pPrivKey == nil {
		return nil, p2p.ErrNilP2PprivateKey
	}

	if sendDataThrottler == nil {
		return nil, p2p.ErrNilDataThrottler
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.Identity(p2pPrivKey),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	}

	if conMgr != nil {
		opts = append(opts, libp2p.ConnectionManager(conMgr))
	}

	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	p2pNode, err := createLibP2P(ctx, h, true, sendDataThrottler)
	if err != nil {
		return nil, err
	}

	if conMgr != nil {
		p2pNode.connManager = conMgr
	}

	return p2pNode, nil
}

func createLibP2P(
	ctx context.Context,
	h host.Host,
	withSigning bool,
	sendDataThrottler p2p.DataThrottler,
) (*libp2pMessenger, error) {

	pb, err := createPubSub(ctx, h, withSigning)
	if err != nil {
		log.LogIfError(h.Close())
		return nil, err
	}

	kad, discoverer, err := createKadDHT(ctx, h, elrondRandezVousString)
	if err != nil {
		log.LogIfError(h.Close())
		return nil, err
	}

	libP2PMes := libp2pMessenger{
		hostP2P:          h,
		pb:               pb,
		topics:           make(map[string]func(message p2p.MessageP2P) error),
		ctx:              ctx,
		kadDHT:           kad,
		discoverer:       discoverer,
		sendDataThrottle: sendDataThrottler,
	}

	ds, err := NewDirectSender(ctx, h, libP2PMes.directMessageHandler)
	if err != nil {
		log.LogIfError(h.Close())
		return nil, err
	}
	libP2PMes.ds = ds

	go func(pubsub *pubsub.PubSub, dt p2p.DataThrottler) {
		for {
			dataToBeSent := dt.CollectFromPipes()

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
	}(pb, libP2PMes.sendDataThrottle)

	for _, address := range libP2PMes.hostP2P.Addrs() {
		fmt.Println(address.String() + "/p2p/" + libP2PMes.ID().Pretty())
	}

	return &libP2PMes, nil
}

func createPubSub(ctx context.Context, host host.Host, withSigning bool) (*pubsub.PubSub, error) {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(withSigning),
	}

	ps, err := pubsub.NewFloodSub(ctx, host, optsPS...)
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
func (p2pMes *libp2pMessenger) Close() error {
	err := p2pMes.kadDHT.Close()
	log.LogIfError(err)

	return p2pMes.hostP2P.Close()
}

// ID returns the messenger's ID
func (p2pMes *libp2pMessenger) ID() p2p.PeerID {
	return p2p.PeerID(p2pMes.hostP2P.ID())
}

// Peers returns the list of all known peers ID (including self)
func (p2pMes *libp2pMessenger) Peers() []p2p.PeerID {
	peers := make([]p2p.PeerID, 0)

	for _, p := range p2pMes.hostP2P.Peerstore().Peers() {
		peers = append(peers, p2p.PeerID(p))
	}
	return peers
}

// Addresses returns all addresses found in peerstore
func (p2pMes *libp2pMessenger) Addresses() []string {
	addrs := make([]string, 0)

	for _, address := range p2pMes.hostP2P.Addrs() {
		addrs = append(addrs, address.String()+"/p2p/"+p2pMes.ID().Pretty())
	}

	return addrs
}

// ConnectToPeer tries to open a new connection to a peer
func (p2pMes *libp2pMessenger) ConnectToPeer(address string) error {
	multiAddr, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return err
	}

	pInfo, err := peerstore.InfoFromP2pAddr(multiAddr)
	if err != nil {
		return err
	}

	if p2pMes.preconnectPeerHandler != nil {
		p2pMes.preconnectPeerHandler(*pInfo)
	}

	return p2pMes.hostP2P.Connect(p2pMes.ctx, *pInfo)
}

// DiscoverNewPeers starts a blocking function that searches for all known peers querying all connected peers
// The default libp2p implementation tries to connect to all of them
func (p2pMes *libp2pMessenger) DiscoverNewPeers() error {
	if p2pMes.discoverer == nil {
		return p2p.ErrNilDiscoverer
	}

	peerChan, err := p2pMes.discoverer.FindPeers(p2pMes.ctx, elrondRandezVousString)
	if err != nil {
		return err
	}

	for {
		pInfo, more := <-peerChan

		if !more {
			//discovered peers channel closed
			break
		}

		handler := p2pMes.PeerDiscoveredHandler

		if handler != nil {
			handler(pInfo)
		}
	}

	return nil
}

// TrimConnections will trigger a manual sweep onto current connection set reducing the
// number of connections if needed
func (p2pMes *libp2pMessenger) TrimConnections() error {
	if p2pMes.connManager == nil {
		return p2p.ErrNilConnManager
	}

	p2pMes.connManager.TrimOpenConns(p2pMes.ctx)
	return nil
}

// IsConnected returns true if current node is connected to provided peer
func (p2pMes *libp2pMessenger) IsConnected(peerID p2p.PeerID) bool {
	connectedness := p2pMes.hostP2P.Network().Connectedness(peer.ID(peerID))

	return connectedness == net.Connected
}

// ConnectedPeers returns the current connected peers list
func (p2pMes *libp2pMessenger) ConnectedPeers() []p2p.PeerID {
	peerList := make([]p2p.PeerID, 0)

	for _, conn := range p2pMes.hostP2P.Network().Conns() {
		p := p2p.PeerID(conn.RemotePeer())

		if p2pMes.IsConnected(p) {
			peerList = append(peerList, p)
		}
	}

	return peerList
}

// CreateTopic opens a new topic using pubsub infrastructure
func (p2pMes *libp2pMessenger) CreateTopic(name string, createPipeForTopic bool) error {
	p2pMes.mutTopics.Lock()
	_, found := p2pMes.topics[name]
	if found {
		p2pMes.mutTopics.Unlock()
		return p2p.ErrTopicAlreadyExists
	}

	p2pMes.topics[name] = nil
	subscrRequest, err := p2pMes.pb.Subscribe(name)
	if err != nil {
		p2pMes.mutTopics.Unlock()
		return err
	}
	p2pMes.mutTopics.Unlock()

	if createPipeForTopic {
		err = p2pMes.sendDataThrottle.AddPipe(name)
	}

	//just a dummy func to consume messages received by the newly created topic
	go func() {
		for {
			_, _ = subscrRequest.Next(p2pMes.ctx)
		}
	}()

	return err
}

// HasTopic returns true if the topic has been created
func (p2pMes *libp2pMessenger) HasTopic(name string) bool {
	p2pMes.mutTopics.RLock()
	_, found := p2pMes.topics[name]
	p2pMes.mutTopics.RUnlock()

	return found
}

// HasTopicValidator returns true if the topic has a validator set
func (p2pMes *libp2pMessenger) HasTopicValidator(name string) bool {
	p2pMes.mutTopics.RLock()
	validator, _ := p2pMes.topics[name]
	p2pMes.mutTopics.RUnlock()

	return validator != nil
}

// SendDataThrottler returns the data throttler object used by the messenger to send data
func (p2pMes *libp2pMessenger) SendDataThrottler() p2p.DataThrottler {
	return p2pMes.sendDataThrottle
}

// BroadcastData tries to send a byte buffer onto a topic
func (p2pMes *libp2pMessenger) BroadcastData(pipe string, topic string, buff []byte) {
	go func() {
		sendable := &p2p.SendableData{
			Buff:  buff,
			Topic: topic,
		}
		p2pMes.sendDataThrottle.GetChannelOrDefault(pipe) <- sendable
	}()
}

// SetTopicValidator sets a validator on a topic
func (p2pMes *libp2pMessenger) SetTopicValidator(topic string, handler func(message p2p.MessageP2P) error) error {
	p2pMes.mutTopics.Lock()
	defer p2pMes.mutTopics.Unlock()
	validator, found := p2pMes.topics[topic]

	if !found {
		return p2p.ErrNilTopic
	}

	if handler == nil && validator != nil {
		err := p2pMes.pb.UnregisterTopicValidator(topic)
		if err != nil {
			return err
		}

		p2pMes.topics[topic] = nil
		return nil
	}

	if handler != nil && validator == nil {
		err := p2pMes.pb.RegisterTopicValidator(topic, func(i context.Context, message *pubsub.Message) bool {
			err := handler(NewMessage(message))

			return err == nil
		})

		if err != nil {
			return err
		}

		p2pMes.topics[topic] = handler
		return nil
	}

	return p2p.ErrTopicValidatorOperationNotSupported
}

// SendDirectToConnectedPeer sends a direct message to a connected peer
func (p2pMes *libp2pMessenger) SendDirectToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	return p2pMes.ds.SendDirectToConnectedPeer(topic, buff, peerID)
}

func (p2pMes *libp2pMessenger) directMessageHandler(message p2p.MessageP2P) error {
	var validator func(message p2p.MessageP2P) error

	p2pMes.mutTopics.RLock()
	validator = p2pMes.topics[message.TopicIDs()[0]]
	p2pMes.mutTopics.RUnlock()

	if validator == nil {
		return p2p.ErrNilValidator
	}

	go func(msg p2p.MessageP2P) {
		log.LogIfError(validator(msg))
	}(message)

	return nil
}
