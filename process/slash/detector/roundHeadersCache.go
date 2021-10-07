package detector

import (
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

type headerHashList []headerHash
type headerHash struct {
	hash   string
	header data.HeaderHandler
}

type roundHeadersCache struct {
	cache       map[uint64]headerHashList
	mutexCache  sync.RWMutex
	oldestRound uint64
	cacheSize   uint64
}

func newRoundHeadersCache(maxRounds uint64) *roundHeadersCache {
	return &roundHeadersCache{
		cache:       make(map[uint64]headerHashList),
		mutexCache:  sync.RWMutex{},
		oldestRound: math.MaxUint64,
		cacheSize:   maxRounds,
	}
}

func (rdc *roundHeadersCache) Add(round uint64, hash []byte, header data.HeaderHandler) {
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

	if _, exists := rdc.cache[round]; exists {
		rdc.cache[round] = append(rdc.cache[round],
			headerHash{
				hash:   string(hash),
				header: header,
			},
		)
	} else {
		rdc.cache[round] = headerHashList{
			headerHash{
				hash:   string(hash),
				header: header,
			},
		}
	}
}

func (rdc *roundHeadersCache) Contains(round uint64, hash []byte) bool {
	rdc.mutexCache.RLock()
	defer rdc.mutexCache.RUnlock()

	hashHeaderList, exist := rdc.cache[round]
	if !exist {
		return false
	}

	for _, currData := range hashHeaderList {
		if currData.hash == string(hash) {
			return true
		}
	}

	return false
}

func (rdc *roundHeadersCache) isCacheFull(currRound uint64) bool {
	_, currRoundInCache := rdc.cache[currRound]
	return len(rdc.cache) >= int(rdc.cacheSize) && !currRoundInCache
}
