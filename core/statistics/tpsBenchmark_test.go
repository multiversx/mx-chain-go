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

func updateTpsBenchmark(tpsBenchmark *statistics.TpsBenchmark, txCount uint32, nonce uint64) {
	shardData := block.ShardData{
		ShardID:    1,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:    block.PeerBlock,
				TxCount: 5000,
			},
			{
				Type:            block.TxBlock,
				TxCount:         txCount,
				ReceiverShardID: 1,
			},
		},
	}
	metaBlock := &block.MetaBlock{
		Nonce:     nonce,
		Round:     nonce,
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

func TestTpsBenchmark_UpdateTotalNumberOfTx(t *testing.T) {
	t.Parallel()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(1, 1)
	round := uint64(1)
	blockNumber := round
	totalTxCount := big.NewInt(int64(20))

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
					{
						Type:            block.TxBlock,
						TxCount:         2000,
						SenderShardID:   0,
						ReceiverShardID: 1,
					},
					{
						Type:    block.SmartContractResultBlock,
						TxCount: 200,
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
			{
				Type:    block.SmartContractResultBlock,
				TxCount: 200,
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
					{
						Type:    block.SmartContractResultBlock,
						TxCount: 200,
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
					{
						Type:    block.SmartContractResultBlock,
						TxCount: 200,
					},
				},
				PrevRandSeed:  []byte{1},
				PubKeysBitmap: []byte{1},
				Signature:     []byte{1},
				TxCount:       5210,
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
					{
						Type:    block.SmartContractResultBlock,
						TxCount: 200,
					},
				},
				PrevRandSeed:  []byte{1},
				PubKeysBitmap: []byte{1},
				Signature:     []byte{1},
				TxCount:       5220,
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
			{
				Type:    block.SmartContractResultBlock,
				TxCount: 200,
			},
		},
		TxCount: 5210,
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
			{
				Type:    block.SmartContractResultBlock,
				TxCount: 200,
			},
		},
		TxCount: 5205,
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
				Type:            block.TxBlock,
				TxCount:         5,
				ReceiverShardID: 1,
			},
			{
				Type:    block.SmartContractResultBlock,
				TxCount: 200,
			},
		},
		TxCount: 5205,
	}
	metaBlock := &block.MetaBlock{
		Nonce:     blockNumber,
		Round:     round,
		TxCount:   10410,
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
				Type:            block.TxBlock,
				TxCount:         txCount,
				ReceiverShardID: 1,
			},
			{
				Type:            block.RewardsBlock,
				TxCount:         txCount,
				ReceiverShardID: 1,
			},
			{
				Type:    block.SmartContractResultBlock,
				TxCount: 200,
			},
		},
		TxCount: 5205,
	}
	metaBlock := &block.MetaBlock{
		Nonce:     1,
		Round:     2,
		TxCount:   5205,
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

	bigTxCount := big.NewInt(int64(2 * txCount))
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
			{
				Type:    block.SmartContractResultBlock,
				TxCount: 200,
			},
		},
		TxCount: 5210,
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

func TestTpsBenchmark_TpsShouldUpdateButTxsCountShouldNot(t *testing.T) {
	t.Parallel()

	nrOfShards := uint32(1)
	roundDuration := uint64(6)
	tpsBenchmark, _ := statistics.NewTPSBenchmark(nrOfShards, roundDuration)
	txCount := uint32(60)

	shardData := block.ShardData{
		ShardID:    0,
		HeaderHash: []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{
			{
				Type:    block.PeerBlock,
				TxCount: 5000,
			},
			{
				Type:            block.TxBlock,
				TxCount:         txCount,
				ReceiverShardID: 1,
			},
			{
				Type:    block.SmartContractResultBlock,
				TxCount: 200,
			},
		},
		TxCount: 5210,
	}
	metaBlock := &block.MetaBlock{
		Nonce:     1,
		Round:     2,
		TxCount:   txCount,
		ShardInfo: []block.ShardData{shardData},
	}
	tpsBenchmark.Update(metaBlock)

	assert.Equal(t, big.NewInt(0), tpsBenchmark.TotalProcessedTxCount())
	assert.Equal(t, float64(txCount)/float64(roundDuration), tpsBenchmark.LiveTPS())
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

	for i := 1; i <= nrGoroutines; i++ {
		go func(nonce int) {
			time.Sleep(time.Millisecond)
			updateTpsBenchmark(tpsBenchmark, txCount, uint64(nonce))
			wg.Done()
		}(i)
	}
	wg.Wait()

	bigTxCount := big.NewInt(int64(txCount))
	bigTxCount.Mul(bigTxCount, big.NewInt(int64(nrGoroutines)))
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

func TestTpsBenchmark_ShardStatisticConcurrentAccess(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("code should not panic")
		}
	}()

	tpsBenchmark, _ := statistics.NewTPSBenchmark(12, 4)

	wg := sync.WaitGroup{}
	wg.Add(200)
	for i := uint64(0); i < 100; i++ {
		go func(nonce uint64) {
			time.Sleep(time.Millisecond)
			metaBlock := &block.MetaBlock{
				Nonce:   1,
				Round:   2,
				TxCount: 0,
				ShardInfo: []block.ShardData{
					{Nonce: nonce, ShardID: 0},
					{Nonce: nonce, ShardID: 1},
					{Nonce: nonce, ShardID: 2},
					{Nonce: nonce, ShardID: 3},
					{Nonce: nonce, ShardID: 4},
					{Nonce: nonce, ShardID: 5},
					{Nonce: nonce, ShardID: 6},
					{Nonce: nonce, ShardID: 7},
					{Nonce: nonce, ShardID: 8},
					{Nonce: nonce, ShardID: 9},
					{Nonce: nonce, ShardID: 10},
					{Nonce: nonce, ShardID: 11},
				},
			}
			tpsBenchmark.Update(metaBlock)
			wg.Done()
		}(i)
		go func() {
			for shardID, stats := range tpsBenchmark.ShardStatistics() {
				time.Sleep(time.Millisecond)
				_, _ = shardID, stats
			}

			wg.Done()
		}()
	}
	wg.Wait()
}
