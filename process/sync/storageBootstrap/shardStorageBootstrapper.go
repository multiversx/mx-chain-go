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
	return ssb == nil
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

func (ssb *shardStorageBootstrapper) cleanupNotarizedStorage(shardHeaderHash []byte) {
	log.Debug("cleanup notarized storage")

	shardHeader, err := process.GetShardHeaderFromStorage(shardHeaderHash, ssb.marshalizer, ssb.store)
	if err != nil {
		log.Debug("shard header is not found in BlockHeaderUnit storage",
			"hash", shardHeaderHash)
		return
	}

	for _, metaBlockHash := range shardHeader.MetaBlockHashes {
		metaBlock, err := process.GetMetaHeaderFromStorage(metaBlockHash, ssb.marshalizer, ssb.store)
		if err != nil {
			log.Debug("meta block is not found in MetaBlockUnit storage",
				"hash", metaBlockHash)
			continue
		}

		nonceToByteSlice := ssb.uint64Converter.ToByteSlice(metaBlock.GetNonce())
		err = ssb.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit).Remove(nonceToByteSlice)
		if err != nil {
			log.Debug("meta block was not removed from MetaHdrNonceHashDataUnit storage",
				"shradId", metaBlock.GetShardID(),
				"nonce", metaBlock.GetNonce(),
				"hash", metaBlockHash,
				"error", err.Error())
		}

		err = ssb.store.GetStorer(dataRetriever.MetaBlockUnit).Remove(metaBlockHash)
		if err != nil {
			log.Debug("meta block was not removed from MetaBlockUnit storage",
				"shradId", metaBlock.GetShardID(),
				"nonce", metaBlock.GetNonce(),
				"hash", metaBlockHash,
				"error", err.Error())
			continue
		}

		log.Debug("meta block was removed from storage",
			"shradId", metaBlock.GetShardID(),
			"nonce", metaBlock.GetNonce(),
			"hash", metaBlockHash)
	}
}
