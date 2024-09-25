package dataRetriever

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// ProofsPoolMock -
type ProofsPoolMock struct {
	AddProofCalled                 func(headerProof data.HeaderProofHandler) error
	CleanupProofsBehindNonceCalled func(shardID uint32, nonce uint64) error
	GetProofCalled                 func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error)
	HasProofCalled                 func(shardID uint32, headerHash []byte) bool
}

// AddProof -
func (p *ProofsPoolMock) AddProof(headerProof data.HeaderProofHandler) error {
	if p.AddProofCalled != nil {
		return p.AddProofCalled(headerProof)
	}

	return nil
}

// CleanupProofsBehindNonce -
func (p *ProofsPoolMock) CleanupProofsBehindNonce(shardID uint32, nonce uint64) error {
	if p.CleanupProofsBehindNonceCalled != nil {
		return p.CleanupProofsBehindNonceCalled(shardID, nonce)
	}

	return nil
}

// GetProof -
func (p *ProofsPoolMock) GetProof(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
	if p.GetProofCalled != nil {
		return p.GetProofCalled(shardID, headerHash)
	}

	return &block.HeaderProof{}, nil
}

// HasProof -
func (p *ProofsPoolMock) HasProof(shardID uint32, headerHash []byte) bool {
	if p.HasProofCalled != nil {
		return p.HasProofCalled(shardID, headerHash)
	}

	return false
}

// IsInterfaceNil -
func (p *ProofsPoolMock) IsInterfaceNil() bool {
	return p == nil
}
