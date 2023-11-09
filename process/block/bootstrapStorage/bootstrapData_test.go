package bootstrapStorage

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/assert"
)

func TestMiniBlocksInMeta_IsFullyProcessedShouldWork(t *testing.T) {
	t.Parallel()

	mbim := MiniBlocksInMeta{}

	isFullyProcessed := mbim.IsFullyProcessed(0)
	assert.True(t, isFullyProcessed)

	mbim.FullyProcessed = make([]bool, 0)
	isFullyProcessed = mbim.IsFullyProcessed(0)
	assert.True(t, isFullyProcessed)

	mbim.FullyProcessed = append(mbim.FullyProcessed, true)
	isFullyProcessed = mbim.IsFullyProcessed(0)
	assert.True(t, isFullyProcessed)

	mbim.FullyProcessed = append(mbim.FullyProcessed, false)
	isFullyProcessed = mbim.IsFullyProcessed(1)
	assert.False(t, isFullyProcessed)

	isFullyProcessed = mbim.IsFullyProcessed(2)
	assert.True(t, isFullyProcessed)
}

func TestMiniBlocksInMeta_GetIndexOfLastTxProcessedInMiniBlock(t *testing.T) {
	t.Parallel()

	mbim := MiniBlocksInMeta{}

	index := mbim.GetIndexOfLastTxProcessedInMiniBlock(0)
	assert.Equal(t, common.MaxIndexOfTxInMiniBlock, index)

	mbim.FullyProcessed = make([]bool, 0)
	index = mbim.GetIndexOfLastTxProcessedInMiniBlock(0)
	assert.Equal(t, common.MaxIndexOfTxInMiniBlock, index)

	mbim.IndexOfLastTxProcessed = append(mbim.IndexOfLastTxProcessed, 1)
	index = mbim.GetIndexOfLastTxProcessedInMiniBlock(0)
	assert.Equal(t, int32(1), index)

	index = mbim.GetIndexOfLastTxProcessedInMiniBlock(1)
	assert.Equal(t, common.MaxIndexOfTxInMiniBlock, index)
}
