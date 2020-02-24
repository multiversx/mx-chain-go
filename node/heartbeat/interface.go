package heartbeat

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
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

// EligibleListProvider defines what an eligible list provider should do
type EligibleListProvider interface {
	GetEligiblePublicKeysPerShard(epoch uint32) (map[uint32][][]byte, error)
	GetWaitingPublicKeysPerShard(epoch uint32) (map[uint32][][]byte, error)
	IsInterfaceNil() bool
}

// PeerTypeProviderHandler defines what a component which computes the type of a peer should do
type PeerTypeProviderHandler interface {
	ComputeForPubKey(pubKey []byte, shardID uint32) (core.ValidatorList, error)
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
