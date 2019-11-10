package timecache

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
)

// TimeCache can retain an amount of string keys for a defined period of time
// sweeping (clean-up) is triggered each time a new item is added or a key is present in the time cache
// This data structure is concurrent safe.
type TimeCache struct {
	mut  sync.Mutex
	data map[string]time.Time
	keys []string
	span time.Duration
}

// NewTimeCache creates a new time cache data structure instance
func NewTimeCache(span time.Duration) *TimeCache {
	return &TimeCache{
		data: make(map[string]time.Time),
		keys: make([]string, 0),
		span: span,
	}
}

// Add will store the key in the time cache
// Double adding the key is not permitted by the time cache. Also, add will trigger sweeping.
func (tc *TimeCache) Add(key string) error {
	if len(key) == 0 {
		return storage.ErrEmptyKey
	}

	tc.mut.Lock()
	defer tc.mut.Unlock()

	tc.sweep()

	_, ok := tc.data[key]
	if ok {
		return storage.ErrDuplicateKeyToAdd
	}

	tc.data[key] = time.Now()
	tc.keys = append(tc.keys, key)
	return nil
}

// sweep starts from the oldest element and will search each element if it is still valid to be kept.
// sweep ends when it finds an element that is still valid
func (tc *TimeCache) sweep() {
	for {
		if len(tc.keys) == 0 {
			return
		}

		firstElement := tc.keys[0]
		t, ok := tc.data[firstElement]
		if !ok {
			//inconsistency handled
			tc.keys = tc.keys[1:]
			continue
		}

		isOldElement := time.Since(t) > tc.span
		if isOldElement {
			tc.keys = tc.keys[1:]
			delete(tc.data, firstElement)
		} else {
			return
		}
	}
}

// Has returns if the key is still found in the time cache
func (tc *TimeCache) Has(key string) bool {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	tc.sweep()
	_, ok := tc.data[key]

	return ok
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *TimeCache) IsInterfaceNil() bool {
	return tc == nil
}
