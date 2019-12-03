package complexStructures

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
)

var log = logger.GetOrCreate("core/complexStructures")

// ProcessedMiniBlocks is used to store all processed min blocks hashes for e meta block hash
type ProcessedMiniBlocks struct {
	processedMiniBlocks    map[string]map[string]struct{}
	mutProcessedMiniBlocks sync.RWMutex
}

// NewProcessedMiniBlocks will create a complex type of processedMiniBlocks
func NewProcessedMiniBlocks() *ProcessedMiniBlocks {
	return &ProcessedMiniBlocks{
		processedMiniBlocks: make(map[string]map[string]struct{}),
	}
}

// AddMiniBlockHash will add a miniblock hash
func (pmb *ProcessedMiniBlocks) AddMiniBlockHash(metaBlockHash []byte, miniBlockHash []byte) {
	pmb.mutProcessedMiniBlocks.Lock()
	defer pmb.mutProcessedMiniBlocks.Unlock()

	miniBlocksProcessed, ok := pmb.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		miniBlocksProcessed := make(map[string]struct{})
		miniBlocksProcessed[string(miniBlockHash)] = struct{}{}
		pmb.processedMiniBlocks[string(metaBlockHash)] = miniBlocksProcessed

		return
	}

	miniBlocksProcessed[string(miniBlockHash)] = struct{}{}
}

// RemoveMetaBlockHash will remove a meta block hash
func (pmb *ProcessedMiniBlocks) RemoveMetaBlockHash(metaBlockHash []byte) {
	pmb.mutProcessedMiniBlocks.Lock()
	delete(pmb.processedMiniBlocks, string(metaBlockHash))
	pmb.mutProcessedMiniBlocks.Unlock()
}

// RemoveMiniBlockHash will remove a mini block hash
func (pmb *ProcessedMiniBlocks) RemoveMiniBlockHash(miniBlockHash []byte) {
	pmb.mutProcessedMiniBlocks.Lock()
	for metaHash, miniBlocksProcessed := range pmb.processedMiniBlocks {
		_, isProcessed := miniBlocksProcessed[string(miniBlockHash)]
		if isProcessed {
			delete(miniBlocksProcessed, string(miniBlockHash))
		}

		if len(miniBlocksProcessed) == 0 {
			delete(pmb.processedMiniBlocks, metaHash)
		}
	}
	pmb.mutProcessedMiniBlocks.Unlock()
}

// GetProcessedMiniBlocksHashes will return all processed miniblocks for a metablock
func (pmb *ProcessedMiniBlocks) GetProcessedMiniBlocksHashes(metaBlockHash []byte) map[string]struct{} {
	pmb.mutProcessedMiniBlocks.RLock()
	processedMiniBlocksHashes := pmb.processedMiniBlocks[string(metaBlockHash)]
	pmb.mutProcessedMiniBlocks.RUnlock()

	return processedMiniBlocksHashes
}

// IsMiniBlockProcessed will return true if a mini block is processed
func (pmb *ProcessedMiniBlocks) IsMiniBlockProcessed(metaBlockHash []byte, miniBlockHash []byte) bool {
	pmb.mutProcessedMiniBlocks.RLock()
	defer pmb.mutProcessedMiniBlocks.RUnlock()

	miniBlocksProcessed, ok := pmb.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		return false
	}

	_, isProcessed := miniBlocksProcessed[string(miniBlockHash)]
	return isProcessed
}

// ConvertProcessedMiniBlocksMapToSlice will convert a map[string]map[string]struct{} in a slice of MiniBlocksInMeta
func (pmb *ProcessedMiniBlocks) ConvertProcessedMiniBlocksMapToSlice() []bootstrapStorage.MiniBlocksInMeta {
	miniBlocksInMetaBlocks := make([]bootstrapStorage.MiniBlocksInMeta, 0)

	pmb.mutProcessedMiniBlocks.RLock()
	for metaHash, miniBlocksHashes := range pmb.processedMiniBlocks {
		miniBlocksInMeta := bootstrapStorage.MiniBlocksInMeta{MetaHash: []byte(metaHash), MiniBlocksHashes: make([][]byte, 0)}
		for miniBlockHash := range miniBlocksHashes {
			miniBlocksInMeta.MiniBlocksHashes = append(miniBlocksInMeta.MiniBlocksHashes, []byte(miniBlockHash))
		}
		miniBlocksInMetaBlocks = append(miniBlocksInMetaBlocks, miniBlocksInMeta)
	}
	pmb.mutProcessedMiniBlocks.RUnlock()

	return miniBlocksInMetaBlocks
}

// ConvertSliceToProcessedMiniBlocksMap will convert a slice of MiniBlocksInMeta in an map[string]map[string]struct{}
func (pmb *ProcessedMiniBlocks) ConvertSliceToProcessedMiniBlocksMap(miniBlocksInMetaBlocks []bootstrapStorage.MiniBlocksInMeta) {
	pmb.mutProcessedMiniBlocks.Lock()
	for _, miniBlocksInMeta := range miniBlocksInMetaBlocks {
		miniBlocksHashes := make(map[string]struct{})
		for _, miniBlockHash := range miniBlocksInMeta.MiniBlocksHashes {
			miniBlocksHashes[string(miniBlockHash)] = struct{}{}
		}
		pmb.processedMiniBlocks[string(miniBlocksInMeta.MetaHash)] = miniBlocksHashes
	}
	pmb.mutProcessedMiniBlocks.Unlock()
}

// DisplayProcessedMiniBlocks will display all miniblocks hashes and meta block hash from the map
func (pmb *ProcessedMiniBlocks) DisplayProcessedMiniBlocks() {
	log.Debug("processed mini blocks applied")

	pmb.mutProcessedMiniBlocks.RLock()
	for metaBlockHash, miniBlocksHashes := range pmb.processedMiniBlocks {
		log.Debug("processed",
			"meta hash", []byte(metaBlockHash))
		for miniBlockHash := range miniBlocksHashes {
			log.Debug("processed",
				"mini block hash", []byte(miniBlockHash))
		}
	}
	pmb.mutProcessedMiniBlocks.RUnlock()
}
