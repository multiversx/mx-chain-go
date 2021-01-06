package metachain

// SetInCache -
func (sdp *stakingDataProvider) SetInCache(key []byte, ownerData *ownerStats) {
	sdp.mutStakingData.Lock()
	sdp.cache[string(key)] = ownerData
	sdp.mutStakingData.Unlock()
}

// GetFromCache -
func (sdp *stakingDataProvider) GetFromCache(key []byte) *ownerStats {
	sdp.mutStakingData.Lock()
	defer sdp.mutStakingData.Unlock()

	return sdp.cache[string(key)]
}
