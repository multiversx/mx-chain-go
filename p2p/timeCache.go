package p2p

//struct adapted from github.com/whyrusleeping/timecache/timecache.go by the Elrond Team

import (
	"container/list"
	"time"
)

// TimeCache is used for keeping a key for a specified amount of time
type TimeCache struct {
	keys *list.List
	data map[string]time.Time
	span time.Duration
}

// NewTimeCache is used to create a new instance of TimeCache object
func NewTimeCache(span time.Duration) *TimeCache {
	return &TimeCache{
		keys: list.New(),
		data: make(map[string]time.Time),
		span: span,
	}
}

// Add a new key
func (tc *TimeCache) Add(key string) {
	_, ok := tc.data[key]
	if ok {
		return
	}

	tc.sweep()

	tc.data[key] = time.Now()
	tc.keys.PushFront(key)
}

// clean-up func
func (tc *TimeCache) sweep() {
	for {
		back := tc.keys.Back()
		if back == nil {
			return
		}

		v := back.Value.(string)
		t, ok := tc.data[v]
		if !ok {
			panic("inconsistent cache state")
		}

		if time.Since(t) > tc.span {
			tc.keys.Remove(back)
			delete(tc.data, v)
		} else {
			return
		}
	}
}

// Has returns the existence of a key making a sweep cleanup before testing
func (tc *TimeCache) Has(key string) bool {
	tc.sweep()
	_, ok := tc.data[key]
	return ok
}
