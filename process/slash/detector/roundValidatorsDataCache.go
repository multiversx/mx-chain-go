package detector

import (
	"bytes"
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
)

type dataList []process.InterceptedData
type validatorDataMap map[string]dataList

type roundValidatorsDataCache struct {
	cache       map[uint64]validatorDataMap
	cacheMutex  sync.RWMutex
	oldestRound uint64
	cacheSize   uint64
}

// NewRoundValidatorDataCache creates a new instance of roundValidatorsDataCache, which
// is a round-based(per validator data) cache
func NewRoundValidatorDataCache(maxRounds uint64) *roundValidatorsDataCache {
	return &roundValidatorsDataCache{
		cache:       make(map[uint64]validatorDataMap),
		cacheMutex:  sync.RWMutex{},
		oldestRound: math.MaxUint64,
		cacheSize:   maxRounds,
	}
}

// Add adds in cache an intercepted data for a public key, in a given round.
// It has an eviction mechanism which always removes the oldest round entry when cache is full
func (rdc *roundValidatorsDataCache) Add(round uint64, pubKey []byte, data process.InterceptedData) {
	pubKeyStr := string(pubKey)
	rdc.cacheMutex.Lock()
	defer rdc.cacheMutex.Unlock()

	if rdc.isCacheFull(round) {
		if round < rdc.oldestRound {
			return
		}
		delete(rdc.cache, rdc.oldestRound)
		rdc.updateOldestRound()
	}
	if round < rdc.oldestRound {
		rdc.oldestRound = round
	}

	validatorsMap, exists := rdc.cache[round]
	if !exists {
		rdc.cache[round] = validatorDataMap{pubKeyStr: dataList{data}}
		return
	}

	validatorsMap[pubKeyStr] = append(validatorsMap[pubKeyStr], data)

}

func (rdc *roundValidatorsDataCache) isCacheFull(currRound uint64) bool {
	_, currRoundInCache := rdc.cache[currRound]
	return len(rdc.cache) >= int(rdc.cacheSize) && !currRoundInCache
}

func (rdc *roundValidatorsDataCache) updateOldestRound() {
	min := uint64(math.MaxUint64)

	for round := range rdc.cache {
		if round < min {
			min = round
		}
	}

	rdc.oldestRound = min
}

// GetData returns all cached data for a public key, in a given round
func (rdc *roundValidatorsDataCache) GetData(round uint64, pubKey []byte) []process.InterceptedData {
	pubKeyStr := string(pubKey)
	rdc.cacheMutex.RLock()
	defer rdc.cacheMutex.RUnlock()

	data, exists := rdc.cache[round]
	if !exists {
		return nil
	}

	return data[pubKeyStr]
}

// GetPubKeys returns all cached public keys in a given round
func (rdc *roundValidatorsDataCache) GetPubKeys(round uint64) [][]byte {
	ret := make([][]byte, 0)
	rdc.cacheMutex.RLock()
	defer rdc.cacheMutex.RUnlock()

	if _, exists := rdc.cache[round]; exists {
		for pubKey := range rdc.cache[round] {
			ret = append(ret, []byte(pubKey))
		}
	}

	return ret
}

// Contains checks if a public key, in a given round has any data cached.
func (rdc *roundValidatorsDataCache) Contains(round uint64, pubKey []byte, data process.InterceptedData) bool {
	rdc.cacheMutex.RLock()
	defer rdc.cacheMutex.RUnlock()

	validatorsMap, exists := rdc.cache[round]
	if !exists {
		return false
	}

	dataList, exists := validatorsMap[string(pubKey)]
	if !exists {
		return false
	}

	for _, currData := range dataList {
		if bytes.Equal(currData.Hash(), data.Hash()) {
			return true
		}
	}

	return false
}

// IsInterfaceNil checks if the underlying pointer is nil
func (rdc *roundValidatorsDataCache) IsInterfaceNil() bool {
	return rdc == nil
}
