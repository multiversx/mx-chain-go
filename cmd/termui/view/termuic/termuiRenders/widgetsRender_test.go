package termuiRenders

import (
	"fmt"
	"testing"

	"github.com/gizak/termui/v3/widgets"
	"github.com/multiversx/mx-chain-go/cmd/termui/presenter"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWidgetsRender_prepareLogLines(t *testing.T) {
	t.Parallel()

	wr := &WidgetsRender{}
	numLogLines := 25
	testLogLines25 := make([]string, 0, numLogLines)
	for i := 0; i < numLogLines; i++ {
		testLogLines25 = append(testLogLines25, fmt.Sprintf("test line %d", i))
	}

	t.Run("small size should return empty", func(t *testing.T) {
		t.Parallel()

		for i := 0; i <= 2; i++ {
			result := wr.prepareLogLines(testLogLines25, i)
			assert.Empty(t, result)
		}
	})
	t.Run("equal size should return the same slice", func(t *testing.T) {
		t.Parallel()

		result := wr.prepareLogLines(testLogLines25, numLogLines+2)
		assert.Equal(t, testLogLines25, result)
	})
	t.Run("should trim", func(t *testing.T) {
		t.Parallel()

		result := wr.prepareLogLines(testLogLines25, numLogLines)
		assert.Equal(t, testLogLines25[2:], result)
		assert.Equal(t, 23, len(result))

		result = wr.prepareLogLines(testLogLines25, 10)
		assert.Equal(t, testLogLines25[17:], result)
		assert.Equal(t, 8, len(result))
	})
}

func TestPrepareBlockInfo(t *testing.T) {
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricNonce, 42)
	presenterStatusHandler.SetUInt64Value(common.MetricMiniBlocksSize, 2000)
	presenterStatusHandler.SetUInt64Value(common.MetricHeaderSize, 48)
	presenterStatusHandler.SetUInt64Value(common.MetricNumTxInBlock, 5)
	presenterStatusHandler.SetUInt64Value(common.MetricNumMiniBlocks, 2)
	presenterStatusHandler.SetStringValue(common.MetricCurrentBlockHash, "hash")
	presenterStatusHandler.SetStringValue(common.MetricCrossCheckBlockHeight, "cross123")
	presenterStatusHandler.SetUInt64Value(common.MetricHighestFinalBlock, 99)
	presenterStatusHandler.SetUInt64Value(common.MetricShardId, 0) // not metachain
	presenterStatusHandler.SetStringValue(common.MetricConsensusState, "ProposedBlock")
	presenterStatusHandler.SetUInt64Value(common.MetricIsSyncing, 0)
	presenterStatusHandler.SetStringValue(common.MetricConsensusRoundState, "Success")
	presenterStatusHandler.SetUInt64Value(common.MetricCurrentRoundTimestamp, 123456)
	presenterStatusHandler.SetUInt64Value(common.MetricReceivedOrSentProposedBlock, 2_000_000) // ns
	presenterStatusHandler.SetUInt64Value(common.MetricReceivedProof, 3_000_000)
	presenterStatusHandler.SetUInt64Value(common.MetricAvgReceivedOrSentProposedBlock, 1_000_000)
	presenterStatusHandler.SetUInt64Value(common.MetricAvgReceivedProof, 1_500_000)

	wr := &WidgetsRender{
		presenter: presenterStatusHandler,
		blockInfo: &widgets.Table{},
	}

	wr.prepareBlockInfo()

	require.Equal(t, "Block info:", wr.blockInfo.Title)
	require.False(t, wr.blockInfo.RowSeparator)
	require.Len(t, wr.blockInfo.Rows, 9)

	// Example: check one row
	require.Contains(t, wr.blockInfo.Rows[0][0], fmt.Sprintf("Current block height: %d, size: %s", 42, "2.00 KB"))
	require.Contains(t, wr.blockInfo.Rows[2][0], "hash")
	require.Contains(t, wr.blockInfo.Rows[4][0], "Consensus state: ProposedBlock")
	require.Contains(t, wr.blockInfo.Rows[5][0], "Consensus round state: Success")
	require.Contains(t, wr.blockInfo.Rows[6][0], "Received proposed block: 0.002000 sec | Received signatures: 0.003000 sec")
	require.Contains(t, wr.blockInfo.Rows[7][0], "Avg Received proposed block: 0.001000 sec | Avg Received signatures: 0.001500 sec")
	require.Contains(t, wr.blockInfo.Rows[8][0], "Current round timestamp: 123456")
	require.Contains(t, wr.blockInfo.Rows[3][0], "Cross check: cross123, final nonce: 99")

}
