package shardedData

func (sd *shardedData) AddedDataHandlers() []func(key []byte) {
	return sd.addedDataHandlers
}
