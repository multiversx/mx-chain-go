package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
)

// MultipleHeaderProposalProofStub -
type MultipleHeaderProposalProofStub struct {
	GetProofTxDataCalled func() (*coreSlash.ProofTxData, error)
	GetLevelCalled       func() coreSlash.ThreatLevel
	GetHeadersCalled     func() []data.HeaderHandler
}

// GetProofTxData -
func (mps *MultipleHeaderProposalProofStub) GetProofTxData() (*coreSlash.ProofTxData, error) {
	if mps.GetProofTxDataCalled != nil {
		return mps.GetProofTxDataCalled()
	}
	return &coreSlash.ProofTxData{ProofID: coreSlash.MultipleProposalProofID}, nil
}

// GetLevel -
func (mps *MultipleHeaderProposalProofStub) GetLevel() coreSlash.ThreatLevel {
	if mps.GetLevelCalled != nil {
		return mps.GetLevelCalled()
	}
	return coreSlash.Low
}

// GetHeaders -
func (mps *MultipleHeaderProposalProofStub) GetHeaders() []data.HeaderHandler {
	if mps.GetHeadersCalled != nil {
		return mps.GetHeadersCalled()
	}
	return nil
}
