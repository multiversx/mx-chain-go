package chainSimulator

import "github.com/multiversx/mx-chain-core-go/core"

// SyncedBroadcastNetworkHandler defines the synced network interface
type SyncedBroadcastNetworkHandler interface {
	RegisterMessageReceiver(handler messageReceiver, pid core.PeerID)
	Broadcast(pid core.PeerID, topic string, buff []byte)
	SendDirectly(from core.PeerID, topic string, buff []byte, to core.PeerID) error
	GetConnectedPeers() []core.PeerID
	GetConnectedPeersOnTopic(topic string) []core.PeerID
	IsInterfaceNil() bool
}
