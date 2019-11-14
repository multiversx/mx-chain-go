package timecache

import "time"

func (tc *TimeCache) Keys() []string {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	return tc.keys
}

func (tc *TimeCache) KeyTime(key string) (time.Time, bool) {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	val, ok := tc.data[key]

	return val, ok
}

func (tc *TimeCache) ClearMap() {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	tc.data = make(map[string]time.Time)
}
