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

func (rhc *roundHeadersCache) Add(round uint64, hash []byte, header data.HeaderHandler) {
	rhc.mutexCache.Lock()
	defer rhc.mutexCache.Unlock()

	if rhc.isCacheFull(round) {
		if round < rhc.oldestRound {
			return
		}
		delete(rhc.cache, rhc.oldestRound)
	}
	if round < rhc.oldestRound {
		rhc.oldestRound = round
	}

	if _, exists := rhc.cache[round]; exists {
		rhc.cache[round] = append(rhc.cache[round],
			headerHash{
				hash:   string(hash),
				header: header,
			},
		)
	} else {
		rhc.cache[round] = headerHashList{
			headerHash{
				hash:   string(hash),
				header: header,
			},
		}
	}
}

func (rhc *roundHeadersCache) Contains(round uint64, hash []byte) bool {
	rhc.mutexCache.RLock()
	defer rhc.mutexCache.RUnlock()

	hashHeaderList, exist := rhc.cache[round]
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

func (rhc *roundHeadersCache) isCacheFull(currRound uint64) bool {
	_, currRoundInCache := rhc.cache[currRound]
	return len(rhc.cache) >= int(rhc.cacheSize) && !currRoundInCache
}

func (rhc *roundHeadersCache) IsInterfaceNil() bool {
	return rhc == nil
}
