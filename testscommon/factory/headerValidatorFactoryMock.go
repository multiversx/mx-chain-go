package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
)

// HeaderValidatorFactoryMock -
type HeaderValidatorFactoryMock struct {
	CreateHeaderValidatorCalled func(args processBlock.ArgsHeaderValidator) (process.HeaderConstructionValidator, error)
}

// CreateHeaderValidator -
func (h *HeaderValidatorFactoryMock) CreateHeaderValidator(args processBlock.ArgsHeaderValidator) (process.HeaderConstructionValidator, error) {
	if h.CreateHeaderValidatorCalled != nil {
		return h.CreateHeaderValidatorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (h *HeaderValidatorFactoryMock) IsInterfaceNil() bool {
	return h == nil
}
