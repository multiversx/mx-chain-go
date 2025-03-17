package proofscache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
)

type proofNonceMapping struct {
	headerHash string
	nonce      uint64
}

type proofsCache struct {
	mutProofsByNonce     sync.RWMutex
	proofsByNonceBuckets []*proofNonceBucket
	bucketSize           int

	proofsByHash    map[string]data.HeaderProofHandler
	mutProofsByHash sync.RWMutex
}

func newProofsCache(bucketSize int) *proofsCache {

	return &proofsCache{
		proofsByNonceBuckets: make([]*proofNonceBucket, 0),
		bucketSize:           bucketSize,
		proofsByHash:         make(map[string]data.HeaderProofHandler),
	}
}

func (pc *proofsCache) getProofByHash(headerHash []byte) (data.HeaderProofHandler, error) {
	pc.mutProofsByHash.RLock()
	defer pc.mutProofsByHash.RUnlock()

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

	pc.insertProofByNonce(proof)

	pc.mutProofsByHash.Lock()
	pc.proofsByHash[string(proof.GetHeaderHash())] = proof
	pc.mutProofsByHash.Unlock()
}

func (pc *proofsCache) insertProofByNonce(proof data.HeaderProofHandler) {
	pc.mutProofsByNonce.Lock()
	defer pc.mutProofsByNonce.Unlock()

	if len(pc.proofsByNonceBuckets) == 0 {
		pc.insertInNewBucket(proof)
		return
	}

	headBucket := pc.proofsByNonceBuckets[0]

	if headBucket.isFull() {
		pc.insertInNewBucket(proof)
		return
	}

	headBucket.insert(proof)
}

func (pc *proofsCache) insertInNewBucket(proof data.HeaderProofHandler) {
	bucket := newProofBucket(pc.bucketSize)
	bucket.insert(proof)

	pc.proofsByNonceBuckets = append([]*proofNonceBucket{bucket}, pc.proofsByNonceBuckets...)
}

func (pc *proofsCache) cleanupProofsBehindNonce(nonce uint64) {
	if nonce == 0 {
		return
	}

	pc.mutProofsByNonce.Lock()
	defer pc.mutProofsByNonce.Unlock()

	buckets := make([]*proofNonceBucket, 0)

	wg := &sync.WaitGroup{}

	for _, bucket := range pc.proofsByNonceBuckets {
		if nonce > bucket.maxNonce {
			wg.Add(1)

			go func(bucket *proofNonceBucket) {
				pc.cleanupProofsInBucket(bucket)
				wg.Done()
			}(bucket)

			continue
		}

		buckets = append(buckets, bucket)
	}

	wg.Wait()

	pc.proofsByNonceBuckets = buckets
}

func (pc *proofsCache) cleanupProofsInBucket(bucket *proofNonceBucket) {
	pc.mutProofsByHash.Lock()
	defer pc.mutProofsByHash.Unlock()

	for _, proofInfo := range bucket.proofsByNonce {
		delete(pc.proofsByHash, proofInfo.headerHash)
	}
}
