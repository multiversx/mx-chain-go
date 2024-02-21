package block

import "github.com/multiversx/mx-chain-go/process"

type shardHeaderValidatorFactory struct {
}

// NewShardHeaderValidatorFactory creates a new shard header validator factory
func NewShardHeaderValidatorFactory() (*shardHeaderValidatorFactory, error) {
	return &shardHeaderValidatorFactory{}, nil
}

// CreateHeaderValidator creates a new header validator for the chain run type normal
func (shvf *shardHeaderValidatorFactory) CreateHeaderValidator(argsHeaderValidator ArgsHeaderValidator) (process.HeaderConstructionValidator, error) {
	return NewHeaderValidator(argsHeaderValidator)
}

// IsInterfaceNil returns true if there is no value under the interface
func (shvf *shardHeaderValidatorFactory) IsInterfaceNil() bool {
	return shvf == nil
}
