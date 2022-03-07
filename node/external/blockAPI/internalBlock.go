package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

type internalBlockProcessor struct {
	*baseAPIBlockProcessor
}

// newInternalBlockProcessor will create a new instance of internal block processor
func newInternalBlockProcessor(arg *ArgAPIBlockProcessor, emptyReceiptsHash []byte) *internalBlockProcessor {
	hasDbLookupExtensions := arg.HistoryRepo.IsEnabled()

	return &internalBlockProcessor{
		baseAPIBlockProcessor: &baseAPIBlockProcessor{
			hasDbLookupExtensions:    hasDbLookupExtensions,
			selfShardID:              arg.SelfShardID,
			store:                    arg.Store,
			marshalizer:              arg.Marshalizer,
			uint64ByteSliceConverter: arg.Uint64ByteSliceConverter,
			historyRepo:              arg.HistoryRepo,
			txUnmarshaller:           arg.TxUnmarshaller,
			txStatusComputer:         arg.StatusComputer,
			hasher:                   arg.Hasher,
			addressPubKeyConverter:   arg.AddressPubkeyConverter,
			emptyReceiptsHash:        emptyReceiptsHash,
		},
	}
}

// GetInternalShardBlockByNonce wil return a shard block by nonce
func (ibp *internalBlockProcessor) GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if ibp.selfShardID == core.MetachainShardId {
		return nil, ErrShardOnlyEndpoint
	}

	storerUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(ibp.selfShardID)

	nonceToByteSlice := ibp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := ibp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	blockBytes, err := ibp.getFromStorer(dataRetriever.BlockHeaderUnit, headerHash)
	if err != nil {
		return nil, err
	}

	return ibp.convertShardBlockBytesByOutputFormat(format, blockBytes)
}

// GetInternalShardBlockByHash wil return a shard block by hash
func (ibp *internalBlockProcessor) GetInternalShardBlockByHash(format common.ApiOutputFormat, hash []byte) (interface{}, error) {
	if ibp.selfShardID == core.MetachainShardId {
		return nil, ErrShardOnlyEndpoint
	}

	blockBytes, err := ibp.getFromStorer(dataRetriever.BlockHeaderUnit, hash)
	if err != nil {
		return nil, err
	}

	return ibp.convertShardBlockBytesByOutputFormat(format, blockBytes)
}

// GetInternalShardBlockByRound wil return a shard block by round
func (ibp *internalBlockProcessor) GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if ibp.selfShardID == core.MetachainShardId {
		return nil, ErrShardOnlyEndpoint
	}

	_, blockBytes, err := ibp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	return ibp.convertShardBlockBytesByOutputFormat(format, blockBytes)
}

func (ibp *internalBlockProcessor) convertShardBlockBytesToInternalBlock(blockBytes []byte) (interface{}, error) {
	return process.CreateShardHeader(ibp.marshalizer, blockBytes)
}

func (ibp *internalBlockProcessor) convertShardBlockBytesByOutputFormat(format common.ApiOutputFormat, blockBytes []byte) (interface{}, error) {
	switch format {
	case common.ApiOutputFormatJSON:
		return ibp.convertShardBlockBytesToInternalBlock(blockBytes)
	case common.ApiOutputFormatProto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutputFormat
	}
}

// GetInternalMetaBlockByNonce wil return a meta block by nonce
func (ibp *internalBlockProcessor) GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if ibp.selfShardID != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	storerUnit := dataRetriever.MetaHdrNonceHashDataUnit

	nonceToByteSlice := ibp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := ibp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	blockBytes, err := ibp.getFromStorer(dataRetriever.MetaBlockUnit, headerHash)
	if err != nil {
		return nil, err
	}

	return ibp.convertMetaBlockBytesByOutputFormat(format, blockBytes)
}

// GetInternalMetaBlockByHash wil return a meta block by hash
func (ibp *internalBlockProcessor) GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash []byte) (interface{}, error) {
	if ibp.selfShardID != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	blockBytes, err := ibp.getFromStorer(dataRetriever.MetaBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	return ibp.convertMetaBlockBytesByOutputFormat(format, blockBytes)
}

// GetInternalMetaBlockByRound wil return a meta block by round
func (ibp *internalBlockProcessor) GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if ibp.selfShardID != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	_, blockBytes, err := ibp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	return ibp.convertMetaBlockBytesByOutputFormat(format, blockBytes)
}

// GetInternalStartOfEpochMetaBlock wil return the epoch start meta block for the provided epoch
func (ibp *internalBlockProcessor) GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error) {
	if ibp.selfShardID != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	storer := ibp.store.GetStorer(dataRetriever.MetaBlockUnit)

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	blockBytes, err := storer.GetFromEpoch([]byte(epochStartIdentifier), epoch)
	if err != nil {
		return nil, err
	}

	return ibp.convertMetaBlockBytesByOutputFormat(format, blockBytes)
}

func (ibp *internalBlockProcessor) convertMetaBlockBytesToInternalBlock(blockBytes []byte) (*block.MetaBlock, error) {
	blockHeader := &block.MetaBlock{}
	err := ibp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	return blockHeader, nil
}

func (ibp *internalBlockProcessor) convertMetaBlockBytesByOutputFormat(format common.ApiOutputFormat, blockBytes []byte) (interface{}, error) {
	switch format {
	case common.ApiOutputFormatJSON:
		return ibp.convertMetaBlockBytesToInternalBlock(blockBytes)
	case common.ApiOutputFormatProto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutputFormat
	}
}

// GetInternalMiniBlock will return the miniblock based on the hash
func (ibp *internalBlockProcessor) GetInternalMiniBlock(format common.ApiOutputFormat, hash []byte, epoch uint32) (interface{}, error) {
	storer := ibp.store.GetStorer(dataRetriever.MiniBlockUnit)
	blockBytes, err := storer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	return ibp.convertMiniBlockBytesByOutportFormat(format, blockBytes)
}

func (ibp *internalBlockProcessor) convertMiniBlockBytesByOutportFormat(format common.ApiOutputFormat, blockBytes []byte) (interface{}, error) {
	switch format {
	case common.ApiOutputFormatJSON:
		miniBlock := &block.MiniBlock{}
		err := ibp.marshalizer.Unmarshal(miniBlock, blockBytes)
		if err != nil {
			return nil, err
		}
		return miniBlock, nil
	case common.ApiOutputFormatProto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutputFormat
	}
}

// IsInterfaceNil returns true if underlying object is nil
func (ibp *internalBlockProcessor) IsInterfaceNil() bool {
	return ibp == nil
}
