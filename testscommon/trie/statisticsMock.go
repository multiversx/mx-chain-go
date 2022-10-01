package trie

import "github.com/ElrondNetwork/elrond-go/common"

// MockStatistics -
type MockStatistics struct {
	size                           uint64
	numDataTries                   uint64
	numNodes                       uint64
	WaitForSnapshotsToFinishCalled func()
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

// WaitForSnapshotsToFinish -
func (m *MockStatistics) WaitForSnapshotsToFinish() {
	if m.WaitForSnapshotsToFinishCalled != nil {
		m.WaitForSnapshotsToFinishCalled()
	}
}

// AddTrieStats -
func (m *MockStatistics) AddTrieStats(_ common.TrieStatisticsHandler) {
}
