package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessengerStub -
type MessengerStub struct {
	ConnectedPeersCalled func() []p2p.PeerID
}

// Close -
func (m *MessengerStub) Close() error {
	return nil
}

// ID -
func (m *MessengerStub) ID() p2p.PeerID {
	return "peer_id"
}

// Peers -
func (m *MessengerStub) Peers() []p2p.PeerID {
	return nil
}

// Addresses -
func (m *MessengerStub) Addresses() []string {
	return nil
}

// ConnectToPeer -
func (m *MessengerStub) ConnectToPeer(address string) error {
	return nil
}

// IsConnected -
func (m *MessengerStub) IsConnected(peerID p2p.PeerID) bool {
	return true
}

// ConnectedPeers -
func (m *MessengerStub) ConnectedPeers() []p2p.PeerID {
	if m.ConnectedPeersCalled != nil {
		return m.ConnectedPeersCalled()
	}

	return []p2p.PeerID{"peer0", "peer1", "peer2", "peer3", "peer4", "peer5"}
}

// ConnectedAddresses -
func (m *MessengerStub) ConnectedAddresses() []string {
	return nil
}

// PeerAddress -
func (m *MessengerStub) PeerAddress(pid p2p.PeerID) string {
	return "addr"
}

// ConnectedPeersOnTopic -
func (m *MessengerStub) ConnectedPeersOnTopic(topic string) []p2p.PeerID {
	return nil
}

// Bootstrap -
func (m *MessengerStub) Bootstrap() error {
	return nil
}

// CreateTopic -
func (m *MessengerStub) CreateTopic(name string, createChannelForTopic bool) error {
	return nil
}

// HasTopic -
func (m *MessengerStub) HasTopic(name string) bool {
	return false
}

// HasTopicValidator -
func (m *MessengerStub) HasTopicValidator(name string) bool {
	return false
}

// RegisterMessageProcessor -
func (m *MessengerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	return nil
}

// UnregisterAllMessageProcessors -
func (m *MessengerStub) UnregisterAllMessageProcessors() error {
	return nil
}

// UnregisterMessageProcessor -
func (m *MessengerStub) UnregisterMessageProcessor(topic string) error {
	return nil
}

// OutgoingChannelLoadBalancer -
func (m *MessengerStub) OutgoingChannelLoadBalancer() p2p.ChannelLoadBalancer {
	return nil
}

// BroadcastOnChannelBlocking -
func (m *MessengerStub) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	return nil
}

// BroadcastOnChannel -
func (m *MessengerStub) BroadcastOnChannel(channel string, topic string, buff []byte) {
}

// Broadcast -
func (m *MessengerStub) Broadcast(topic string, buff []byte) {
}

// SendToConnectedPeer -
func (m *MessengerStub) SendToConnectedPeer(topic string, buff []byte, peerID p2p.PeerID) error {
	return nil
}

// IsConnectedToTheNetwork -
func (m *MessengerStub) IsConnectedToTheNetwork() bool {
	return false
}

// ThresholdMinConnectedPeers -
func (m *MessengerStub) ThresholdMinConnectedPeers() int {
	return 0
}

// SetThresholdMinConnectedPeers -
func (m *MessengerStub) SetThresholdMinConnectedPeers(minConnectedPeers int) error {
	return nil
}

// SetPeerShardResolver -
func (m *MessengerStub) SetPeerShardResolver(peerShardResolver p2p.PeerShardResolver) error {
	return nil
}

// SetPeerBlackListHandler -
func (m *MessengerStub) SetPeerBlackListHandler(handler p2p.BlacklistHandler) error {
	return nil
}

// GetConnectedPeersInfo -
func (m *MessengerStub) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	return nil
}

// IsInterfaceNil -
func (m *MessengerStub) IsInterfaceNil() bool {
	return m == nil
}
