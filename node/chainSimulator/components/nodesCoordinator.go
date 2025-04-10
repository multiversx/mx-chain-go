package components

import "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"

type nodesCoordinatorWrapperHandler interface {
	nodesCoordinator.NodesCoordinator
	SetCustomPubKeys(pubKeys []string)
}

type nodesCoordinatorWrapper struct {
	nodesCoordinator.NodesCoordinator
	customPubKeys []string
}

// CreateNodesCoordinatorWrapper -
func CreateNodesCoordinatorWrapper(nodesCoordinator nodesCoordinator.NodesCoordinator) nodesCoordinatorWrapperHandler {
	return &nodesCoordinatorWrapper{
		NodesCoordinator: nodesCoordinator,
	}
}

// GetAllEligibleValidatorsPublicKeys -
func (ncw *nodesCoordinatorWrapper) GetAllEligibleValidatorsPublicKeysForShard(epoch uint32, shardID uint32) ([]string, error) {
	eligibleValidators, err := ncw.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	shardEligible := eligibleValidators[shardID]

	validatorsPubKeys := make([]string, 0, len(shardEligible))
	for i := 0; i < len(shardEligible); i++ {
		validatorsPubKeys = append(validatorsPubKeys, string(shardEligible[i]))
	}

	if len(ncw.customPubKeys) > 0 {
		validatorsPubKeys = append(validatorsPubKeys, ncw.customPubKeys...)
	}

	return validatorsPubKeys, nil
}

// SetCustomPubKeys -
func (ncw *nodesCoordinatorWrapper) SetCustomPubKeys(pubKeys []string) {
	ncw.customPubKeys = pubKeys
}
