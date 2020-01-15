package heartbeat

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

// PeerMessenger defines a subset of the p2p.Messenger interface
type PeerMessenger interface {
	Broadcast(topic string, buff []byte)
	IsInterfaceNil() bool
}

// MessageHandler defines what a message processor for heartbeat should do
type MessageHandler interface {
	CreateHeartbeatFromP2PMessage(message p2p.MessageP2P) (*Heartbeat, error)
	IsInterfaceNil() bool
}

//Timer defines an interface for tracking time
type Timer interface {
	Now() time.Time
	IsInterfaceNil() bool
}

// HeartbeatStorageHandler defines what a heartbeat's storer should do
type HeartbeatStorageHandler interface {
	LoadGenesisTime() (time.Time, error)
	UpdateGenesisTime(genesisTime time.Time) error
	LoadHbmiDTO(pubKey string) (*HeartbeatDTO, error)
	SavePubkeyData(pubkey []byte, heartbeat *HeartbeatDTO) error
	LoadKeys() ([][]byte, error)
	SaveKeys(peersSlice [][]byte) error
	IsInterfaceNil() bool
}

// P2PAntifloodHandler defines the behavior of a component able to signal that the system is too busy (or flooded) processing
// p2p messages
type P2PAntifloodHandler interface {
	CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error
	CanProcessMessageOnTopic(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID, topic string) error
	IsInterfaceNil() bool
}
