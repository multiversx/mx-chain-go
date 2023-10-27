package sovereign

import (
	"sort"
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
	hashStr := string(hash)

	op.mutex.Lock()
	defer op.mutex.Unlock()

	if _, exists := op.cache[hashStr]; exists {
		return
	}

	op.cache[hashStr] = cacheEntry{
		data:     data,
		expireAt: time.Now().Add(op.timeout),
	}
}

func (op *outGoingOperationsPool) Get(hash []byte) []byte {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	return op.cache[string(hash)].data
}

func (op *outGoingOperationsPool) Delete(hash []byte) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	delete(op.cache, string(hash))
}

func (op *outGoingOperationsPool) GetUnconfirmedOperations() [][]byte {
	expiredEntries := make([]cacheEntry, 0)

	op.mutex.Lock()
	for _, entry := range op.cache {
		if time.Now().After(entry.expireAt) {
			expiredEntries = append(expiredEntries, entry)
		}
	}
	op.mutex.Unlock()

	sort.Slice(expiredEntries, func(i, j int) bool {
		return expiredEntries[i].expireAt.Before(expiredEntries[j].expireAt)
	})

	ret := make([][]byte, len(expiredEntries))
	for i, entry := range expiredEntries {
		ret[i] = entry.data
	}

	return ret
}

func (op *outGoingOperationsPool) IsInterfaceNil() bool {
	return op == nil
}
