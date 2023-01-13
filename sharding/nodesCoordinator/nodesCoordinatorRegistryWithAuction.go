//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. nodesCoordinatorRegistryWithAuction.proto
package nodesCoordinator

func protoValidatorsMapToSliceMap(validators map[string]Validators) map[string][]*SerializableValidator {
	ret := make(map[string][]*SerializableValidator)

	for shardID, val := range validators {
		ret[shardID] = val.GetData()
	}

	return ret
}

// GetEligibleValidators returns all eligible validators from all shards
func (m *EpochValidatorsWithAuction) GetEligibleValidators() map[string][]*SerializableValidator {
	return protoValidatorsMapToSliceMap(m.GetEligible())
}

// GetWaitingValidators returns all waiting validators from all shards
func (m *EpochValidatorsWithAuction) GetWaitingValidators() map[string][]*SerializableValidator {
	return protoValidatorsMapToSliceMap(m.GetWaiting())
}

// GetLeavingValidators returns all leaving validators from all shards
func (m *EpochValidatorsWithAuction) GetLeavingValidators() map[string][]*SerializableValidator {
	return protoValidatorsMapToSliceMap(m.GetLeaving())
}

// GetShuffledOutValidators returns all shuffled out validators from all shards
func (m *EpochValidatorsWithAuction) GetShuffledOutValidators() map[string][]*SerializableValidator {
	return protoValidatorsMapToSliceMap(m.GetShuffledOut())
}

// GetEpochsConfig returns epoch-validators configuration
func (m *NodesCoordinatorRegistryWithAuction) GetEpochsConfig() map[string]EpochValidatorsHandler {
	ret := make(map[string]EpochValidatorsHandler)
	for epoch, config := range m.GetEpochsConfigWithAuction() {
		ret[epoch] = config
	}

	return ret
}

// SetCurrentEpoch sets internally the current epoch
func (m *NodesCoordinatorRegistryWithAuction) SetCurrentEpoch(epoch uint32) {
	m.CurrentEpoch = epoch
}
