package proofscache

import "github.com/multiversx/mx-chain-core-go/data"

// NewProofsCache -
func NewProofsCache(bucketSize int) *proofsCache {
	return newProofsCache(bucketSize)
}

// HeadBucketSize -
func (pc *proofsCache) FullProofsByNonceSize() int {
	size := 0

	for _, bucket := range pc.proofsByNonceBuckets {
		size += bucket.size()
	}

	return size
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
