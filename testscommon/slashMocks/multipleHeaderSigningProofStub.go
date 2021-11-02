package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
)

// MultipleHeaderSigningProofStub -
type MultipleHeaderSigningProofStub struct {
	GetTypeCalled    func() coreSlash.SlashingType
	GetPubKeysCalled func() [][]byte
	GetHeadersCalled func(pubKey []byte) []data.HeaderInfoHandler
	GetLevelCalled   func(pubKey []byte) coreSlash.ThreatLevel
}

// GetType -
func (mps *MultipleHeaderSigningProofStub) GetType() coreSlash.SlashingType {
	if mps.GetTypeCalled != nil {
		return mps.GetTypeCalled()
	}
	return coreSlash.MultipleSigning
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
func (mps *MultipleHeaderSigningProofStub) GetHeaders(pubKey []byte) []data.HeaderInfoHandler {
	if mps.GetHeadersCalled != nil {
		return mps.GetHeadersCalled(pubKey)
	}
	return nil
}
