package processMocks

import "github.com/multiversx/mx-chain-core-go/data"

// EquivalentProofsPoolMock -
type EquivalentProofsPoolMock struct {
	AddNotarizedProofCalled func(headerProof data.HeaderProofHandler)
	GetNotarizedProofCalled func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error)
}

// AddNotarizedProof -
func (mock *EquivalentProofsPoolMock) AddNotarizedProof(headerProof data.HeaderProofHandler) {
	if mock.AddNotarizedProofCalled != nil {
		mock.AddNotarizedProofCalled(headerProof)
	}
}

// GetNotarizedProof -
func (mock *EquivalentProofsPoolMock) GetNotarizedProof(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
	if mock.GetNotarizedProofCalled != nil {
		return mock.GetNotarizedProofCalled(shardID, headerHash)
	}
	return nil, nil
}

// IsInterfaceNil -
func (mock *EquivalentProofsPoolMock) IsInterfaceNil() bool {
	return mock == nil
}
