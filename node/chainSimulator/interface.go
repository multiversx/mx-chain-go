package chainSimulator

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
)

// SyncedBroadcastNetworkHandler defines the synced network interface
type SyncedBroadcastNetworkHandler interface {
	RegisterMessageReceiver(handler messageReceiver, pid core.PeerID)
	Broadcast(pid core.PeerID, message p2p.MessageP2P)
	SendDirectly(from core.PeerID, topic string, buff []byte, to core.PeerID) error
	GetConnectedPeers() []core.PeerID
	GetConnectedPeersOnTopic(topic string) []core.PeerID
	IsInterfaceNil() bool
}
