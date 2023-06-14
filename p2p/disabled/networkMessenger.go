package disabled

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
)

type networkMessenger struct {
}

// NewNetworkMessenger creates a new disabled Messenger implementation
func NewNetworkMessenger() *networkMessenger {
	return &networkMessenger{}
}

// Close returns nil as it is disabled
func (netMes *networkMessenger) Close() error {
	return nil
}

// CreateTopic returns nil as it is disabled
func (netMes *networkMessenger) CreateTopic(_ string, _ bool) error {
	return nil
}

// HasTopic returns true as it is disabled
func (netMes *networkMessenger) HasTopic(_ string) bool {
	return true
}

// RegisterMessageProcessor returns nil as it is disabled
func (netMes *networkMessenger) RegisterMessageProcessor(_ string, _ string, _ p2p.MessageProcessor) error {
	return nil
}

// UnregisterAllMessageProcessors returns nil as it is disabled
func (netMes *networkMessenger) UnregisterAllMessageProcessors() error {
	return nil
}

// UnregisterMessageProcessor returns nil as it is disabled
func (netMes *networkMessenger) UnregisterMessageProcessor(_ string, _ string) error {
	return nil
}

// Broadcast does nothing as it is disabled
func (netMes *networkMessenger) Broadcast(_ string, _ []byte) {
}

// BroadcastOnChannel does nothing as it is disabled
func (netMes *networkMessenger) BroadcastOnChannel(_ string, _ string, _ []byte) {
}

// BroadcastUsingPrivateKey does nothing as it is disabled
func (netMes *networkMessenger) BroadcastUsingPrivateKey(_ string, _ []byte, _ core.PeerID, _ []byte) {
}

// BroadcastOnChannelUsingPrivateKey does nothing as it is disabled
func (netMes *networkMessenger) BroadcastOnChannelUsingPrivateKey(_ string, _ string, _ []byte, _ core.PeerID, _ []byte) {
}

// SendToConnectedPeer returns nil as it is disabled
func (netMes *networkMessenger) SendToConnectedPeer(_ string, _ []byte, _ core.PeerID) error {
	return nil
}

// UnJoinAllTopics returns nil as it is disabled
func (netMes *networkMessenger) UnJoinAllTopics() error {
	return nil
}

// Bootstrap returns nil as it is disabled
func (netMes *networkMessenger) Bootstrap() error {
	return nil
}

// Peers returns an empty slice as it is disabled
func (netMes *networkMessenger) Peers() []core.PeerID {
	return make([]core.PeerID, 0)
}

// Addresses returns an empty slice as it is disabled
func (netMes *networkMessenger) Addresses() []string {
	return make([]string, 0)
}

// ConnectToPeer returns nil as it is disabled
func (netMes *networkMessenger) ConnectToPeer(_ string) error {
	return nil
}

// IsConnected returns false as it is disabled
func (netMes *networkMessenger) IsConnected(_ core.PeerID) bool {
	return false
}

// ConnectedPeers returns an empty slice as it is disabled
func (netMes *networkMessenger) ConnectedPeers() []core.PeerID {
	return make([]core.PeerID, 0)
}

// ConnectedAddresses returns an empty slice as it is disabled
func (netMes *networkMessenger) ConnectedAddresses() []string {
	return make([]string, 0)
}

// PeerAddresses returns an empty slice as it is disabled
func (netMes *networkMessenger) PeerAddresses(_ core.PeerID) []string {
	return make([]string, 0)
}

// ConnectedPeersOnTopic returns an empty slice as it is disabled
func (netMes *networkMessenger) ConnectedPeersOnTopic(_ string) []core.PeerID {
	return make([]core.PeerID, 0)
}

// SetPeerShardResolver returns nil as it is disabled
func (netMes *networkMessenger) SetPeerShardResolver(_ p2p.PeerShardResolver) error {
	return nil
}

// GetConnectedPeersInfo returns an empty structure as it is disabled
func (netMes *networkMessenger) GetConnectedPeersInfo() *p2p.ConnectedPeersInfo {
	return &p2p.ConnectedPeersInfo{}
}

// WaitForConnections does nothing as it is disabled
func (netMes *networkMessenger) WaitForConnections(_ time.Duration, _ uint32) {
}

// IsConnectedToTheNetwork returns true as it is disabled
func (netMes *networkMessenger) IsConnectedToTheNetwork() bool {
	return true
}

// ThresholdMinConnectedPeers returns 0 as it is disabled
func (netMes *networkMessenger) ThresholdMinConnectedPeers() int {
	return 0
}

// SetThresholdMinConnectedPeers returns nil as it is disabled
func (netMes *networkMessenger) SetThresholdMinConnectedPeers(_ int) error {
	return nil
}

// SetPeerDenialEvaluator returns nil as it is disabled
func (netMes *networkMessenger) SetPeerDenialEvaluator(_ p2p.PeerDenialEvaluator) error {
	return nil
}

// ID returns an empty peerID as it is disabled
func (netMes *networkMessenger) ID() core.PeerID {
	return ""
}

// Port returns 0 as it is disabled
func (netMes *networkMessenger) Port() int {
	return 0
}

// Sign returns an empty slice and nil as it is disabled
func (netMes *networkMessenger) Sign(_ []byte) ([]byte, error) {
	return make([]byte, 0), nil
}

// Verify returns nil as it is disabled
func (netMes *networkMessenger) Verify(_ []byte, _ core.PeerID, _ []byte) error {
	return nil
}

// SignUsingPrivateKey returns an empty slice and nil as it is disabled
func (netMes *networkMessenger) SignUsingPrivateKey(_ []byte, _ []byte) ([]byte, error) {
	return make([]byte, 0), nil
}

// AddPeerTopicNotifier returns nil as it is disabled
func (netMes *networkMessenger) AddPeerTopicNotifier(_ p2p.PeerTopicNotifier) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (netMes *networkMessenger) IsInterfaceNil() bool {
	return netMes == nil
}
