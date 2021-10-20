package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

// MultipleHeaderProposalProofStub -
type MultipleHeaderProposalProofStub struct {
	GetTypeCalled    func() slash.SlashingType
	GetLevelCalled   func() slash.ThreatLevel
	GetHeadersCalled func() []*interceptedBlocks.InterceptedHeader
}

// GetType -
func (mps *MultipleHeaderProposalProofStub) GetType() slash.SlashingType {
	if mps.GetTypeCalled != nil {
		return mps.GetTypeCalled()
	}
	return slash.MultipleProposal
}

// GetLevel -
func (mps *MultipleHeaderProposalProofStub) GetLevel() slash.ThreatLevel {
	if mps.GetLevelCalled != nil {
		return mps.GetLevelCalled()
	}
	return slash.Low
}

// GetHeaders -
func (mps *MultipleHeaderProposalProofStub) GetHeaders() []*interceptedBlocks.InterceptedHeader {
	if mps.GetHeadersCalled != nil {
		return mps.GetHeadersCalled()
	}
	return nil
}
