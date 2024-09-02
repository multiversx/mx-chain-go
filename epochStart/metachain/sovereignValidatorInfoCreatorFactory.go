package metachain

import "github.com/multiversx/mx-chain-go/process"

type sovereignValidatorInfoCreatorFactory struct {
}

// NewSovereignValidatorInfoCreatorFactory creates a validator info creator factory for sovereign run type chain
func NewSovereignValidatorInfoCreatorFactory() *sovereignValidatorInfoCreatorFactory {
	return &sovereignValidatorInfoCreatorFactory{}
}

// CreateValidatorInfoCreator creates a validator info creator for sovereign run type chain
func (f *sovereignValidatorInfoCreatorFactory) CreateValidatorInfoCreator(args ArgsNewValidatorInfoCreator) (process.EpochStartValidatorInfoCreator, error) {
	vic, err := NewValidatorInfoCreator(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignValidatorInfoCreator(vic)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignValidatorInfoCreatorFactory) IsInterfaceNil() bool {
	return f == nil
}
