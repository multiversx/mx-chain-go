package mock

import "github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"

// NodesCoordinatorStub -
type NodesCoordinatorStub struct {
	GetAllEligibleValidatorsPublicKeysCalled func(epoch uint32) (map[uint32][][]byte, error)
	GetValidatorWithPublicKeyCalled          func(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error)
}

// GetAllEligibleValidatorsPublicKeys -
func (nc *NodesCoordinatorStub) GetAllEligibleValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	if nc.GetAllEligibleValidatorsPublicKeysCalled != nil {
		return nc.GetAllEligibleValidatorsPublicKeysCalled(epoch)
	}

	return nil, nil
}

// GetValidatorWithPublicKey -
func (nc *NodesCoordinatorStub) GetValidatorWithPublicKey(publicKey []byte) (validator nodesCoordinator.Validator, shardId uint32, err error) {
	if nc.GetValidatorWithPublicKeyCalled != nil {
		return nc.GetValidatorWithPublicKeyCalled(publicKey)
	}

	return nil, 0, nil
}

// IsInterfaceNil -
func (nc *NodesCoordinatorStub) IsInterfaceNil() bool {
	return nc == nil
}
