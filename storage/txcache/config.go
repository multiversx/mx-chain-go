package txcache

// CacheConfig holds cache configuration
type CacheConfig struct {
	Name                       string
	NumChunksHint              uint32
	EvictionEnabled            bool
	NumBytesThreshold          uint32
	NumBytesPerSenderThreshold uint32
	CountThreshold             uint32
	CountPerSenderThreshold    uint32
	NumSendersToEvictInOneStep uint32
	MinGasPriceMicroErd        uint32
}
