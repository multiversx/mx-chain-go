package components

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/crypto"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const virtualAddressTemplate = "/virtual/p2p/%s"

var (
	log                    = logger.GetOrCreate("node/chainSimulator")
	p2pInstanceCreator, _  = crypto.NewIdentityGenerator(log)
	hasher                 = blake2b.NewBlake2b()
	errNilNetwork          = errors.New("nil network")
	errTopicAlreadyCreated = errors.New("topic already created")
	errNilMessageProcessor = errors.New("nil message processor")
	errTopicNotCreated     = errors.New("topic not created")
	errTopicHasProcessor   = errors.New("there is already a message processor for provided topic and identifier")
	errInvalidSignature    = errors.New("invalid signature")
)

type syncedMessenger struct {
	mutOperation sync.RWMutex
	topics       map[string]map[string]p2p.MessageProcessor
	network      SyncedBroadcastNetworkHandler
	pid          core.PeerID
}

// NewSyncedMessenger creates a new synced network messenger
func NewSyncedMessenger(network SyncedBroadcastNetworkHandler) (*syncedMessenger, error) {
	if check.IfNil(network) {
		return nil, errNilNetwork
	}

	_, pid, err := p2pInstanceCreator.CreateRandomP2PIdentity()
	if err != nil {
		return nil, err
	}

	messenger := &syncedMessenger{
		network: network,
		topics:  make(map[string]map[string]p2p.MessageProcessor),
		pid:     pid,
	}

	log.Debug("created syncedMessenger", "pid", pid.Pretty())

	network.RegisterMessageReceiver(messenger, pid)

	return messenger, nil
}

// HasCompatibleProtocolID returns false as it is disabled
func (messenger *syncedMessenger) HasCompatibleProtocolID(_ string) bool {
	return false
}

func (messenger *syncedMessenger) receive(fromConnectedPeer core.PeerID, message p2p.MessageP2P) {
	if check.IfNil(message) {
		return
	}

	messenger.mutOperation.RLock()
	handlers := messenger.topics[message.Topic()]
	messenger.mutOperation.RUnlock()

	for _, handler := range handlers {
		err := handler.ProcessReceivedMessage(message, fromConnectedPeer, messenger)
		if err != nil {
			log.Trace("received message syncedMessenger",
				"error", err, "topic", message.Topic(), "from connected peer", fromConnectedPeer.Pretty())
		}
	}
}

// ProcessReceivedMessage does nothing and returns nil
func (messenger *syncedMessenger) ProcessReceivedMessage(_ p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) error {
	return nil
}

// CreateTopic will create a topic for receiving data
func (messenger *syncedMessenger) CreateTopic(name string, _ bool) error {
	messenger.mutOperation.Lock()
	defer messenger.mutOperation.Unlock()

	_, found := messenger.topics[name]
	if found {
		return fmt.Errorf("programming error in syncedMessenger.CreateTopic, %w for topic %s", errTopicAlreadyCreated, name)
	}

	messenger.topics[name] = make(map[string]p2p.MessageProcessor)

	return nil
}

// HasTopic returns true if the topic was registered
func (messenger *syncedMessenger) HasTopic(name string) bool {
	messenger.mutOperation.RLock()
	defer messenger.mutOperation.RUnlock()

	_, found := messenger.topics[name]

	return found
}

// RegisterMessageProcessor will try to register a message processor on the provided topic & identifier
func (messenger *syncedMessenger) RegisterMessageProcessor(topic string, identifier string, handler p2p.MessageProcessor) error {
	if check.IfNil(handler) {
		return fmt.Errorf("programming error in syncedMessenger.RegisterMessageProcessor, "+
			"%w for topic %s and identifier %s", errNilMessageProcessor, topic, identifier)
	}

	messenger.mutOperation.Lock()
	defer messenger.mutOperation.Unlock()

	handlers, found := messenger.topics[topic]
	if !found {
		handlers = make(map[string]p2p.MessageProcessor)
		messenger.topics[topic] = handlers
	}

	_, found = handlers[identifier]
	if found {
		return fmt.Errorf("programming error in syncedMessenger.RegisterMessageProcessor, %w, topic %s, identifier %s",
			errTopicHasProcessor, topic, identifier)
	}

	handlers[identifier] = handler

	return nil
}

// UnregisterAllMessageProcessors will unregister all message processors
func (messenger *syncedMessenger) UnregisterAllMessageProcessors() error {
	messenger.mutOperation.Lock()
	defer messenger.mutOperation.Unlock()

	for topic := range messenger.topics {
		messenger.topics[topic] = make(map[string]p2p.MessageProcessor)
	}

	return nil
}

// UnregisterMessageProcessor will unregister the message processor for the provided topic and identifier
func (messenger *syncedMessenger) UnregisterMessageProcessor(topic string, identifier string) error {
	messenger.mutOperation.Lock()
	defer messenger.mutOperation.Unlock()

	handlers, found := messenger.topics[topic]
	if !found {
		return fmt.Errorf("programming error in syncedMessenger.UnregisterMessageProcessor, %w for topic %s",
			errTopicNotCreated, topic)
	}

	delete(handlers, identifier)

	return nil
}

// Broadcast will broadcast the provided buffer on the topic in a synchronous manner
func (messenger *syncedMessenger) Broadcast(topic string, buff []byte) {
	if !messenger.HasTopic(topic) {
		return
	}

	messenger.network.Broadcast(messenger.pid, topic, buff)
}

// BroadcastOnChannel calls the Broadcast method
func (messenger *syncedMessenger) BroadcastOnChannel(_ string, topic string, buff []byte) {
	messenger.Broadcast(topic, buff)
}

// BroadcastUsingPrivateKey calls the Broadcast method
func (messenger *syncedMessenger) BroadcastUsingPrivateKey(topic string, buff []byte, _ core.PeerID, _ []byte) {
	messenger.Broadcast(topic, buff)
}

// BroadcastOnChannelUsingPrivateKey calls the Broadcast method
func (messenger *syncedMessenger) BroadcastOnChannelUsingPrivateKey(_ string, topic string, buff []byte, _ core.PeerID, _ []byte) {
	messenger.Broadcast(topic, buff)
}

// SendToConnectedPeer will send the message to the peer
func (messenger *syncedMessenger) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	if !messenger.HasTopic(topic) {
		return nil
	}

	log.Trace("syncedMessenger.SendToConnectedPeer",
		"from", messenger.pid.Pretty(),
		"to", peerID.Pretty(),
		"data", buff)

	return messenger.network.SendDirectly(messenger.pid, topic, buff, peerID)
}

// UnJoinAllTopics will unjoin all topics
func (messenger *syncedMessenger) UnJoinAllTopics() error {
	messenger.mutOperation.Lock()
	defer messenger.mutOperation.Unlock()

	messenger.topics = make(map[string]map[string]p2p.MessageProcessor)

	return nil
}

// Bootstrap does nothing and returns nil
func (messenger *syncedMessenger) Bootstrap() error {
	return nil
}

// Peers returns the network's peer ID
func (messenger *syncedMessenger) Peers() []core.PeerID {
	return messenger.network.GetConnectedPeers()
}

// Addresses returns the addresses this messenger was bound to. It returns a virtual address
func (messenger *syncedMessenger) Addresses() []string {
	return []string{fmt.Sprintf(virtualAddressTemplate, messenger.pid.Pretty())}
}

// ConnectToPeer does nothing and returns nil
func (messenger *syncedMessenger) ConnectToPeer(_ string) error {
	return nil
}

// IsConnected returns true if the peer ID is found on the network
func (messenger *syncedMessenger) IsConnected(peerID core.PeerID) bool {
	peers := messenger.network.GetConnectedPeers()
	for _, peer := range peers {
		if peer == peerID {
			return true
		}
	}

	return false
}

// ConnectedPeers returns the same list as the function Peers
func (messenger *syncedMessenger) ConnectedPeers() []core.PeerID {
	return messenger.Peers()
}

// ConnectedAddresses returns all connected addresses
func (messenger *syncedMessenger) ConnectedAddresses() []string {
	peers := messenger.network.GetConnectedPeers()
	addresses := make([]string, 0, len(peers))
	for _, peer := range peers {
		addresses = append(addresses, fmt.Sprintf(virtualAddressTemplate, peer.Pretty()))
	}

	return addresses
}

// PeerAddresses returns the virtual peer address
func (messenger *syncedMessenger) PeerAddresses(pid core.PeerID) []string {
	return []string{fmt.Sprintf(virtualAddressTemplate, pid.Pretty())}
}

// ConnectedPeersOnTopic returns the connected peers on the provided topic
func (messenger *syncedMessenger) ConnectedPeersOnTopic(topic string) []core.PeerID {
	return messenger.network.GetConnectedPeersOnTopic(topic)
}

// SetPeerShardResolver does nothing and returns nil
func (messenger *syncedMessenger) SetPeerShardResolver(_ p2p.PeerShardResolver) error {
	return nil
}

// GetConnectedPeersInfo return current connected peers info
func (messenger *syncedMessenger) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	peersInfo := &p2p.ConnectedPeersInfo{}
	peers := messenger.network.GetConnectedPeers()
	for _, peer := range peers {
		peersInfo.UnknownPeers = append(peersInfo.UnknownPeers, peer.Pretty())
	}

	return peersInfo
}

// WaitForConnections does nothing
func (messenger *syncedMessenger) WaitForConnections(_ time.Duration, _ uint32) {
}

// IsConnectedToTheNetwork returns true
func (messenger *syncedMessenger) IsConnectedToTheNetwork() bool {
	return true
}

// ThresholdMinConnectedPeers returns 0
func (messenger *syncedMessenger) ThresholdMinConnectedPeers() int {
	return 0
}

// SetThresholdMinConnectedPeers does nothing and returns nil
func (messenger *syncedMessenger) SetThresholdMinConnectedPeers(_ int) error {
	return nil
}

// SetPeerDenialEvaluator does nothing and returns nil
func (messenger *syncedMessenger) SetPeerDenialEvaluator(_ p2p.PeerDenialEvaluator) error {
	return nil
}

// ID returns the peer ID
func (messenger *syncedMessenger) ID() core.PeerID {
	return messenger.pid
}

// Port returns 0
func (messenger *syncedMessenger) Port() int {
	return 0
}

// Sign will return the hash(messenger.ID + payload)
func (messenger *syncedMessenger) Sign(payload []byte) ([]byte, error) {
	return hasher.Compute(messenger.pid.Pretty() + string(payload)), nil
}

// Verify will check if the provided signature === hash(pid + payload)
func (messenger *syncedMessenger) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	sig := hasher.Compute(pid.Pretty() + string(payload))
	if bytes.Equal(sig, signature) {
		return nil
	}

	return errInvalidSignature
}

// SignUsingPrivateKey will return an empty byte slice
func (messenger *syncedMessenger) SignUsingPrivateKey(_ []byte, _ []byte) ([]byte, error) {
	return make([]byte, 0), nil
}

// AddPeerTopicNotifier does nothing and returns nil
func (messenger *syncedMessenger) AddPeerTopicNotifier(_ p2p.PeerTopicNotifier) error {
	return nil
}

// SetDebugger will set the provided debugger
func (messenger *syncedMessenger) SetDebugger(_ p2p.Debugger) error {
	return nil
}

// Close does nothing and returns nil
func (messenger *syncedMessenger) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (messenger *syncedMessenger) IsInterfaceNil() bool {
	return messenger == nil
}
