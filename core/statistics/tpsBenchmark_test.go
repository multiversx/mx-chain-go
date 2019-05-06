package statistics_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)


func TestTpsBenchmark_NewTPSBenchmarkReturnsErrorOnInvalidDuration(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(10)
	roundDuration := uint64(0)
	tpsBenchmark, err := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	assert.Nil(t, tpsBenchmark)
	assert.Equal(t, err, core.ErrInvalidRoundDuration)
}

func TestTpsBenchmark_NewTPSBenchmark(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(10)
	roundDuration := uint64(4)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	gotNrOfShards := uint32(len(tpsBenchmark.ShardStatistics()))

	assert.Equal(t, gotNrOfShards, nrOfShards)
	assert.Equal(t, tpsBenchmark.RoundTime(), roundDuration)
}

func TestTpsBenchmark_BlockNumber(t *testing.T) {
	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)
	blockNumber := uint64(1)
	metaBlock := &block.MetaBlock{
		Nonce: blockNumber,
		ShardInfo: []block.ShardData{
			{0, []byte{1}, []block.ShardMiniBlockHeader{}, 10},
		},
	}
	assert.Equal(t, tpsBenchmark.BlockNumber(), uint64(0))
	tpsBenchmark.Update(metaBlock)
	assert.Equal(t, tpsBenchmark.BlockNumber(), blockNumber)
}

func TestTpsBenchmark_UpdateIrrelevantBlock(t *testing.T) {
	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)

	tpsBenchmark.Update(nil)
	assert.Equal(t, tpsBenchmark.BlockNumber(), uint64(0))
}

func TestTpsBenchmark_UpdateSmallerNonce(t *testing.T) {
	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)

	blockNumber := uint64(2)

	metaBlock := &block.MetaBlock{
		Nonce: blockNumber - 1,
		ShardInfo: []block.ShardData{
			{0, []byte{1}, []block.ShardMiniBlockHeader{}, 10},
		},
	}
	metaBlock2 := &block.MetaBlock{
		Nonce: blockNumber,
		ShardInfo: []block.ShardData{
			{0, []byte{1}, []block.ShardMiniBlockHeader{}, 10},
		},
	}
	// Start with block with nonce 1 so it would be processed
	tpsBenchmark.Update(metaBlock)
	// Add second block, again, it would be processed
	tpsBenchmark.Update(metaBlock2)
	// Try adding the first block again, it should not be processed
	tpsBenchmark.Update(metaBlock)

	assert.Equal(t, tpsBenchmark.BlockNumber(), blockNumber)
}

func TestTpsBenchmark_UpdateEmptyShardInfoInMiniblock(t *testing.T) {
	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)
	blockNumber := uint64(1)

	metaBlock := &block.MetaBlock{
		Nonce:     blockNumber,
		ShardInfo: make([]block.ShardData, 0),
	}

	tpsBenchmark.Update(metaBlock)
	assert.Equal(t, tpsBenchmark.BlockNumber(), uint64(0))
}
