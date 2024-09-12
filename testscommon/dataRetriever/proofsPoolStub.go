package dataRetriever

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// ProofsPoolStub -
type ProofsPoolStub struct {
	AddNotarizedProofCalled                 func(headerProof data.HeaderProofHandler) error
	CleanupNotarizedProofsBehindNonceCalled func(shardID uint32, nonce uint64) error
	GetNotarizedProofCalled                 func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error)
	GetAllNotarizedProofsCalled             func(shardID uint32) (map[string]data.HeaderProofHandler, error)
}

// AddNotarizedProof -
func (p *ProofsPoolStub) AddNotarizedProof(headerProof data.HeaderProofHandler) error {
	if p.AddNotarizedProofCalled != nil {
		return p.AddNotarizedProofCalled(headerProof)
	}

	return nil
}

// CleanupNotarizedProofsBehindNonce -
func (p *ProofsPoolStub) CleanupNotarizedProofsBehindNonce(shardID uint32, nonce uint64) error {
	if p.CleanupNotarizedProofsBehindNonceCalled != nil {
		return p.CleanupNotarizedProofsBehindNonceCalled(shardID, nonce)
	}

	return nil
}

// GetNotarizedProof -
func (p *ProofsPoolStub) GetNotarizedProof(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
	if p.GetNotarizedProofCalled != nil {
		return p.GetNotarizedProofCalled(shardID, headerHash)
	}

	return &block.HeaderProof{}, nil
}

// GetAllNotarizedProofs -
func (p *ProofsPoolStub) GetAllNotarizedProofs(shardID uint32) (map[string]data.HeaderProofHandler, error) {
	if p.GetAllNotarizedProofsCalled != nil {
		return p.GetAllNotarizedProofsCalled(shardID)
	}

	return make(map[string]data.HeaderProofHandler), nil
}

// IsInterfaceNil -
func (p *ProofsPoolStub) IsInterfaceNil() bool {
	return p == nil
}
