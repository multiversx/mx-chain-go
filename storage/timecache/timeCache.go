package timecache

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ dataRetriever.RequestedItemsHandler = (*TimeCache)(nil)
var _ p2p.BlacklistHandler = (*TimeCache)(nil)

type span struct {
	timestamp time.Time
	span      time.Duration
}

// TimeCache can retain an amount of string keys for a defined period of time
// sweeping (clean-up) is triggered each time a new item is added or a key is present in the time cache
// This data structure is concurrent safe.
type TimeCache struct {
	mut         sync.Mutex
	data        map[string]span
	keys        []string
	defaultSpan time.Duration
}

// NewTimeCache creates a new time cache data structure instance
func NewTimeCache(defaultSpan time.Duration) *TimeCache {
	return &TimeCache{
		data:        make(map[string]span),
		keys:        make([]string, 0),
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

	_, ok := tc.data[key]
	if ok {
		return storage.ErrDuplicateKeyToAdd
	}

	tc.data[key] = span{
		timestamp: time.Now(),
		span:      duration,
	}
	tc.keys = append(tc.keys, key)
	return nil
}

// AddWithSpan will store the key in the time cache with the provided span duration
// Double adding the key is not permitted by the time cache. Also, add will trigger sweeping.
func (tc *TimeCache) AddWithSpan(key string, duration time.Duration) error {
	return tc.add(key, duration)
}

// Sweep starts from the oldest element and will search each element if it is still valid to be kept. Sweep ends when
// it finds an element that is still valid
func (tc *TimeCache) Sweep() {
	tc.mut.Lock()
	defer tc.mut.Unlock()

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

		isOldElement := time.Since(t.timestamp) > t.span
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

	_, ok := tc.data[key]

	return ok
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *TimeCache) IsInterfaceNil() bool {
	return tc == nil
}
