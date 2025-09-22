package processMocks

import "github.com/multiversx/mx-chain-core-go/data"

// HeaderValidatorMock -
type HeaderValidatorMock struct {
	IsHeaderConstructionValidCalled func(currHdr, prevHdr data.HeaderHandler) error
}

// IsHeaderConstructionValid -
func (hvm *HeaderValidatorMock) IsHeaderConstructionValid(currHdr, prevHdr data.HeaderHandler) error {
	if hvm.IsHeaderConstructionValidCalled != nil {
		return hvm.IsHeaderConstructionValidCalled(currHdr, prevHdr)
	}
	return nil
}

// IsInterfaceNil returns if underlying object is true
func (hvm *HeaderValidatorMock) IsInterfaceNil() bool {
	return hvm == nil
}
