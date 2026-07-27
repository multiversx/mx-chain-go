package proofscache

import (
	"bytes"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

// maxProofsPerNonce bounds how many different-hash proofs are kept at one nonce (more than one is
// equivocation evidence); on overflow the highest (round, hash) proof is evicted
const maxProofsPerNonce = 4

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

// getProofByNonce returns the canonical proof at the given nonce: the lowest round, with the
// lowest hash as tie-break
func (pc *proofsCache) getProofByNonce(headerNonce uint64) (data.HeaderProofHandler, error) {
	pc.mutProofsCache.RLock()
	defer pc.mutProofsCache.RUnlock()

	proof := pc.lowestProofAtNonce(headerNonce)
	if proof == nil {
		return nil, ErrMissingProof
	}

	return proof, nil
}

// lowestProofAtNonce must be called under mutex protection; allocation-free on purpose, it sits
// on the hot read path
func (pc *proofsCache) lowestProofAtNonce(headerNonce uint64) data.HeaderProofHandler {
	bucket, ok := pc.proofsByNonceBuckets[pc.getBucketKey(headerNonce)]
	if !ok {
		return nil
	}

	var lowest data.HeaderProofHandler
	for _, headerHash := range bucket.hashesAt(headerNonce) {
		proof, hasProof := pc.proofsByHash[headerHash]
		if !hasProof {
			continue
		}
		if lowest == nil || lessProof(proof, lowest) {
			lowest = proof
		}
	}

	return lowest
}

// lessProof reports whether a orders before b by (round, hash) ascending
func lessProof(a, b data.HeaderProofHandler) bool {
	if a.GetHeaderRound() != b.GetHeaderRound() {
		return a.GetHeaderRound() < b.GetHeaderRound()
	}
	return bytes.Compare(a.GetHeaderHash(), b.GetHeaderHash()) < 0
}

// getProofsByNonce returns all proofs at the given nonce, ordered by (round, hash) ascending
func (pc *proofsCache) getProofsByNonce(headerNonce uint64) []data.HeaderProofHandler {
	pc.mutProofsCache.RLock()
	defer pc.mutProofsCache.RUnlock()

	return pc.sortedProofsAtNonce(headerNonce)
}

// sortedProofsAtNonce must be called under mutex protection
func (pc *proofsCache) sortedProofsAtNonce(headerNonce uint64) []data.HeaderProofHandler {
	bucket, ok := pc.proofsByNonceBuckets[pc.getBucketKey(headerNonce)]
	if !ok {
		return nil
	}

	hashes := bucket.hashesAt(headerNonce)
	proofs := make([]data.HeaderProofHandler, 0, len(hashes))
	for _, headerHash := range hashes {
		proof, hasProof := pc.proofsByHash[headerHash]
		if hasProof {
			proofs = append(proofs, proof)
		}
	}

	if len(proofs) > 1 {
		sort.Slice(proofs, func(i, j int) bool {
			return lessProof(proofs[i], proofs[j])
		})
	}

	return proofs
}

// addProof stores the proof, keeping different-hash proofs at the same nonce as equivocation
// evidence; returns the competing proofs only on a newly stored hash, so equivocation reports once
func (pc *proofsCache) addProof(proof data.HeaderProofHandler) []data.HeaderProofHandler {
	if check.IfNil(proof) {
		return nil
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	nonce := proof.GetHeaderNonce()
	newHash := string(proof.GetHeaderHash())
	bucket := pc.getOrCreateBucket(nonce)

	alreadyStored := false
	var competingProofs []data.HeaderProofHandler
	for _, existingHash := range bucket.hashesAt(nonce) {
		if existingHash == newHash {
			alreadyStored = true
			continue
		}

		existingProof, hasProof := pc.proofsByHash[existingHash]
		if hasProof {
			competingProofs = append(competingProofs, existingProof)
		}
	}

	bucket.insert(nonce, newHash)
	pc.proofsByHash[newHash] = proof

	pc.evictExcessProofsAtNonce(bucket, nonce)

	if alreadyStored {
		return nil
	}

	return competingProofs
}

// addProofIfNoneAtNonce adds the proof only if its nonce slot is free; an occupied slot (same or
// different hash) rejects the add and returns the pre-existing canonical proof, never overwriting it
func (pc *proofsCache) addProofIfNoneAtNonce(proof data.HeaderProofHandler) (bool, data.HeaderProofHandler) {
	if check.IfNil(proof) {
		return false, nil
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	nonce := proof.GetHeaderNonce()
	existingProof := pc.lowestProofAtNonce(nonce)
	if existingProof != nil {
		return false, existingProof
	}

	bucket := pc.getOrCreateBucket(nonce)
	bucket.insert(nonce, string(proof.GetHeaderHash()))
	pc.proofsByHash[string(proof.GetHeaderHash())] = proof

	return true, nil
}

// evictExcessProofsAtNonce must be called under mutex protection
func (pc *proofsCache) evictExcessProofsAtNonce(bucket *proofNonceBucket, nonce uint64) {
	if len(bucket.hashesAt(nonce)) <= maxProofsPerNonce {
		return
	}

	proofs := pc.sortedProofsAtNonce(nonce)
	for len(proofs) > maxProofsPerNonce {
		evictedProof := proofs[len(proofs)-1]
		proofs = proofs[:len(proofs)-1]
		evictedHash := string(evictedProof.GetHeaderHash())
		bucket.remove(nonce, evictedHash)
		delete(pc.proofsByHash, evictedHash)

		log.Warn("proofsCache: too many proofs at the same nonce, evicted the highest round one",
			"nonce", nonce,
			"evicted hash", evictedProof.GetHeaderHash(),
			"evicted round", evictedProof.GetHeaderRound(),
		)
	}
}

// getBucketKey will return bucket key as lower bound window value
func (pc *proofsCache) getBucketKey(index uint64) uint64 {
	return (index / pc.bucketSize) * pc.bucketSize
}

// getOrCreateBucket must be called under mutex protection
func (pc *proofsCache) getOrCreateBucket(nonce uint64) *proofNonceBucket {
	bucketKey := pc.getBucketKey(nonce)

	bucket, ok := pc.proofsByNonceBuckets[bucketKey]
	if !ok {
		bucket = newProofBucket()
		pc.proofsByNonceBuckets[bucketKey] = bucket
	}

	return bucket
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
	for _, headerHashes := range bucket.proofsByNonce {
		for _, headerHash := range headerHashes {
			delete(pc.proofsByHash, headerHash)
		}
	}
}
