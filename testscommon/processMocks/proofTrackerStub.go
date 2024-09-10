package processMocks

import "github.com/multiversx/mx-chain-core-go/data"

// ProofTrackerStub -
type ProofTrackerStub struct {
	AddNotarizedProofCalled                 func(notarizedHeaderHash []byte, notarizedProof data.HeaderProof, nonce uint64)
	CleanupNotarizedProofsBehindNonceCalled func(shardID uint32, nonce uint64)
	GetNotarizedProofCalled                 func(headerHash []byte) (data.HeaderProof, error)
}

// AddNotarizedProof -
func (p *ProofTrackerStub) AddNotarizedProof(notarizedHeaderHash []byte, notarizedProof data.HeaderProof, nonce uint64) {
	if p.AddNotarizedProofCalled != nil {
		p.AddNotarizedProofCalled(notarizedHeaderHash, notarizedProof, nonce)
	}
}

// CleanupNotarizedProofsBehindNonce -
func (p *ProofTrackerStub) CleanupNotarizedProofsBehindNonce(shardID uint32, nonce uint64) {
	if p.CleanupNotarizedProofsBehindNonceCalled != nil {
		p.CleanupNotarizedProofsBehindNonceCalled(shardID, nonce)
	}
}

// GetNotarizedProof -
func (p *ProofTrackerStub) GetNotarizedProof(headerHash []byte) (data.HeaderProof, error) {
	if p.GetNotarizedProofCalled != nil {
		return p.GetNotarizedProofCalled(headerHash)
	}

	return data.HeaderProof{}, nil
}

// IsInterfaceNil -
func (p *ProofTrackerStub) IsInterfaceNil() bool {
	return p == nil
}
