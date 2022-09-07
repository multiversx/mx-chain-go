package timecache

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ dataRetriever.RequestedItemsHandler = (*TimeCache)(nil)

// TimeCache can retain an amount of string keys for a defined period of time
// sweeping (clean-up) is triggered each time a new item is added or a key is present in the time cache
// This data structure is concurrent safe.
type TimeCache struct {
	timeCache *timeCacheCore
}

// NewTimeCache creates a new time cache data structure instance
func NewTimeCache(defaultSpan time.Duration) *TimeCache {
	return &TimeCache{
		timeCache: newTimeCacheCore(defaultSpan),
	}
}

// Add will store the key in the time cache
// Double adding the key is permitted. It will replace the data, if existing. It does not trigger sweep.
func (tc *TimeCache) Add(key string) error {
	return tc.add(key, tc.timeCache.defaultSpan)
}

func (tc *TimeCache) add(key string, duration time.Duration) error {
	if len(key) == 0 {
		return storage.ErrEmptyKey
	}

	tc.timeCache.Lock()
	defer tc.timeCache.Unlock()

	tc.timeCache.data[key] = &entry{
		timestamp: time.Now(),
		span:      duration,
	}
	return nil
}

// AddWithSpan will store the key in the time cache with the provided span duration
// Double adding the key is permitted. It will replace the data, if existing. It does not trigger sweep.
func (tc *TimeCache) AddWithSpan(key string, duration time.Duration) error {
	return tc.add(key, duration)
}

// Upsert will add the key and provided duration if not exists
// If the record exists, will update the duration if the provided duration is larger than existing
// Also, it will reset the contained timestamp to time.Now
func (tc *TimeCache) Upsert(key string, duration time.Duration) error {
	_, err := tc.timeCache.upsert(key, nil, duration)

	return err
}

// Sweep starts from the oldest element and will search each element if it is still valid to be kept. Sweep ends when
// it finds an element that is still valid
func (tc *TimeCache) Sweep() {
	tc.timeCache.sweep()
}

// Has returns if the key is still found in the time cache
func (tc *TimeCache) Has(key string) bool {
	return tc.timeCache.has(key)
}

// Len returns the number of elements which are still stored in the time cache
func (tc *TimeCache) Len() int {
	return tc.timeCache.len()
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *TimeCache) IsInterfaceNil() bool {
	return tc == nil
}
