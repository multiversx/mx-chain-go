package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
)

// HeaderValidatorFactoryStub -
type HeaderValidatorFactoryStub struct {
	CreateHeaderValidatorCalled func(args processBlock.ArgsHeaderValidator) (process.HeaderConstructionValidator, error)
}

// NewHeaderValidatorFactoryStub -
func NewHeaderValidatorFactoryStub() *HeaderValidatorFactoryStub {
	return &HeaderValidatorFactoryStub{}
}

// CreateHeaderValidator -
func (h *HeaderValidatorFactoryStub) CreateHeaderValidator(args processBlock.ArgsHeaderValidator) (process.HeaderConstructionValidator, error) {
	if h.CreateHeaderValidatorCalled != nil {
		return h.CreateHeaderValidatorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (h *HeaderValidatorFactoryStub) IsInterfaceNil() bool {
	return false
}
