package getters

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("dataretriever/getters")

// ArgMiniBlockDataGetter represents a miniblock data getter
type ArgMiniBlockDataGetter struct {
	MiniBlockPool    storage.Cacher
	MiniBlockStorage storage.Storer
	Marshalizer      marshal.Marshalizer
}

// miniBlockDataGetter is able to get data from cache or from storage
type miniBlockDataGetter struct {
	miniBlockPool    storage.Cacher
	miniBlockStorage storage.Storer
	marshalizer      marshal.Marshalizer
}

// NewMiniBlockDataGetter creates a new miniBlockDataGetter instance
func NewMiniBlockDataGetter(arg ArgMiniBlockDataGetter) (*miniBlockDataGetter, error) {
	if check.IfNil(arg.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(arg.MiniBlockPool) {
		return nil, dataRetriever.ErrNilMiniblocksPool
	}
	if check.IfNil(arg.MiniBlockStorage) {
		return nil, dataRetriever.ErrNilMiniblocksStorage
	}

	return &miniBlockDataGetter{
		miniBlockPool:    arg.MiniBlockPool,
		miniBlockStorage: arg.MiniBlockStorage,
		marshalizer:      arg.Marshalizer,
	}, nil
}

// GetMiniBlocks method returns a list of deserialized mini blocks from a given hash list either from data pool or from storage
func (mbdg *miniBlockDataGetter) GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	miniBlocks, missingMiniBlocksHashes := mbdg.GetMiniBlocksFromPool(hashes)
	if len(missingMiniBlocksHashes) == 0 {
		return miniBlocks, missingMiniBlocksHashes
	}

	miniBlocksFromStorer, missingMiniBlocksHashes := mbdg.getMiniBlocksFromStorer(missingMiniBlocksHashes)
	miniBlocks = append(miniBlocks, miniBlocksFromStorer...)

	return miniBlocks, missingMiniBlocksHashes
}

// GetMiniBlocksFromPool method returns a list of deserialized mini blocks from a given hash list from data pool
func (mbdg *miniBlockDataGetter) GetMiniBlocksFromPool(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	miniBlocks := make(block.MiniBlockSlice, 0)
	missingMiniBlocksHashes := make([][]byte, 0)

	for i := 0; i < len(hashes); i++ {
		obj, ok := mbdg.miniBlockPool.Peek(hashes[i])
		if !ok {
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlock, ok := obj.(*block.MiniBlock)
		if !ok {
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return miniBlocks, missingMiniBlocksHashes
}

// getMiniBlocksFromStorer returns a list of mini blocks from storage and a list of missing hashes
func (mbdg *miniBlockDataGetter) getMiniBlocksFromStorer(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	miniBlocks := make(block.MiniBlockSlice, 0)
	missingMiniBlocksHashes := make([][]byte, 0)

	for i := 0; i < len(hashes); i++ {
		buff, err := mbdg.miniBlockStorage.SearchFirst(hashes[i])
		if err != nil {
			log.Trace("missing miniblock",
				"error", err.Error(),
				"hash", hashes[i])
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlock := &block.MiniBlock{}
		err = mbdg.marshalizer.Unmarshal(miniBlock, buff)
		if err != nil {
			log.Warn("marshal error", "error", err.Error())

			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return miniBlocks, missingMiniBlocksHashes
}

// IsInterfaceNil returns true if there is no value under the interface
func (mbdg *miniBlockDataGetter) IsInterfaceNil() bool {
	return mbdg == nil
}
