package memp2p

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type MemP2PNetwork struct {
	Mutex sync.Mutex
	Peers map[string]MemP2PMessenger
}

func NewMemP2PNetwork() *MemP2PNetwork {
	return &MemP2PNetwork{
		Mutex: sync.Mutex{},
		Peers: make(map[string]MemP2PMessenger),
	}
}

func (network *MemP2PNetwork) ListPeers() []string {
	peerIDs := make([]string, len(network.Peers))
	i := 0
	for key, _ := range network.Peers {
		peerIDs[i] = key
		i++
	}
	return peerIDs
}

func (network *MemP2PNetwork) ListAddresses() []string {
	addresses := make([]string, len(network.Peers))
	i := 0
	for key, _ := range network.Peers {
		addresses[i] = fmt.Sprintf("/memp2p/%s", key)
		i++
	}
	return addresses
}

func (network *MemP2PNetwork) RegisterPeer(messenger MemP2PMessenger) {
	network.Peers[string(messenger.ID())] = messenger
}

type MemP2PMessenger struct {
	Network *MemP2PNetwork
	P2P_ID  p2p.PeerID
	Address string
}

func NewMemP2PMessenger(network *MemP2PNetwork) MemP2PMessenger {
	ID := fmt.Sprintf("Peer%d", len(network.Peers)+1)
	Address := fmt.Sprintf("/memp2p/%s", string(ID))

	messenger := MemP2PMessenger{
		Network: network,
		P2P_ID:  p2p.PeerID(ID),
		Address: Address,
	}

	network.RegisterPeer(messenger)

	return messenger
}

func (messenger MemP2PMessenger) ID() p2p.PeerID {
	return messenger.P2P_ID
}

func (messenger MemP2PMessenger) Peers() []string {
	return messenger.Network.ListPeers()
}

func (messenger MemP2PMessenger) Addresses() []string {
	addresses := make([]string, 1)
	addresses[0] = messenger.Address
	return addresses
}

func (messenger MemP2PMessenger) ConnectToPeer() {
	// Do nothing, all peers are connected to each other already.
}

func (messenger MemP2PMessenger) IsConnected(peerID p2p.PeerID) bool {
	return true
}

func (messenger MemP2PMessenger) ConnectedPeers() []string {
	return messenger.Network.ListPeers()
}

func (messenger MemP2PMessenger) ConnectedAddresses() []string {
	return messenger.Network.ListAddresses()
}

func (messenger MemP2PMessenger) PeerAddress(pid p2p.PeerID) string {
	return fmt.Sprintf("/memp2p/%s", string(pid))
}

func (messenger MemP2PMessenger) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
}

func (messenger MemP2PMessenger) Close() {
	// Remove messenger from the network it references.
}
