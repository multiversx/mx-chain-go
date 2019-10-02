package heartbeat

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"time"
)

// PeerMessenger defines a subset of the p2p.Messenger interface
type PeerMessenger interface {
	Broadcast(topic string, buff []byte)
	IsInterfaceNil() bool
}

// MessageHandler defines what a message processor for heartbeat should do
type MessageHandler interface {
	CreateHeartbeatFromP2pMessage(message p2p.MessageP2P) (*Heartbeat, error)
	IsInterfaceNil() bool
}

type HeartbeatStorageHandler interface {
	LoadGenesisTime() (time.Time, error)
	UpdateGenesisTime(genesisTime time.Time) error
	LoadPubkeysData() (map[string]*HeartbeatDTO, error)
	LoadHbmiDTO(pubKey string) (*HeartbeatDTO, error)
	SavePubkeyData(pubkey []byte, heartbeat *HeartbeatDTO) error
	LoadKeys() ([][]byte, error)
	SaveKeys(peersSlice [][]byte) error
	IsInterfaceNil() bool
}
