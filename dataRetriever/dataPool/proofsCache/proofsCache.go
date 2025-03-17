package proofscache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
)

const bucketSize = 100

type proofNonceMapping struct {
	headerHash string
	nonce      uint64
}

type proofsCache struct {
	mutProofsCache       sync.RWMutex
	proofsByNonceBuckets []*proofNonceBucket
	bucketSize           int
	proofsByHash         map[string]data.HeaderProofHandler
}

func newProofsCache(bucketSize int) *proofsCache {

	return &proofsCache{
		mutProofsCache:       sync.RWMutex{},
		proofsByNonceBuckets: make([]*proofNonceBucket, 0),
		proofsByHash:         make(map[string]data.HeaderProofHandler),
		bucketSize:           bucketSize,
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
	if proof == nil {
		return
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	pc.insertProofByNonce(proof)

	pc.proofsByHash[string(proof.GetHeaderHash())] = proof
}

func (pc *proofsCache) insertProofByNonce(proof data.HeaderProofHandler) {
	if len(pc.proofsByNonceBuckets) == 0 {
		pc.insertInNewBucket(proof)
		return
	}

	headBucket := pc.proofsByNonceBuckets[0]

	if headBucket.isFull() {
		pc.insertInNewBucket(proof)
		return
	}

	headBucket.insertInExisting(proof)
}

func (pc *proofsCache) insertInNewBucket(proof data.HeaderProofHandler) {
	bucket := newProofBucket(pc.bucketSize)
	bucket.insertInNew(proof)

	pc.proofsByNonceBuckets = append([]*proofNonceBucket{bucket}, pc.proofsByNonceBuckets...)
}

func (pc *proofsCache) cleanupProofsBehindNonce(nonce uint64) {
	if nonce == 0 {
		return
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	buckets := make([]*proofNonceBucket, 0)

	for _, bucket := range pc.proofsByNonceBuckets {
		log.Info("nnn", "nonce", nonce, "maxNonce", bucket.maxNonce)
		if nonce > bucket.maxNonce {
			pc.cleanupProofsInBucket(bucket)
			continue
		}

		buckets = append(buckets, bucket)
	}

	pc.proofsByNonceBuckets = buckets
}

func (pc *proofsCache) cleanupProofsInBucket(bucket *proofNonceBucket) {
	for _, proofInfo := range bucket.proofsByNonce {
		delete(pc.proofsByHash, proofInfo.headerHash)
	}
}
