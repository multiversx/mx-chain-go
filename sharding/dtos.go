package sharding

// NodesSetupDTO is the data transfer object used to map the nodes' configuration in regard to the genesis nodes setup
type NodesSetupDTO struct {
	StartTime     int64   `json:"startTime"`
	RoundDuration uint64  `json:"roundDuration"`
	Hysteresis    float32 `json:"hysteresis"`
	Adaptivity    bool    `json:"adaptivity"`

	InitialNodes []*InitialNode `json:"initialNodes"`
}
