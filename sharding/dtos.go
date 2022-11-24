package sharding

// ConsensusConfiguration holds the consensus configuration that can be used by both the shard and the metachain
type ConsensusConfiguration struct {
	EnableEpoch        uint32 `json:"enableEpoch"`
	MinNodes           uint32 `json:"minNodes"`
	ConsensusGroupSize uint32 `json:"consensusGroupSize"`
}

// NodesSetupDTO is the data transfer object used to map the node's configuration in regard to the genesis nodes setup
type NodesSetupDTO struct {
	StartTime      int64                    `json:"startTime"`
	RoundDuration  uint64                   `json:"roundDuration"`
	ShardConsensus []ConsensusConfiguration `json:"shardConsensusConfiguration"`
	MetaConsensus  []ConsensusConfiguration `json:"metachainConsensusConfiguration"`
	Hysteresis     float32                  `json:"hysteresis"`
	Adaptivity     bool                     `json:"adaptivity"`

	InitialNodes []*InitialNode `json:"initialNodes"`
}
