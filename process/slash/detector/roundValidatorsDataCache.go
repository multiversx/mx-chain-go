package detector

import (
	"math"

	"github.com/ElrondNetwork/elrond-go/process"
)

type dataList []process.InterceptedData
type validatorDataMap map[string]dataList

type roundValidatorsDataCache struct {
	cache       map[uint64]validatorDataMap
	oldestRound uint64
	cacheSize   uint64
}

func newRoundProposerDataCache(maxRounds uint64) *roundValidatorsDataCache {
	return &roundValidatorsDataCache{
		cache:       make(map[uint64]validatorDataMap),
		oldestRound: math.MaxUint64,
		cacheSize:   maxRounds,
	}
}

func (rpd *roundValidatorsDataCache) add(round uint64, pubKey []byte, data process.InterceptedData) {
	pubKeyStr := string(pubKey)

	if rpd.isCacheFull(round) {
		if round < rpd.oldestRound {
			return
		}
		delete(rpd.cache, rpd.oldestRound)
	}
	if round < rpd.oldestRound {
		rpd.oldestRound = round
	}

	if _, exists := rpd.cache[round]; exists {
		if _, exists = rpd.cache[round][pubKeyStr]; exists {
			rpd.cache[round][pubKeyStr] = append(rpd.cache[round][pubKeyStr], data)
		} else {
			rpd.cache[round][pubKeyStr] = dataList{data}
		}
	} else {
		list := dataList{data}
		proposerMap := validatorDataMap{pubKeyStr: list}

		rpd.cache[round] = proposerMap
	}
}

func (rpd *roundValidatorsDataCache) isCacheFull(currRound uint64) bool {
	_, currRoundInCache := rpd.cache[currRound]
	return len(rpd.cache) >= int(rpd.cacheSize) && !currRoundInCache
}

func (rpd *roundValidatorsDataCache) data(round uint64, pubKey []byte) dataList {
	pubKeyStr := string(pubKey)

	if _, exists := rpd.cache[round]; exists {
		if _, exists = rpd.cache[round][pubKeyStr]; exists {
			return rpd.cache[round][pubKeyStr]
		}
	}

	return nil
}

func (rpd *roundValidatorsDataCache) validators(round uint64) [][]byte {
	ret := make([][]byte, 0)

	if _, exists := rpd.cache[round]; exists {
		for pubKey, _ := range rpd.cache[round] {
			ret = append(ret, []byte(pubKey))
		}
	}

	return ret
}
