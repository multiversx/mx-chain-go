package slashMocks

import (
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
)

// SlashingProofStub -
type SlashingProofStub struct {
	GetTypeCalled func() coreSlash.SlashingType
}

// GetType -
func (sps *SlashingProofStub) GetType() coreSlash.SlashingType {
	if sps.GetTypeCalled != nil {
		return sps.GetTypeCalled()
	}

	return coreSlash.None
}
