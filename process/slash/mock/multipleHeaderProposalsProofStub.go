package mock

import (
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

// MultipleProposalProofStub -
type MultipleProposalProofStub struct {
	GetTypeCalled    func() slash.SlashingType
	GetLevelCalled   func() slash.SlashingLevel
	GetHeadersCalled func() []*interceptedBlocks.InterceptedHeader
}

// GetLevel -
func (mps *MultipleProposalProofStub) GetLevel() slash.SlashingLevel {
	if mps.GetLevelCalled != nil {
		return mps.GetLevelCalled()
	}
	return slash.Level0
}

// GetHeaders -
func (mps *MultipleProposalProofStub) GetHeaders() []*interceptedBlocks.InterceptedHeader {
	if mps.GetHeadersCalled != nil {
		return mps.GetHeadersCalled()
	}
	return nil
}

// GetType -
func (mps *MultipleProposalProofStub) GetType() slash.SlashingType {
	if mps.GetTypeCalled != nil {
		return mps.GetTypeCalled()
	}
	return slash.None
}
