package processedMb_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/stretchr/testify/assert"
)

func TestProcessedMiniBlocks_SetProcessedMiniBlockInfoShouldWork(t *testing.T) {
	t.Parallel()

	pmbt := processedMb.NewProcessedMiniBlocksTracker()

	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")
	mtbHash1 := []byte("meta1")
	mtbHash2 := []byte("meta2")

	pmbt.SetProcessedMiniBlockInfo(mtbHash1, mbHash1, &processedMb.ProcessedMiniBlockInfo{FullyProcessed: true})
	assert.True(t, pmbt.IsMiniBlockFullyProcessed(mtbHash1, mbHash1))

	pmbt.SetProcessedMiniBlockInfo(mtbHash2, mbHash1, &processedMb.ProcessedMiniBlockInfo{FullyProcessed: true})
	assert.True(t, pmbt.IsMiniBlockFullyProcessed(mtbHash2, mbHash1))

	pmbt.SetProcessedMiniBlockInfo(mtbHash1, mbHash2, &processedMb.ProcessedMiniBlockInfo{FullyProcessed: true})
	assert.True(t, pmbt.IsMiniBlockFullyProcessed(mtbHash1, mbHash2))

	pmbt.RemoveMiniBlockHash(mbHash1)
	assert.False(t, pmbt.IsMiniBlockFullyProcessed(mtbHash1, mbHash1))

	pmbt.RemoveMiniBlockHash(mbHash1)
	assert.False(t, pmbt.IsMiniBlockFullyProcessed(mtbHash1, mbHash1))

	pmbt.RemoveMetaBlockHash(mtbHash2)
	assert.False(t, pmbt.IsMiniBlockFullyProcessed(mtbHash2, mbHash1))
}

func TestProcessedMiniBlocks_GetProcessedMiniBlocksInfo(t *testing.T) {
	t.Parallel()

	pmbt := processedMb.NewProcessedMiniBlocksTracker()

	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")
	mtbHash1 := []byte("meta1")
	mtbHash2 := []byte("meta2")

	pmbt.SetProcessedMiniBlockInfo(mtbHash1, mbHash1, &processedMb.ProcessedMiniBlockInfo{FullyProcessed: true})
	pmbt.SetProcessedMiniBlockInfo(mtbHash1, mbHash2, &processedMb.ProcessedMiniBlockInfo{FullyProcessed: true})
	pmbt.SetProcessedMiniBlockInfo(mtbHash2, mbHash2, &processedMb.ProcessedMiniBlockInfo{FullyProcessed: true})

	mapData := pmbt.GetProcessedMiniBlocksInfo(mtbHash1)
	assert.NotNil(t, mapData[string(mbHash1)])
	assert.NotNil(t, mapData[string(mbHash2)])

	mapData = pmbt.GetProcessedMiniBlocksInfo(mtbHash2)
	assert.NotNil(t, mapData[string(mbHash2)])
}

func TestProcessedMiniBlocks_ConvertSliceToProcessedMiniBlocksMap(t *testing.T) {
	t.Parallel()

	pmbt := processedMb.NewProcessedMiniBlocksTracker()

	mbHash1 := []byte("hash1")
	mtbHash1 := []byte("meta1")

	data1 := bootstrapStorage.MiniBlocksInMeta{
		MetaHash:               mtbHash1,
		MiniBlocksHashes:       [][]byte{mbHash1},
		FullyProcessed:         []bool{true},
		IndexOfLastTxProcessed: []int32{69},
	}

	miniBlocksInMeta := []bootstrapStorage.MiniBlocksInMeta{data1}
	pmbt.ConvertSliceToProcessedMiniBlocksMap(miniBlocksInMeta)
	assert.True(t, pmbt.IsMiniBlockFullyProcessed(mtbHash1, mbHash1))

	convertedData := pmbt.ConvertProcessedMiniBlocksMapToSlice()
	assert.Equal(t, miniBlocksInMeta, convertedData)
}

func TestProcessedMiniBlocks_GetProcessedMiniBlockInfo(t *testing.T) {
	t.Parallel()

	mbHash := []byte("mb_hash")
	metaHash := []byte("meta_hash")
	processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
		FullyProcessed:         true,
		IndexOfLastTxProcessed: 69,
	}
	pmbt := processedMb.NewProcessedMiniBlocksTracker()
	pmbt.SetProcessedMiniBlockInfo(metaHash, mbHash, processedMbInfo)

	processedMiniBlockInfo, processedMetaHash := pmbt.GetProcessedMiniBlockInfo(nil)
	assert.Nil(t, processedMetaHash)
	assert.False(t, processedMiniBlockInfo.FullyProcessed)
	assert.Equal(t, int32(-1), processedMiniBlockInfo.IndexOfLastTxProcessed)

	processedMiniBlockInfo, processedMetaHash = pmbt.GetProcessedMiniBlockInfo(mbHash)
	assert.Equal(t, metaHash, processedMetaHash)
	assert.Equal(t, processedMbInfo.FullyProcessed, processedMiniBlockInfo.FullyProcessed)
	assert.Equal(t, processedMbInfo.IndexOfLastTxProcessed, processedMiniBlockInfo.IndexOfLastTxProcessed)
}
