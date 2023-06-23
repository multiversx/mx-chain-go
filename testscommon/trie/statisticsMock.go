package trie

import (
	"github.com/multiversx/mx-chain-go/common"
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
func (m *MockStatistics) AddTrieStats(_ common.TrieStatisticsHandler, _ common.TrieType) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *MockStatistics) IsInterfaceNil() bool {
	return m == nil
}
