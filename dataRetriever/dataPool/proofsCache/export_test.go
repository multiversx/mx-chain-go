package proofscache

import "github.com/multiversx/mx-chain-core-go/data"

// MaxProofsPerNonce -
const MaxProofsPerNonce = maxProofsPerNonce

// NewProofsCache -
func NewProofsCache(bucketSize int) *proofsCache {
	return newProofsCache(bucketSize)
}

// NewProofBucket -
func NewProofBucket() *proofNonceBucket {
	return newProofBucket()
}

// Insert -
func (p *proofNonceBucket) Insert(nonce uint64, hash string) {
	p.insert(nonce, hash)
}

// Remove -
func (p *proofNonceBucket) Remove(nonce uint64, hash string) {
	p.remove(nonce, hash)
}

// Size -
func (p *proofNonceBucket) Size() int {
	return p.size()
}

// HashesAt -
func (p *proofNonceBucket) HashesAt(nonce uint64) []string {
	return p.hashesAt(nonce)
}

// MaxNonce -
func (p *proofNonceBucket) MaxNonce() uint64 {
	return p.maxNonce
}

// TrackedNonces -
func (p *proofNonceBucket) TrackedNonces() int {
	return len(p.proofsByNonce)
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

// GetProofByHash -
func (pc *proofsCache) GetProofByHash(headerHash []byte) (data.HeaderProofHandler, error) {
	return pc.getProofByHash(headerHash)
}

// GetProofByNonce -
func (pc *proofsCache) GetProofByNonce(headerNonce uint64) (data.HeaderProofHandler, error) {
	return pc.getProofByNonce(headerNonce)
}
