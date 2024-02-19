package sharding

type sovereignShardCoordinator struct {
	*multiShardCoordinator
}

// NewSovereignShardCoordinator creates a new sovereign shard coordinator
func NewSovereignShardCoordinator(selfId uint32) *sovereignShardCoordinator {
	sr := &sovereignShardCoordinator{
		multiShardCoordinator: &multiShardCoordinator{
			selfId:         selfId,
			numberOfShards: 1,
		},
	}
	sr.maskHigh, sr.maskLow = sr.calculateMasks()
	return sr
}

// ComputeId returns the self sovereign shard id
func (ssc *sovereignShardCoordinator) ComputeId(_ []byte) uint32 {
	return ssc.selfId
}

// SameShard returns true if the addresses are in the same shard
func (msc *sovereignShardCoordinator) SameShard(_, _ []byte) bool {
	return true
}
