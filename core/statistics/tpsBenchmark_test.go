package statistics_test

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
)

func updateTpsBenchmark(tpsBenchmark *statistics.TpsBenchmark, txCount uint32) {
	shardData := block.ShardData{
		ShardID:    1,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:    block.PeerBlock,
				TxCount: 5000,
			},
			{
				Type:    block.TxBlock,
				TxCount: txCount,
			},
		},
	}
	metaBlock := &block.MetaBlock{
		Nonce:     1,
		Round:     2,
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
	assert.False(t, check.IfNil(tpsBenchmark))
	assert.False(t, tpsBenchmark.ShardStatistics()[0].IsInterfaceNil())
}

func TestTpsBenchmark_BlockNumber(t *testing.T) {
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)
	blockNumber := uint64(1)
	metaBlock := &block.MetaBlock{
		Nonce: blockNumber,
		Round: blockNumber,
		ShardInfo: []block.ShardData{
			{
				ShardID:               0,
				HeaderHash:            []byte{1},
				ShardMiniBlockHeaders: []block.MiniBlockHeader{},
				PrevRandSeed:          []byte{1},
				PubKeysBitmap:         []byte{1},
				Signature:             []byte{1},
				TxCount:               10,
				Round:                 uint64(1),
				PrevHash:              []byte{1},
				Nonce:                 uint64(1),
			},
		},
	}
	assert.Equal(t, tpsBenchmark.BlockNumber(), uint64(0))
	tpsBenchmark.Update(metaBlock)
	assert.Equal(t, tpsBenchmark.BlockNumber(), blockNumber)
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

	round := uint64(2)
	blockNumber := round

	metaBlock := &block.MetaBlock{
		Nonce: blockNumber - 1,
		Round: round - 1,
		ShardInfo: []block.ShardData{
			{
				ShardID:               0,
				HeaderHash:            []byte{1},
				ShardMiniBlockHeaders: []block.MiniBlockHeader{},
				PrevRandSeed:          []byte{1},
				PubKeysBitmap:         []byte{1},
				Signature:             []byte{1},
				TxCount:               10,
				Round:                 uint64(1),
				PrevHash:              []byte{1},
				Nonce:                 uint64(1),
			},
		},
	}
	metaBlock2 := &block.MetaBlock{
		Nonce: blockNumber,
		Round: round,
		ShardInfo: []block.ShardData{
			{
				ShardID:               0,
				HeaderHash:            []byte{1},
				ShardMiniBlockHeaders: []block.MiniBlockHeader{},
				PrevRandSeed:          []byte{1},
				PubKeysBitmap:         []byte{1},
				Signature:             []byte{1},
				TxCount:               10,
				Round:                 uint64(1),
				PrevHash:              []byte{1},
				Nonce:                 uint64(1),
			},
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
	round := uint64(1)
	blockNumber := round
	totalTxCount := big.NewInt(int64(25))

	metaBlock1 := &block.MetaBlock{
		Nonce: blockNumber,
		Round: round,
		ShardInfo: []block.ShardData{
			{
				ShardID:    0,
				HeaderHash: []byte{1},
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						Type:    block.PeerBlock,
						TxCount: 1000,
					},
					{
						Type:    block.TxBlock,
						TxCount: 5,
					},
					{
						Type:    block.TxBlock,
						TxCount: 5,
					},
				},
			},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:    block.PeerBlock,
				TxCount: 1000,
			},
			{
				Type:    block.TxBlock,
				TxCount: 5,
			},
		},
	}

	metaBlock2 := &block.MetaBlock{
		Nonce: blockNumber + 1,
		Round: round + 1,
		ShardInfo: []block.ShardData{
			{
				ShardID:    0,
				HeaderHash: []byte{1},
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						Type:    block.PeerBlock,
						TxCount: 5000,
					},
					{
						Type:    block.TxBlock,
						TxCount: 5,
					},
					{
						Type:    block.TxBlock,
						TxCount: 5,
					},
				},
			},
		},
	}

	tpsBenchmark.Update(metaBlock1)
	tpsBenchmark.Update(metaBlock2)
	assert.Equal(t, tpsBenchmark.TotalProcessedTxCount(), totalTxCount)
}

func TestTpsBenchmark_UpdatePeakTps(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(1)
	roundDuration := uint64(1)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	round := uint64(1)
	blockNumber := round
	peakTps := uint32(20)

	metaBlock1 := &block.MetaBlock{
		Nonce:   blockNumber,
		Round:   round,
		TxCount: 5005,
		ShardInfo: []block.ShardData{
			{
				ShardID:    0,
				HeaderHash: []byte{1},
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						Type:    block.PeerBlock,
						TxCount: 5000,
					},
					{
						Type:    block.TxBlock,
						TxCount: 5,
					},
					{
						Type:    block.TxBlock,
						TxCount: 5,
					},
				},
				PrevRandSeed:  []byte{1},
				PubKeysBitmap: []byte{1},
				Signature:     []byte{1},
				TxCount:       5010,
				Round:         uint64(1),
				PrevHash:      []byte{1},
				Nonce:         uint64(1),
			},
		},
	}

	metaBlock2 := &block.MetaBlock{
		Nonce:   blockNumber + 1,
		Round:   round + 1,
		TxCount: 5005,
		ShardInfo: []block.ShardData{
			{
				ShardID:    0,
				HeaderHash: []byte{1},
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						Type:    block.PeerBlock,
						TxCount: 5000,
					},
					{
						Type:    block.TxBlock,
						TxCount: 10,
					},
					{
						Type:    block.TxBlock,
						TxCount: 10,
					},
				},
				PrevRandSeed:  []byte{1},
				PubKeysBitmap: []byte{1},
				Signature:     []byte{1},
				TxCount:       5020,
				Round:         uint64(1),
				PrevHash:      []byte{1},
				Nonce:         uint64(1),
			},
		},
	}

	tpsBenchmark.Update(metaBlock1)
	tpsBenchmark.Update(metaBlock2)
	assert.Equal(t, float64(peakTps), tpsBenchmark.PeakTPS())
}

func TestTPSBenchmark_GettersAndSetters(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(1)
	roundDuration := uint64(1)
	shardId := uint32(0)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	round := uint64(1)
	blockNumber := round

	totalTxs := uint32(10)
	shardData := block.ShardData{
		ShardID:    shardId,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:    block.PeerBlock,
				TxCount: 5000,
			},
			{
				Type:    block.TxBlock,
				TxCount: 5,
			},
			{
				Type:    block.TxBlock,
				TxCount: 5,
			},
		},
		TxCount: 5100,
	}
	metaBlock := &block.MetaBlock{
		Nonce:     blockNumber,
		Round:     round,
		TxCount:   totalTxs,
		ShardInfo: []block.ShardData{shardData},
	}

	tpsBenchmark.Update(metaBlock)

	assert.Equal(t, nrOfShards, tpsBenchmark.NrOfShards())
	assert.Equal(t, roundDuration, tpsBenchmark.RoundTime())
	assert.Equal(t, blockNumber, tpsBenchmark.BlockNumber())
	assert.Equal(t, blockNumber, tpsBenchmark.BlockNumber())
	assert.Equal(t, float64(totalTxs), tpsBenchmark.PeakTPS())
	assert.Equal(t, totalTxs, tpsBenchmark.LastBlockTxCount())
	assert.Equal(t, big.NewInt(int64(totalTxs)), tpsBenchmark.AverageBlockTxCount())
	assert.Equal(t, big.NewInt(int64(totalTxs)), tpsBenchmark.TotalProcessedTxCount())
	assert.Equal(t, totalTxs, tpsBenchmark.ShardStatistic(shardId).LastBlockTxCount())
}

func TestTPSBenchmarkShardStatistics_GettersAndSetters(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(2)
	roundDuration := uint64(1)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	round := uint64(1)
	blockNumber := round
	txCount := uint32(5)

	shardDataShard0 := block.ShardData{
		ShardID:    0,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:    block.PeerBlock,
				TxCount: 5000,
			},
			{
				Type:    block.TxBlock,
				TxCount: 5,
			},
		},
		TxCount: 5005,
	}
	shardDataShard1 := block.ShardData{
		ShardID:    1,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:    block.PeerBlock,
				TxCount: 5000,
			},
			{
				Type:    block.TxBlock,
				TxCount: 5,
			},
		},
		TxCount: 5005,
	}
	metaBlock := &block.MetaBlock{
		Nonce:     blockNumber,
		Round:     round,
		TxCount:   10010,
		ShardInfo: []block.ShardData{shardDataShard0, shardDataShard1},
	}

	tpsBenchmark.Update(metaBlock)

	shardStatistics := tpsBenchmark.ShardStatistics()
	assert.NotNil(t, shardStatistics)

	for shardId, shardStats := range shardStatistics {
		assert.Equal(t, 1, shardStats.AverageTPS().Cmp(big.NewInt(0)))
		assert.True(t, shardStats.LiveTPS() > 0)
		assert.Equal(t, round, shardStats.CurrentBlockNonce())
		assert.Equal(t, shardId, shardStats.ShardID())
		assert.Equal(t, float64(txCount), shardStats.PeakTPS())
		assert.Equal(t, txCount, shardStats.LastBlockTxCount())
		assert.Equal(t, big.NewInt(int64(txCount)), shardStats.TotalProcessedTxCount())
	}
}

func TestTpsBenchmark_ShouldUpdateSameNonceOnlyOnce(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(2)
	roundDuration := uint64(6)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	txCount := uint32(10)

	shardData := block.ShardData{
		ShardID:    1,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:    block.PeerBlock,
				TxCount: 5000,
			},
			{
				Type:    block.TxBlock,
				TxCount: txCount,
			},
		},
		TxCount: 5005,
	}
	metaBlock := &block.MetaBlock{
		Nonce:     1,
		Round:     2,
		TxCount:   5005,
		ShardInfo: []block.ShardData{shardData},
	}
	tpsBenchmark.Update(metaBlock)

	shardData2 := block.ShardData{
		ShardID:               1,
		HeaderHash:            []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{},
		TxCount:               5005,
	}
	metaBlock2 := &block.MetaBlock{
		Nonce:     1,
		Round:     2,
		TxCount:   5005,
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
		ShardID:    0,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:    block.PeerBlock,
				TxCount: 5000,
			},
			{
				Type:    block.TxBlock,
				TxCount: txCount,
			},
		},
		TxCount: 5010,
	}
	shard2Data := block.ShardData{
		ShardID:               1,
		HeaderHash:            []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{},
		TxCount:               0,
	}
	metaBlock := &block.MetaBlock{
		Nonce:     1,
		Round:     2,
		TxCount:   txCount,
		ShardInfo: []block.ShardData{shardData, shard2Data},
	}
	tpsBenchmark.Update(metaBlock)

	bigTxCount := big.NewInt(int64(txCount))
	assert.Equal(t, bigTxCount, tpsBenchmark.TotalProcessedTxCount())
}

func TestTpsBenchmark_Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	t.Parallel()

	numOfTests := 25
	for i := 0; i < numOfTests; i++ {
		testTpsBenchmarkConcurrent(t)
	}
}

func testTpsBenchmarkConcurrent(t *testing.T) {
	nrOfShards := uint32(2)
	roundDuration := uint64(6)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	txCount := uint32(10)
	nrGoroutines := 8000

	wg := sync.WaitGroup{}
	wg.Add(nrGoroutines)

	for i := 0; i < nrGoroutines; i++ {
		go func() {
			time.Sleep(time.Millisecond)
			updateTpsBenchmark(tpsBenchmark, txCount)
			wg.Done()
		}()
	}
	wg.Wait()

	bigTxCount := big.NewInt(int64(txCount))
	assert.Equal(t, bigTxCount, tpsBenchmark.TotalProcessedTxCount())
}

func TestTpsBenchmark_ZeroTxMetaBlockAndShardHeader(t *testing.T) {
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 4)

	shardData := block.ShardData{
		ShardID:               0,
		HeaderHash:            []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{},
		TxCount:               0,
	}
	metaBlock := &block.MetaBlock{
		Nonce:     1,
		Round:     2,
		TxCount:   0,
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
		Nonce:     1,
		Round:     2,
		TxCount:   0,
		ShardInfo: []block.ShardData{},
	}

	tpsBenchmark.Update(metaBlock)

	bigTxCount := big.NewInt(0)
	assert.Equal(t, bigTxCount, tpsBenchmark.TotalProcessedTxCount())
}
