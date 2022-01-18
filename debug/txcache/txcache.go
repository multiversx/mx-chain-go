package txcache

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const evictionTime = time.Minute * 10

var TXCACHE = NewTxCache()

type txCache struct {
	mut  sync.Mutex
	data map[string]time.Time
}

func NewTxCache() *txCache {
	return &txCache{
		data: make(map[string]time.Time),
	}
}

func (cache *txCache) AddNewTxHash(txHash []byte) {
	cache.mut.Lock()
	defer cache.mut.Unlock()

	t, found := cache.data[string(txHash)]
	if !found {
		cache.data[string(txHash)] = time.Now()
		return
	}

	if time.Since(t) < evictionTime {
		panic(fmt.Sprintf("eviction before time! txHash %s, time %v", hex.EncodeToString(txHash), time.Since(t)))
	}

	cache.data[string(txHash)] = time.Now()
}
