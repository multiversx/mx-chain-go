package shardedData

func (sd *ShardedData) AddedDataHandlers() []func(key []byte) {
	return sd.addedDataHandlers
}
