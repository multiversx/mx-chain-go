package detector

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

type headerList []process.InterceptedData
type proposerHeadersMap map[string]headerList

type roundProposerDataCache struct {
	cache     map[uint64]proposerHeadersMap
	cacheSize uint64
}

func newRoundProposerDataCache(maxRounds uint64) *roundProposerDataCache {
	return &roundProposerDataCache{
		cache:     make(map[uint64]proposerHeadersMap),
		cacheSize: maxRounds,
	}
}

func (rpd *roundProposerDataCache) add(round uint64, pubKey []byte, data process.InterceptedData) {
	pubKeyStr := string(pubKey)

	if rpd.isCacheFull(round) {
		rpd.removeFirstKey()
	}

	if roundData, exists := rpd.cache[round]; exists {
		if _, proposerDataExists := roundData[pubKeyStr]; proposerDataExists {
			rpd.cache[round][pubKeyStr] = append(rpd.cache[round][pubKeyStr], data)
		} else {
			rpd.cache[round][pubKeyStr] = headerList{data}
		}
	} else {
		header := headerList{data}
		proposerMap := proposerHeadersMap{pubKeyStr: header}

		rpd.cache[round] = proposerMap
	}
}

func (rpd *roundProposerDataCache) getProposedHeaders(round uint64, pubKey []byte) headerList {
	pubKeyStr := string(pubKey)

	if _, exists := rpd.cache[round]; exists {
		if _, exists = rpd.cache[round][pubKeyStr]; exists {
			return rpd.cache[round][pubKeyStr]
		}
	}

	return nil
}

func (rpd *roundProposerDataCache) removeFirstKey() {
	var min uint64
	for min = range rpd.cache {
		break
	}

	for n := range rpd.cache {
		if n < min {
			min = n
		}
	}

	delete(rpd.cache, min)
}

func (rpd *roundProposerDataCache) isCacheFull(currRound uint64) bool {
	_, currRoundInCache := rpd.cache[currRound]
	return len(rpd.cache) >= int(rpd.cacheSize) && !currRoundInCache
}
