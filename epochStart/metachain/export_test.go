package metachain

// SetInCache -
func (rsp *rewardsStakingProvider) SetInCache(key []byte, ownerData *ownerStats) {
	rsp.mutCache.Lock()
	rsp.cache[string(key)] = ownerData
	rsp.mutCache.Unlock()
}

// GetFromCache -
func (rsp *rewardsStakingProvider) GetFromCache(key []byte) *ownerStats {
	rsp.mutCache.Lock()
	defer rsp.mutCache.Unlock()

	return rsp.cache[string(key)]
}
