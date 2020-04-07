package shardedData

func (sd *shardedData) AddedDataHandlers() []func(key []byte, value interface{}) {
	return sd.addedDataHandlers
}
