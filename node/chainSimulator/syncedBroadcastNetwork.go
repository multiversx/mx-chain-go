package chainSimulator

import (
	"errors"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	p2pMessage "github.com/multiversx/mx-chain-communication-go/p2p/message"
	"github.com/multiversx/mx-chain-core-go/core"
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
	mutOperation sync.RWMutex
	peers        map[core.PeerID]messageReceiver
}

// NewSyncedBroadcastNetwork creates a new synced broadcast network
func NewSyncedBroadcastNetwork() *syncedBroadcastNetwork {
	return &syncedBroadcastNetwork{
		peers: make(map[core.PeerID]messageReceiver),
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

// Broadcast will iterate through peers and send the message
func (network *syncedBroadcastNetwork) Broadcast(pid core.PeerID, topic string, buff []byte) {
	_, handlers := network.getPeersAndHandlers()

	for _, handler := range handlers {
		message := &p2pMessage.Message{
			FromField:            pid.Bytes(),
			DataField:            buff,
			TopicField:           topic,
			BroadcastMethodField: p2p.Broadcast,
		}

		handler.receive(pid, message)
	}
}

// SendDirectly will try to send directly to the provided peer
func (network *syncedBroadcastNetwork) SendDirectly(from core.PeerID, topic string, buff []byte, to core.PeerID) error {
	network.mutOperation.RLock()
	handler, found := network.peers[to]
	if !found {
		network.mutOperation.RUnlock()

		return fmt.Errorf("syncedBroadcastNetwork.SendDirectly: %w, pid %s", errUnknownPeer, to.Pretty())
	}
	network.mutOperation.RUnlock()

	message := &p2pMessage.Message{
		FromField:            from.Bytes(),
		DataField:            buff,
		TopicField:           topic,
		BroadcastMethodField: p2p.Direct,
	}

	handler.receive(from, message)

	return nil
}

// GetConnectedPeers returns all connected peers
func (network *syncedBroadcastNetwork) GetConnectedPeers() []core.PeerID {
	peers, _ := network.getPeersAndHandlers()

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

	return peersOnTopic
}

// IsInterfaceNil returns true if there is no value under the interface
func (network *syncedBroadcastNetwork) IsInterfaceNil() bool {
	return network == nil
}
