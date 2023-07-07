package disabled

type stateStatistics struct{}

// NewStateStatistics will create a new disabled statistics component
func NewStateStatistics() *stateStatistics {
	return &stateStatistics{}
}

// Reset does nothing
func (s *stateStatistics) Reset() {
}

// IncrCacheOp does nothing
func (s *stateStatistics) IncrCacheOp() {
}

// CacheOp returns zero
func (s *stateStatistics) CacheOp() uint64 {
	return 0
}

// IncrPersisterOp does nothing
func (s *stateStatistics) IncrPersisterOp() {
}

// PersisterOp returns zero
func (s *stateStatistics) PersisterOp() uint64 {
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
