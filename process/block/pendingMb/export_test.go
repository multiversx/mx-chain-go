package pendingMb

// SetInMapPendingMbShard -
func (p *pendingMiniBlocks) SetInMapPendingMbShard(hash string, shardID uint32) {
	p.mutPendingMbShard.Lock()
	defer p.mutPendingMbShard.Unlock()

	p.mapPendingMbShard[hash] = shardID
}

// SetInBeforeRevertPendingMbShard -
func (p *pendingMiniBlocks) SetInBeforeRevertPendingMbShard(hash string, shardID uint32) {
	p.mutPendingMbShard.Lock()
	defer p.mutPendingMbShard.Unlock()

	p.beforeRevertPendingMbShard[hash] = shardID
}

// GetMapPendingMbShard -
func (p *pendingMiniBlocks) GetMapPendingMbShard() map[string]uint32 {
	p.mutPendingMbShard.RLock()
	defer p.mutPendingMbShard.RUnlock()

	return copyMap(p.mapPendingMbShard)
}

// GetBeforeRevertPendingMbShard -
func (p *pendingMiniBlocks) GetBeforeRevertPendingMbShard() map[string]uint32 {
	p.mutPendingMbShard.RLock()
	defer p.mutPendingMbShard.RUnlock()

	return copyMap(p.beforeRevertPendingMbShard)
}

func copyMap(src map[string]uint32) map[string]uint32 {
	newMap := make(map[string]uint32)
	for hash, shardID := range src {
		newMap[hash] = shardID
	}

	return newMap
}
