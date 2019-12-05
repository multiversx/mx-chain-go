package processedMb

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
)

var log = logger.GetOrCreate("core/processedMb")

// ProcessedMiniBlockTracker is used to store all processed min blocks hashes for e meta block hash
type ProcessedMiniBlockTracker struct {
	processedMiniBlocks    map[string]map[string]struct{}
	mutProcessedMiniBlocks sync.RWMutex
}

// NewProcessedMiniBlocks will create a complex type of processedMb
func NewProcessedMiniBlocks() *ProcessedMiniBlockTracker {
	return &ProcessedMiniBlockTracker{
		processedMiniBlocks: make(map[string]map[string]struct{}),
	}
}

// AddMiniBlockHash will add a miniblock hash
func (pmb *ProcessedMiniBlockTracker) AddMiniBlockHash(metaBlockHash string, miniBlockHash string) {
	pmb.mutProcessedMiniBlocks.Lock()
	defer pmb.mutProcessedMiniBlocks.Unlock()

	miniBlocksProcessed, ok := pmb.processedMiniBlocks[metaBlockHash]
	if !ok {
		miniBlocksProcessed := make(map[string]struct{})
		miniBlocksProcessed[miniBlockHash] = struct{}{}
		pmb.processedMiniBlocks[metaBlockHash] = miniBlocksProcessed

		return
	}

	miniBlocksProcessed[miniBlockHash] = struct{}{}
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
		_, isProcessed := miniBlocksProcessed[miniBlockHash]
		if isProcessed {
			delete(miniBlocksProcessed, miniBlockHash)
		}

		if len(miniBlocksProcessed) == 0 {
			delete(pmb.processedMiniBlocks, metaHash)
		}
	}
	pmb.mutProcessedMiniBlocks.Unlock()
}

// GetProcessedMiniBlocksHashes will return all processed miniblocks for a metablock
func (pmb *ProcessedMiniBlockTracker) GetProcessedMiniBlocksHashes(metaBlockHash string) map[string]struct{} {
	pmb.mutProcessedMiniBlocks.RLock()
	processedMiniBlocksHashes := pmb.processedMiniBlocks[metaBlockHash]
	pmb.mutProcessedMiniBlocks.RUnlock()

	return processedMiniBlocksHashes
}

// IsMiniBlockProcessed will return true if a mini block is processed
func (pmb *ProcessedMiniBlockTracker) IsMiniBlockProcessed(metaBlockHash string, miniBlockHash string) bool {
	pmb.mutProcessedMiniBlocks.RLock()
	defer pmb.mutProcessedMiniBlocks.RUnlock()

	miniBlocksProcessed, ok := pmb.processedMiniBlocks[metaBlockHash]
	if !ok {
		return false
	}

	_, isProcessed := miniBlocksProcessed[miniBlockHash]
	return isProcessed
}

// ConvertProcessedMiniBlocksMapToSlice will convert a map[string]map[string]struct{} in a slice of MiniBlocksInMeta
func (pmb *ProcessedMiniBlockTracker) ConvertProcessedMiniBlocksMapToSlice() []bootstrapStorage.MiniBlocksInMeta {
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
func (pmb *ProcessedMiniBlockTracker) ConvertSliceToProcessedMiniBlocksMap(miniBlocksInMetaBlocks []bootstrapStorage.MiniBlocksInMeta) {
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
func (pmb *ProcessedMiniBlockTracker) DisplayProcessedMiniBlocks() {
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
