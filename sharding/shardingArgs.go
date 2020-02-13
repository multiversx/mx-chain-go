package sharding

import (
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgNodesCoordinator holds all dependencies required by the nodes coordinator in order to create new instances
type ArgNodesCoordinator struct {
	ShardConsensusGroupSize int
	MetaConsensusGroupSize  int
	Hasher                  hashing.Hasher
	Shuffler                NodesShuffler
	EpochStartSubscriber    EpochStartSubscriber
	ListIndexUpdater        ListIndexUpdaterHandler
	BootStorer              storage.Storer
	ShardId                 uint32
	NbShards                uint32
	EligibleNodes           map[uint32][]Validator
	WaitingNodes            map[uint32][]Validator
	SelfPublicKey           []byte
	Epoch                   uint32
	ConsensusGroupCache     Cacher
}
