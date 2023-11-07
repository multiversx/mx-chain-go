package stateMetrics_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/stateMetrics"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewStateMetrics(t *testing.T) {
	t.Parallel()

	t.Run("nil app status handler", func(t *testing.T) {
		t.Parallel()

		sm, err := stateMetrics.NewStateMetrics(stateMetrics.ArgsStateMetrics{}, nil)
		assert.Nil(t, sm)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		setSnapshotInProgressKey := false
		snapshotInProgressKey := "snapshotInProgressKey"

		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetUInt64ValueHandler: func(key string, value uint64) {
				setSnapshotInProgressKey = true
				assert.Equal(t, snapshotInProgressKey, key)
				assert.Equal(t, uint64(0), value)
			},
		}

		args := stateMetrics.ArgsStateMetrics{
			SnapshotInProgressKey: snapshotInProgressKey,
		}

		sm, err := stateMetrics.NewStateMetrics(args, appStatusHandler)
		assert.NotNil(t, sm)
		assert.Nil(t, err)

		assert.True(t, setSnapshotInProgressKey)
	})
}

func TestStateMetrics_UpdateMetricsOnSnapshotStart(t *testing.T) {
	t.Parallel()

	numSetUInt64ValueHandlerCalls := 0
	setInt64ValueCalled := false
	snapshotInProgressKey := "snapshotInProgressKey"
	lastSnapshotDurationKey := "lastSnapshotDurationKey"

	appStatusHandler := &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			assert.Equal(t, snapshotInProgressKey, key)
			assert.Equal(t, uint64(numSetUInt64ValueHandlerCalls), value)
			numSetUInt64ValueHandlerCalls++
		},
		SetInt64ValueHandler: func(key string, value int64) {
			assert.Equal(t, lastSnapshotDurationKey, key)
			assert.Equal(t, int64(0), value)
			setInt64ValueCalled = true
		},
	}

	args := stateMetrics.ArgsStateMetrics{
		SnapshotInProgressKey:   snapshotInProgressKey,
		LastSnapshotDurationKey: lastSnapshotDurationKey,
	}

	sm, _ := stateMetrics.NewStateMetrics(args, appStatusHandler)
	sm.UpdateMetricsOnSnapshotStart()

	assert.Equal(t, 2, numSetUInt64ValueHandlerCalls)
	assert.True(t, setInt64ValueCalled)
}

func TestStateMetrics_UpdateMetricsOnSnapshotCompletion(t *testing.T) {
	t.Parallel()

	numSetUInt64ValueHandlerCalls := 0
	setInt64ValueCalled := false
	setNumNodesCalled := false
	snapshotInProgressKey := "snapshotInProgressKey"
	lastSnapshotDurationKey := "lastSnapshotDurationKey"

	appStatusHandler := &statusHandler.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			if bytes.Equal([]byte(snapshotInProgressKey), []byte(key)) {
				assert.Equal(t, uint64(0), value)
				numSetUInt64ValueHandlerCalls++
			}

			if bytes.Equal([]byte(common.MetricAccountsSnapshotNumNodes), []byte(key)) {
				setNumNodesCalled = true
			}
		},
		SetInt64ValueHandler: func(key string, value int64) {
			assert.Equal(t, lastSnapshotDurationKey, key)
			assert.Equal(t, int64(0), value)
			setInt64ValueCalled = true
		},
	}

	args := stateMetrics.ArgsStateMetrics{
		SnapshotInProgressKey:   snapshotInProgressKey,
		LastSnapshotDurationKey: lastSnapshotDurationKey,
		SnapshotMessage:         stateMetrics.UserTrieSnapshotMsg,
	}

	sm, _ := stateMetrics.NewStateMetrics(args, appStatusHandler)
	sm.UpdateMetricsOnSnapshotCompletion(&trie.MockStatistics{})

	assert.Equal(t, 2, numSetUInt64ValueHandlerCalls)
	assert.True(t, setInt64ValueCalled)
	assert.True(t, setNumNodesCalled)
}
