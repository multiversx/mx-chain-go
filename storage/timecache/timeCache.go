package timecache

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ dataRetriever.RequestedItemsHandler = (*TimeCache)(nil)

type span struct {
	timestamp time.Time
	span      time.Duration
}

// TimeCache can retain an amount of string keys for a defined period of time
// sweeping (clean-up) is triggered each time a new item is added or a key is present in the time cache
// This data structure is concurrent safe.
type TimeCache struct {
	mut         sync.RWMutex
	data        map[string]*span
	defaultSpan time.Duration
}

// NewTimeCache creates a new time cache data structure instance
func NewTimeCache(defaultSpan time.Duration) *TimeCache {
	return &TimeCache{
		data:        make(map[string]*span),
		defaultSpan: defaultSpan,
	}
}

// Add will store the key in the time cache
// Double adding the key is not permitted by the time cache. Also, add will trigger sweeping.
func (tc *TimeCache) Add(key string) error {
	return tc.add(key, tc.defaultSpan)
}

func (tc *TimeCache) add(key string, duration time.Duration) error {
	if len(key) == 0 {
		return storage.ErrEmptyKey
	}

	tc.mut.Lock()
	defer tc.mut.Unlock()

	tc.data[key] = &span{
		timestamp: time.Now(),
		span:      duration,
	}
	return nil
}

// AddWithSpan will store the key in the time cache with the provided span duration
// Double adding the key is not permitted by the time cache. Also, add will trigger sweeping.
func (tc *TimeCache) AddWithSpan(key string, duration time.Duration) error {
	return tc.add(key, duration)
}

// Upsert will add the key and provided duration if not exists
// If the record exists, will update the duration if the provided duration is larger than existing
// Also, it will reset the contained timestamp to time.Now
func (tc *TimeCache) Upsert(key string, duration time.Duration) error {
	if len(key) == 0 {
		return storage.ErrEmptyKey
	}

	tc.mut.Lock()
	defer tc.mut.Unlock()

	existing, found := tc.data[key]
	if found {
		if existing.span < duration {
			existing.span = duration
		}
		existing.timestamp = time.Now()

		return nil
	}

	tc.data[key] = &span{
		timestamp: time.Now(),
		span:      duration,
	}
	return nil
}

// Sweep starts from the oldest element and will search each element if it is still valid to be kept. Sweep ends when
// it finds an element that is still valid
func (tc *TimeCache) Sweep() {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	for key, element := range tc.data {
		isOldElement := time.Since(element.timestamp) > element.span
		if isOldElement {
			delete(tc.data, key)
		}
	}
}

// Has returns if the key is still found in the time cache
func (tc *TimeCache) Has(key string) bool {
	tc.mut.RLock()
	defer tc.mut.RUnlock()

	_, ok := tc.data[key]

	return ok
}

// Len returns the number of elements which are still stored in the time cache
func (tc *TimeCache) Len() int {
	tc.mut.RLock()
	defer tc.mut.RUnlock()

	return len(tc.data)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *TimeCache) IsInterfaceNil() bool {
	return tc == nil
}
