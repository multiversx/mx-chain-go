package mock

import "github.com/ElrondNetwork/elrond-go/data"

type HeaderValidatorStub struct {
	IsHeaderConstructionValidCalled func(currHdr, prevHdr data.HeaderHandler) error
}

func (hvs *HeaderValidatorStub) IsHeaderConstructionValid(currHdr, prevHdr data.HeaderHandler) error {
	if hvs.IsHeaderConstructionValidCalled != nil {
		return hvs.IsHeaderConstructionValidCalled(currHdr, prevHdr)
	}
	return nil
}

// IsInterfaceNil returns if underlying object is true
func (hvs *HeaderValidatorStub) IsInterfaceNil() bool {
	return hvs == nil
}
