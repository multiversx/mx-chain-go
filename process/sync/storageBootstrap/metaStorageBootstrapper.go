package storageBootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
)

var _ process.BootstrapperFromStorage = (*metaStorageBootstrapper)(nil)

type metaStorageBootstrapper struct {
	*storageBootstrapper
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler
}

// NewMetaStorageBootstrapper is method used to create a new storage bootstrapper
func NewMetaStorageBootstrapper(arguments ArgsMetaStorageBootstrapper) (*metaStorageBootstrapper, error) {
	err := checkMetaStorageBootstrapperArgs(arguments)
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
		enableEpochsHandler:          arguments.EnableEpochsHandler,
		proofsPool:                   arguments.ProofsPool,
	}

	boot := metaStorageBootstrapper{
		storageBootstrapper:      base,
		pendingMiniBlocksHandler: arguments.PendingMiniBlocksHandler,
	}

	base.bootstrapper = &boot
	base.headerNonceHashStore, err = boot.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if err != nil {
		return nil, err
	}

	return &boot, nil
}

// LoadFromStorage will load all blocks from storage
func (msb *metaStorageBootstrapper) LoadFromStorage() error {
	return msb.loadBlocks()
}

// IsInterfaceNil returns true if there is no value under the interface
func (msb *metaStorageBootstrapper) IsInterfaceNil() bool {
	return msb == nil
}

func (msb *metaStorageBootstrapper) applyCrossNotarizedHeaders(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) error {
	for _, crossNotarizedHeader := range crossNotarizedHeaders {
		header, err := process.GetShardHeaderFromStorage(crossNotarizedHeader.Hash, msb.marshalizer, msb.store)
		if err != nil {
			return err
		}

		err = msb.getAndApplyProofForHeader(crossNotarizedHeader.Hash, header)
		if err != nil {
			return err
		}

		log.Debug("added cross notarized header in block tracker",
			"shard", crossNotarizedHeader.ShardId,
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"hash", crossNotarizedHeader.Hash)

		msb.blockTracker.AddCrossNotarizedHeader(crossNotarizedHeader.ShardId, header, crossNotarizedHeader.Hash)
		msb.blockTracker.AddTrackedHeader(header, crossNotarizedHeader.Hash)
	}

	return nil
}

func (msb *metaStorageBootstrapper) getHeader(hash []byte) (data.HeaderHandler, error) {
	return process.GetMetaHeaderFromStorage(hash, msb.marshalizer, msb.store)
}

func (msb *metaStorageBootstrapper) getHeaderWithNonce(nonce uint64, _ uint32) (data.HeaderHandler, []byte, error) {
	return process.GetMetaHeaderFromStorageWithNonce(nonce, msb.store, msb.uint64Converter, msb.marshalizer)
}

func (msb *metaStorageBootstrapper) cleanupNotarizedStorage(metaBlockHash []byte) {
	log.Debug("cleanup notarized storage")

	metaBlock, err := process.GetMetaHeaderFromStorage(metaBlockHash, msb.marshalizer, msb.store)
	if err != nil {
		log.Debug("meta block is not found in MetaBlockUnit storage",
			"hash", metaBlockHash)
		return
	}

	shardHeaderHashes := make([][]byte, len(metaBlock.ShardInfo))
	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		shardHeaderHashes[i] = metaBlock.ShardInfo[i].HeaderHash
	}

	for _, shardHeaderHash := range shardHeaderHashes {
		var shardHeader data.HeaderHandler
		shardHeader, err = process.GetShardHeaderFromStorage(shardHeaderHash, msb.marshalizer, msb.store)
		if err != nil {
			log.Debug("shard header is not found in BlockHeaderUnit storage",
				"hash", shardHeaderHash)
			continue
		}

		log.Debug("removing shard header from ShardHdrNonceHashDataUnit storage",
			"shardId", shardHeader.GetShardID(),
			"nonce", shardHeader.GetNonce(),
			"hash", shardHeaderHash)

		hdrNonceHashDataUnit := dataRetriever.GetHdrNonceHashDataUnit(shardHeader.GetShardID())
		storer, err := msb.store.GetStorer(hdrNonceHashDataUnit)
		if err != nil {
			log.Debug("could not get storage unit",
				"unit", hdrNonceHashDataUnit,
				"error", err.Error())
			return
		}

		nonceToByteSlice := msb.uint64Converter.ToByteSlice(shardHeader.GetNonce())
		err = storer.Remove(nonceToByteSlice)
		if err != nil {
			log.Debug("shard header was not removed from ShardHdrNonceHashDataUnit storage",
				"shardId", shardHeader.GetShardID(),
				"nonce", shardHeader.GetNonce(),
				"hash", shardHeaderHash,
				"error", err.Error())
		}
	}
}

func (msb *metaStorageBootstrapper) cleanupNotarizedStorageForHigherNoncesIfExist(_ []bootstrapStorage.BootstrapHeaderInfo) {
}

func (msb *metaStorageBootstrapper) applySelfNotarizedHeaders(
	bootstrapHeadersInfo []bootstrapStorage.BootstrapHeaderInfo,
) ([]data.HeaderHandler, [][]byte, error) {

	for _, bootstrapHeaderInfo := range bootstrapHeadersInfo {
		selfNotarizedHeader, err := msb.getHeader(bootstrapHeaderInfo.Hash)
		if err != nil {
			return nil, nil, err
		}

		err = msb.getAndApplyProofForHeader(bootstrapHeaderInfo.Hash, selfNotarizedHeader)
		if err != nil {
			return nil, nil, err
		}

		log.Debug("added self notarized header in block tracker",
			"shard", bootstrapHeaderInfo.ShardId,
			"round", selfNotarizedHeader.GetRound(),
			"nonce", selfNotarizedHeader.GetNonce(),
			"hash", bootstrapHeaderInfo.Hash)

		msb.blockTracker.AddSelfNotarizedHeader(bootstrapHeaderInfo.ShardId, selfNotarizedHeader, bootstrapHeaderInfo.Hash)
	}

	return make([]data.HeaderHandler, 0), make([][]byte, 0), nil
}

const maxSelfNotarizedLookback = 50

func (msb *metaStorageBootstrapper) completeSelfNotarizedHeaders(lastMetaBlockHash []byte) error {
	numShards := msb.shardCoordinator.NumberOfShards()
	missingShards := make(map[uint32]bool)

	for shardID := uint32(0); shardID < numShards; shardID++ {
		lastSelfNotarized, _, err := msb.blockTracker.GetLastSelfNotarizedHeader(shardID)
		if err != nil || check.IfNil(lastSelfNotarized) || lastSelfNotarized.GetNonce() == 0 {
			missingShards[shardID] = true
		}
	}

	if len(missingShards) == 0 {
		return nil
	}

	log.Debug("completeSelfNotarizedHeaders: deriving missing per-shard headers",
		"numMissing", len(missingShards))

	currentHash := lastMetaBlockHash
	for i := 0; i < maxSelfNotarizedLookback && len(missingShards) > 0 && len(currentHash) > 0; i++ {
		metaBlock, err := process.GetMetaHeaderFromStorage(currentHash, msb.marshalizer, msb.store)
		if err != nil {
			log.Debug("completeSelfNotarizedHeaders: could not load metablock",
				"hash", currentHash, "error", err.Error())
			break
		}

		msb.findSelfNotarizedForMissingShards(metaBlock, missingShards)

		currentHash = metaBlock.GetPrevHash()
	}

	if len(missingShards) > 0 {
		log.Warn("completeSelfNotarizedHeaders: could not derive all self-notarized headers",
			"numStillMissing", len(missingShards))
	}

	return nil
}

func (msb *metaStorageBootstrapper) findSelfNotarizedForMissingShards(
	metaBlock *block.MetaBlock,
	missingShards map[uint32]bool,
) {
	for shardID := range missingShards {
		var bestNonce uint64
		var bestHeader data.HeaderHandler
		var bestHash []byte
		hadLoadErrors := false

		for i := range metaBlock.ShardInfo {
			if metaBlock.ShardInfo[i].ShardID != shardID {
				continue
			}

			shardHeader, err := process.GetShardHeaderFromStorage(metaBlock.ShardInfo[i].HeaderHash, msb.marshalizer, msb.store)
			if err != nil {
				log.Warn("completeSelfNotarizedHeaders: could not load shard header",
					"shardID", shardID,
					"headerHash", metaBlock.ShardInfo[i].HeaderHash,
					"error", err.Error())
				hadLoadErrors = true
				continue
			}

			for _, metaHash := range shardHeader.GetMetaBlockHashes() {
				metaHeader, err := process.GetMetaHeaderFromStorage(metaHash, msb.marshalizer, msb.store)
				if err != nil {
					continue
				}

				if metaHeader.GetNonce() > bestNonce {
					bestNonce = metaHeader.GetNonce()
					bestHeader = metaHeader
					bestHash = metaHash
				}
			}
		}

		if hadLoadErrors {
			continue
		}

		if bestHeader != nil {
			log.Debug("completeSelfNotarizedHeaders: derived self-notarized header for shard",
				"shardID", shardID,
				"metaNonce", bestNonce,
				"metaHash", bestHash)

			msb.blockTracker.AddSelfNotarizedHeader(shardID, bestHeader, bestHash)
			delete(missingShards, shardID)
		}
	}
}

func (msb *metaStorageBootstrapper) applyNumPendingMiniBlocks(pendingMiniBlocksInfo []bootstrapStorage.PendingMiniBlocksInfo) {
	for _, pendingMiniBlockInfo := range pendingMiniBlocksInfo {
		msb.pendingMiniBlocksHandler.SetPendingMiniBlocks(pendingMiniBlockInfo.ShardID, pendingMiniBlockInfo.MiniBlocksHashes)

		log.Debug("set pending miniblocks",
			"shard", pendingMiniBlockInfo.ShardID,
			"num", len(pendingMiniBlockInfo.MiniBlocksHashes))

		for _, hash := range pendingMiniBlockInfo.MiniBlocksHashes {
			log.Trace("miniblock", "hash", hash)
		}
	}
}

func (msb *metaStorageBootstrapper) getRootHash(metaBlockHash []byte) []byte {
	metaBlock, err := process.GetMetaHeaderFromStorage(metaBlockHash, msb.marshalizer, msb.store)
	if err != nil {
		log.Debug("meta block is not found in MetaBlockUnit storage",
			"hash", metaBlockHash)
		return nil
	}

	return metaBlock.GetRootHash()
}

func checkMetaStorageBootstrapperArgs(args ArgsMetaStorageBootstrapper) error {
	err := checkBaseStorageBootstrapperArguments(args.ArgsBaseStorageBootstrapper)
	if err != nil {
		return err
	}
	if check.IfNil(args.PendingMiniBlocksHandler) {
		return process.ErrNilPendingMiniBlocksHandler
	}

	return nil
}
