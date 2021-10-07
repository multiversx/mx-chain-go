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

func NewRoundValidatorDataCache(maxRounds uint64) *roundValidatorsDataCache {
	return &roundValidatorsDataCache{
		cache:       make(map[uint64]validatorDataMap),
		mutexCache:  sync.RWMutex{},
		oldestRound: math.MaxUint64,
		cacheSize:   maxRounds,
	}
}

func (rdc *roundValidatorsDataCache) Add(round uint64, pubKey []byte, data process.InterceptedData) {
	pubKeyStr := string(pubKey)
	rdc.mutexCache.Lock()
	defer rdc.mutexCache.Unlock()

	if rdc.isCacheFull(round) {
		if round < rdc.oldestRound {
			return
		}
		delete(rdc.cache, rdc.oldestRound)
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

func (rdc *roundValidatorsDataCache) GetData(round uint64, pubKey []byte) []process.InterceptedData {
	pubKeyStr := string(pubKey)
	rdc.mutexCache.RLock()
	defer rdc.mutexCache.RUnlock()

	data, exists := rdc.cache[round]
	if !exists {
		return nil
	}

	return data[pubKeyStr]
}

func (rdc *roundValidatorsDataCache) GetPubKeys(round uint64) [][]byte {
	ret := make([][]byte, 0)
	rdc.mutexCache.RLock()
	defer rdc.mutexCache.RUnlock()

	if _, exists := rdc.cache[round]; exists {
		for pubKey := range rdc.cache[round] {
			ret = append(ret, []byte(pubKey))
		}
	}

	return ret
}

func (rdc *roundValidatorsDataCache) Contains(round uint64, pubKey []byte, data process.InterceptedData) bool {
	rdc.mutexCache.RLock()
	defer rdc.mutexCache.RUnlock()

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
func (rdc *roundValidatorsDataCache) IsInterfaceNil() bool {
	return rdc == nil
}
