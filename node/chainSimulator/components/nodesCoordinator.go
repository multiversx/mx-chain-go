package components

import (
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
)

type PubKeyShard struct {
	PubKey  string
	ShardID uint32
}

type NodesCoordinatorWrapperHandler interface {
	nodesCoordinator.NodesCoordinator
	SetCustomPubKeys(pubKeys []PubKeyShard)
}

type nodesCoordinatorWrapper struct {
	nodesCoordinator.NodesCoordinator
	customPubKeys []PubKeyShard
}

// CreateNodesCoordinatorWrapper -
func CreateNodesCoordinatorWrapper(nodesCoordinator nodesCoordinator.NodesCoordinator) NodesCoordinatorWrapperHandler {
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

	for _, pubKey := range ncw.customPubKeys {
		validatorsPubKeys = append(validatorsPubKeys, pubKey.PubKey)
	}

	return validatorsPubKeys, nil
}

func (ncw *nodesCoordinatorWrapper) GetValidatorWithPublicKey(publicKey []byte) (nodesCoordinator.Validator, uint32, error) {
	validKey := false
	shardID := uint32(0)
	for _, pubKey := range ncw.customPubKeys {
		if pubKey.PubKey == string(publicKey) {
			validKey = true
			shardID = pubKey.ShardID
			break
		}
	}

	if !validKey {
		return nil, 0, nodesCoordinator.ErrValidatorNotFound
	}

	validator := shardingMocks.NewValidatorMock(publicKey, 1, 1)
	return validator, shardID, nil
}

// SetCustomPubKeys -
func (ncw *nodesCoordinatorWrapper) SetCustomPubKeys(pubKeys []PubKeyShard) {
	ncw.customPubKeys = pubKeys
}
