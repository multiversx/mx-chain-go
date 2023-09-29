package state

import (
	"github.com/multiversx/mx-chain-go/common"
)

// StateMetricsStub -
type StateMetricsStub struct {
	UpdateMetricsOnSnapshotStartCalled      func()
	UpdateMetricsOnSnapshotCompletionCalled func(stats common.SnapshotStatisticsHandler)
	GetSnapshotMessageCalled                func() string
}

// UpdateMetricsOnSnapshotStart -
func (s *StateMetricsStub) UpdateMetricsOnSnapshotStart() {
	if s.UpdateMetricsOnSnapshotStartCalled != nil {
		s.UpdateMetricsOnSnapshotStartCalled()
	}
}

// UpdateMetricsOnSnapshotCompletion -
func (s *StateMetricsStub) UpdateMetricsOnSnapshotCompletion(stats common.SnapshotStatisticsHandler) {
	if s.UpdateMetricsOnSnapshotCompletionCalled != nil {
		s.UpdateMetricsOnSnapshotCompletionCalled(stats)
	}
}

// GetSnapshotMessage -
func (s *StateMetricsStub) GetSnapshotMessage() string {
	if s.GetSnapshotMessageCalled != nil {
		return s.GetSnapshotMessageCalled()
	}
	return "snapshot state"
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *StateMetricsStub) IsInterfaceNil() bool {
	return s == nil
}
