package headersCashe

import "sync"

type headersCounter struct {
	hdrsCounter    map[uint32]uint64
	mutHdrsCounter sync.RWMutex
}

func newHeadersCounter() *headersCounter {
	return &headersCounter{
		hdrsCounter:    make(map[uint32]uint64),
		mutHdrsCounter: sync.RWMutex{},
	}
}

func (hdc *headersCounter) increment(shardId uint32) {
	hdc.mutHdrsCounter.Lock()
	defer hdc.mutHdrsCounter.Unlock()

	if _, ok := hdc.hdrsCounter[shardId]; !ok {
		hdc.hdrsCounter[shardId] = 1

		return
	}

	hdc.hdrsCounter[shardId]++
}

func (hdc *headersCounter) decrement(shardId uint32) {
	hdc.mutHdrsCounter.Lock()
	defer hdc.mutHdrsCounter.Unlock()

	if _, ok := hdc.hdrsCounter[shardId]; !ok {
		return
	}

	hdc.hdrsCounter[shardId]--
}

func (hdc *headersCounter) getNumHeaderFromCache(shardId uint32) int64 {
	hdc.mutHdrsCounter.RLock()
	defer hdc.mutHdrsCounter.RUnlock()

	numShardHeaders, ok := hdc.hdrsCounter[shardId]
	if !ok {
		return 0
	}

	return int64(numShardHeaders)
}
