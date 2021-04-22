package sharding

// IntraShardValidatorDistributor - distributes validators from source to destination inside the same shard
type IntraShardValidatorDistributor struct{}

// DistributeValidators will handle the moving of the nodes to the map for intra shard validator distributor
func (vd *IntraShardValidatorDistributor) DistributeValidators(
	destination map[uint32][]Validator,
	source map[uint32][]Validator,
	_ []byte,
	_ bool,
) error {
	return moveNodesToMap(destination, source)
}

// IsInterfaceNil - verifies if the interface is nil
func (vd *IntraShardValidatorDistributor) IsInterfaceNil() bool {
	return vd == nil
}

// CrossShardValidatorDistributor - distributes validators from source to destination cross shards
type CrossShardValidatorDistributor struct{}

// DistributeValidators will handle the moving of the nodes to the map for cross shard validator distributor
func (vd *CrossShardValidatorDistributor) DistributeValidators(
	destination map[uint32][]Validator,
	source map[uint32][]Validator,
	rand []byte,
	balanced bool,
) error {
	allValidators := make([]Validator, 0)
	for _, list := range source {
		allValidators = append(allValidators, list...)
	}

	return distributeValidators(destination, allValidators, rand, balanced)
}

// IsInterfaceNil - verifies if the interface is nil
func (vd *CrossShardValidatorDistributor) IsInterfaceNil() bool {
	return vd == nil
}
