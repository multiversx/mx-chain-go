package components

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	p2pMessage "github.com/multiversx/mx-chain-communication-go/p2p/message"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
)

var (
	errNilHandler           = errors.New("nil handler")
	errHandlerAlreadyExists = errors.New("handler already exists")
	errUnknownPeer          = errors.New("unknown peer")
)

type messageReceiver interface {
	receive(fromConnectedPeer core.PeerID, message p2p.MessageP2P)
	HasTopic(name string) bool
}

type syncedBroadcastNetwork struct {
	*blockBodyDeliveryRegistry
	*blockHeaderDeliveryRegistry

	mutOperation    sync.RWMutex
	peers           map[core.PeerID]messageReceiver
	peerAliases     map[core.PeerID]messageReceiver
	headerNotifiers map[uint32][]func(header data.HeaderHandler)
	proofsBroadcast map[string]struct{}
	peerSignatures  *sharedPeerSignatureVerifier
}

// NewSyncedBroadcastNetwork creates a new synced broadcast network
func NewSyncedBroadcastNetwork() *syncedBroadcastNetwork {
	return &syncedBroadcastNetwork{
		blockBodyDeliveryRegistry:   newBlockBodyDeliveryRegistry(),
		blockHeaderDeliveryRegistry: newBlockHeaderDeliveryRegistry(),
		peers:                       make(map[core.PeerID]messageReceiver),
		peerAliases:                 make(map[core.PeerID]messageReceiver),
		headerNotifiers:             make(map[uint32][]func(header data.HeaderHandler)),
		proofsBroadcast:             make(map[string]struct{}),
		peerSignatures:              newSharedPeerSignatureVerifier(),
	}
}

func (network *syncedBroadcastNetwork) wrapPeerSignatureHandler(
	handler crypto.PeerSignatureHandler,
) crypto.PeerSignatureHandler {
	return network.peerSignatures.wrap(handler)
}

// RegisterHeaderNotifier registers the epoch-aware header callback of one simulated node.
func (network *syncedBroadcastNetwork) RegisterHeaderNotifier(shardID uint32, notifier func(header data.HeaderHandler)) {
	if notifier == nil {
		return
	}

	network.mutOperation.Lock()
	network.headerNotifiers[shardID] = append(network.headerNotifiers[shardID], notifier)
	network.mutOperation.Unlock()
}

// NotifyHeader synchronously notifies all simulated nodes in one shard about a header.
func (network *syncedBroadcastNetwork) NotifyHeader(shardID uint32, header data.HeaderHandler) {
	network.mutOperation.RLock()
	notifiers := append([]func(header data.HeaderHandler){}, network.headerNotifiers[shardID]...)
	network.mutOperation.RUnlock()

	for _, notifier := range notifiers {
		notifier(header)
	}
}

// RegisterMessageReceiver registers the message receiver
func (network *syncedBroadcastNetwork) RegisterMessageReceiver(handler messageReceiver, pid core.PeerID) {
	if handler == nil {
		log.Error("programming error in syncedBroadcastNetwork.RegisterMessageReceiver: %w", errNilHandler)
		return
	}

	network.mutOperation.Lock()
	defer network.mutOperation.Unlock()

	_, found := network.peers[pid]
	if found {
		log.Error("programming error in syncedBroadcastNetwork.RegisterMessageReceiver", "pid", pid.Pretty(), "error", errHandlerAlreadyExists)
		return
	}

	network.peers[pid] = handler
}

// RegisterPeerAlias routes direct messages addressed to a managed validator key's virtual peer ID
// to the physical node that owns the key. Aliases are deliberately kept outside peers: broadcasts
// must visit a physical receiver only once even when that receiver manages multiple validator keys.
func (network *syncedBroadcastNetwork) RegisterPeerAlias(alias core.PeerID, target core.PeerID) error {
	network.mutOperation.Lock()
	defer network.mutOperation.Unlock()

	handler, found := network.peers[target]
	if !found {
		return fmt.Errorf("syncedBroadcastNetwork.RegisterPeerAlias: %w, target pid %s", errUnknownPeer, target.Pretty())
	}
	if _, found = network.peers[alias]; found {
		return fmt.Errorf("syncedBroadcastNetwork.RegisterPeerAlias: %w, alias pid %s", errHandlerAlreadyExists, alias.Pretty())
	}
	if existing, aliasFound := network.peerAliases[alias]; aliasFound && existing != handler {
		return fmt.Errorf("syncedBroadcastNetwork.RegisterPeerAlias: %w, alias pid %s", errHandlerAlreadyExists, alias.Pretty())
	}

	network.peerAliases[alias] = handler

	return nil
}

// Broadcast will iterate through peers and send the message
func (network *syncedBroadcastNetwork) Broadcast(pid core.PeerID, topic string, buff []byte) {
	if network.proofAlreadyBroadcast(topic, buff) {
		return
	}

	peers, handlers := network.getPeersAndHandlers()

	for idx, handler := range handlers {
		// Production pubsub does not feed a node's own consensus broadcast back through its inbound
		// worker. The consensus subround records the sender's local job before broadcasting, so
		// simulator loopback only repeats message parsing and BLS validation. Keep historical
		// loopback for every non-consensus topic.
		if strings.HasPrefix(topic, common.ConsensusTopic) && peers[idx] == pid {
			continue
		}

		message := &p2pMessage.Message{
			FromField:            pid.Bytes(),
			DataField:            buff,
			TopicField:           topic,
			BroadcastMethodField: p2p.Broadcast,
			PeerField:            pid,
			// the libp2p envelope signature is produced and verified by pubsub, which this in-memory
			// network replaces; the consensus worker only checks the envelope signature is present
			// (the message's own BLS signatures remain real and are verified separately), so a
			// non-empty stand-in derived from the sender is enough for multi-party consensus messages
			SignatureField: pid.Bytes(),
		}

		handler.receive(pid, message)
	}
}

// proofAlreadyBroadcast suppresses identical equivalent-proof floods. All consensus members can
// assemble the same proof at once under the simulator's synchronized drive; production pubsub and
// network latency naturally make the first propagated proof stop the other senders, while the
// in-memory network would otherwise synchronously deliver N identical proofs to every node. Besides
// unnecessary work, that flood can starve metachain finality callbacks behind DEBUG log I/O.
func (network *syncedBroadcastNetwork) proofAlreadyBroadcast(topic string, buff []byte) bool {
	if !strings.HasPrefix(topic, common.EquivalentProofsTopic) {
		return false
	}

	key := topic + "\x00" + string(buff)
	network.mutOperation.Lock()
	defer network.mutOperation.Unlock()

	if _, exists := network.proofsBroadcast[key]; exists {
		return true
	}

	network.proofsBroadcast[key] = struct{}{}
	return false
}

// SendDirectly will try to send directly to the provided peer
func (network *syncedBroadcastNetwork) SendDirectly(from core.PeerID, topic string, buff []byte, to core.PeerID) error {
	network.mutOperation.RLock()
	handler, found := network.peers[to]
	if !found {
		handler, found = network.peerAliases[to]
		if !found {
			network.mutOperation.RUnlock()

			return fmt.Errorf("syncedBroadcastNetwork.SendDirectly: %w, pid %s", errUnknownPeer, to.Pretty())
		}
	}
	network.mutOperation.RUnlock()

	message := &p2pMessage.Message{
		FromField:            from.Bytes(),
		DataField:            buff,
		TopicField:           topic,
		BroadcastMethodField: p2p.Direct,
		PeerField:            from,
		// see Broadcast for why the envelope signature is stamped with a sender-derived stand-in
		SignatureField: from.Bytes(),
	}

	handler.receive(from, message)

	return nil
}

// GetConnectedPeers returns all connected peers
func (network *syncedBroadcastNetwork) GetConnectedPeers() []core.PeerID {
	peers, _ := network.getPeersAndHandlers()

	network.mutOperation.RLock()
	for alias := range network.peerAliases {
		peers = append(peers, alias)
	}
	network.mutOperation.RUnlock()

	return peers
}

func (network *syncedBroadcastNetwork) getPeersAndHandlers() ([]core.PeerID, []messageReceiver) {
	network.mutOperation.RLock()
	defer network.mutOperation.RUnlock()

	peers := make([]core.PeerID, 0, len(network.peers))
	handlers := make([]messageReceiver, 0, len(network.peers))

	for p, handler := range network.peers {
		peers = append(peers, p)
		handlers = append(handlers, handler)
	}

	return peers, handlers
}

// GetConnectedPeersOnTopic will find suitable peers connected on the provided topic
func (network *syncedBroadcastNetwork) GetConnectedPeersOnTopic(topic string) []core.PeerID {
	peers, handlers := network.getPeersAndHandlers()

	peersOnTopic := make([]core.PeerID, 0, len(peers))
	for idx, p := range peers {
		if handlers[idx].HasTopic(topic) {
			peersOnTopic = append(peersOnTopic, p)
		}
	}

	network.mutOperation.RLock()
	for alias, handler := range network.peerAliases {
		if handler.HasTopic(topic) {
			peersOnTopic = append(peersOnTopic, alias)
		}
	}
	network.mutOperation.RUnlock()

	return peersOnTopic
}

// IsInterfaceNil returns true if there is no value under the interface
func (network *syncedBroadcastNetwork) IsInterfaceNil() bool {
	return network == nil
}
