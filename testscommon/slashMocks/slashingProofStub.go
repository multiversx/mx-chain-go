package slashMocks

import (
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
)

// SlashingProofStub -
type SlashingProofStub struct {
	GetProofTxDataCalled func() (*coreSlash.ProofTxData, error)
}

// GetProofTxData -
func (sps *SlashingProofStub) GetProofTxData() (*coreSlash.ProofTxData, error) {
	if sps.GetProofTxDataCalled != nil {
		return sps.GetProofTxDataCalled()
	}
	return &coreSlash.ProofTxData{}, nil
}
