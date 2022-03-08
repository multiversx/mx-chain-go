//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. nodesCoordinatorRegistryWithAuction.proto
package sharding

func protoValidatorsMapToSliceMap(validators map[string]Validators) map[string][]*SerializableValidator {
	ret := make(map[string][]*SerializableValidator)

	for shardID, val := range validators {
		ret[shardID] = val.GetData()
	}

	return ret
}

func sliceMapToProtoMap(validators map[string][]*SerializableValidator) map[string]Validators {
	ret := make(map[string]Validators)

	for shardID, val := range validators {
		ret[shardID] = Validators{Data: val}
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

// SetEpochsConfig sets internally epoch-validators configuration
func (m *NodesCoordinatorRegistryWithAuction) SetEpochsConfig(epochsConfig map[string]EpochValidatorsHandler) {
	m.EpochsConfigWithAuction = make(map[string]*EpochValidatorsWithAuction)

	for epoch, config := range epochsConfig {
		shuffledOut := make(map[string]Validators)
		configWithAuction, castOk := config.(EpochValidatorsHandlerWithAuction)
		if castOk {
			shuffledOut = sliceMapToProtoMap(configWithAuction.GetShuffledOutValidators())
		}

		m.EpochsConfigWithAuction[epoch] = &EpochValidatorsWithAuction{
			Eligible:    sliceMapToProtoMap(config.GetEligibleValidators()),
			Waiting:     sliceMapToProtoMap(config.GetWaitingValidators()),
			Leaving:     sliceMapToProtoMap(config.GetLeavingValidators()),
			ShuffledOut: shuffledOut,
		}
	}
}
