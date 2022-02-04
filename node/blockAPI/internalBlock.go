package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type internalBlockProcessor struct {
	*baseAPIBlockProcessor
}

// NewInternalBlockProcessor will create a new instance of internal block processor
func NewInternalBlockProcessor(arg *APIBlockProcessorArg) *internalBlockProcessor {
	hasDbLookupExtensions := arg.HistoryRepo.IsEnabled()

	return &internalBlockProcessor{
		baseAPIBlockProcessor: &baseAPIBlockProcessor{
			hasDbLookupExtensions:    hasDbLookupExtensions,
			selfShardID:              arg.SelfShardID,
			store:                    arg.Store,
			marshalizer:              arg.Marshalizer,
			uint64ByteSliceConverter: arg.Uint64ByteSliceConverter,
			historyRepo:              arg.HistoryRepo,
			unmarshalTx:              arg.UnmarshalTx,
			txStatusComputer:         arg.StatusComputer,
		},
	}
}

// GetInternalShardBlockByNonce wil return a shard block by nonce
func (ibp *internalBlockProcessor) GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
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
	blockBytes, err := ibp.getFromStorer(dataRetriever.BlockHeaderUnit, hash)
	if err != nil {
		return nil, err
	}

	return ibp.convertShardBlockBytesByOutputFormat(format, blockBytes)
}

// GetInternalShardBlockByRound wil return a shard block by round
func (ibp *internalBlockProcessor) GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	_, blockBytes, err := ibp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	return ibp.convertShardBlockBytesByOutputFormat(format, blockBytes)
}

func (ibp *internalBlockProcessor) convertShardBlockBytesToInternalBlock(blockBytes []byte) (*block.Header, error) {
	blockHeader := &block.Header{}
	err := ibp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	return blockHeader, nil
}

func (ibp *internalBlockProcessor) convertShardBlockBytesByOutputFormat(format common.ApiOutputFormat, blockBytes []byte) (interface{}, error) {
	switch format {
	case common.ApiOutputFormatInternal:
		return ibp.convertShardBlockBytesToInternalBlock(blockBytes)
	case common.ApiOutputFormatProto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutputFormat
	}
}

// GetInternalMetaBlockByNonce wil return a meta block by nonce
func (ibp *internalBlockProcessor) GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
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
	blockBytes, err := ibp.getFromStorer(dataRetriever.MetaBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	return ibp.convertMetaBlockBytesByOutputFormat(format, blockBytes)
}

// GetInternalMetaBlockByRound wil return a meta block by round
func (ibp *internalBlockProcessor) GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	_, blockBytes, err := ibp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.MetaBlockUnit)
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
	case common.ApiOutputFormatInternal:
		return ibp.convertMetaBlockBytesToInternalBlock(blockBytes)
	case common.ApiOutputFormatProto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutputFormat
	}
}
