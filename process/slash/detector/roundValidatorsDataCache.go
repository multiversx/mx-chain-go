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
	mutexCache  sync.RWMutex
	oldestRound uint64
	cacheSize   uint64
}

func newRoundProposerDataCache(maxRounds uint64) *roundValidatorsDataCache {
	return &roundValidatorsDataCache{
		cache:       make(map[uint64]validatorDataMap),
		mutexCache:  sync.RWMutex{},
		oldestRound: math.MaxUint64,
		cacheSize:   maxRounds,
	}
}

func (rpd *roundValidatorsDataCache) add(round uint64, pubKey []byte, data process.InterceptedData) {
	pubKeyStr := string(pubKey)
	rpd.mutexCache.Lock()
	defer rpd.mutexCache.Unlock()

	if rpd.isCacheFull(round) {
		if round < rpd.oldestRound {
			return
		}
		delete(rpd.cache, rpd.oldestRound)
	}
	if round < rpd.oldestRound {
		rpd.oldestRound = round
	}

	validatorsMap, exists := rpd.cache[round]
	if !exists {
		rpd.cache[round] = validatorDataMap{pubKeyStr: dataList{data}}
		return
	}

	validatorsMap[pubKeyStr] = append(validatorsMap[pubKeyStr], data)

}

func (rpd *roundValidatorsDataCache) isCacheFull(currRound uint64) bool {
	_, currRoundInCache := rpd.cache[currRound]
	return len(rpd.cache) >= int(rpd.cacheSize) && !currRoundInCache
}

func (rpd *roundValidatorsDataCache) data(round uint64, pubKey []byte) dataList {
	pubKeyStr := string(pubKey)
	rpd.mutexCache.RLock()
	defer rpd.mutexCache.RUnlock()

	data, exists := rpd.cache[round]
	if !exists {
		return nil
	}

	return data[pubKeyStr]
}

func (rpd *roundValidatorsDataCache) validators(round uint64) [][]byte {
	ret := make([][]byte, 0)
	rpd.mutexCache.RLock()
	defer rpd.mutexCache.RUnlock()

	if _, exists := rpd.cache[round]; exists {
		for pubKey := range rpd.cache[round] {
			ret = append(ret, []byte(pubKey))
		}
	}

	return ret
}

func (rpd *roundValidatorsDataCache) contains(round uint64, pubKey []byte, data process.InterceptedData) bool {
	rpd.mutexCache.RLock()
	defer rpd.mutexCache.RUnlock()

	validatorsMap, exists := rpd.cache[round]
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
