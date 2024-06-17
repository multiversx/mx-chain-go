package sharding

type sovereignShardCoordinatorFactory struct {
}

// NewSovereignShardCoordinatorFactory creates a shard coordinator factory for sovereign chain
func NewSovereignShardCoordinatorFactory() *sovereignShardCoordinatorFactory {
	return &sovereignShardCoordinatorFactory{}
}

// CreateShardCoordinator creates a shard coordinator for sovereign chain
func (ssc *sovereignShardCoordinatorFactory) CreateShardCoordinator(_, _ uint32) (Coordinator, error) {
	return NewSovereignShardCoordinator(), nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ssc *sovereignShardCoordinatorFactory) IsInterfaceNil() bool {
	return ssc == nil
}
