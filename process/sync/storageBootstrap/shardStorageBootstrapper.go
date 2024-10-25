package storageBootstrap

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/sync"
)

var _ process.BootstrapperFromStorage = (*shardStorageBootstrapper)(nil)

type shardStorageBootstrapper struct {
	*storageBootstrapper
}

// NewShardStorageBootstrapper is method used to create a new storage bootstrapper
func NewShardStorageBootstrapper(arguments ArgsShardStorageBootstrapper) (*shardStorageBootstrapper, error) {
	err := checkShardStorageBootstrapperArgs(arguments)
	if err != nil {
		return nil, err
	}

	base := &storageBootstrapper{
		bootStorer:                   arguments.BootStorer,
		forkDetector:                 arguments.ForkDetector,
		blkExecutor:                  arguments.BlockProcessor,
		blkc:                         arguments.ChainHandler,
		marshalizer:                  arguments.Marshalizer,
		store:                        arguments.Store,
		shardCoordinator:             arguments.ShardCoordinator,
		nodesCoordinator:             arguments.NodesCoordinator,
		epochStartTrigger:            arguments.EpochStartTrigger,
		blockTracker:                 arguments.BlockTracker,
		uint64Converter:              arguments.Uint64Converter,
		bootstrapRoundIndex:          arguments.BootstrapRoundIndex,
		chainID:                      arguments.ChainID,
		scheduledTxsExecutionHandler: arguments.ScheduledTxsExecutionHandler,
		miniBlocksProvider:           arguments.MiniblocksProvider,
		epochNotifier:                arguments.EpochNotifier,
		processedMiniBlocksTracker:   arguments.ProcessedMiniBlocksTracker,
		appStatusHandler:             arguments.AppStatusHandler,
	}

	boot := shardStorageBootstrapper{
		storageBootstrapper: base,
	}

	base.bootstrapper = &boot
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(boot.shardCoordinator.SelfId())
	base.headerNonceHashStore, err = boot.store.GetStorer(hdrNonceHashDataUnit)
	if err != nil {
		return nil, err
	}

	base.getScheduledRootHashMethod = base.getScheduledRootHash
	base.setScheduledInfoMethod = base.setScheduledInfo

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
	return process.GetShardHeaderFromStorage(hash, ssb.marshalizer, ssb.store)
}

func (ssb *shardStorageBootstrapper) getHeaderWithNonce(nonce uint64, shardID uint32) (data.HeaderHandler, []byte, error) {
	return process.GetShardHeaderFromStorageWithNonce(nonce, shardID, ssb.store, ssb.uint64Converter, ssb.marshalizer)
}

func (ssb *shardStorageBootstrapper) applyCrossNotarizedHeaders(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) error {
	for _, crossNotarizedHeader := range crossNotarizedHeaders {
		if crossNotarizedHeader.ShardId != core.MetachainShardId {
			continue
		}

		metaBlock, err := process.GetMetaHeaderFromStorage(crossNotarizedHeader.Hash, ssb.marshalizer, ssb.store)
		if err != nil {
			return err
		}

		log.Debug("added cross notarized header in block tracker",
			"shard", core.MetachainShardId,
			"round", metaBlock.GetRound(),
			"nonce", metaBlock.GetNonce(),
			"hash", crossNotarizedHeader.Hash)

		ssb.blockTracker.AddCrossNotarizedHeader(core.MetachainShardId, metaBlock, crossNotarizedHeader.Hash)
		ssb.blockTracker.AddTrackedHeader(metaBlock, crossNotarizedHeader.Hash)
	}

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

	for _, metaBlockHash := range shardHeader.GetMetaBlockHashes() {
		var metaBlock *block.MetaBlock
		metaBlock, err = process.GetMetaHeaderFromStorage(metaBlockHash, ssb.marshalizer, ssb.store)
		if err != nil {
			log.Debug("meta block is not found in MetaBlockUnit storage",
				"hash", metaBlockHash)
			continue
		}

		log.Debug("removing meta block from storage",
			"shardId", metaBlock.GetShardID(),
			"nonce", metaBlock.GetNonce(),
			"hash", metaBlockHash)

		ssb.removeHdrFromHeaderNonceToHashUnit(metaBlock, metaBlockHash, dataRetriever.MetaHdrNonceHashDataUnit)
		ssb.removeBlockFromBlockUnit(metaBlock, metaBlockHash, dataRetriever.MetaBlockUnit)
	}
}

func (ssb *shardStorageBootstrapper) cleanupNotarizedStorageForHigherNoncesIfExist(
	crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo,
) {
	var numConsecutiveNoncesNotFound int

	lastCrossNotarizedNonce, err := getLastCrossNotarizedHeaderNonce(crossNotarizedHeaders, core.MetachainShardId)
	if err != nil {
		log.Warn("cleanupNotarizedStorageForHigherNoncesIfExist", "error", err.Error())
		return
	}

	log.Debug("cleanup notarized storage has been started", "from nonce", lastCrossNotarizedNonce+1)
	nonce := lastCrossNotarizedNonce

	for {
		nonce++

		metaBlock, metaBlockHash, err := process.GetMetaHeaderFromStorageWithNonce(
			nonce,
			ssb.store,
			ssb.uint64Converter,
			ssb.marshalizer,
		)
		if err != nil {
			log.Debug("meta block is not found in MetaHdrNonceHashDataUnit storage",
				"nonce", nonce, "error", err.Error())

			numConsecutiveNoncesNotFound++
			if numConsecutiveNoncesNotFound > maxNumOfConsecutiveNoncesNotFoundAccepted {
				log.Debug("cleanup notarized storage has been finished",
					"from nonce", lastCrossNotarizedNonce+1,
					"to nonce", nonce)
				break
			}

			continue
		}

		numConsecutiveNoncesNotFound = 0

		log.Debug("removing meta block from storage",
			"shardId", metaBlock.GetShardID(),
			"nonce", metaBlock.GetNonce(),
			"hash", metaBlockHash)

		ssb.removeHdrFromHeaderNonceToHashUnit(metaBlock, metaBlockHash, dataRetriever.MetaHdrNonceHashDataUnit)
		ssb.removeBlockFromBlockUnit(metaBlock, metaBlockHash, dataRetriever.MetaBlockUnit)
	}
}

func (ssb *shardStorageBootstrapper) removeHdrFromHeaderNonceToHashUnit(
	block data.HeaderHandler,
	hash []byte,
	unitType dataRetriever.UnitType,
) {
	nonceToByteSlice := ssb.uint64Converter.ToByteSlice(block.GetNonce())
	hdrNonceHashStorer, err := ssb.store.GetStorer(unitType)
	if err != nil {
		log.Debug("could not get storage unit",
			"unit", unitType,
			"error", err.Error())
		return
	}

	err = hdrNonceHashStorer.Remove(nonceToByteSlice)
	if err != nil {
		log.Debug(fmt.Sprintf("block was not removed from %s storage", unitType.String()),
			"shardId", block.GetShardID(),
			"nonce", block.GetNonce(),
			"hash", hash,
			"error", err.Error())
	}
}

func (ssb *shardStorageBootstrapper) removeBlockFromBlockUnit(
	block data.HeaderHandler,
	hash []byte,
	unitType dataRetriever.UnitType,
) {
	blockStorer, err := ssb.store.GetStorer(unitType)
	if err != nil {
		log.Debug("could not get storage unit",
			"unit", unitType,
			"error", err.Error())
		return
	}

	err = blockStorer.Remove(hash)
	if err != nil {
		log.Debug(fmt.Sprintf("block was not removed from %s storage", unitType.String()),
			"shardId", block.GetShardID(),
			"nonce", block.GetNonce(),
			"hash", hash,
			"error", err.Error())
	}
}

func getLastCrossNotarizedHeaderNonce(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo, shardID uint32) (uint64, error) {
	for _, crossNotarizedHeader := range crossNotarizedHeaders {
		if crossNotarizedHeader.ShardId != shardID {
			continue
		}

		log.Debug("last cross notarized header",
			"shard", crossNotarizedHeader.ShardId,
			"epoch", crossNotarizedHeader.Epoch,
			"nonce", crossNotarizedHeader.Nonce,
			"hash", crossNotarizedHeader.Hash)

		return crossNotarizedHeader.Nonce, nil
	}

	return 0, sync.ErrHeaderNotFound
}

func (ssb *shardStorageBootstrapper) applySelfNotarizedHeaders(
	bootstrapHeadersInfo []bootstrapStorage.BootstrapHeaderInfo,
) ([]data.HeaderHandler, [][]byte, error) {

	selfNotarizedHeadersHashes := make([][]byte, len(bootstrapHeadersInfo))
	for index, selfNotarizedHeader := range bootstrapHeadersInfo {
		selfNotarizedHeadersHashes[index] = selfNotarizedHeader.Hash
	}

	selfNotarizedHeaders := make([]data.HeaderHandler, len(selfNotarizedHeadersHashes))
	for index, selfNotarizedHeaderHash := range selfNotarizedHeadersHashes {
		selfNotarizedHeader, err := ssb.getHeader(selfNotarizedHeaderHash)
		if err != nil {
			return nil, nil, err
		}

		selfNotarizedHeaders[index] = selfNotarizedHeader

		log.Debug("added self notarized header in block tracker",
			"shard", core.MetachainShardId,
			"round", selfNotarizedHeader.GetRound(),
			"nonce", selfNotarizedHeader.GetNonce(),
			"hash", selfNotarizedHeaderHash)

		ssb.blockTracker.AddSelfNotarizedHeader(core.MetachainShardId, selfNotarizedHeader, selfNotarizedHeaderHash)
	}

	return selfNotarizedHeaders, selfNotarizedHeadersHashes, nil
}

func (ssb *shardStorageBootstrapper) applyNumPendingMiniBlocks(_ []bootstrapStorage.PendingMiniBlocksInfo) {
}

func (ssb *shardStorageBootstrapper) getRootHash(shardHeaderHash []byte) []byte {
	shardHeader, err := process.GetShardHeaderFromStorage(shardHeaderHash, ssb.marshalizer, ssb.store)
	if err != nil {
		log.Debug("shard header is not found in BlockHeaderUnit storage",
			"hash", shardHeaderHash)
		return nil
	}

	return shardHeader.GetRootHash()
}

func checkShardStorageBootstrapperArgs(args ArgsShardStorageBootstrapper) error {
	return checkBaseStorageBootstrapperArguments(args.ArgsBaseStorageBootstrapper)
}
