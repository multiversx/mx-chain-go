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
		delete(rhc.cache, rhc.oldestRound)
		rhc.updateOldestRound()
	}

	if round < rhc.oldestRound {
		rhc.oldestRound = round
	}

	rhc.cache[round] = append(rhc.cache[round], hash)
	return nil
}

func (rhc *roundHashCache) contains(round uint64, hash []byte) bool {
	hashHeaderList, exist := rhc.cache[round]
	if !exist {
		return false
	}

	for _, currData := range hashHeaderList {
		if bytes.Equal(currData, hash) {
			return true
		}
	}

	return false
}

func (rhc *roundHashCache) isCacheFull(currRound uint64) bool {
	_, currRoundInCache := rhc.cache[currRound]
	return len(rhc.cache) >= int(rhc.cacheSize) && !currRoundInCache
}

func (rhc *roundHashCache) updateOldestRound() {
	min := uint64(math.MaxUint64)

	for round := range rhc.cache {
		if round < min {
			min = round
		}
	}

	rhc.oldestRound = min
}

// IsInterfaceNil checks if the underlying pointer is nil
func (rhc *roundHashCache) IsInterfaceNil() bool {
	return rhc == nil
}
