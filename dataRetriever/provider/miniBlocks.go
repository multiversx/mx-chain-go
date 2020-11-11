package provider

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.MiniBlockProvider = (*miniBlockProvider)(nil)

var log = logger.GetOrCreate("dataretriever/getters")

// ArgMiniBlockProvider represents a miniblock data provider argument
type ArgMiniBlockProvider struct {
	MiniBlockPool    storage.Cacher
	MiniBlockStorage storage.Storer
	Marshalizer      marshal.Marshalizer
}

// miniBlockProvider is able to get data from cache or from storage
type miniBlockProvider struct {
	miniBlockPool    storage.Cacher
	miniBlockStorage storage.Storer
	marshalizer      marshal.Marshalizer
}

// NewMiniBlockProvider creates a new miniBlockDataGetter instance
func NewMiniBlockProvider(arg ArgMiniBlockProvider) (*miniBlockProvider, error) {
	if check.IfNil(arg.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(arg.MiniBlockPool) {
		return nil, dataRetriever.ErrNilMiniblocksPool
	}
	if check.IfNil(arg.MiniBlockStorage) {
		return nil, dataRetriever.ErrNilMiniblocksStorage
	}

	return &miniBlockProvider{
		miniBlockPool:    arg.MiniBlockPool,
		miniBlockStorage: arg.MiniBlockStorage,
		marshalizer:      arg.Marshalizer,
	}, nil
}

// GetMiniBlocks method returns a list of deserialized mini blocks from a given hash list either from data pool or from storage
func (mbp *miniBlockProvider) GetMiniBlocks(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
	miniBlocksAndHashes := make([]*block.MiniblockAndHash, 0)
	missingMiniBlocksHashes := make([][]byte, 0)

	for _, hash := range hashes {
		miniblock := mbp.getMiniblockFromPool(hash)
		if miniblock != nil {
			miniBlockAndHash := &block.MiniblockAndHash{
				Miniblock: miniblock,
				Hash:      hash,
			}
			miniBlocksAndHashes = append(miniBlocksAndHashes, miniBlockAndHash)
			continue
		}

		miniblock = mbp.getMiniBlockFromStorer(hash)
		if miniblock != nil {
			miniBlockAndHash := &block.MiniblockAndHash{
				Miniblock: miniblock,
				Hash:      hash,
			}
			miniBlocksAndHashes = append(miniBlocksAndHashes, miniBlockAndHash)
			continue
		}

		missingMiniBlocksHashes = append(missingMiniBlocksHashes, hash)
	}

	return miniBlocksAndHashes, missingMiniBlocksHashes
}

// GetMiniBlocksFromPool method returns a list of deserialized mini blocks from a given hash list from data pool
func (mbp *miniBlockProvider) GetMiniBlocksFromPool(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
	miniBlocksAndHashes := make([]*block.MiniblockAndHash, 0)
	missingMiniBlocksHashes := make([][]byte, 0)

	for i := 0; i < len(hashes); i++ {
		miniblock := mbp.getMiniblockFromPool(hashes[i])
		if miniblock == nil {
			missingMiniBlocksHashes = append(missingMiniBlocksHashes, hashes[i])
			continue
		}

		miniBlockAndHash := &block.MiniblockAndHash{
			Miniblock: miniblock,
			Hash:      hashes[i],
		}
		miniBlocksAndHashes = append(miniBlocksAndHashes, miniBlockAndHash)
	}

	return miniBlocksAndHashes, missingMiniBlocksHashes
}

func (mbp *miniBlockProvider) getMiniblockFromPool(hash []byte) *block.MiniBlock {
	obj, ok := mbp.miniBlockPool.Peek(hash)
	if !ok {
		log.Trace("missing miniblock in cache",
			"hash", hash,
		)
		return nil
	}

	miniBlock, ok := obj.(*block.MiniBlock)
	if !ok {
		log.Warn("miniBlockProvider.getMiniblockFromPool does not contain a miniblock instance",
			"hash", hash,
		)
	}

	return miniBlock
}

// getMiniBlocksFromStorer returns a list of mini blocks from storage and a list of missing hashes
func (mbp *miniBlockProvider) getMiniBlockFromStorer(hash []byte) *block.MiniBlock {
	buff, err := mbp.miniBlockStorage.Get(hash)
	if err != nil {
		log.Trace("miniBlockProvider.getMiniBlockFromStorer missing miniblock in storer",
			"error", err.Error(),
			"hash", hash,
		)
		return nil
	}

	miniBlock := &block.MiniBlock{}
	err = mbp.marshalizer.Unmarshal(miniBlock, buff)
	if err != nil {
		log.Warn("miniBlockProvider.getMiniBlockFromStorer marshal error", "error", err.Error())

		return nil
	}

	return miniBlock
}

// IsInterfaceNil returns true if there is no value under the interface
func (mbp *miniBlockProvider) IsInterfaceNil() bool {
	return mbp == nil
}
