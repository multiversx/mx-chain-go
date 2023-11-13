package components

import (
	"context"
	"fmt"
	"strings"
	"sync"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const tcpInterface = "0.0.0.0"       // bind on all interfaces
const pubSubMaxMessageSize = 1 << 21 // 2 MB
var log = logger.GetOrCreate("p2p")

// ArgsNetMessenger defines the arguments to instantiate a network messenger wrapper struct
type ArgsNetMessenger struct {
	InitialPeerList []string
	PrivateKeyBytes []byte
	ProtocolID      string
	Port            int
}

type netMessenger struct {
	*bootstrapper
	host          host.Host
	pb            *pubsub.PubSub
	mutTopics     sync.RWMutex
	topics        map[string]*pubsub.Topic
	subscriptions map[string]*pubsub.Subscription
	processors    map[string]MessageProcessor
	cancel        func()
}

// GeneratePrivateKeyBytes will generate a byte slice that can be used as a private key
func GeneratePrivateKeyBytes() ([]byte, error) {
	privKey, err := secp.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	return privKey.Serialize(), nil
}

// NewNetMessenger creates a new instance of type netMessenger
func NewNetMessenger(args ArgsNetMessenger) (*netMessenger, error) {
	privKeyBytes := args.PrivateKeyBytes
	var err error
	if len(privKeyBytes) == 0 {
		log.Info("provided empty private key bytes, generating a new private key")
		privKeyBytes, err = GeneratePrivateKeyBytes()
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := crypto.UnmarshalSecp256k1PrivateKey(privKeyBytes)
	if err != nil {
		return nil, err
	}

	transport := libp2p.Transport(tcp.NewTCPTransport)

	resourceManager, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
	if err != nil {
		return nil, err
	}

	// always get a free port
	address := fmt.Sprintf("/ip4/%s/tcp/%d", tcpInterface, args.Port)
	options := []libp2p.Option{
		libp2p.ListenAddrStrings(address),
		libp2p.Identity(privateKey),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		// we need to disable relay option in order to save the node's bandwidth as much as possible
		libp2p.DisableRelay(),
		libp2p.NATPortMap(),
		libp2p.ResourceManager(resourceManager),
	}
	options = append(options, transport)

	h, err := libp2p.New(options...)
	if err != nil {
		return nil, err
	}

	instance := &netMessenger{
		host:          h,
		topics:        make(map[string]*pubsub.Topic),
		subscriptions: make(map[string]*pubsub.Subscription),
		processors:    make(map[string]MessageProcessor),
	}
	instance.bootstrapper, err = newBootstrapper(h, args.InitialPeerList, protocol.ID(args.ProtocolID))
	if err != nil {
		return nil, err
	}

	optsPS := []pubsub.Option{
		pubsub.WithPeerFilter(instance.newPeerFound),
		pubsub.WithMaxMessageSize(pubSubMaxMessageSize),
	}

	var ctx context.Context
	ctx, instance.cancel = context.WithCancel(context.Background())

	instance.pb, err = pubsub.NewGossipSub(ctx, h, optsPS...)
	if err != nil {
		return nil, err
	}

	log.Info("Listening on the following interfaces: " + strings.Join(instance.Addresses(), ", "))

	return instance, nil
}

func (netMes *netMessenger) newPeerFound(_ peer.ID, _ string) bool {
	return true
}

// Addresses returns the addresses that the current messenger was able to bind to
func (netMes *netMessenger) Addresses() []string {
	addresses := make([]string, 0)
	for _, ma := range netMes.host.Addrs() {
		addresses = append(addresses, ma.String()+"/p2p/"+netMes.ID().String())
	}

	return addresses
}

// ID returns the peer ID
func (netMes *netMessenger) ID() peer.ID {
	return netMes.host.ID()
}

// Bootstrap will start the bootstrapping process
func (netMes *netMessenger) Bootstrap() {
	netMes.bootstrapper.bootstrap()
}

// ConnectedAddresses returns all connected peer's addresses
func (netMes *netMessenger) ConnectedAddresses() []string {
	conns := make([]string, 0)
	for _, c := range netMes.h.Network().Conns() {
		conns = append(conns, c.RemoteMultiaddr().String()+"/p2p/"+c.RemotePeer().String())
	}
	return conns
}

// GetConnectedness returns the connectedness with the provided peer ID
func (netMes *netMessenger) GetConnectedness(pid peer.ID) network.Connectedness {
	return netMes.host.Network().Connectedness(pid)
}

// Peers returns the list of all known peers ID (including self)
func (netMes *netMessenger) Peers() []peer.ID {
	peers := make([]peer.ID, 0)

	for _, p := range netMes.h.Peerstore().Peers() {
		peers = append(peers, p)
	}
	return peers
}

// CreateTopic opens a new topic using pubsub infrastructure
func (netMes *netMessenger) CreateTopic(name string) error {
	netMes.mutTopics.Lock()
	defer netMes.mutTopics.Unlock()
	_, found := netMes.topics[name]
	if found {
		return nil
	}

	topic, err := netMes.pb.Join(name)
	if err != nil {
		return fmt.Errorf("%w for topic %s", err, name)
	}

	netMes.topics[name] = topic
	subscrRequest, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("%w for topic %s", err, name)
	}

	netMes.subscriptions[name] = subscrRequest

	// just a dummy func to consume messages received by the newly created topic
	go func() {
		var errSubscrNext error
		for {
			_, errSubscrNext = subscrRequest.Next(context.Background())
			if errSubscrNext != nil {
				log.Debug("closed subscription",
					"topic", subscrRequest.Topic(),
					"err", errSubscrNext,
				)
				return
			}
		}
	}()

	return err
}

// Broadcast tries to send a byte buffer onto a topic
func (netMes *netMessenger) Broadcast(topic string, buff []byte) {
	netMes.mutTopics.RLock()
	defer netMes.mutTopics.RUnlock()

	topicInstance := netMes.topics[topic]

	err := topicInstance.Publish(context.Background(), buff)
	if err != nil {
		log.Warn("error publishing message",
			"pid", netMes.ID().String(),
			"topic", topic,
			"error", err)
	}
}

// RegisterMessageProcessor registers a message process on a topic. The function allows registering multiple handlers
// on a topic. Each handler should be associated with a new identifier on the same topic. Using same identifier on different
// topics is allowed. The order of handler calling on a particular topic is not deterministic.
func (netMes *netMessenger) RegisterMessageProcessor(topic string, handler MessageProcessor) error {
	netMes.mutTopics.Lock()
	defer netMes.mutTopics.Unlock()

	_, found := netMes.processors[topic]
	if found {
		return fmt.Errorf("topic %s is already taken by a message processor", topic)
	}

	netMes.processors[topic] = handler

	return netMes.pb.RegisterTopicValidator(topic, netMes.pubsubCallback(topic))
}

func (netMes *netMessenger) pubsubCallback(topic string) func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
	return func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
		netMes.mutTopics.RLock()
		handler := netMes.processors[topic]
		netMes.mutTopics.RUnlock()

		err := handler.ProcessMessage(message)
		if err != nil {
			log.Trace("p2p validator",
				"error", err.Error(),
				"topic", topic,
				"originator", peer.ID(message.From).String(),
				"from connected peer", pid.String(),
				"seq no", message.Seqno,
			)
		}

		return err == nil
	}

}

// Close will call Close on all inner components
func (netMes *netMessenger) Close() error {
	netMes.cancel()
	netMes.bootstrapper.close()

	netMes.mutTopics.RLock()
	for _, subscription := range netMes.subscriptions {
		subscription.Cancel()
	}
	netMes.mutTopics.RUnlock()

	return netMes.host.Close()
}
