package complexStructures

import (
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessedMiniBlocks_AddMiniBlockHashShouldWork(t *testing.T) {
	t.Parallel()

	pmb := NewProcessedMiniBlocks()

	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")
	mtbHash1 := []byte("meta1")
	mtbHash2 := []byte("meta2")

	pmb.AddMiniBlockHash(mtbHash1, mbHash1)
	assert.True(t, pmb.IsMiniBlockProcessed(mtbHash1, mbHash1))

	pmb.AddMiniBlockHash(mtbHash2, mbHash1)
	assert.True(t, pmb.IsMiniBlockProcessed(mtbHash2, mbHash1))

	pmb.AddMiniBlockHash(mtbHash1, mbHash2)
	assert.True(t, pmb.IsMiniBlockProcessed(mtbHash1, mbHash2))

	pmb.RemoveMiniBlockHash(mbHash1)
	assert.False(t, pmb.IsMiniBlockProcessed(mtbHash1, mbHash1))

	pmb.RemoveMiniBlockHash(mbHash1)
	assert.False(t, pmb.IsMiniBlockProcessed(mtbHash1, mbHash1))

	pmb.RemoveMetaBlockHash(mtbHash2)
	assert.False(t, pmb.IsMiniBlockProcessed(mtbHash2, mbHash1))
}

func TestProcessedMiniBlocks_GetProcessedMiniBlocksHashes(t *testing.T) {
	t.Parallel()

	pmb := NewProcessedMiniBlocks()

	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")
	mtbHash1 := []byte("meta1")
	mtbHash2 := []byte("meta2")

	pmb.AddMiniBlockHash(mtbHash1, mbHash1)
	pmb.AddMiniBlockHash(mtbHash1, mbHash2)
	pmb.AddMiniBlockHash(mtbHash2, mbHash2)

	mapData := pmb.GetProcessedMiniBlocksHashes(mtbHash1)
	assert.NotNil(t, mapData[string(mbHash1)])
	assert.NotNil(t, mapData[string(mbHash2)])

	mapData = pmb.GetProcessedMiniBlocksHashes(mtbHash2)
	assert.NotNil(t, mapData[string(mbHash1)])
}

func TestProcessedMiniBlocks_ConvertSliceToProcessedMiniBlocksMap(t *testing.T) {
	t.Parallel()

	pmb := NewProcessedMiniBlocks()

	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")
	mtbHash1 := []byte("meta1")
	mtbHash2 := []byte("meta2")

	data1 := bootstrapStorage.MiniBlocksInMeta{
		MetaHash:         mtbHash1,
		MiniBlocksHashes: [][]byte{mbHash1, mbHash2},
	}

	data2 := bootstrapStorage.MiniBlocksInMeta{
		MetaHash:         mtbHash2,
		MiniBlocksHashes: [][]byte{mbHash1, mbHash2},
	}

	miniBlocksInMeta := []bootstrapStorage.MiniBlocksInMeta{data1, data2}
	pmb.ConvertSliceToProcessedMiniBlocksMap(miniBlocksInMeta)
	assert.True(t, pmb.IsMiniBlockProcessed(mtbHash1, mbHash1))
	assert.True(t, pmb.IsMiniBlockProcessed(mtbHash1, mbHash2))
	assert.True(t, pmb.IsMiniBlockProcessed(mtbHash2, mbHash1))
	assert.True(t, pmb.IsMiniBlockProcessed(mtbHash2, mbHash2))

	convertedData := pmb.ConvertProcessedMiniBlocksMapToSlice()
	assert.Equal(t, miniBlocksInMeta, convertedData)
}
