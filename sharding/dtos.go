package sharding

// ConsensusConfiguration holds the consensus configuration that can be used by both the shard and the metachain
type ConsensusConfiguration struct {
	EnableEpoch        uint32
	MinNodes           uint32
	ConsensusGroupSize uint32
}

// NodesSetupDTO is the data transfer object used to map the nodes' configuration in regard to the genesis nodes setup
type NodesSetupDTO struct {
	StartTime     int64   `json:"startTime"`
	RoundDuration uint64  `json:"roundDuration"`
	Hysteresis    float32 `json:"hysteresis"`
	Adaptivity    bool    `json:"adaptivity"`

	InitialNodes []*InitialNode `json:"initialNodes"`
}
