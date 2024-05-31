package sharding

type multiShardCoordinatorFactory struct {
}

// NewMultiShardCoordinatorFactory creates a shard coordinator factory for regular chain(meta+shards)
func NewMultiShardCoordinatorFactory() *multiShardCoordinatorFactory {
	return &multiShardCoordinatorFactory{}
}

// CreateShardCoordinator creates a shard coordinator for regular chain
func (msc *multiShardCoordinatorFactory) CreateShardCoordinator(numberOfShards, selfId uint32) (Coordinator, error) {
	return NewMultiShardCoordinator(numberOfShards, selfId)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (msc *multiShardCoordinatorFactory) IsInterfaceNil() bool {
	return msc == nil
}
