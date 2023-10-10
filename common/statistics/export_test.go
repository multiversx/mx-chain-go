package statistics

// GetSyncStats -
func (ss *stateStatistics) GetSyncStats() (uint64, map[uint32]uint64) {
	return ss.getSyncStats()
}
