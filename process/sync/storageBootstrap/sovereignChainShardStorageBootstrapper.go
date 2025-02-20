package storageBootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
)

type sovereignChainShardStorageBootstrapper struct {
	*shardStorageBootstrapper
}

// NewSovereignChainShardStorageBootstrapper creates a new instance of sovereignChainShardStorageBootstrapper
func NewSovereignChainShardStorageBootstrapper(shardStorageBootstrapper *shardStorageBootstrapper) (*sovereignChainShardStorageBootstrapper, error) {
	if shardStorageBootstrapper == nil {
		return nil, process.ErrNilShardStorageBootstrapper
	}

	scssb := &sovereignChainShardStorageBootstrapper{
		shardStorageBootstrapper,
	}

	scssb.getScheduledRootHashMethod = scssb.sovereignChainGetScheduledRootHash
	scssb.setScheduledInfoMethod = scssb.sovereignChainSetScheduledInfo

	scssb.bootstrapper = scssb
	return scssb, nil
}

func (ssb *sovereignChainShardStorageBootstrapper) applyCrossNotarizedHeaders(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) error {
	for _, crossNotarizedHeader := range crossNotarizedHeaders {
		if crossNotarizedHeader.ShardId != core.MainChainShardId {
			continue
		}

		extendedHeader, err := process.GetExtendedShardHeaderFromStorage(crossNotarizedHeader.Hash, ssb.marshalizer, ssb.store)
		if err != nil {
			return err
		}

		log.Debug("added cross notarized header in block tracker",
			"shard", core.MainChainShardId,
			"round", extendedHeader.GetRound(),
			"nonce", extendedHeader.GetNonce(),
			"hash", crossNotarizedHeader.Hash)

		ssb.blockTracker.AddCrossNotarizedHeader(core.MainChainShardId, extendedHeader, crossNotarizedHeader.Hash)
		ssb.blockTracker.AddTrackedHeader(extendedHeader, crossNotarizedHeader.Hash)
	}

	return nil
}

func (ssb *sovereignChainShardStorageBootstrapper) cleanupNotarizedStorage(shardHeaderHash []byte) {
	log.Debug("sovereign cleanup notarized storage")

	shardHeader, err := process.GetShardHeaderFromStorage(shardHeaderHash, ssb.marshalizer, ssb.store)
	if err != nil {
		log.Debug("sovereign shard header is not found in BlockHeaderUnit storage",
			"hash", shardHeaderHash)
		return
	}

	sovereignHeader, castOk := shardHeader.(data.SovereignChainHeaderHandler)
	if !castOk {
		log.Warn("sovereignChainShardStorageBootstrapper.cleanupNotarizedStorage",
			"error", process.ErrWrongTypeAssertion,
			"expected shard header of type", "SovereignChainHeaderHandler",
		)
		return
	}

	var extendedHeader data.HeaderHandler
	for _, extendedHeaderHash := range sovereignHeader.GetExtendedShardHeaderHashes() {
		extendedHeader, err = process.GetExtendedShardHeaderFromStorage(extendedHeaderHash, ssb.marshalizer, ssb.store)
		if err != nil {
			log.Debug("extended block is not found in ExtendedShardHeadersUnit storage",
				"hash", extendedHeaderHash)
			continue
		}

		log.Debug("removing extended header from storage",
			"shardId", extendedHeader.GetShardID(),
			"nonce", extendedHeader.GetNonce(),
			"hash", extendedHeaderHash)

		ssb.removeHdrFromHeaderNonceToHashUnit(extendedHeader, extendedHeaderHash, dataRetriever.ExtendedShardHeadersNonceHashDataUnit)
		ssb.removeBlockFromBlockUnit(extendedHeader, extendedHeaderHash, dataRetriever.ExtendedShardHeadersUnit)
	}
}

func (ssb *sovereignChainShardStorageBootstrapper) cleanupNotarizedStorageForHigherNoncesIfExist(
	crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo,
) {
	var numConsecutiveNoncesNotFound int

	lastCrossNotarizedNonce, err := getLastCrossNotarizedHeaderNonce(crossNotarizedHeaders, core.MainChainShardId)
	if err != nil {
		log.Warn("cleanupNotarizedStorageForHigherNoncesIfExist", "error", err.Error())
		return
	}

	log.Debug("cleanup notarized storage has been started", "from nonce", lastCrossNotarizedNonce+1)
	nonce := lastCrossNotarizedNonce

	for {
		nonce++

		extendedBlock, extendedBlockHash, err := process.GetExtendedHeaderFromStorageWithNonce(
			nonce,
			ssb.store,
			ssb.uint64Converter,
			ssb.marshalizer,
		)
		if err != nil {
			log.Debug("sovereignChainShardStorageBootstrapper.cleanupNotarizedStorageForHigherNoncesIfExist:"+
				"trying to cleanup an extended header from storage that is not found",
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

		log.Debug("removing extended block from storage",
			"shardId", extendedBlock.GetShardID(),
			"nonce", extendedBlock.GetNonce(),
			"hash", extendedBlockHash)

		ssb.removeHdrFromHeaderNonceToHashUnit(extendedBlock, extendedBlockHash, dataRetriever.ExtendedShardHeadersNonceHashDataUnit)
		ssb.removeBlockFromBlockUnit(extendedBlock, extendedBlockHash, dataRetriever.ExtendedShardHeadersUnit)
	}
}
