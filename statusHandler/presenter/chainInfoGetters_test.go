package presenter

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestPresenterStatusHandler_GetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNonce, nonce)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, nonce, result)
}

func TestPresenterStatusHandler_GetIsSyncing(t *testing.T) {
	t.Parallel()

	isSyncing := uint64(1)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricIsSyncing, isSyncing)
	result := presenterStatusHandler.GetIsSyncing()

	assert.Equal(t, isSyncing, result)
}

func TestPresenterStatusHandler_GetTxPoolLoad(t *testing.T) {
	t.Parallel()

	txPoolLoad := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricTxPoolLoad, txPoolLoad)
	result := presenterStatusHandler.GetTxPoolLoad()

	assert.Equal(t, txPoolLoad, result)
}

func TestPresenterStatusHandler_GetProbableHighestNonce(t *testing.T) {
	t.Parallel()

	probableHighestNonce := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricProbableHighestNonce, probableHighestNonce)
	result := presenterStatusHandler.GetProbableHighestNonce()

	assert.Equal(t, probableHighestNonce, result)
}

func TestPresenterStatusHandler_GetSynchronizedRound(t *testing.T) {
	t.Parallel()

	synchronizedRound := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricSynchronizedRound, synchronizedRound)
	result := presenterStatusHandler.GetSynchronizedRound()

	assert.Equal(t, synchronizedRound, result)
}

func TestPresenterStatusHandler_GetRoundTime(t *testing.T) {
	t.Parallel()

	roundTime := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricRoundTime, roundTime)
	result := presenterStatusHandler.GetRoundTime()

	assert.Equal(t, roundTime, result)
}

func TestPresenterStatusHandler_GetLiveValidatorNodes(t *testing.T) {
	t.Parallel()

	numLiveValidatorNodes := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricLiveValidatorNodes, numLiveValidatorNodes)
	result := presenterStatusHandler.GetLiveValidatorNodes()

	assert.Equal(t, numLiveValidatorNodes, result)
}

func TestPresenterStatusHandler_GetConnectedNodes(t *testing.T) {
	t.Parallel()

	numConnectedNodes := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricConnectedNodes, numConnectedNodes)
	result := presenterStatusHandler.GetConnectedNodes()

	assert.Equal(t, numConnectedNodes, result)
}

func TestPresenterStatusHandler_GetNumConnectedPeers(t *testing.T) {
	t.Parallel()

	numConnectedPeers := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNumConnectedPeers, numConnectedPeers)
	result := presenterStatusHandler.GetNumConnectedPeers()

	assert.Equal(t, numConnectedPeers, result)
}

func TestPresenterStatusHandler_GetCurrentRound(t *testing.T) {
	t.Parallel()

	currentRound := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricCurrentRound, currentRound)
	result := presenterStatusHandler.GetCurrentRound()

	assert.Equal(t, currentRound, result)
}
