package trie

// MockStatistics -
type MockStatistics struct {
	size         uint64
	numDataTries uint64
	numNodes     uint64
}

// AddSize -
func (m *MockStatistics) AddSize(size uint64) {
	m.size += size
	m.numNodes++
}

// SnapshotFinished -
func (m *MockStatistics) SnapshotFinished() {
}

// NewSnapshotStarted -
func (m *MockStatistics) NewSnapshotStarted() {
}

// NewDataTrie -
func (m *MockStatistics) NewDataTrie() {
	m.numDataTries++
}
