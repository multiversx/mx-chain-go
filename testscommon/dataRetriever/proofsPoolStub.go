package dataRetriever

import "github.com/multiversx/mx-chain-core-go/data"

// ProofsPoolStub -
type ProofsPoolStub struct {
	AddNotarizedProofCalled                 func(headerProof data.HeaderProofHandler)
	CleanupNotarizedProofsBehindNonceCalled func(shardID uint32, nonce uint64) error
	GetNotarizedProofCalled                 func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error)
}

// AddNotarizedProof -
func (p *ProofsPoolStub) AddNotarizedProof(headerProof data.HeaderProofHandler) {
	if p.AddNotarizedProofCalled != nil {
		p.AddNotarizedProofCalled(headerProof)
	}
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

	return nil, nil
}

// IsInterfaceNil -
func (p *ProofsPoolStub) IsInterfaceNil() bool {
	return p == nil
}
