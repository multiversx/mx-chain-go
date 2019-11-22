package storageBootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/sync"
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
	if msb == nil {
		return true
	}
	return false
}

func (msb *metaStorageBootstrapper) applyNotarizedBlocks(
	lastNotarized map[uint32]*sync.HdrInfo,
) error {
	for i := uint32(0); i < msb.shardCoordinator.NumberOfShards(); i++ {
		if lastNotarized[i] == nil {
			continue
		}
		if lastNotarized[i].Hash == nil {
			return sync.ErrNilHash
		}

		headerHandler, err := process.GetShardHeaderFromStorage(lastNotarized[i].Hash, msb.marshalizer, msb.store)
		if err != nil {
			return err
		}

		msb.blkExecutor.AddLastNotarizedHdr(i, headerHandler)
	}

	return nil
}

func (msb *metaStorageBootstrapper) getHeader(hash []byte) (data.HeaderHandler, error) {
	return msb.getMetaHeaderFromStorage(hash)
}

func (msb *metaStorageBootstrapper) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	return block.Body{}, nil
}

func (msb *metaStorageBootstrapper) cleanupNotarizedStorage(lastNotarized map[uint32]*sync.HdrInfo) {
	for shardId, hdrInfo := range lastNotarized {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardId)
		storer := msb.store.GetStorer(hdrNonceHashDataUnit)
		nonceToByteSlice := msb.uint64Converter.ToByteSlice(hdrInfo.Nonce)

		err := storer.Remove(nonceToByteSlice)
		if err != nil {
			log.Info("header cannot be removed", "error", err.Error(),
				"nonce", hdrInfo.Nonce, "shardId", shardId)
		}
	}
}
