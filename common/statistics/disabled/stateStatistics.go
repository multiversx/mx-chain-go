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

// ResetSnapshot does nothing
func (s *stateStatistics) ResetSnapshot() {
}

// IncrementCache does nothing
func (s *stateStatistics) IncrementCache() {
}

// Cache returns zero
func (s *stateStatistics) Cache() uint64 {
	return 0
}

// IncrementSnapshotCache does nothing
func (ss *stateStatistics) IncrementSnapshotCache() {
}

// SnapshotCache returns the number of cached operations
func (ss *stateStatistics) SnapshotCache() uint64 {
	return 0
}

// IncrementPersister does nothing
func (s *stateStatistics) IncrementPersister(epoch uint32) {
}

// Persister returns zero
func (s *stateStatistics) Persister(epoch uint32) uint64 {
	return 0
}

// IncrementSnapshotPersister does nothing
func (ss *stateStatistics) IncrementSnapshotPersister(epoch uint32) {
}

// SnapshotPersister returns the number of persister operations
func (ss *stateStatistics) SnapshotPersister(epoch uint32) uint64 {
	return 0
}

// IncrementTrie does nothing
func (s *stateStatistics) IncrementTrie() {
}

// Trie returns zero
func (s *stateStatistics) Trie() uint64 {
	return 0
}

// ProcessingStats returns nil
func (s *stateStatistics) ProcessingStats() []string {
	return nil
}

// SnapshotStats returns nil
func (s *stateStatistics) SnapshotStats() []string {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *stateStatistics) IsInterfaceNil() bool {
	return s == nil
}
