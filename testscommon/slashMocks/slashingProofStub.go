package slashMocks

import "github.com/ElrondNetwork/elrond-go/process/slash"

// SlashingProofStub -
type SlashingProofStub struct {
	GetTypeCalled func() slash.SlashingType
}

// GetType -
func (sps *SlashingProofStub) GetType() slash.SlashingType {
	if sps.GetTypeCalled != nil {
		return sps.GetTypeCalled()
	}

	return slash.None
}
