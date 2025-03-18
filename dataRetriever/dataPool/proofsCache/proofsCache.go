package proofscache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type proofNonceMapping struct {
	headerHash string
	nonce      uint64
}

type proofsCache struct {
	mutProofsCache       sync.RWMutex
	proofsByNonceBuckets sync.Map
	bucketSize           uint64
	proofsByHash         map[string]data.HeaderProofHandler
}

func newProofsCache(bucketSize int) *proofsCache {
	return &proofsCache{
		proofsByNonceBuckets: sync.Map{},
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

func (pc *proofsCache) addProof(proof data.HeaderProofHandler) {
	if check.IfNil(proof) {
		return
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	pc.insertProofByNonce(proof)

	pc.proofsByHash[string(proof.GetHeaderHash())] = proof
}

// getBucketKey will return bucket key as lower bound window value
func (pc *proofsCache) getBucketKey(index uint64) uint64 {
	return (index / pc.bucketSize) * pc.bucketSize
}

func (pc *proofsCache) insertProofByNonce(proof data.HeaderProofHandler) {
	bucketKey := pc.getBucketKey(proof.GetHeaderNonce())

	bucket, _ := pc.proofsByNonceBuckets.LoadOrStore(bucketKey, newProofBucket())

	b := bucket.(*proofNonceBucket)
	b.insert(proof)
}

func (pc *proofsCache) cleanupProofsBehindNonce(nonce uint64) {
	if nonce == 0 {
		return
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	bucketsToDelete := make([]uint64, 0)

	pc.proofsByNonceBuckets.Range(func(key, value interface{}) bool {
		bucketKey := key.(uint64)
		bucket := value.(*proofNonceBucket)

		if nonce > bucket.maxNonce {
			pc.cleanupProofsInBucket(bucket)
			bucketsToDelete = append(bucketsToDelete, bucketKey)
			pc.proofsByNonceBuckets.Delete(key)
		}

		return true
	})

	for _, key := range bucketsToDelete {
		pc.proofsByNonceBuckets.Delete(key)
	}
}

func (pc *proofsCache) cleanupProofsInBucket(bucket *proofNonceBucket) {
	for _, headerHash := range bucket.proofsByNonce {
		delete(pc.proofsByHash, headerHash)
	}
}
