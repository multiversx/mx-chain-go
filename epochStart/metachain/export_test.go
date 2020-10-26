package metachain

// SetInCache -
func (sdr *stakingDataProvider) SetInCache(key []byte, ownerData *ownerStats) {
	sdr.mutCache.Lock()
	sdr.cache[string(key)] = ownerData
	sdr.mutCache.Unlock()
}

// GetFromCache -
func (sdr *stakingDataProvider) GetFromCache(key []byte) *ownerStats {
	sdr.mutCache.Lock()
	defer sdr.mutCache.Unlock()

	return sdr.cache[string(key)]
}
