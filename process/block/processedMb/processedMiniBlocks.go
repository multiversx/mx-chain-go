package processedMb

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
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
func (pmb *ProcessedMiniBlockTracker) SetProcessedMiniBlockInfo(metaBlockHash []byte, miniBlockHash []byte, processedMbInfo *ProcessedMiniBlockInfo) {
	pmb.mutProcessedMiniBlocks.Lock()
	defer pmb.mutProcessedMiniBlocks.Unlock()

	miniBlocksProcessed, ok := pmb.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		miniBlocksProcessed = make(MiniBlocksInfo)
		pmb.processedMiniBlocks[string(metaBlockHash)] = miniBlocksProcessed
	}

	miniBlocksProcessed[string(miniBlockHash)] = &ProcessedMiniBlockInfo{
		IsFullyProcessed:       processedMbInfo.IsFullyProcessed,
		IndexOfLastTxProcessed: processedMbInfo.IndexOfLastTxProcessed,
	}
}

// RemoveMetaBlockHash will remove a meta block hash
func (pmb *ProcessedMiniBlockTracker) RemoveMetaBlockHash(metaBlockHash []byte) {
	pmb.mutProcessedMiniBlocks.Lock()
	defer pmb.mutProcessedMiniBlocks.Unlock()

	delete(pmb.processedMiniBlocks, string(metaBlockHash))
}

// RemoveMiniBlockHash will remove a mini block hash
func (pmb *ProcessedMiniBlockTracker) RemoveMiniBlockHash(miniBlockHash []byte) {
	pmb.mutProcessedMiniBlocks.Lock()
	defer pmb.mutProcessedMiniBlocks.Unlock()

	for metaHash, miniBlocksProcessed := range pmb.processedMiniBlocks {
		delete(miniBlocksProcessed, string(miniBlockHash))

		if len(miniBlocksProcessed) == 0 {
			delete(pmb.processedMiniBlocks, metaHash)
		}
	}
}

// GetProcessedMiniBlocksInfo will return all processed miniblocks info for a metablock
func (pmb *ProcessedMiniBlockTracker) GetProcessedMiniBlocksInfo(metaBlockHash []byte) map[string]*ProcessedMiniBlockInfo {
	pmb.mutProcessedMiniBlocks.RLock()
	defer pmb.mutProcessedMiniBlocks.RUnlock()

	processedMiniBlocksInfo := make(map[string]*ProcessedMiniBlockInfo)
	for miniBlockHash, processedMiniBlockInfo := range pmb.processedMiniBlocks[string(metaBlockHash)] {
		processedMiniBlocksInfo[miniBlockHash] = &ProcessedMiniBlockInfo{
			IsFullyProcessed:       processedMiniBlockInfo.IsFullyProcessed,
			IndexOfLastTxProcessed: processedMiniBlockInfo.IndexOfLastTxProcessed,
		}
	}

	return processedMiniBlocksInfo
}

// GetProcessedMiniBlockInfo will return all processed info for a miniblock
func (pmb *ProcessedMiniBlockTracker) GetProcessedMiniBlockInfo(miniBlockHash []byte) (*ProcessedMiniBlockInfo, []byte) {
	pmb.mutProcessedMiniBlocks.RLock()
	defer pmb.mutProcessedMiniBlocks.RUnlock()

	for metaBlockHash, miniBlocksInfo := range pmb.processedMiniBlocks {
		processedMiniBlockInfo, hashExists := miniBlocksInfo[string(miniBlockHash)]
		if !hashExists {
			continue
		}

		return &ProcessedMiniBlockInfo{
			IsFullyProcessed:       processedMiniBlockInfo.IsFullyProcessed,
			IndexOfLastTxProcessed: processedMiniBlockInfo.IndexOfLastTxProcessed,
		}, []byte(metaBlockHash)
	}

	return &ProcessedMiniBlockInfo{
		IsFullyProcessed:       false,
		IndexOfLastTxProcessed: -1,
	}, nil
}

// IsMiniBlockFullyProcessed will return true if a mini block is fully processed
func (pmb *ProcessedMiniBlockTracker) IsMiniBlockFullyProcessed(metaBlockHash []byte, miniBlockHash []byte) bool {
	pmb.mutProcessedMiniBlocks.RLock()
	defer pmb.mutProcessedMiniBlocks.RUnlock()

	miniBlocksProcessed, ok := pmb.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		return false
	}

	processedMbInfo, hashExists := miniBlocksProcessed[string(miniBlockHash)]
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
			if miniBlocksInMeta.IsFullyProcessed != nil && index < len(miniBlocksInMeta.IsFullyProcessed) {
				isFullyProcessed = miniBlocksInMeta.IsFullyProcessed[index]
			}

			//TODO: Check if needed, how to set the real index (metaBlock -> ShardInfo -> ShardMiniBlockHeaders -> TxCount)
			indexOfLastTxProcessed := common.MaxIndexOfTxInMiniBlock
			if miniBlocksInMeta.IndexOfLastTxProcessed != nil && index < len(miniBlocksInMeta.IndexOfLastTxProcessed) {
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
	pmb.mutProcessedMiniBlocks.RLock()
	defer pmb.mutProcessedMiniBlocks.RUnlock()

	log.Debug("processed mini blocks applied")
	for metaBlockHash, miniBlocksInfo := range pmb.processedMiniBlocks {
		log.Debug("processed",
			"meta hash", []byte(metaBlockHash))
		for miniBlockHash, processedMiniBlockInfo := range miniBlocksInfo {
			log.Debug("processed",
				"mini block hash", []byte(miniBlockHash),
				"index of last tx processed", processedMiniBlockInfo.IndexOfLastTxProcessed,
				"is fully processed", processedMiniBlockInfo.IsFullyProcessed,
			)
		}
	}
}
