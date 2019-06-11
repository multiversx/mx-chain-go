package statistics_test

import (
	"math/big"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func updateTpsBenchmark(tpsBenchmark *statistics.TpsBenchmark, txCount uint32) {
	shardData := block.ShardData{
		ShardId: 1,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{},
		TxCount: txCount,
	}
	metaBlock := &block.MetaBlock{
		Nonce: 1,
		Round: 2,
		TxCount: txCount,
		ShardInfo: []block.ShardData{shardData},
	}
	tpsBenchmark.Update(metaBlock)

	_ = tpsBenchmark.ActiveNodes()
	_ = tpsBenchmark.RoundTime()
	_ = tpsBenchmark.BlockNumber()
	_ = tpsBenchmark.RoundNumber()
	_ = tpsBenchmark.AverageBlockTxCount()
	_ = tpsBenchmark.LastBlockTxCount()
	_ = tpsBenchmark.TotalProcessedTxCount()
	_ = tpsBenchmark.LiveTPS()
	_ = tpsBenchmark.PeakTPS()
	_ = tpsBenchmark.NrOfShards()
	_ = tpsBenchmark.ShardStatistics()
}

func TestTpsBenchmark_NewTPSBenchmarkReturnsErrorOnInvalidDuration(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(10)
	roundDuration := uint64(0)
	tpsBenchmark, err := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	assert.Nil(t, tpsBenchmark)
	assert.Equal(t, err, statistics.ErrInvalidRoundDuration)
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
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)
	blockNumber := uint32(1)
	metaBlock := &block.MetaBlock{
		Nonce: uint64(blockNumber),
		Round: blockNumber,
		ShardInfo: []block.ShardData{
			{0, []byte{1}, []block.ShardMiniBlockHeader{}, 10},
		},
	}
	assert.Equal(t, tpsBenchmark.BlockNumber(), uint64(0))
	tpsBenchmark.Update(metaBlock)
	assert.Equal(t, tpsBenchmark.BlockNumber(), uint64(blockNumber))
}

func TestTpsBenchmark_UpdateIrrelevantBlock(t *testing.T) {
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)

	tpsBenchmark.Update(nil)
	assert.Equal(t, tpsBenchmark.BlockNumber(), uint64(0))
}

func TestTpsBenchmark_UpdateSmallerNonce(t *testing.T) {
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)

	round := uint32(2)
	blockNumber := uint64(round)

	metaBlock := &block.MetaBlock{
		Nonce: blockNumber - 1,
		Round: round - 1,
		ShardInfo: []block.ShardData{
			{0, []byte{1}, []block.ShardMiniBlockHeader{}, 10},
		},
	}
	metaBlock2 := &block.MetaBlock{
		Nonce: blockNumber,
		Round: round,
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
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)
	blockNumber := uint64(1)

	metaBlock := &block.MetaBlock{
		Nonce:     blockNumber,
		ShardInfo: make([]block.ShardData, 0),
	}

	tpsBenchmark.Update(metaBlock)
	assert.Equal(t, tpsBenchmark.BlockNumber(), uint64(0))
}

func TestTpsBenchmark_UpdateTotalNumberOfTx(t *testing.T) {
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)
	round := uint32(1)
	blockNumber := uint64(round)
	txCount := uint32(10)
	totalTxCount := big.NewInt(int64(txCount * 2))

	metaBlock := &block.MetaBlock{
		Nonce: blockNumber,
		Round: round,
		TxCount: txCount,
		ShardInfo: []block.ShardData{
			{0, []byte{1}, []block.ShardMiniBlockHeader{}, txCount},
		},
	}

	metaBlock2 := &block.MetaBlock{
		Nonce: blockNumber + 1,
		Round: round + 1,
		TxCount: txCount,
		ShardInfo: []block.ShardData{
			{0, []byte{1}, []block.ShardMiniBlockHeader{}, txCount},
		},
	}

	tpsBenchmark.Update(metaBlock)
	tpsBenchmark.Update(metaBlock2)
	assert.Equal(t, tpsBenchmark.TotalProcessedTxCount(), totalTxCount)
}

func TestTpsBenchmark_UpdatePeakTps(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(1)
	roundDuration := uint64(1)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	round := uint32(1)
	blockNumber := uint64(round)
	txCount := uint32(10)
	peakTps := uint32(20)

	metaBlock := &block.MetaBlock{
		Nonce: blockNumber,
		Round: round,
		TxCount: peakTps,
		ShardInfo: []block.ShardData{
			{0, []byte{1}, []block.ShardMiniBlockHeader{}, peakTps},
		},
	}

	metaBlock2 := &block.MetaBlock{
		Nonce: blockNumber + 1,
		Round: round + 1,
		TxCount: txCount,
		ShardInfo: []block.ShardData{
			{0, []byte{1}, []block.ShardMiniBlockHeader{}, txCount},
		},
	}

	tpsBenchmark.Update(metaBlock)
	tpsBenchmark.Update(metaBlock2)
	assert.Equal(t, float64(peakTps), tpsBenchmark.PeakTPS())
}

func TestTPSBenchmark_GettersAndSetters(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(1)
	roundDuration := uint64(1)
	shardId := uint32(0)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	round := uint32(1)
	blockNumber := uint64(round)
	txCount := uint32(10)

	shardData := block.ShardData{
		ShardId: shardId,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{},
		TxCount: txCount,
	}
	metaBlock := &block.MetaBlock{
		Nonce: blockNumber,
		Round: round,
		TxCount: txCount,
		ShardInfo: []block.ShardData{shardData},
	}

	tpsBenchmark.Update(metaBlock)

	assert.Equal(t, nrOfShards, tpsBenchmark.NrOfShards())
	assert.Equal(t, roundDuration, tpsBenchmark.RoundTime())
	assert.Equal(t, blockNumber, tpsBenchmark.BlockNumber())
	assert.Equal(t, blockNumber, tpsBenchmark.BlockNumber())
	assert.Equal(t, float64(txCount), tpsBenchmark.PeakTPS())
	assert.Equal(t, txCount, tpsBenchmark.LastBlockTxCount())
	assert.Equal(t, big.NewInt(int64(txCount)), tpsBenchmark.AverageBlockTxCount())
	assert.Equal(t, big.NewInt(int64(txCount)), tpsBenchmark.TotalProcessedTxCount())
	assert.Equal(t, shardData.TxCount, tpsBenchmark.ShardStatistic(shardId).LastBlockTxCount())
}

func TestTpsBenchmark_ShouldUpdateSameNonceOnlyOnce(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(2)
	roundDuration := uint64(6)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	txCount := uint32(10)

	shardData := block.ShardData{
		ShardId: 1,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{},
		TxCount: txCount,
	}
	metaBlock := &block.MetaBlock{
		Nonce: 1,
		Round: 2,
		TxCount: txCount,
		ShardInfo: []block.ShardData{shardData},
	}
	tpsBenchmark.Update(metaBlock)

	shardData2 := block.ShardData{
		ShardId: 1,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{},
		TxCount: txCount,
	}
	metaBlock2 := &block.MetaBlock{
		Nonce: 1,
		Round: 2,
		TxCount: txCount,
		ShardInfo: []block.ShardData{shardData2},
	}
	tpsBenchmark.Update(metaBlock2)

	bigTxCount := big.NewInt(int64(txCount))
	assert.Equal(t, bigTxCount, tpsBenchmark.TotalProcessedTxCount())
}

func TestTpsBenchmark_EmptyBlocksShouldNotUpdateMultipleTimes(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(2)
	roundDuration := uint64(6)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	txCount := uint32(10)

	shardData := block.ShardData{
		ShardId: 0,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{},
		TxCount: txCount,
	}
	shard2Data := block.ShardData{
		ShardId: 1,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{},
		TxCount: 0,
	}
	metaBlock := &block.MetaBlock{
		Nonce: 1,
		Round: 2,
		TxCount: txCount,
		ShardInfo: []block.ShardData{shardData, shard2Data},
	}
	tpsBenchmark.Update(metaBlock)

	bigTxCount := big.NewInt(int64(txCount))
	assert.Equal(t, bigTxCount, tpsBenchmark.TotalProcessedTxCount())
}

func TestTpsBenchmark_Concurrent(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(2)
	roundDuration := uint64(6)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	txCount := uint32(10)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		updateTpsBenchmark(tpsBenchmark, txCount)
		wg.Done()
	}()

	go func() {
		updateTpsBenchmark(tpsBenchmark, txCount)
		wg.Done()
	}()

	wg.Wait()

	bigTxCount := big.NewInt(int64(txCount))
	assert.Equal(t, bigTxCount, tpsBenchmark.TotalProcessedTxCount())
}

func TestTpsBenchmark_ZeroTxMetaBlockAndShardHeader(t *testing.T) {
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 4)

	shardData := block.ShardData{
		ShardId: 0,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{},
		TxCount: 0,
	}
	metaBlock := &block.MetaBlock{
		Nonce: 1,
		Round: 2,
		TxCount: 0,
		ShardInfo: []block.ShardData{shardData},
	}

	tpsBenchmark.Update(metaBlock)

	bigTxCount := big.NewInt(0)
	assert.Equal(t, bigTxCount, tpsBenchmark.TotalProcessedTxCount())
}

func TestTpsBenchmark_ZeroTxMetaBlockAndEmptyShardHeader(t *testing.T) {
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 4)

	metaBlock := &block.MetaBlock{
		Nonce: 1,
		Round: 2,
		TxCount: 0,
		ShardInfo: []block.ShardData{},
	}

	tpsBenchmark.Update(metaBlock)

	bigTxCount := big.NewInt(0)
	assert.Equal(t, bigTxCount, tpsBenchmark.TotalProcessedTxCount())
}
