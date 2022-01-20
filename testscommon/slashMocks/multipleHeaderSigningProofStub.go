package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
)

// MultipleHeaderSigningProofStub -
type MultipleHeaderSigningProofStub struct {
	GetTypeCalled        func() coreSlash.SlashingType
	GetProofTxDataCalled func() (*coreSlash.ProofTxData, error)
	GetPubKeysCalled     func() [][]byte
	GetHeadersCalled     func(pubKey []byte) []data.HeaderHandler
	GetLevelCalled       func(pubKey []byte) coreSlash.ThreatLevel
}

// GetType -
func (mps *MultipleHeaderSigningProofStub) GetType() coreSlash.SlashingType {
	if mps.GetTypeCalled != nil {
		return mps.GetTypeCalled()
	}
	return coreSlash.MultipleSigning
}

// GetProofTxData -
func (mps *MultipleHeaderSigningProofStub) GetProofTxData() (*coreSlash.ProofTxData, error) {
	if mps.GetProofTxDataCalled != nil {
		return mps.GetProofTxDataCalled()
	}
	return &coreSlash.ProofTxData{}, nil
}

// GetPubKeys -
func (mps *MultipleHeaderSigningProofStub) GetPubKeys() [][]byte {
	if mps.GetPubKeysCalled != nil {
		return mps.GetPubKeysCalled()
	}
	return nil
}

// GetLevel -
func (mps *MultipleHeaderSigningProofStub) GetLevel(pubKey []byte) coreSlash.ThreatLevel {
	if mps.GetLevelCalled != nil {
		return mps.GetLevelCalled(pubKey)
	}
	return coreSlash.Low
}

// GetHeaders -
func (mps *MultipleHeaderSigningProofStub) GetHeaders(pubKey []byte) []data.HeaderHandler {
	if mps.GetHeadersCalled != nil {
		return mps.GetHeadersCalled(pubKey)
	}
	return nil
}
