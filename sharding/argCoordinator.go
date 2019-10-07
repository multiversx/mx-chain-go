package sharding

import "github.com/ElrondNetwork/elrond-go/hashing"

// ArgNodesCoordinator holds all dependencies required by the nodes coordinator in order to create
//// new instances
type ArgNodesCoordinator struct {
	ShardConsensusGroupSize int
	MetaConsensusGroupSize  int
	Hasher                  hashing.Hasher
	ShardId                 uint32
	NbShards                uint32
	Nodes                   map[uint32][]Validator
	SelfPublicKey           []byte
}
