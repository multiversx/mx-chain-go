package presenter

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestPresenterStatusHandler_GetNumTxInBlock(t *testing.T) {
	t.Parallel()

	numTxInBlock := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, numTxInBlock)
	result := presenterStatusHandler.GetNumTxInBlock()

	assert.Equal(t, numTxInBlock, result)
}

func TestPresenterStatusHandler_GetNumTxInBlockShouldBeZero(t *testing.T) {
	t.Parallel()

	numTxInBlock := "1000"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricNumTxInBlock, numTxInBlock)
	result := presenterStatusHandler.GetNumTxInBlock()

	assert.Equal(t, uint64(0), result)
}

func TestPresenterStatusHandler_GetNumTxShouldZeroIfIsNotSet(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := NewPresenterStatusHandler()
	result := presenterStatusHandler.GetNumTxInBlock()

	assert.Equal(t, uint64(0), result)
}

func TestPresenterStatusHandler_GetNumMiniBLocks(t *testing.T) {
	t.Parallel()

	numMiniBlocks := uint64(100)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, numMiniBlocks)
	result := presenterStatusHandler.GetNumMiniBlocks()

	assert.Equal(t, numMiniBlocks, result)
}

func TestPresenterStatusHandler_GetCrossCheckBlockHeight(t *testing.T) {
	t.Parallel()

	crossCheckBlockHeight := "meta:1000"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricCrossCheckBlockHeight, crossCheckBlockHeight)
	result := presenterStatusHandler.GetCrossCheckBlockHeight()

	assert.Equal(t, crossCheckBlockHeight, result)
}

func TestPresenterStatusHandler_GetConsensusState(t *testing.T) {
	t.Parallel()

	consensusState := "not in consensus group"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricConsensusState, consensusState)
	result := presenterStatusHandler.GetConsensusState()

	assert.Equal(t, consensusState, result)
}

func TestPresenterStatusHandler_GetConsensusStateShouldReturnErrorMessageInvalidType(t *testing.T) {
	t.Parallel()

	consensusState := uint64(1)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricConsensusState, consensusState)
	result := presenterStatusHandler.GetConsensusState()

	assert.Equal(t, invalidType, result)
}

func TestPresenterStatusHandler_GetConsensusStateShouldReturnErrorMessageInvalidKey(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := NewPresenterStatusHandler()
	result := presenterStatusHandler.GetConsensusState()

	assert.Equal(t, invalidKey, result)
}

func TestPresenterStatusHandler_GetConsensusRoundStateState(t *testing.T) {
	t.Parallel()

	consensusRoundState := "participant"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricConsensusRoundState, consensusRoundState)
	result := presenterStatusHandler.GetConsensusRoundState()

	assert.Equal(t, consensusRoundState, result)
}
