package disabled

type stateStatistics struct{}

// NewStateStatistics will create a new disabled statistics component
func NewStateStatistics() *stateStatistics {
	return &stateStatistics{}
}

// Reset does nothing
func (s *stateStatistics) Reset() {
}

// Reset does nothing
func (s *stateStatistics) ResetSync() {
}

// PrintSync returns empty string
func (s *stateStatistics) PrintSync() string {
	return ""
}

// Reset does nothing
func (s *stateStatistics) ResetSnapshot() {
}

// IncrCacheOp does nothing
func (s *stateStatistics) IncrCacheOp() {
}

// CacheOp returns zero
func (s *stateStatistics) CacheOp() uint64 {
	return 0
}

// IncrSyncCacheOp will increment cache counter
func (s *stateStatistics) IncrSyncCacheOp() {
}

// SyncCacheOp returns the number of cached operations
func (ss *stateStatistics) SyncCacheOp() uint64 {
	return 0
}

// IncrSnapshotCacheOp will increment cache counter
func (ss *stateStatistics) IncrSnapshotCacheOp() {
}

// SnapshotCacheOp returns the number of cached operations
func (ss *stateStatistics) SnapshotCacheOp() uint64 {
	return 0
}

// IncrPersisterOp does nothing
func (s *stateStatistics) IncrPersisterOp(epoch uint32) {
}

// PersisterOp returns zero
func (s *stateStatistics) PersisterOp(epoch uint32) uint64 {
	return 0
}

// IncrSyncPersisterOp will increment persister counter
func (ss *stateStatistics) IncrSyncPersisterOp(epoch uint32) {
}

// SyncPersisterOp returns the number of persister operations
func (ss *stateStatistics) SyncPersisterOp(epoch uint32) uint64 {
	return 0
}

// IncrSnapshotPersisterOp will increment persister counter
func (ss *stateStatistics) IncrSnapshotPersisterOp(epoch uint32) {
}

// SyncPersisterOp returns the number of persister operations
func (ss *stateStatistics) SnapshotPersisterOp(epoch uint32) uint64 {
	return 0
}

// IncrTrieOp does nothing
func (s *stateStatistics) IncrTrieOp() {
}

// TrieOp returns zero
func (s *stateStatistics) TrieOp() uint64 {
	return 0
}

// ToString returns empty string
func (s *stateStatistics) ToString() string {
	return ""
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *stateStatistics) IsInterfaceNil() bool {
	return s == nil
}
