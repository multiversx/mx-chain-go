package processedMb

import (
	"sync"

	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/processedMb")

// ProcessedMiniBlockInfo will keep the info about a processed mini block
type ProcessedMiniBlockInfo struct {
	FullyProcessed         bool
	IndexOfLastTxProcessed int32
}

// miniBlocksInfo will keep a list of mini blocks hashes as keys, with mini blocks info as value
type miniBlocksInfo map[string]*ProcessedMiniBlockInfo

// processedMiniBlocksTracker is used to store all processed mini blocks hashes grouped by a meta hash
type processedMiniBlocksTracker struct {
	processedMiniBlocks    map[string]miniBlocksInfo
	mutProcessedMiniBlocks sync.RWMutex
}

// NewProcessedMiniBlocksTracker will create a processed mini blocks tracker object
func NewProcessedMiniBlocksTracker() *processedMiniBlocksTracker {
	return &processedMiniBlocksTracker{
		processedMiniBlocks: make(map[string]miniBlocksInfo),
	}
}

// SetProcessedMiniBlockInfo will set a processed mini block info for the given meta block hash and mini block hash
func (pmbt *processedMiniBlocksTracker) SetProcessedMiniBlockInfo(metaBlockHash []byte, miniBlockHash []byte, processedMbInfo *ProcessedMiniBlockInfo) {
	pmbt.mutProcessedMiniBlocks.Lock()
	defer pmbt.mutProcessedMiniBlocks.Unlock()

	miniBlocksProcessed, ok := pmbt.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		miniBlocksProcessed = make(miniBlocksInfo)
		pmbt.processedMiniBlocks[string(metaBlockHash)] = miniBlocksProcessed
	}

	miniBlocksProcessed[string(miniBlockHash)] = &ProcessedMiniBlockInfo{
		FullyProcessed:         processedMbInfo.FullyProcessed,
		IndexOfLastTxProcessed: processedMbInfo.IndexOfLastTxProcessed,
	}
}

// RemoveHeaderHash will remove a header hash
func (pmbt *processedMiniBlocksTracker) RemoveHeaderHash(metaBlockHash []byte) {
	pmbt.mutProcessedMiniBlocks.Lock()
	defer pmbt.mutProcessedMiniBlocks.Unlock()

	delete(pmbt.processedMiniBlocks, string(metaBlockHash))
}

// RemoveMiniBlockHash will remove a mini block hash
func (pmbt *processedMiniBlocksTracker) RemoveMiniBlockHash(miniBlockHash []byte) {
	pmbt.mutProcessedMiniBlocks.Lock()
	defer pmbt.mutProcessedMiniBlocks.Unlock()

	for metaHash, miniBlocksProcessed := range pmbt.processedMiniBlocks {
		delete(miniBlocksProcessed, string(miniBlockHash))

		if len(miniBlocksProcessed) == 0 {
			delete(pmbt.processedMiniBlocks, metaHash)
		}
	}
}

// GetProcessedMiniBlocksInfo will return all processed mini blocks info for a meta block hash
func (pmbt *processedMiniBlocksTracker) GetProcessedMiniBlocksInfo(metaBlockHash []byte) map[string]*ProcessedMiniBlockInfo {
	pmbt.mutProcessedMiniBlocks.RLock()
	defer pmbt.mutProcessedMiniBlocks.RUnlock()

	processedMiniBlocksInfo := make(map[string]*ProcessedMiniBlockInfo)
	for miniBlockHash, processedMiniBlockInfo := range pmbt.processedMiniBlocks[string(metaBlockHash)] {
		processedMiniBlocksInfo[miniBlockHash] = &ProcessedMiniBlockInfo{
			FullyProcessed:         processedMiniBlockInfo.FullyProcessed,
			IndexOfLastTxProcessed: processedMiniBlockInfo.IndexOfLastTxProcessed,
		}
	}

	return processedMiniBlocksInfo
}

// GetProcessedMiniBlockInfo will return processed mini block info for a mini block hash
func (pmbt *processedMiniBlocksTracker) GetProcessedMiniBlockInfo(miniBlockHash []byte) (*ProcessedMiniBlockInfo, []byte) {
	pmbt.mutProcessedMiniBlocks.RLock()
	defer pmbt.mutProcessedMiniBlocks.RUnlock()

	for metaBlockHash, miniBlocksInfo := range pmbt.processedMiniBlocks {
		processedMiniBlockInfo, hashExists := miniBlocksInfo[string(miniBlockHash)]
		if !hashExists {
			continue
		}

		return &ProcessedMiniBlockInfo{
			FullyProcessed:         processedMiniBlockInfo.FullyProcessed,
			IndexOfLastTxProcessed: processedMiniBlockInfo.IndexOfLastTxProcessed,
		}, []byte(metaBlockHash)
	}

	return &ProcessedMiniBlockInfo{
		FullyProcessed:         false,
		IndexOfLastTxProcessed: -1,
	}, nil
}

// IsMiniBlockFullyProcessed will return true if a mini block is fully processed
func (pmbt *processedMiniBlocksTracker) IsMiniBlockFullyProcessed(metaBlockHash []byte, miniBlockHash []byte) bool {
	pmbt.mutProcessedMiniBlocks.RLock()
	defer pmbt.mutProcessedMiniBlocks.RUnlock()

	miniBlocksProcessed, ok := pmbt.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		return false
	}

	processedMbInfo, hashExists := miniBlocksProcessed[string(miniBlockHash)]
	if !hashExists {
		return false
	}

	return processedMbInfo.FullyProcessed
}

// ConvertProcessedMiniBlocksMapToSlice will convert a map[string]map[string]struct{} in a slice of MiniBlocksInMeta
func (pmbt *processedMiniBlocksTracker) ConvertProcessedMiniBlocksMapToSlice() []bootstrapStorage.MiniBlocksInMeta {
	pmbt.mutProcessedMiniBlocks.RLock()
	defer pmbt.mutProcessedMiniBlocks.RUnlock()

	if len(pmbt.processedMiniBlocks) == 0 {
		return nil
	}

	miniBlocksInMetaBlocks := make([]bootstrapStorage.MiniBlocksInMeta, 0, len(pmbt.processedMiniBlocks))

	for metaHash, miniBlocksInfo := range pmbt.processedMiniBlocks {
		miniBlocksInMeta := bootstrapStorage.MiniBlocksInMeta{
			MetaHash:               []byte(metaHash),
			MiniBlocksHashes:       make([][]byte, 0, len(miniBlocksInfo)),
			FullyProcessed:         make([]bool, 0, len(miniBlocksInfo)),
			IndexOfLastTxProcessed: make([]int32, 0, len(miniBlocksInfo)),
		}

		for miniBlockHash, processedMiniBlockInfo := range miniBlocksInfo {
			miniBlocksInMeta.MiniBlocksHashes = append(miniBlocksInMeta.MiniBlocksHashes, []byte(miniBlockHash))
			miniBlocksInMeta.FullyProcessed = append(miniBlocksInMeta.FullyProcessed, processedMiniBlockInfo.FullyProcessed)
			miniBlocksInMeta.IndexOfLastTxProcessed = append(miniBlocksInMeta.IndexOfLastTxProcessed, processedMiniBlockInfo.IndexOfLastTxProcessed)
		}

		miniBlocksInMetaBlocks = append(miniBlocksInMetaBlocks, miniBlocksInMeta)
	}

	return miniBlocksInMetaBlocks
}

// ConvertSliceToProcessedMiniBlocksMap will convert a slice of MiniBlocksInMeta in a map[string]MiniBlockHashes
func (pmbt *processedMiniBlocksTracker) ConvertSliceToProcessedMiniBlocksMap(miniBlocksInMetaBlocks []bootstrapStorage.MiniBlocksInMeta) {
	pmbt.mutProcessedMiniBlocks.Lock()
	defer pmbt.mutProcessedMiniBlocks.Unlock()

	for _, miniBlocksInMeta := range miniBlocksInMetaBlocks {
		pmbt.processedMiniBlocks[string(miniBlocksInMeta.MetaHash)] = getMiniBlocksInfo(miniBlocksInMeta)
	}
}

func getMiniBlocksInfo(miniBlocksInMeta bootstrapStorage.MiniBlocksInMeta) miniBlocksInfo {
	mbsInfo := make(miniBlocksInfo)

	for index, miniBlockHash := range miniBlocksInMeta.MiniBlocksHashes {
		fullyProcessed := miniBlocksInMeta.IsFullyProcessed(index)
		indexOfLastTxProcessed := miniBlocksInMeta.GetIndexOfLastTxProcessedInMiniBlock(index)

		mbsInfo[string(miniBlockHash)] = &ProcessedMiniBlockInfo{
			FullyProcessed:         fullyProcessed,
			IndexOfLastTxProcessed: indexOfLastTxProcessed,
		}
	}

	return mbsInfo
}

// DisplayProcessedMiniBlocks will display all mini blocks hashes and meta block hash from the map
func (pmbt *processedMiniBlocksTracker) DisplayProcessedMiniBlocks() {
	pmbt.mutProcessedMiniBlocks.RLock()
	defer pmbt.mutProcessedMiniBlocks.RUnlock()

	log.Debug("processed mini blocks applied")
	for metaBlockHash, miniBlocksInfo := range pmbt.processedMiniBlocks {
		log.Debug("processed",
			"meta hash", []byte(metaBlockHash))
		for miniBlockHash, processedMiniBlockInfo := range miniBlocksInfo {
			log.Debug("processed",
				"mini block hash", []byte(miniBlockHash),
				"index of last tx processed", processedMiniBlockInfo.IndexOfLastTxProcessed,
				"fully processed", processedMiniBlockInfo.FullyProcessed,
			)
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (pmbt *processedMiniBlocksTracker) IsInterfaceNil() bool {
	return pmbt == nil
}
