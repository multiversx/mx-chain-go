package trie

import (
	"github.com/multiversx/mx-chain-go/common"
)

// MockStatistics -
type MockStatistics struct {
	AddTrieStatsCalled             func(common.TrieStatisticsHandler, common.TrieType)
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
func (m *MockStatistics) AddTrieStats(stats common.TrieStatisticsHandler, t common.TrieType) {
	if m.AddTrieStatsCalled != nil {
		m.AddTrieStatsCalled(stats, t)
	}
}

// GetSnapshotDuration -
func (m *MockStatistics) GetSnapshotDuration() int64 {
	return 0
}

// GetSnapshotNumNodes -
func (m *MockStatistics) GetSnapshotNumNodes() uint64 {
	return 0
}

// IncrementThrottlerWaits -
func (m *MockStatistics) IncrementThrottlerWaits() {}

// IsInterfaceNil returns true if there is no value under the interface
func (m *MockStatistics) IsInterfaceNil() bool {
	return m == nil
}
