package proofscache

import "github.com/multiversx/mx-chain-core-go/data"

// NewProofsCache -
func NewProofsCache(bucketSize int) *proofsCache {
	return newProofsCache(bucketSize)
}

// HeadBucketSize -
func (pc *proofsCache) FullProofsByNonceSize() int {
	size := 0

	pc.proofsByNonceBuckets.Range(func(key, value interface{}) bool {
		bucket := value.(*proofNonceBucket)
		size += bucket.size()

		return true
	})

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
