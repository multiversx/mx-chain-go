package shardchain

// SaveEpochStartInfoToStaticStorer -
func (t *trigger) SaveEpochStartInfoToStaticStorer() error {
	t.mutTrigger.Lock()
	defer t.mutTrigger.Unlock()

	return t.saveEpochStartInfoToStaticStorer()
}
