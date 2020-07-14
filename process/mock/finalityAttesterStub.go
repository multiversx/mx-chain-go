package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ValidityAttesterStub -
type ValidityAttesterStub struct {
	CheckBlockAgainstRounderCalled   func(headerHandler data.HeaderHandler) error
	CheckBlockAgainstFinalCalled     func(headerHandler data.HeaderHandler) error
	CheckBlockAgainstWhitelistCalled func(interceptedData process.InterceptedData) bool
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

// CheckBlockAgainstWhitelist -
func (vas *ValidityAttesterStub) CheckBlockAgainstWhitelist(interceptedData process.InterceptedData) bool {
	if vas.CheckBlockAgainstWhitelistCalled != nil {
		return vas.CheckBlockAgainstWhitelistCalled(interceptedData)
	}

	return false
}

// IsInterfaceNil -
func (vas *ValidityAttesterStub) IsInterfaceNil() bool {
	return vas == nil
}
