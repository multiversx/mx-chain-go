package txcache

// CacheConfig holds cache configuration
type CacheConfig struct {
	Name                            string
	NumChunksHint                   uint32
	EvictionEnabled                 bool
	NumBytesThreshold               uint32
	CountThreshold                  uint32
	NumSendersToEvictInOneStep      uint32
	ALotOfTransactionsForASender    uint32
	NumTxsToEvictForASenderWithALot uint32
	MinGasPriceMicroErd             uint32
}
