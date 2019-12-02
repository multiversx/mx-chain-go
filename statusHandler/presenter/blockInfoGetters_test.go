package presenter

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/constants"

	"github.com/stretchr/testify/assert"
)

func TestPresenterStatusHandler_GetNumTxInBlock(t *testing.T) {
	t.Parallel()

	numTxInBlock := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(constants.MetricNumTxInBlock, numTxInBlock)
	result := presenterStatusHandler.GetNumTxInBlock()

	assert.Equal(t, numTxInBlock, result)
}

func TestPresenterStatusHandler_GetNumTxInBlockShouldBeZero(t *testing.T) {
	t.Parallel()

	numTxInBlock := "1000"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(constants.MetricNumTxInBlock, numTxInBlock)
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
	presenterStatusHandler.SetUInt64Value(constants.MetricNumMiniBlocks, numMiniBlocks)
	result := presenterStatusHandler.GetNumMiniBlocks()

	assert.Equal(t, numMiniBlocks, result)
}

func TestPresenterStatusHandler_GetCrossCheckBlockHeight(t *testing.T) {
	t.Parallel()

	crossCheckBlockHeight := "meta:1000"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(constants.MetricCrossCheckBlockHeight, crossCheckBlockHeight)
	result := presenterStatusHandler.GetCrossCheckBlockHeight()

	assert.Equal(t, crossCheckBlockHeight, result)
}

func TestPresenterStatusHandler_GetConsensusState(t *testing.T) {
	t.Parallel()

	consensusState := "not in consensus group"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(constants.MetricConsensusState, consensusState)
	result := presenterStatusHandler.GetConsensusState()

	assert.Equal(t, consensusState, result)
}

func TestPresenterStatusHandler_GetConsensusStateShouldReturnErrorMessageInvalidType(t *testing.T) {
	t.Parallel()

	consensusState := uint64(1)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(constants.MetricConsensusState, consensusState)
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
	presenterStatusHandler.SetStringValue(constants.MetricConsensusRoundState, consensusRoundState)
	result := presenterStatusHandler.GetConsensusRoundState()

	assert.Equal(t, consensusRoundState, result)
}

func TestPresenterStatusHandler_GetCurrentBlockHash(t *testing.T) {
	t.Parallel()

	currentBlockHash := "hash"
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(constants.MetricCurrentBlockHash, currentBlockHash)
	result := presenterStatusHandler.GetCurrentBlockHash()

	assert.Equal(t, currentBlockHash, result)
}

func TestPresenterStatusHandler_GetCurrentRoundTimestamp(t *testing.T) {
	t.Parallel()

	currentRoundTimestamp := uint64(time.Now().Unix())
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(constants.MetricCurrentRoundTimestamp, currentRoundTimestamp)
	result := presenterStatusHandler.GetCurrentRoundTimestamp()

	assert.Equal(t, currentRoundTimestamp, result)
}

func TestPresenterStatusHandler_GetBlockSize(t *testing.T) {
	t.Parallel()

	miniBlocksSize := uint64(100)
	headerSize := uint64(50)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(constants.MetricMiniBlocksSize, miniBlocksSize)
	presenterStatusHandler.SetUInt64Value(constants.MetricHeaderSize, headerSize)
	result := presenterStatusHandler.GetBlockSize()

	blockExpectedSize := miniBlocksSize + headerSize
	assert.Equal(t, blockExpectedSize, result)
}

func TestPresenterStatusHandler_GetHighestFinalBlockInShard(t *testing.T) {
	t.Parallel()

	highestFinalBlockNonce := uint64(100)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(constants.MetricHighestFinalBlockInShard, highestFinalBlockNonce)
	result := presenterStatusHandler.GetHighestFinalBlockInShard()

	assert.Equal(t, highestFinalBlockNonce, result)
}
