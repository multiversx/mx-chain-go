package stateMetrics

import (
	"errors"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"sync"
)

const (
	// UserTrieSnapshotMsg - user trie snapshot message
	UserTrieSnapshotMsg = "user trie"

	// PeerTrieSnapshotMsg - peer trie snapshot message
	PeerTrieSnapshotMsg = "peer trie"
)

// ArgsStateMetrics are the arguments need for creating a new state metrics object
type ArgsStateMetrics struct {
	SnapshotInProgressKey   string
	LastSnapshotDurationKey string
	SnapshotMessage         string
}

type stateMetrics struct {
	appStatusHandler core.AppStatusHandler
	mutex            sync.RWMutex

	snapshotInProgressKey   string
	lastSnapshotDurationKey string
	snapshotMessage         string
}

// NewStateMetrics creates a new state metrics
func NewStateMetrics(args ArgsStateMetrics, appStatusHandler core.AppStatusHandler) (*stateMetrics, error) {
	if check.IfNil(appStatusHandler) {
		return nil, errors.New("nil app status handler")
	}

	appStatusHandler.SetUInt64Value(args.SnapshotInProgressKey, 0)

	return &stateMetrics{
		appStatusHandler:        appStatusHandler,
		snapshotInProgressKey:   args.SnapshotInProgressKey,
		lastSnapshotDurationKey: args.LastSnapshotDurationKey,
		snapshotMessage:         args.SnapshotMessage,
		mutex:                   sync.RWMutex{},
	}, nil
}

// UpdateMetricsOnSnapshotStart will update the metrics on snapshot start
func (sm *stateMetrics) UpdateMetricsOnSnapshotStart() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.appStatusHandler.SetUInt64Value(sm.snapshotInProgressKey, 1)
	sm.appStatusHandler.SetInt64Value(sm.lastSnapshotDurationKey, 0)
}

// UpdateMetricsOnSnapshotCompletion will update the metrics on snapshot completion
func (sm *stateMetrics) UpdateMetricsOnSnapshotCompletion(stats common.SnapshotStatisticsHandler) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.appStatusHandler.SetUInt64Value(sm.snapshotInProgressKey, 0)
	sm.appStatusHandler.SetInt64Value(sm.lastSnapshotDurationKey, stats.GetSnapshotDuration())
	if sm.snapshotMessage == UserTrieSnapshotMsg {
		sm.appStatusHandler.SetUInt64Value(common.MetricAccountsSnapshotNumNodes, stats.GetSnapshotNumNodes())
	}
}

// GetSnapshotMessage returns the snapshot message
func (sm *stateMetrics) GetSnapshotMessage() string {
	return sm.snapshotMessage
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *stateMetrics) IsInterfaceNil() bool {
	return sm == nil
}
