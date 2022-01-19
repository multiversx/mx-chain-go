package detector

import (
	"bytes"
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
)

type hashList [][]byte

type roundHashCache struct {
	cache       map[uint64]hashList
	cacheMutex  sync.RWMutex
	oldestRound uint64
	cacheSize   uint64
}

// NewRoundHashCache creates an instance of roundHashCache, which is a header-hash-based cache
func NewRoundHashCache(maxRounds uint64) *roundHashCache {
	return &roundHashCache{
		cache:       make(map[uint64]hashList),
		cacheMutex:  sync.RWMutex{},
		oldestRound: math.MaxUint64,
		cacheSize:   maxRounds,
	}
}

// Add adds a header-hash in cache, in a given round.
// It has an eviction mechanism which always removes the oldest round entry when cache is full
func (rhc *roundHashCache) Add(round uint64, hash []byte) error {
	rhc.cacheMutex.Lock()
	defer rhc.cacheMutex.Unlock()

	if rhc.contains(round, hash) {
		return process.ErrHeadersNotDifferentHashes
	}

	if rhc.isCacheFull(round) {
		if round < rhc.oldestRound {
			return process.ErrHeaderRoundNotRelevant
		}
		rhc.updateOldestRound()
	}

	if round < rhc.oldestRound {
		rhc.oldestRound = round
	}

	rhc.cache[round] = append(rhc.cache[round], hash)
	return nil
}

func (rhc *roundHashCache) contains(round uint64, hash []byte) bool {
	hashes, exist := rhc.cache[round]
	if !exist {
		return false
	}

	for _, currHash := range hashes {
		if bytes.Equal(currHash, hash) {
			return true
		}
	}

	return false
}

func (rhc *roundHashCache) isCacheFull(currRound uint64) bool {
	_, isCurrRoundInCache := rhc.cache[currRound]
	return len(rhc.cache) >= int(rhc.cacheSize) && !isCurrRoundInCache
}

func (rhc *roundHashCache) updateOldestRound() {
	delete(rhc.cache, rhc.oldestRound)
	min := uint64(math.MaxUint64)

	for round := range rhc.cache {
		if round < min {
			min = round
		}
	}

	rhc.oldestRound = min
}

// Remove removes the given hash in the given round. If there is no match for given data, it does nothing
func (rhc *roundHashCache) Remove(round uint64, hash []byte) {
	rhc.cacheMutex.Lock()
	defer rhc.cacheMutex.Unlock()

	hashes, exists := rhc.cache[round]
	if !exists {
		return
	}

	for idx, currHash := range hashes {
		if bytes.Equal(currHash, hash) {
			if len(hashes) == 1 {
				// If only one hash is cached in the given round, completely delete the entry in cache
				delete(rhc.cache, round)

			} else {
				// Otherwise, delete the hash from the hashes list in the given round
				rhc.cache[round] = rhc.remove(hashes, idx)
			}

			return
		}
	}
}

func (rhc *roundHashCache) remove(hashes hashList, idx int) hashList {
	lastIdx := len(hashes) - 1

	// Indexing here is safe, since this is called only when len(hashes) >= 1
	hashes[idx] = hashes[lastIdx]
	hashes[lastIdx] = nil
	return hashes[:lastIdx]
}

// IsInterfaceNil checks if the underlying pointer is nil
func (rhc *roundHashCache) IsInterfaceNil() bool {
	return rhc == nil
}
