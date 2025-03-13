package proofscache

import "github.com/multiversx/mx-chain-core-go/data"

// NewProofsCache -
func NewProofsCache() *proofsCache {
	return newProofsCache()
}

// ProofsByNonceSize -
func (pc *proofsCache) ProofsByNonceSize() int {
	return len(pc.proofsByNonce)
}

// ProofsByHashSize -
func (pc *proofsCache) ProofsByHashSize() int {
	return len(pc.proofsByHash)
}

// AddProof -
func (pc *proofsCache) AddProof(proof data.HeaderProofHandler) {
	pc.addProof(proof)
}

// CleanupProofsBehindNonce -
func (pc *proofsCache) CleanupProofsBehindNonce(nonce uint64) {
	pc.cleanupProofsBehindNonce(nonce)
}
