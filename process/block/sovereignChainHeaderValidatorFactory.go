package block

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignHeaderValidatorFactory struct {
	headerValidatorCreator HeaderValidatorCreator
}

// NewSovereignHeaderValidatorFactory creates a new shard header validator factory
func NewSovereignHeaderValidatorFactory(headerValidatorCreator HeaderValidatorCreator) (*sovereignHeaderValidatorFactory, error) {
	if check.IfNil(headerValidatorCreator) {
		return nil, process.ErrNilHeaderValidatorCreator
	}

	return &sovereignHeaderValidatorFactory{
		headerValidatorCreator: headerValidatorCreator,
	}, nil
}

// CreateHeaderValidator creates a new header validator for the chain run type sovereign
func (shvf *sovereignHeaderValidatorFactory) CreateHeaderValidator(argsHeaderValidator ArgsHeaderValidator) (process.HeaderConstructionValidator, error) {
	hv, err := shvf.headerValidatorCreator.CreateHeaderValidator(argsHeaderValidator)
	if err != nil {
		return nil, err
	}

	shardHeaderValidator, ok := hv.(*headerValidator)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return NewSovereignChainHeaderValidator(shardHeaderValidator)
}

// IsInterfaceNil returns true if there is no value under the interface
func (shvf *sovereignHeaderValidatorFactory) IsInterfaceNil() bool {
	return shvf == nil
}
