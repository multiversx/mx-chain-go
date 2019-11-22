package storageBootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type shardStorageBootstrapper struct {
	*storageBootstrapper
	miniBlocksResolver dataRetriever.MiniBlocksResolver
}

// NewShardStorageBootstrapper is method used to create a nes storage bootstrapper
func NewShardStorageBootstrapper(arguments ArgsStorageBootstrapper) (*shardStorageBootstrapper, error) {
	base := &storageBootstrapper{
		bootStorer:       arguments.BootStorer,
		forkDetector:     arguments.ForkDetector,
		blkExecutor:      arguments.BlockProcessor,
		blkc:             arguments.ChainHandler,
		marshalizer:      arguments.Marshalizer,
		store:            arguments.Store,
		shardCoordinator: arguments.ShardCoordinator,

		uint64Converter:     arguments.Uint64Converter,
		bootstrapRoundIndex: arguments.BootstrapRoundIndex,
	}

	miniBlocksResolver, err := arguments.ResolversFinder.IntraShardResolver(factory.MiniBlocksTopic)
	if err != nil {
		return nil, err
	}

	miniBlocksRes, ok := miniBlocksResolver.(dataRetriever.MiniBlocksResolver)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	boot := shardStorageBootstrapper{
		storageBootstrapper: base,
		miniBlocksResolver:  miniBlocksRes,
	}

	base.bootstrapper = &boot
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(boot.shardCoordinator.SelfId())
	base.headerNonceHashStore = boot.store.GetStorer(hdrNonceHashDataUnit)

	return &boot, nil
}

// LoadFromStorage will load all blocks from storage
func (ssb *shardStorageBootstrapper) LoadFromStorage() error {
	return ssb.loadBlocks()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ssb *shardStorageBootstrapper) IsInterfaceNil() bool {
	if ssb == nil {
		return true
	}
	return false
}

func (ssb *shardStorageBootstrapper) getHeader(hash []byte) (data.HeaderHandler, error) {
	return ssb.getShardHeaderFromStorage(hash)
}

func (ssb *shardStorageBootstrapper) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	header, ok := headerHandler.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	hashes := make([][]byte, len(header.MiniBlockHeaders))
	for i := 0; i < len(header.MiniBlockHeaders); i++ {
		hashes[i] = header.MiniBlockHeaders[i].Hash
	}

	miniBlocks, missingMiniBlocksHashes := ssb.miniBlocksResolver.GetMiniBlocks(hashes)
	if len(missingMiniBlocksHashes) > 0 {
		return nil, process.ErrMissingBody
	}

	return block.Body(miniBlocks), nil
}

func (ssb *shardStorageBootstrapper) applyNotarizedBlocks(
	lastNotarized map[uint32]*sync.HdrInfo,
) error {
	if len(lastNotarized) == 0 {
		return nil
	}

	if lastNotarized[sharding.MetachainShardId] == nil {
		return sync.ErrNilNotarizedHeader
	}
	if lastNotarized[sharding.MetachainShardId].Hash == nil {
		return sync.ErrNilHash
	}

	metaBlock, err := process.GetMetaHeaderFromStorage(lastNotarized[sharding.MetachainShardId].Hash, ssb.marshalizer, ssb.store)
	if err != nil {
		return err
	}

	ssb.blkExecutor.AddLastNotarizedHdr(sharding.MetachainShardId, metaBlock)

	return nil
}

func (ssb *shardStorageBootstrapper) cleanupNotarizedStorage(lastNotarized map[uint32]*sync.HdrInfo) {
	for shardId, hdrInfo := range lastNotarized {
		storer := ssb.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
		nonceToByteSlice := ssb.uint64Converter.ToByteSlice(hdrInfo.Nonce)

		err := storer.Remove(nonceToByteSlice)
		if err != nil {
			log.Debug("header cannot be removed", "error", err.Error(),
				"nonce", hdrInfo.Nonce, "shardId", shardId)
		}
	}
}
