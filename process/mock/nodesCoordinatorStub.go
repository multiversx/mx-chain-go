package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// NodesCoordinatorStub -
type NodesCoordinatorStub struct {
	GetValidatorWithPublicKeyCalled func(publicKey []byte) (validator core.Validator, shardId uint32, err error)
}

// GetValidatorWithPublicKey -
func (nc *NodesCoordinatorStub) GetValidatorWithPublicKey(publicKey []byte) (validator core.Validator, shardId uint32, err error) {
	if nc.GetValidatorWithPublicKeyCalled != nil {
		return nc.GetValidatorWithPublicKeyCalled(publicKey)
	}
	return nil, 0, nil
}

// IsInterfaceNil -
func (nc *NodesCoordinatorStub) IsInterfaceNil() bool {
	return false
}
