package timecache

import "time"

func (tc *TimeCache) Keys() []string {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	keys := make([]string, 0, len(tc.data))
	for key := range tc.data {
		keys = append(keys, key)
	}

	return keys
}

func (tc *TimeCache) KeyTime(key string) (time.Time, bool) {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	val, ok := tc.data[key]

	return val.timestamp, ok
}

func (tc *TimeCache) KeySpan(key string) (time.Duration, bool) {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	val, ok := tc.data[key]

	return val.span, ok
}

func (tc *TimeCache) ClearMap() {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	tc.data = make(map[string]span)
}
