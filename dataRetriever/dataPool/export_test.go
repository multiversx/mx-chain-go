package dataPool

func (nspc *nonceSyncMapCacher) GetAddedHandlers() []func(nonce uint64, shardId uint32, value []byte) {
	nspc.mutAddedDataHandlers.RLock()
	defer nspc.mutAddedDataHandlers.RUnlock()

	handlers := make([]func(nonce uint64, shardId uint32, value []byte), len(nspc.addedDataHandlers))
	copy(handlers, nspc.addedDataHandlers)

	return handlers
}
