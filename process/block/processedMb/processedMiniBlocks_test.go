package processedMb_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
	"github.com/stretchr/testify/assert"
)

func TestProcessedMiniBlocks_SetProcessedMiniBlockInfoShouldWork(t *testing.T) {
	t.Parallel()

	pmb := processedMb.NewProcessedMiniBlocks()

	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")
	mtbHash1 := []byte("meta1")
	mtbHash2 := []byte("meta2")

	pmb.SetProcessedMiniBlockInfo(mtbHash1, mbHash1, &processedMb.ProcessedMiniBlockInfo{IsFullyProcessed: true})
	assert.True(t, pmb.IsMiniBlockFullyProcessed(mtbHash1, mbHash1))

	pmb.SetProcessedMiniBlockInfo(mtbHash2, mbHash1, &processedMb.ProcessedMiniBlockInfo{IsFullyProcessed: true})
	assert.True(t, pmb.IsMiniBlockFullyProcessed(mtbHash2, mbHash1))

	pmb.SetProcessedMiniBlockInfo(mtbHash1, mbHash2, &processedMb.ProcessedMiniBlockInfo{IsFullyProcessed: true})
	assert.True(t, pmb.IsMiniBlockFullyProcessed(mtbHash1, mbHash2))

	pmb.RemoveMiniBlockHash(mbHash1)
	assert.False(t, pmb.IsMiniBlockFullyProcessed(mtbHash1, mbHash1))

	pmb.RemoveMiniBlockHash(mbHash1)
	assert.False(t, pmb.IsMiniBlockFullyProcessed(mtbHash1, mbHash1))

	pmb.RemoveMetaBlockHash(mtbHash2)
	assert.False(t, pmb.IsMiniBlockFullyProcessed(mtbHash2, mbHash1))
}

func TestProcessedMiniBlocks_GetProcessedMiniBlocksInfo(t *testing.T) {
	t.Parallel()

	pmb := processedMb.NewProcessedMiniBlocks()

	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")
	mtbHash1 := []byte("meta1")
	mtbHash2 := []byte("meta2")

	pmb.SetProcessedMiniBlockInfo(mtbHash1, mbHash1, &processedMb.ProcessedMiniBlockInfo{IsFullyProcessed: true})
	pmb.SetProcessedMiniBlockInfo(mtbHash1, mbHash2, &processedMb.ProcessedMiniBlockInfo{IsFullyProcessed: true})
	pmb.SetProcessedMiniBlockInfo(mtbHash2, mbHash2, &processedMb.ProcessedMiniBlockInfo{IsFullyProcessed: true})

	mapData := pmb.GetProcessedMiniBlocksInfo(mtbHash1)
	assert.NotNil(t, mapData[string(mbHash1)])
	assert.NotNil(t, mapData[string(mbHash2)])

	mapData = pmb.GetProcessedMiniBlocksInfo(mtbHash2)
	assert.NotNil(t, mapData[string(mbHash2)])
}

func TestProcessedMiniBlocks_ConvertSliceToProcessedMiniBlocksMap(t *testing.T) {
	t.Parallel()

	pmb := processedMb.NewProcessedMiniBlocks()

	mbHash1 := []byte("hash1")
	mtbHash1 := []byte("meta1")

	data1 := bootstrapStorage.MiniBlocksInMeta{
		MetaHash:               mtbHash1,
		MiniBlocksHashes:       [][]byte{mbHash1},
		IsFullyProcessed:       []bool{true},
		IndexOfLastTxProcessed: []int32{69},
	}

	miniBlocksInMeta := []bootstrapStorage.MiniBlocksInMeta{data1}
	pmb.ConvertSliceToProcessedMiniBlocksMap(miniBlocksInMeta)
	assert.True(t, pmb.IsMiniBlockFullyProcessed(mtbHash1, mbHash1))

	convertedData := pmb.ConvertProcessedMiniBlocksMapToSlice()
	assert.Equal(t, miniBlocksInMeta, convertedData)
}
