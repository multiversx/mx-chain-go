package detector

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

type dataList []process.InterceptedData
type proposerDataMap map[string]dataList

type roundProposerDataCache struct {
	cache     map[uint64]proposerDataMap
	rounds    []uint64
	cacheSize uint64
}

func newRoundProposerDataCache(maxRounds uint64) *roundProposerDataCache {
	return &roundProposerDataCache{
		cache:     make(map[uint64]proposerDataMap),
		rounds:    make([]uint64, 0, maxRounds),
		cacheSize: maxRounds,
	}
}

func (rpd *roundProposerDataCache) add(round uint64, pubKey []byte, data process.InterceptedData) {
	pubKeyStr := string(pubKey)

	if rpd.isCacheFull(round) {
		rpd.removeOldestEntryFromCache()
	}

	if roundData, exists := rpd.cache[round]; exists {
		if _, proposerDataExists := roundData[pubKeyStr]; proposerDataExists {
			rpd.cache[round][pubKeyStr] = append(rpd.cache[round][pubKeyStr], data)
		} else {
			rpd.cache[round][pubKeyStr] = dataList{data}
		}
	} else {
		header := dataList{data}
		proposerMap := proposerDataMap{pubKeyStr: header}

		rpd.cache[round] = proposerMap
	}

	rpd.addRound(round)
}

func (rpd *roundProposerDataCache) addRound(round uint64) {
	if !contains(rpd.rounds, round) {
		rpd.rounds = append(rpd.rounds, round)
	}
}

func contains(slice []uint64, elem uint64) bool {
	for _, val := range slice {
		if val == elem {
			return true
		}
	}
	return false
}

func (rpd *roundProposerDataCache) proposedData(round uint64, pubKey []byte) dataList {
	pubKeyStr := string(pubKey)

	if _, exists := rpd.cache[round]; exists {
		if _, exists = rpd.cache[round][pubKeyStr]; exists {
			return rpd.cache[round][pubKeyStr]
		}
	}

	return nil
}

func (rpd *roundProposerDataCache) removeOldestEntryFromCache() {
	first := rpd.rounds[0]
	rpd.rounds = rpd.rounds[1:]
	delete(rpd.cache, first)
}

func (rpd *roundProposerDataCache) isCacheFull(currRound uint64) bool {
	_, currRoundInCache := rpd.cache[currRound]
	return len(rpd.cache) >= int(rpd.cacheSize) && !currRoundInCache
}
