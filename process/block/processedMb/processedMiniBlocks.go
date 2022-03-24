package processedMb

import (
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
)

var log = logger.GetOrCreate("process/processedMb")

// ProcessedMiniBlockInfo will keep the info about processed mini blocks
type ProcessedMiniBlockInfo struct {
	IsFullyProcessed       bool
	IndexOfLastTxProcessed int32
}

// MiniBlocksInfo will keep a list of miniblocks hashes as keys, with miniblocks info as value
type MiniBlocksInfo map[string]*ProcessedMiniBlockInfo

// ProcessedMiniBlockTracker is used to store all processed mini blocks hashes grouped by a metahash
type ProcessedMiniBlockTracker struct {
	processedMiniBlocks    map[string]MiniBlocksInfo
	mutProcessedMiniBlocks sync.RWMutex
}

// NewProcessedMiniBlocks will create a complex type of processedMb
func NewProcessedMiniBlocks() *ProcessedMiniBlockTracker {
	return &ProcessedMiniBlockTracker{
		processedMiniBlocks: make(map[string]MiniBlocksInfo),
	}
}

// SetProcessedMiniBlockInfo will set a processed miniblock info for the given metablock hash and miniblock hash
func (pmb *ProcessedMiniBlockTracker) SetProcessedMiniBlockInfo(metaBlockHash string, miniBlockHash string, processedMbInfo *ProcessedMiniBlockInfo) {
	pmb.mutProcessedMiniBlocks.Lock()
	defer pmb.mutProcessedMiniBlocks.Unlock()

	miniBlocksProcessed, ok := pmb.processedMiniBlocks[metaBlockHash]
	if !ok {
		miniBlocksProcessed = make(MiniBlocksInfo)
		pmb.processedMiniBlocks[metaBlockHash] = miniBlocksProcessed
	}

	miniBlocksProcessed[miniBlockHash] = &ProcessedMiniBlockInfo{
		IsFullyProcessed:       processedMbInfo.IsFullyProcessed,
		IndexOfLastTxProcessed: processedMbInfo.IndexOfLastTxProcessed,
	}
}

// RemoveMetaBlockHash will remove a meta block hash
func (pmb *ProcessedMiniBlockTracker) RemoveMetaBlockHash(metaBlockHash string) {
	pmb.mutProcessedMiniBlocks.Lock()
	delete(pmb.processedMiniBlocks, metaBlockHash)
	pmb.mutProcessedMiniBlocks.Unlock()
}

// RemoveMiniBlockHash will remove a mini block hash
func (pmb *ProcessedMiniBlockTracker) RemoveMiniBlockHash(miniBlockHash string) {
	pmb.mutProcessedMiniBlocks.Lock()
	for metaHash, miniBlocksProcessed := range pmb.processedMiniBlocks {
		delete(miniBlocksProcessed, miniBlockHash)

		if len(miniBlocksProcessed) == 0 {
			delete(pmb.processedMiniBlocks, metaHash)
		}
	}
	pmb.mutProcessedMiniBlocks.Unlock()
}

// GetProcessedMiniBlocksInfo will return all processed miniblocks info for a metablock
func (pmb *ProcessedMiniBlockTracker) GetProcessedMiniBlocksInfo(metaBlockHash string) map[string]*ProcessedMiniBlockInfo {
	processedMiniBlocksInfo := make(map[string]*ProcessedMiniBlockInfo)

	pmb.mutProcessedMiniBlocks.RLock()
	for miniBlockHash, processedMiniBlockInfo := range pmb.processedMiniBlocks[metaBlockHash] {
		processedMiniBlocksInfo[miniBlockHash] = &ProcessedMiniBlockInfo{
			IsFullyProcessed:       processedMiniBlockInfo.IsFullyProcessed,
			IndexOfLastTxProcessed: processedMiniBlockInfo.IndexOfLastTxProcessed,
		}
	}
	pmb.mutProcessedMiniBlocks.RUnlock()

	return processedMiniBlocksInfo
}

// IsMiniBlockFullyProcessed will return true if a mini block is fully processed
func (pmb *ProcessedMiniBlockTracker) IsMiniBlockFullyProcessed(metaBlockHash string, miniBlockHash string) bool {
	pmb.mutProcessedMiniBlocks.RLock()
	defer pmb.mutProcessedMiniBlocks.RUnlock()

	miniBlocksProcessed, ok := pmb.processedMiniBlocks[metaBlockHash]
	if !ok {
		return false
	}

	processedMbInfo, hashExists := miniBlocksProcessed[miniBlockHash]
	if !hashExists {
		return false
	}

	return processedMbInfo.IsFullyProcessed
}

// ConvertProcessedMiniBlocksMapToSlice will convert a map[string]map[string]struct{} in a slice of MiniBlocksInMeta
func (pmb *ProcessedMiniBlockTracker) ConvertProcessedMiniBlocksMapToSlice() []bootstrapStorage.MiniBlocksInMeta {
	pmb.mutProcessedMiniBlocks.RLock()
	defer pmb.mutProcessedMiniBlocks.RUnlock()

	if len(pmb.processedMiniBlocks) == 0 {
		return nil
	}

	miniBlocksInMetaBlocks := make([]bootstrapStorage.MiniBlocksInMeta, 0, len(pmb.processedMiniBlocks))

	for metaHash, miniBlocksInfo := range pmb.processedMiniBlocks {
		miniBlocksInMeta := bootstrapStorage.MiniBlocksInMeta{
			MetaHash:               []byte(metaHash),
			MiniBlocksHashes:       make([][]byte, 0, len(miniBlocksInfo)),
			IsFullyProcessed:       make([]bool, 0, len(miniBlocksInfo)),
			IndexOfLastTxProcessed: make([]int32, 0, len(miniBlocksInfo)),
		}

		for miniBlockHash, processedMiniBlockInfo := range miniBlocksInfo {
			miniBlocksInMeta.MiniBlocksHashes = append(miniBlocksInMeta.MiniBlocksHashes, []byte(miniBlockHash))
			miniBlocksInMeta.IsFullyProcessed = append(miniBlocksInMeta.IsFullyProcessed, processedMiniBlockInfo.IsFullyProcessed)
			miniBlocksInMeta.IndexOfLastTxProcessed = append(miniBlocksInMeta.IndexOfLastTxProcessed, processedMiniBlockInfo.IndexOfLastTxProcessed)
		}

		miniBlocksInMetaBlocks = append(miniBlocksInMetaBlocks, miniBlocksInMeta)
	}

	return miniBlocksInMetaBlocks
}

// ConvertSliceToProcessedMiniBlocksMap will convert a slice of MiniBlocksInMeta in an map[string]MiniBlockHashes
func (pmb *ProcessedMiniBlockTracker) ConvertSliceToProcessedMiniBlocksMap(miniBlocksInMetaBlocks []bootstrapStorage.MiniBlocksInMeta) {
	pmb.mutProcessedMiniBlocks.Lock()
	defer pmb.mutProcessedMiniBlocks.Unlock()

	for _, miniBlocksInMeta := range miniBlocksInMetaBlocks {
		miniBlocksInfo := make(MiniBlocksInfo)
		for index, miniBlockHash := range miniBlocksInMeta.MiniBlocksHashes {
			isFullyProcessed := true
			if miniBlocksInMeta.IsFullyProcessed != nil && len(miniBlocksInMeta.IsFullyProcessed) > index {
				isFullyProcessed = miniBlocksInMeta.IsFullyProcessed[index]
			}

			//TODO: Check how to set the correct index
			indexOfLastTxProcessed := int32(math.MaxInt32)
			if miniBlocksInMeta.IndexOfLastTxProcessed != nil && len(miniBlocksInMeta.IndexOfLastTxProcessed) > index {
				indexOfLastTxProcessed = miniBlocksInMeta.IndexOfLastTxProcessed[index]
			}

			miniBlocksInfo[string(miniBlockHash)] = &ProcessedMiniBlockInfo{
				IsFullyProcessed:       isFullyProcessed,
				IndexOfLastTxProcessed: indexOfLastTxProcessed,
			}
		}
		pmb.processedMiniBlocks[string(miniBlocksInMeta.MetaHash)] = miniBlocksInfo
	}
}

// DisplayProcessedMiniBlocks will display all miniblocks hashes and meta block hash from the map
func (pmb *ProcessedMiniBlockTracker) DisplayProcessedMiniBlocks() {
	log.Debug("processed mini blocks applied")

	pmb.mutProcessedMiniBlocks.RLock()
	for metaBlockHash, miniBlocksInfo := range pmb.processedMiniBlocks {
		log.Debug("processed",
			"meta hash", []byte(metaBlockHash))
		for miniBlockHash, processedMiniBlockInfo := range miniBlocksInfo {
			log.Debug("processed",
				"mini block hash", []byte(miniBlockHash),
				"is fully processed", processedMiniBlockInfo.IsFullyProcessed,
				"index of last tx processed", processedMiniBlockInfo.IndexOfLastTxProcessed,
			)
		}
	}
	pmb.mutProcessedMiniBlocks.RUnlock()
}
