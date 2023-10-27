package sovereign

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data     []byte
	expireAt time.Time
}

type outGoingOperationsPool struct {
	mutex   sync.RWMutex
	timeout time.Duration
	cache   map[string]cacheEntry
}

func NewOutGoingOperationPool(expiryTime time.Duration) *outGoingOperationsPool {
	return &outGoingOperationsPool{
		timeout: expiryTime,
		cache:   map[string]cacheEntry{},
	}
}

func (op *outGoingOperationsPool) Add(hash []byte, data []byte) {
	if _, exists := op.cache[string(hash)]; exists {
		return
	}

	op.cache[string(hash)] = cacheEntry{
		data:     data,
		expireAt: time.Now().Add(op.timeout),
	}
}

func (op *outGoingOperationsPool) Delete(hash []byte) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	delete(op.cache, string(hash))
}

func (op *outGoingOperationsPool) Get(hash []byte) []byte {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	return op.cache[string(hash)].data
}

func (op *outGoingOperationsPool) GetUnconfirmedOperations() [][]byte {
	ret := make([][]byte, 0)

	op.mutex.Lock()
	for _, entry := range op.cache {
		if time.Now().After(entry.expireAt) {
			ret = append(ret, entry.data)
		}
	}
	op.mutex.Unlock()

	return ret
}
