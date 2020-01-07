package storageBootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

type metaStorageBootstrapper struct {
	*storageBootstrapper
}

// NewMetaStorageBootstrapper is method used to create a nes storage bootstrapper
func NewMetaStorageBootstrapper(arguments ArgsStorageBootstrapper) (*metaStorageBootstrapper, error) {
	base := &storageBootstrapper{
		bootStorer:       arguments.BootStorer,
		forkDetector:     arguments.ForkDetector,
		blkExecutor:      arguments.BlockProcessor,
		blkc:             arguments.ChainHandler,
		marshalizer:      arguments.Marshalizer,
		store:            arguments.Store,
		shardCoordinator: arguments.ShardCoordinator,
		blockTracker:     arguments.BlockTracker,

		uint64Converter:     arguments.Uint64Converter,
		bootstrapRoundIndex: arguments.BootstrapRoundIndex,
	}

	boot := metaStorageBootstrapper{
		storageBootstrapper: base,
	}

	base.bootstrapper = &boot
	base.headerNonceHashStore = boot.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)

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

func (msb *metaStorageBootstrapper) applyCrossNotarizedHeaders(crossNotarizedHeadersHashes map[uint32][]byte) error {
	for i := uint32(0); i < msb.shardCoordinator.NumberOfShards(); i++ {
		hash, ok := crossNotarizedHeadersHashes[i]
		if !ok {
			continue
		}

		header, err := process.GetShardHeaderFromStorage(hash, msb.marshalizer, msb.store)
		if err != nil {
			return err
		}

		log.Debug("added cross notarized header in block tracker",
			"shard", i,
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"hash", hash)

		msb.blockTracker.AddCrossNotarizedHeader(i, header, hash)
		msb.blockTracker.AddTrackedHeader(header, hash)
	}

	return nil
}

func (msb *metaStorageBootstrapper) getHeader(hash []byte) (data.HeaderHandler, error) {
	return process.GetMetaHeaderFromStorage(hash, msb.marshalizer, msb.store)
}

func (msb *metaStorageBootstrapper) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	return block.Body{}, nil
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
		shardHeader, err := process.GetShardHeaderFromStorage(shardHeaderHash, msb.marshalizer, msb.store)
		if err != nil {
			log.Debug("shard header is not found in BlockHeaderUnit storage",
				"hash", shardHeaderHash)
			continue
		}

		log.Debug("removing shard header from ShardHdrNonceHashDataUnit storage",
			"shradId", shardHeader.GetShardID(),
			"nonce", shardHeader.GetNonce(),
			"hash", shardHeaderHash)

		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardHeader.GetShardID())
		storer := msb.store.GetStorer(hdrNonceHashDataUnit)
		nonceToByteSlice := msb.uint64Converter.ToByteSlice(shardHeader.GetNonce())
		err = storer.Remove(nonceToByteSlice)
		if err != nil {
			log.Debug("shard header was not removed from ShardHdrNonceHashDataUnit storage",
				"shradId", shardHeader.GetShardID(),
				"nonce", shardHeader.GetNonce(),
				"hash", shardHeaderHash,
				"error", err.Error())
		}
	}
}

func (msb *metaStorageBootstrapper) applySelfNotarizedHeaders(selfNotarizedHeadersHashes [][]byte) ([]data.HeaderHandler, error) {
	selfNotarizedHeaders := make([]data.HeaderHandler, 0, len(selfNotarizedHeadersHashes))
	return selfNotarizedHeaders, nil
}
