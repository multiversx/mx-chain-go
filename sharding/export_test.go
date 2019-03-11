package sharding

func (sr *shardRegistry) CalculateMasks() (uint32, uint32) {
	return sr.calculateMasks()
}

func (sr *shardRegistry) Masks() (uint32, uint32) {
	return sr.mask1, sr.mask2
}
