package libp2p

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	connMonitorFactory "github.com/ElrondNetwork/elrond-go/p2p/libp2p/connectionMonitor/factory"
	discoveryFactory "github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery/factory"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/metrics"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/networksharding/factory"
	randFactory "github.com/ElrondNetwork/elrond-go/p2p/libp2p/rand/factory"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-pubsub"
)

const durationBetweenSends = time.Microsecond * 10

// ListenAddrWithIp4AndTcp defines the listening address with ip v.4 and TCP
const ListenAddrWithIp4AndTcp = "/ip4/0.0.0.0/tcp/"

// ListenLocalhostAddrWithIp4AndTcp defines the local host listening ip v.4 address and TCP
const ListenLocalhostAddrWithIp4AndTcp = "/ip4/127.0.0.1/tcp/"

// DirectSendID represents the protocol ID for sending and receiving direct P2P messages
const DirectSendID = protocol.ID("/directsend/1.0.0")

const refreshPeersOnTopic = time.Second * 5
const ttlPeersOnTopic = time.Second * 30
const pubsubTimeCacheDuration = 10 * time.Minute
const broadcastGoRoutines = 1000
const secondsBetweenPeerPrints = 20
const defaultThresholdMinConnectedPeers = 3

//TODO remove the header size of the message when commit d3c5ecd3a3e884206129d9f2a9a4ddfd5e7c8951 from
// https://github.com/libp2p/go-libp2p-pubsub/pull/189/commits will be part of a new release
var messageHeader = 64 * 1024 //64kB
var maxSendBuffSize = (1 << 20) - messageHeader

var log = logger.GetOrCreate("p2p/libp2p")

//TODO refactor this struct to have be a wrapper (with logic) over a glue code
type networkMessenger struct {
	ctx                 context.Context
	p2pHost             ConnectableHost
	pb                  *pubsub.PubSub
	ds                  p2p.DirectSender
	connMonitor         ConnectionMonitor
	peerDiscoverer      p2p.PeerDiscoverer
	sharder             p2p.CommonSharder
	peerShardResolver   p2p.PeerShardResolver
	mutTopics           sync.RWMutex
	topics              map[string]p2p.MessageProcessor
	outgoingPLB         p2p.ChannelLoadBalancer
	poc                 *peersOnChannel
	goRoutinesThrottler *throttler.NumGoRoutineThrottler
	ip                  *identityProvider
	connectionsMetric   *metrics.Connections
}

// ArgsNetworkMessenger defines the options used to create a p2p wrapper
type ArgsNetworkMessenger struct {
	Context       context.Context
	ListenAddress string
	P2pConfig     config.P2PConfig
}

// NewNetworkMessenger creates a libP2P messenger by opening a port on the current machine
func NewNetworkMessenger(args ArgsNetworkMessenger) (*networkMessenger, error) {
	err := checkParameters(args)
	if err != nil {
		return nil, err
	}

	p2pPrivKey, err := createP2PPrivKey(args.P2pConfig.Node.Seed)
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf(args.ListenAddress+"%d", args.P2pConfig.Node.Port)
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(address),
		libp2p.Identity(p2pPrivKey),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.DefaultTransports,
		//we need the disable relay option in order to save the node's bandwidth as much as possible
		libp2p.DisableRelay(),
		libp2p.NATPortMap(),
	}

	h, err := libp2p.New(args.Context, opts...)
	if err != nil {
		return nil, err
	}

	p2pNode, err := createMessenger(args, h, true)
	if err != nil {
		log.LogIfError(h.Close())
		return nil, err
	}

	return p2pNode, nil
}

func checkParameters(args ArgsNetworkMessenger) error {
	if args.Context == nil {
		return p2p.ErrNilContext
	}

	return nil
}

func createP2PPrivKey(seed string) (*libp2pCrypto.Secp256k1PrivateKey, error) {
	randReader, err := randFactory.NewRandFactory(seed)
	if err != nil {
		return nil, err
	}

	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), randReader)

	return (*libp2pCrypto.Secp256k1PrivateKey)(prvKey), nil
}

func createMessenger(
	args ArgsNetworkMessenger,
	p2pHost host.Host,
	withMessageSigning bool,
) (*networkMessenger, error) {
	var err error
	netMes := networkMessenger{
		ctx:               args.Context,
		p2pHost:           NewConnectableHost(p2pHost),
		topics:            make(map[string]p2p.MessageProcessor),
		outgoingPLB:       loadBalancer.NewOutgoingChannelLoadBalancer(),
		peerShardResolver: &unknownPeerShardResolver{},
	}

	err = netMes.createPubSub(withMessageSigning)
	if err != nil {
		return nil, err
	}

	err = netMes.createSharder(args.P2pConfig)
	if err != nil {
		return nil, err
	}

	err = netMes.createDiscoverer(args.P2pConfig)
	if err != nil {
		return nil, err
	}

	err = netMes.createConnectionMonitor(args.P2pConfig)
	if err != nil {
		return nil, err
	}

	netMes.createConnectionsMetric()

	netMes.ds, err = NewDirectSender(args.Context, p2pHost, netMes.directMessageHandler)
	if err != nil {
		return nil, err
	}

	netMes.goRoutinesThrottler, err = throttler.NewNumGoRoutineThrottler(broadcastGoRoutines)
	if err != nil {
		return nil, err
	}

	netMes.printLogs()

	return &netMes, nil
}

func (netMes *networkMessenger) createPubSub(withMessageSigning bool) error {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(withMessageSigning),
	}

	pubsub.TimeCacheDuration = pubsubTimeCacheDuration

	var err error
	netMes.pb, err = pubsub.NewGossipSub(netMes.ctx, netMes.p2pHost, optsPS...)
	if err != nil {
		return err
	}

	netMes.poc, err = newPeersOnChannel(
		netMes.pb.ListPeers,
		refreshPeersOnTopic,
		ttlPeersOnTopic)
	if err != nil {
		return err
	}

	go func(pubsub *pubsub.PubSub, plb p2p.ChannelLoadBalancer) {
		for {
			sendableData := plb.CollectOneElementFromChannels()
			if sendableData == nil {
				continue
			}

			errPublish := pubsub.Publish(sendableData.Topic, sendableData.Buff)
			if errPublish != nil {
				log.Trace("error sending data", "error", errPublish)
			}

			time.Sleep(durationBetweenSends)
		}
	}(netMes.pb, netMes.outgoingPLB)

	return nil
}

func (netMes *networkMessenger) createSharder(p2pConfig config.P2PConfig) error {
	args := factory.ArgsSharderFactory{
		PeerShardResolver:  &unknownPeerShardResolver{},
		PrioBits:           p2pConfig.Sharding.PrioBits,
		Pid:                netMes.p2pHost.ID(),
		MaxConnectionCount: p2pConfig.Sharding.TargetPeerCount,
		MaxIntraShard:      int(p2pConfig.Sharding.MaxIntraShard),
		MaxCrossShard:      int(p2pConfig.Sharding.MaxCrossShard),
		Type:               p2pConfig.Sharding.Type,
	}

	var err error
	netMes.sharder, err = factory.NewSharder(args)

	return err
}

func (netMes *networkMessenger) createDiscoverer(p2pConfig config.P2PConfig) error {
	var err error
	netMes.peerDiscoverer, err = discoveryFactory.NewPeerDiscoverer(
		netMes.ctx,
		netMes.p2pHost,
		netMes.sharder,
		p2pConfig,
	)

	return err
}

func (netMes *networkMessenger) createConnectionMonitor(p2pConfig config.P2PConfig) error {
	reconnecter, ok := netMes.peerDiscoverer.(p2p.Reconnecter)
	if !ok {
		return fmt.Errorf("%w when converting peerDiscoverer to reconnecter interface", p2p.ErrWrongTypeAssertion)
	}

	args := connMonitorFactory.ArgsConnectionMonitorFactory{
		Reconnecter:                reconnecter,
		Sharder:                    netMes.sharder,
		ThresholdMinConnectedPeers: defaultThresholdMinConnectedPeers,
		TargetCount:                p2pConfig.Sharding.TargetPeerCount,
	}
	var err error
	netMes.connMonitor, err = connMonitorFactory.NewConnectionMonitor(args)
	if err != nil {
		return err
	}
	netMes.p2pHost.Network().Notify(netMes.connMonitor)

	return nil
}

func (netMes *networkMessenger) createConnectionsMetric() {
	netMes.connectionsMetric = metrics.NewConnections()
	netMes.p2pHost.Network().Notify(netMes.connectionsMetric)
}

func (netMes *networkMessenger) printLogs() {
	addresses := make([]interface{}, 0)
	for i, address := range netMes.p2pHost.Addrs() {
		addresses = append(addresses, fmt.Sprintf("addr%d", i))
		addresses = append(addresses, address.String()+"/p2p/"+netMes.ID().Pretty())
	}
	log.Info("listening on addresses", addresses...)

	go netMes.printLogsStats()
}

func (netMes *networkMessenger) printLogsStats() {
	for {
		time.Sleep(secondsBetweenPeerPrints * time.Second)

		conns := netMes.connectionsMetric.ResetNumConnections()
		disconns := netMes.connectionsMetric.ResetNumDisconnections()

		peersInfo := netMes.GetConnectedPeersInfo()
		log.Debug("network connection status",
			"known peers", len(netMes.Peers()),
			"connected peers", len(netMes.ConnectedPeers()),
			"intra shard", len(peersInfo.IntraShardPeers),
			"cross shard", len(peersInfo.CrossShardPeers),
			"unknown", len(peersInfo.UnknownPeers),
		)

		connsPerSec := conns / secondsBetweenPeerPrints
		disconnsPerSec := disconns / secondsBetweenPeerPrints

		log.Debug("network connection metrics",
			"connections/s", connsPerSec,
			"disconnections/s", disconnsPerSec,
		)
	}
}

// ApplyOptions can set up different configurable options of a networkMessenger instance
func (netMes *networkMessenger) ApplyOptions(opts ...Option) error {
	for _, opt := range opts {
		err := opt(netMes)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the host, connections and streams
func (netMes *networkMessenger) Close() error {
	return netMes.p2pHost.Close()
}

// ID returns the messenger's ID
func (netMes *networkMessenger) ID() p2p.PeerID {
	h := netMes.p2pHost

	return p2p.PeerID(h.ID())
}

// Peers returns the list of all known peers ID (including self)
func (netMes *networkMessenger) Peers() []p2p.PeerID {
	peers := make([]p2p.PeerID, 0)

	for _, p := range netMes.p2pHost.Peerstore().Peers() {
		peers = append(peers, p2p.PeerID(p))
	}
	return peers
}

// Addresses returns all addresses found in peerstore
func (netMes *networkMessenger) Addresses() []string {
	addrs := make([]string, 0)

	for _, address := range netMes.p2pHost.Addrs() {
		addrs = append(addrs, address.String()+"/p2p/"+netMes.ID().Pretty())
	}

	return addrs
}

// ConnectToPeer tries to open a new connection to a peer
func (netMes *networkMessenger) ConnectToPeer(address string) error {
	return netMes.p2pHost.ConnectToPeer(netMes.ctx, address)
}

// Bootstrap will start the peer discovery mechanism
func (netMes *networkMessenger) Bootstrap() error {
	return netMes.peerDiscoverer.Bootstrap()
}

// IsConnected returns true if current node is connected to provided peer
func (netMes *networkMessenger) IsConnected(peerID p2p.PeerID) bool {
	h := netMes.p2pHost

	connectedness := h.Network().Connectedness(peer.ID(peerID))

	return connectedness == network.Connected
}

// ConnectedPeers returns the current connected peers list
func (netMes *networkMessenger) ConnectedPeers() []p2p.PeerID {
	h := netMes.p2pHost

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
	h := netMes.p2pHost
	conns := make([]string, 0)

	for _, c := range h.Network().Conns() {
		conns = append(conns, c.RemoteMultiaddr().String()+"/p2p/"+c.RemotePeer().Pretty())
	}
	return conns
}

// PeerAddress returns the peer's address or empty string if the peer is unknown
func (netMes *networkMessenger) PeerAddress(pid p2p.PeerID) string {
	h := netMes.p2pHost

	//check if the peer is connected to return it's connected address
	for _, c := range h.Network().Conns() {
		if string(c.RemotePeer()) == string(pid.Bytes()) {
			return c.RemoteMultiaddr().String()
		}
	}

	//check in peerstore (maybe it is known but not connected)
	addresses := h.Peerstore().Addrs(peer.ID(pid.Bytes()))
	if len(addresses) == 0 {
		return ""
	}

	//return the first address from multi address slice
	return addresses[0].String()
}

// ConnectedPeersOnTopic returns the connected peers on a provided topic
func (netMes *networkMessenger) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
	return netMes.poc.ConnectedPeersOnChannel(topic)
}

// CreateTopic opens a new topic using pubsub infrastructure
func (netMes *networkMessenger) CreateTopic(name string, createChannelForTopic bool) error {
	netMes.mutTopics.Lock()
	_, found := netMes.topics[name]
	if found {
		netMes.mutTopics.Unlock()
		return p2p.ErrTopicAlreadyExists
	}

	//TODO investigate if calling Subscribe on the pubsub impl does exactly the same thing as Topic.Subscribe
	// after calling pubsub.Join
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
	validator := netMes.topics[name]
	netMes.mutTopics.RUnlock()

	return validator != nil
}

// OutgoingChannelLoadBalancer returns the channel load balancer object used by the messenger to send data
func (netMes *networkMessenger) OutgoingChannelLoadBalancer() p2p.ChannelLoadBalancer {
	return netMes.outgoingPLB
}

// BroadcastOnChannelBlocking tries to send a byte buffer onto a topic using provided channel
// It is a blocking method. It needs to be launched on a go routine
func (netMes *networkMessenger) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	if len(buff) > maxSendBuffSize {
		return p2p.ErrMessageTooLarge
	}

	if !netMes.goRoutinesThrottler.CanProcess() {
		return p2p.ErrTooManyGoroutines
	}

	netMes.goRoutinesThrottler.StartProcessing()

	sendable := &p2p.SendableData{
		Buff:  buff,
		Topic: topic,
	}
	netMes.outgoingPLB.GetChannelOrDefault(channel) <- sendable
	netMes.goRoutinesThrottler.EndProcessing()
	return nil
}

// BroadcastOnChannel tries to send a byte buffer onto a topic using provided channel
func (netMes *networkMessenger) BroadcastOnChannel(channel string, topic string, buff []byte) {
	go func() {
		err := netMes.BroadcastOnChannelBlocking(channel, topic, buff)
		if err != nil {
			log.Debug("p2p broadcast", "error", err.Error())
		}
	}()
}

// Broadcast tries to send a byte buffer onto a topic using the topic name as channel
func (netMes *networkMessenger) Broadcast(topic string, buff []byte) {
	netMes.BroadcastOnChannel(topic, topic, buff)
}

// RegisterMessageProcessor registers a message process on a topic
func (netMes *networkMessenger) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	if check.IfNil(handler) {
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

	broadcastHandler := func(buffToSend []byte) {
		netMes.Broadcast(topic, buffToSend)
	}

	err := netMes.pb.RegisterTopicValidator(topic, func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
		wrappedMsg, err := NewMessage(message)
		if err != nil {
			log.Trace("p2p validator - new message", "error", err.Error(), "topics", message.TopicIDs)
			return false
		}
		err = handler.ProcessReceivedMessage(wrappedMsg, broadcastHandler)
		if err != nil {
			log.Trace("p2p validator",
				"error", err.Error(),
				"topics", message.TopicIDs,
				"pid", p2p.MessageOriginatorPid(wrappedMsg),
				"seq no", p2p.MessageOriginatorSeq(wrappedMsg),
			)

			return false
		}

		return true
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
		err := processor.ProcessReceivedMessage(msg, nil)
		if err != nil {
			log.Trace("p2p validator",
				"error", err.Error(),
				"topics", msg.TopicIDs(),
				"pid", p2p.MessageOriginatorPid(msg),
				"seq no", p2p.MessageOriginatorSeq(msg),
			)
		}
	}(message)

	return nil
}

// IsConnectedToTheNetwork returns true if the current node is connected to the network
func (netMes *networkMessenger) IsConnectedToTheNetwork() bool {
	netw := netMes.p2pHost.Network()
	return netMes.connMonitor.IsConnectedToTheNetwork(netw)
}

// SetThresholdMinConnectedPeers sets the minimum connected peers before triggering a new reconnection
func (netMes *networkMessenger) SetThresholdMinConnectedPeers(minConnectedPeers int) error {
	if minConnectedPeers < 0 {
		return p2p.ErrInvalidValue
	}

	netw := netMes.p2pHost.Network()
	netMes.connMonitor.SetThresholdMinConnectedPeers(minConnectedPeers, netw)

	return nil
}

// ThresholdMinConnectedPeers returns the minimum connected peers before triggering a new reconnection
func (netMes *networkMessenger) ThresholdMinConnectedPeers() int {
	return netMes.connMonitor.ThresholdMinConnectedPeers()
}

// SetPeerShardResolver sets the peer shard resolver component that is able to resolve the link
// between p2p.PeerID and shardId
func (netMes *networkMessenger) SetPeerShardResolver(peerShardResolver p2p.PeerShardResolver) error {
	if check.IfNil(peerShardResolver) {
		return p2p.ErrNilPeerShardResolver
	}

	err := netMes.sharder.SetPeerShardResolver(peerShardResolver)
	if err != nil {
		return err
	}

	netMes.peerShardResolver = peerShardResolver

	return nil
}

// GetConnectedPeersInfo gets the current connected peers information
func (netMes *networkMessenger) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	peers := netMes.p2pHost.Network().Peers()
	peerInfo := &p2p.ConnectedPeersInfo{
		UnknownPeers:    make([]string, 0),
		IntraShardPeers: make([]string, 0),
		CrossShardPeers: make([]string, 0),
	}
	selfShardId := netMes.peerShardResolver.ByID(netMes.ID())

	for _, p := range peers {
		conns := netMes.p2pHost.Network().ConnsToPeer(p)
		connString := "[invalid connection string]"
		if len(conns) > 0 {
			connString = conns[0].RemoteMultiaddr().String() + "/p2p/" + p.Pretty()
		}

		shardId := netMes.peerShardResolver.ByID(p2p.PeerID(p))
		switch shardId {
		case core.UnknownShardId:
			peerInfo.UnknownPeers = append(peerInfo.UnknownPeers, connString)
		case selfShardId:
			peerInfo.IntraShardPeers = append(peerInfo.IntraShardPeers, connString)
		default:
			peerInfo.CrossShardPeers = append(peerInfo.CrossShardPeers, connString)
		}
	}

	return peerInfo
}

// IsInterfaceNil returns true if there is no value under the interface
func (netMes *networkMessenger) IsInterfaceNil() bool {
	return netMes == nil
}
