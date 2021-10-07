package mock

import (
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

// MultipleHeaderProposalProofStub -
type MultipleHeaderProposalProofStub struct {
	GetTypeCalled func() slash.SlashingType
}

// GetType -
func (mps *MultipleHeaderProposalProofStub) GetType() slash.SlashingType {
	if mps.GetTypeCalled != nil {
		return mps.GetTypeCalled()
	}
	return slash.None
}

// GetLevel -
func (mps *MultipleHeaderProposalProofStub) GetLevel() slash.ThreatLevel {
	return slash.Low
}

// GetHeaders -
func (mps *MultipleHeaderProposalProofStub) GetHeaders() []*interceptedBlocks.InterceptedHeader {
	return nil
}
