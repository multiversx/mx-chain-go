package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

// MultipleHeaderSigningProofStub -
type MultipleHeaderSigningProofStub struct {
	GetTypeCalled func() slash.SlashingType
}

// GetType -
func (mps *MultipleHeaderSigningProofStub) GetType() slash.SlashingType {
	if mps.GetTypeCalled != nil {
		return mps.GetTypeCalled()
	}
	return slash.None
}

// GetPubKeys -
func (mps *MultipleHeaderSigningProofStub) GetPubKeys() [][]byte {
	return nil
}

// GetLevel -
func (mps *MultipleHeaderSigningProofStub) GetLevel(pubKey []byte) slash.ThreatLevel {
	return slash.Low
}

// GetHeaders -
func (mps *MultipleHeaderSigningProofStub) GetHeaders(pubKey []byte) []*interceptedBlocks.InterceptedHeader {
	return nil
}
