package timecache

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
)

type entry struct {
	timestamp time.Time
	span      time.Duration
	value     interface{}
}

type timeCacheCore struct {
	*sync.RWMutex
	data        map[string]*entry
	defaultSpan time.Duration
}

func newTimeCacheCore(defaultSpan time.Duration) *timeCacheCore {
	return &timeCacheCore{
		RWMutex:     &sync.RWMutex{},
		data:        make(map[string]*entry),
		defaultSpan: defaultSpan,
	}
}

// upsert will add the key, value and provided duration if not exists
// If the record exists, will update the duration if the provided duration is larger than existing
// Also, it will reset the contained timestamp to time.Now
// It returns if the value existed before this call. It also operates on the locker so the call is concurrent safe
func (tcc *timeCacheCore) upsert(key string, value interface{}, duration time.Duration) (bool, error) {
	if len(key) == 0 {
		return false, storage.ErrEmptyKey
	}

	tcc.Lock()
	defer tcc.Unlock()

	existing, found := tcc.data[key]
	if found {
		if existing.span < duration {
			existing.span = duration
		}
		existing.timestamp = time.Now()

		return found, nil
	}

	tcc.data[key] = &entry{
		timestamp: time.Now(),
		span:      duration,
		value:     value,
	}
	return found, nil
}

// put will add the key, value and provided duration, overriding values if the data already existed
// It returns true if the value existed before this call. It also operates on the locker so the call is concurrent safe
func (tcc *timeCacheCore) put(key string, value interface{}, duration time.Duration) error {
	if len(key) == 0 {
		return storage.ErrEmptyKey
	}

	tcc.Lock()
	defer tcc.Unlock()

	tcc.data[key] = &entry{
		timestamp: time.Now(),
		span:      duration,
		value:     value,
	}
	return nil
}

// hasOrAdd will add the key, value and provided duration, if the key is not found
// It returns true if the value existed before this call and if it has been added or not. It also operates on the locker so the call is concurrent safe
func (tcc *timeCacheCore) hasOrAdd(key string, value interface{}, duration time.Duration) (bool, bool, error) {
	if len(key) == 0 {
		return false, false, storage.ErrEmptyKey
	}

	tcc.Lock()
	defer tcc.Unlock()

	_, found := tcc.data[key]
	if found {
		return true, false, nil
	}

	tcc.data[key] = &entry{
		timestamp: time.Now(),
		span:      duration,
		value:     value,
	}
	return false, true, nil
}

// sweep iterates over all contained elements checking if the element is still valid to be kept
// It also operates on the locker so the call is concurrent safe
func (tcc *timeCacheCore) sweep() {
	tcc.Lock()
	defer tcc.Unlock()

	for key, element := range tcc.data {
		isOldElement := time.Since(element.timestamp) > element.span
		if isOldElement {
			delete(tcc.data, key)
		}
	}
}

// has returns if the key is still found in the time cache
func (tcc *timeCacheCore) has(key string) bool {
	tcc.RLock()
	defer tcc.RUnlock()

	_, ok := tcc.data[key]

	return ok
}

// len returns the number of elements which are still stored in the time cache
func (tcc *timeCacheCore) len() int {
	tcc.RLock()
	defer tcc.RUnlock()

	return len(tcc.data)
}

// clear recreates the map, thus deleting any existing entries
// It also operates on the locker so the call is concurrent safe
func (tcc *timeCacheCore) clear() {
	tcc.Lock()
	tcc.data = make(map[string]*entry)
	tcc.Unlock()
}
