package slashMocks

import (
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
)

// SlashingProofStub -
type SlashingProofStub struct {
	GetTypeCalled        func() coreSlash.SlashingType
	GetProofTxDataCalled func() (*coreSlash.ProofTxData, error)
}

// GetType -
func (sps *SlashingProofStub) GetType() coreSlash.SlashingType {
	if sps.GetTypeCalled != nil {
		return sps.GetTypeCalled()
	}

	return coreSlash.None
}

// GetProofTxData -
func (sps *SlashingProofStub) GetProofTxData() (*coreSlash.ProofTxData, error) {
	if sps.GetProofTxDataCalled != nil {
		return sps.GetProofTxDataCalled()
	}
	return &coreSlash.ProofTxData{}, nil
}
