package disabled

type stateStatistics struct{}

// NewStateStatistics will create a new disabled statistics component
func NewStateStatistics() *stateStatistics {
	return &stateStatistics{}
}

// ResetAll does nothing
func (s *stateStatistics) ResetAll() {
}

// Reset does nothing
func (s *stateStatistics) Reset() {
}

// ResetSync does nothing
func (s *stateStatistics) ResetSync() {
}

// ResetSnapshot does nothing
func (s *stateStatistics) ResetSnapshot() {
}

// IncrCache does nothing
func (s *stateStatistics) IncrCache() {
}

// Cache returns zero
func (s *stateStatistics) Cache() uint64 {
	return 0
}

// IncrSyncCache does nothing
func (s *stateStatistics) IncrSyncCache() {
}

// SyncCache returns the number of cached operations
func (ss *stateStatistics) SyncCache() uint64 {
	return 0
}

// IncrSnapshotCache does nothing
func (ss *stateStatistics) IncrSnapshotCache() {
}

// SnapshotCache returns the number of cached operations
func (ss *stateStatistics) SnapshotCache() uint64 {
	return 0
}

// IncrPersister does nothing
func (s *stateStatistics) IncrPersister(epoch uint32) {
}

// Persister returns zero
func (s *stateStatistics) Persister(epoch uint32) uint64 {
	return 0
}

// IncrSyncPersister does nothing
func (ss *stateStatistics) IncrSyncPersister(epoch uint32) {
}

// SyncPersister returns the number of persister operations
func (ss *stateStatistics) SyncPersister(epoch uint32) uint64 {
	return 0
}

// IncrSnapshotPersister does nothing
func (ss *stateStatistics) IncrSnapshotPersister(epoch uint32) {
}

// SnapshotPersister returns the number of persister operations
func (ss *stateStatistics) SnapshotPersister(epoch uint32) uint64 {
	return 0
}

// IncrTrie does nothing
func (s *stateStatistics) IncrTrie() {
}

// Trie returns zero
func (s *stateStatistics) Trie() uint64 {
	return 0
}

// ProcessingStats returns empty string
func (s *stateStatistics) ProcessingStats() string {
	return ""
}

// SnapshotStats returns empty string
func (s *stateStatistics) SnapshotStats() string {
	return ""
}

// SyncStats returns empty string
func (s *stateStatistics) SyncStats() string {
	return ""
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *stateStatistics) IsInterfaceNil() bool {
	return s == nil
}
