package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// ValidityAttesterStub -
type ValidityAttesterStub struct {
	CheckBlockAgainstRounderCalled func(headerHandler data.HeaderHandler) error
	CheckBlockAgainstFinalCalled   func(headerHandler data.HeaderHandler) error
}

// CheckBlockAgainstRounder -
func (vas *ValidityAttesterStub) CheckBlockAgainstRounder(headerHandler data.HeaderHandler) error {
	if vas.CheckBlockAgainstRounderCalled != nil {
		return vas.CheckBlockAgainstRounderCalled(headerHandler)
	}

	return nil
}

// CheckBlockAgainstFinal -
func (vas *ValidityAttesterStub) CheckBlockAgainstFinal(headerHandler data.HeaderHandler) error {
	if vas.CheckBlockAgainstFinalCalled != nil {
		return vas.CheckBlockAgainstFinalCalled(headerHandler)
	}

	return nil
}

// IsInterfaceNil -
func (vas *ValidityAttesterStub) IsInterfaceNil() bool {
	return vas == nil
}
