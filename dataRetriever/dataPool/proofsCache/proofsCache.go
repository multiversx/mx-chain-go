package proofscache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type proofsCache struct {
	mutProofsCache       sync.RWMutex
	proofsByNonceBuckets map[uint64]*proofNonceBucket
	bucketSize           uint64
	proofsByHash         map[string]data.HeaderProofHandler
}

func newProofsCache(bucketSize int) *proofsCache {
	return &proofsCache{
		proofsByNonceBuckets: make(map[uint64]*proofNonceBucket),
		bucketSize:           uint64(bucketSize),
		proofsByHash:         make(map[string]data.HeaderProofHandler),
	}
}

func (pc *proofsCache) getProofByHash(headerHash []byte) (data.HeaderProofHandler, error) {
	pc.mutProofsCache.RLock()
	defer pc.mutProofsCache.RUnlock()

	proof, ok := pc.proofsByHash[string(headerHash)]
	if !ok {
		return nil, ErrMissingProof
	}

	return proof, nil
}

func (pc *proofsCache) getProofByNonce(headerNonce uint64) (data.HeaderProofHandler, error) {
	pc.mutProofsCache.RLock()
	defer pc.mutProofsCache.RUnlock()

	bucketKey := pc.getBucketKey(headerNonce)
	bucket, ok := pc.proofsByNonceBuckets[bucketKey]
	if !ok {
		return nil, ErrMissingProof
	}

	proofHash, ok := bucket.proofsByNonce[headerNonce]
	if !ok {
		return nil, ErrMissingProof
	}

	proof, ok := pc.proofsByHash[proofHash]
	if !ok {
		return nil, ErrMissingProof
	}

	return proof, nil
}

func (pc *proofsCache) addProof(proof data.HeaderProofHandler) {
	if check.IfNil(proof) {
		return
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	oldHash := pc.insertProofByNonce(proof)
	newHash := string(proof.GetHeaderHash())

	// Delete the old hash from proofsByHash if it's different from the new hash
	if len(oldHash) != 0 && oldHash != newHash {
		delete(pc.proofsByHash, oldHash)
	}

	pc.proofsByHash[newHash] = proof
}

// getBucketKey will return bucket key as lower bound window value
func (pc *proofsCache) getBucketKey(index uint64) uint64 {
	return (index / pc.bucketSize) * pc.bucketSize
}

func (pc *proofsCache) insertProofByNonce(proof data.HeaderProofHandler) string {
	bucketKey := pc.getBucketKey(proof.GetHeaderNonce())

	bucket, ok := pc.proofsByNonceBuckets[bucketKey]
	if !ok {
		bucket = newProofBucket()
		pc.proofsByNonceBuckets[bucketKey] = bucket
	}

	return bucket.insert(proof)
}

func (pc *proofsCache) cleanupProofsBehindNonce(nonce uint64) {
	if nonce == 0 {
		return
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	for key, bucket := range pc.proofsByNonceBuckets {
		if nonce > bucket.maxNonce {
			pc.cleanupProofsInBucket(bucket)
			delete(pc.proofsByNonceBuckets, key)
		}
	}
}

func (pc *proofsCache) cleanupProofsInBucket(bucket *proofNonceBucket) {
	for _, headerHash := range bucket.proofsByNonce {
		delete(pc.proofsByHash, headerHash)
	}
}

func (pc *proofsCache) getNumProofsByHash() int {
	pc.mutProofsCache.RLock()
	defer pc.mutProofsCache.RUnlock()

	return len(pc.proofsByHash)
}
