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

func (m *EpochValidatorsWithAuction) GetEligibleValidators() map[string][]*SerializableValidator {
	return protoValidatorsMapToSliceMap(m.GetEligible())
}

func (m *EpochValidatorsWithAuction) GetWaitingValidators() map[string][]*SerializableValidator {
	return protoValidatorsMapToSliceMap(m.GetWaiting())
}

func (m *EpochValidatorsWithAuction) GetLeavingValidators() map[string][]*SerializableValidator {
	return protoValidatorsMapToSliceMap(m.GetLeaving())
}

func (m *EpochValidatorsWithAuction) GetShuffledOutValidators() map[string][]*SerializableValidator {
	return protoValidatorsMapToSliceMap(m.GetShuffledOut())
}

func (m *NodesCoordinatorRegistryWithAuction) GetEpochsConfig() map[string]EpochValidatorsHandler {
	ret := make(map[string]EpochValidatorsHandler)
	for epoch, config := range m.GetEpochsConfigWithAuction() {
		ret[epoch] = config
	}

	return ret
}

func (m *NodesCoordinatorRegistryWithAuction) SetCurrentEpoch(epoch uint32) {
	m.CurrentEpoch = epoch
}

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
