package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

// MultipleHeaderSigningProofStub -
type MultipleHeaderSigningProofStub struct {
	GetTypeCalled    func() slash.SlashingType
	GetPubKeysCalled func() [][]byte
	GetHeadersCalled func(pubKey []byte) []*interceptedBlocks.InterceptedHeader
	GetLevelCalled   func(pubKey []byte) slash.ThreatLevel
}

// GetType -
func (mps *MultipleHeaderSigningProofStub) GetType() slash.SlashingType {
	if mps.GetTypeCalled != nil {
		return mps.GetTypeCalled()
	}
	return slash.MultipleSigning
}

// GetPubKeys -
func (mps *MultipleHeaderSigningProofStub) GetPubKeys() [][]byte {
	if mps.GetPubKeysCalled != nil {
		return mps.GetPubKeysCalled()
	}
	return nil
}

// GetLevel -
func (mps *MultipleHeaderSigningProofStub) GetLevel(pubKey []byte) slash.ThreatLevel {
	if mps.GetLevelCalled != nil {
		return mps.GetLevelCalled(pubKey)
	}
	return slash.Low
}

// GetHeaders -
func (mps *MultipleHeaderSigningProofStub) GetHeaders(pubKey []byte) []*interceptedBlocks.InterceptedHeader {
	if mps.GetHeadersCalled != nil {
		return mps.GetHeadersCalled(pubKey)
	}
	return nil
}
