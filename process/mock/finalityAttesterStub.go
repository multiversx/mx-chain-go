package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process"
)

// ValidityAttesterStub -
type ValidityAttesterStub struct {
	CheckBlockAgainstRoundHandlerCalled func(headerHandler data.HeaderHandler) error
	CheckBlockAgainstFinalCalled        func(headerHandler data.HeaderHandler) error
	CheckAgainstWhitelistCalled         func(interceptedData process.InterceptedData) bool
	CheckProofAgainstFinalCalled        func(proof data.HeaderProofHandler) error
	CheckProofAgainstRoundHandlerCalled func(proof data.HeaderProofHandler) error
}

// CheckBlockAgainstRoundHandler -
func (vas *ValidityAttesterStub) CheckBlockAgainstRoundHandler(headerHandler data.HeaderHandler) error {
	if vas.CheckBlockAgainstRoundHandlerCalled != nil {
		return vas.CheckBlockAgainstRoundHandlerCalled(headerHandler)
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

// CheckAgainstWhitelist -
func (vas *ValidityAttesterStub) CheckAgainstWhitelist(interceptedData process.InterceptedData) bool {
	if vas.CheckAgainstWhitelistCalled != nil {
		return vas.CheckAgainstWhitelistCalled(interceptedData)
	}

	return false
}

// CheckProofAgainstFinal -
func (vas *ValidityAttesterStub) CheckProofAgainstFinal(proof data.HeaderProofHandler) error {
	if vas.CheckProofAgainstFinalCalled != nil {
		return vas.CheckProofAgainstFinalCalled(proof)
	}

	return nil
}

// CheckProofAgainstRoundHandler -
func (vas *ValidityAttesterStub) CheckProofAgainstRoundHandler(proof data.HeaderProofHandler) error {
	if vas.CheckProofAgainstRoundHandlerCalled != nil {
		return vas.CheckProofAgainstRoundHandlerCalled(proof)
	}

	return nil
}

// IsInterfaceNil -
func (vas *ValidityAttesterStub) IsInterfaceNil() bool {
	return vas == nil
}
