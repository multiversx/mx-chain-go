package nodesCoordinator

// EpochValidators holds one epoch configuration for a nodes coordinator
type EpochValidators struct {
	EligibleValidators map[string][]*SerializableValidator `json:"eligibleValidators"`
	WaitingValidators  map[string][]*SerializableValidator `json:"waitingValidators"`
	LeavingValidators  map[string][]*SerializableValidator `json:"leavingValidators"`
}

// GetEligibleValidators returns all eligible validators from all shards
func (ev *EpochValidators) GetEligibleValidators() map[string][]*SerializableValidator {
	return ev.EligibleValidators
}

// GetWaitingValidators returns all waiting validators from all shards
func (ev *EpochValidators) GetWaitingValidators() map[string][]*SerializableValidator {
	return ev.WaitingValidators
}

// GetLeavingValidators returns all leaving validators from all shards
func (ev *EpochValidators) GetLeavingValidators() map[string][]*SerializableValidator {
	return ev.LeavingValidators
}

// NodesCoordinatorRegistry holds the data that can be used to initialize a nodes coordinator
type NodesCoordinatorRegistry struct {
	EpochsConfig map[string]*EpochValidators `json:"epochConfigs"`
	CurrentEpoch uint32                      `json:"currentEpoch"`
}

// GetCurrentEpoch returns the current epoch
func (ncr *NodesCoordinatorRegistry) GetCurrentEpoch() uint32 {
	return ncr.CurrentEpoch
}

// GetEpochsConfig returns epoch-validators configuration
func (ncr *NodesCoordinatorRegistry) GetEpochsConfig() map[string]EpochValidatorsHandler {
	ret := make(map[string]EpochValidatorsHandler)
	for epoch, config := range ncr.EpochsConfig {
		ret[epoch] = config
	}

	return ret
}

// SetCurrentEpoch sets internally the current epoch
func (ncr *NodesCoordinatorRegistry) SetCurrentEpoch(epoch uint32) {
	ncr.CurrentEpoch = epoch
}
