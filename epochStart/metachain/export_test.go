package metachain

// SetInCache -
func (sdp *stakingDataProvider) SetInCache(key []byte, ownerData *ownerStats) {
	sdp.mutCache.Lock()
	sdp.cache[string(key)] = ownerData
	sdp.mutCache.Unlock()
}

// GetFromCache -
func (sdp *stakingDataProvider) GetFromCache(key []byte) *ownerStats {
	sdp.mutCache.Lock()
	defer sdp.mutCache.Unlock()

	return sdp.cache[string(key)]
}
