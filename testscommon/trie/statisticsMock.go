package trie

import (
	"github.com/ElrondNetwork/elrond-go/trie/statistics"
)

// MockStatistics -
type MockStatistics struct {
	WaitForSnapshotsToFinishCalled func()
}

// SnapshotFinished -
func (m *MockStatistics) SnapshotFinished() {
}

// NewSnapshotStarted -
func (m *MockStatistics) NewSnapshotStarted() {
}

// WaitForSnapshotsToFinish -
func (m *MockStatistics) WaitForSnapshotsToFinish() {
	if m.WaitForSnapshotsToFinishCalled != nil {
		m.WaitForSnapshotsToFinishCalled()
	}
}

// AddTrieStats -
func (m *MockStatistics) AddTrieStats(_ *statistics.TrieStatsDTO) {
}
