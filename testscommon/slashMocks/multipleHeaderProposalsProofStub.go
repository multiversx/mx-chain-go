package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
)

// MultipleHeaderProposalProofStub -
type MultipleHeaderProposalProofStub struct {
	GetTypeCalled    func() coreSlash.SlashingType
	GetLevelCalled   func() coreSlash.ThreatLevel
	GetHeadersCalled func() []data.HeaderHandler
}

// GetType -
func (mps *MultipleHeaderProposalProofStub) GetType() coreSlash.SlashingType {
	if mps.GetTypeCalled != nil {
		return mps.GetTypeCalled()
	}
	return coreSlash.MultipleProposal
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
