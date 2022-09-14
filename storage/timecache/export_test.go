package timecache

// Keys -
func (tc *TimeCache) Keys() []string {
	tc.timeCache.Lock()
	defer tc.timeCache.Unlock()

	keys := make([]string, 0, len(tc.timeCache.data))
	for key := range tc.timeCache.data {
		keys = append(keys, key)
	}

	return keys
}

// Value -
func (tc *TimeCache) Value(key string) (*entry, bool) {
	tc.timeCache.Lock()
	defer tc.timeCache.Unlock()

	val, ok := tc.timeCache.data[key]

	return val, ok
}

// NumRegisteredHandlers -
func (tc *timeCacher) NumRegisteredHandlers() int {
	tc.mutAddedDataHandlers.RLock()
	defer tc.mutAddedDataHandlers.RUnlock()

	return len(tc.mapDataHandlers)
}
