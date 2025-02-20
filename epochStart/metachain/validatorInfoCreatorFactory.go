package metachain

import "github.com/multiversx/mx-chain-go/process"

type validatorInfoCreatorFactory struct {
}

// NewValidatorInfoCreatorFactory creates a validator info creator factory for normal run type chain
func NewValidatorInfoCreatorFactory() *validatorInfoCreatorFactory {
	return &validatorInfoCreatorFactory{}
}

// CreateValidatorInfoCreator creates a validator info creator for normal run type chain
func (f *validatorInfoCreatorFactory) CreateValidatorInfoCreator(args ArgsNewValidatorInfoCreator) (process.EpochStartValidatorInfoCreator, error) {
	return NewValidatorInfoCreator(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *validatorInfoCreatorFactory) IsInterfaceNil() bool {
	return f == nil
}
