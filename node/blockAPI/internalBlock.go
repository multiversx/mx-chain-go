package blockAPI

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ErrInvalidOutportFormat signals that the outport format type is not valid
var ErrInvalidOutportFormat = errors.New("the outport format type is invalid")

type internalBlockProcessor struct {
	*baseAPIBlockProcessor
}

// NewInternalBlockProcessor will create a new instance of raw block processor
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
func (ibp *internalBlockProcessor) GetInternalShardBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error) {
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

	return ibp.convertShardBlockBytesByOutportFormat(format, blockBytes)
}

// GetInternalShardBlockByHash wil return a shard block by hash
func (ibp *internalBlockProcessor) GetInternalShardBlockByHash(format common.OutportFormat, hash []byte) (interface{}, error) {
	blockBytes, err := ibp.getFromStorer(dataRetriever.BlockHeaderUnit, hash)
	if err != nil {
		return nil, err
	}

	return ibp.convertShardBlockBytesByOutportFormat(format, blockBytes)
}

// GetInternalShardBlockByRound wil return a shard block by round
func (ibp *internalBlockProcessor) GetInternalShardBlockByRound(format common.OutportFormat, round uint64) (interface{}, error) {
	_, blockBytes, err := ibp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	return ibp.convertShardBlockBytesByOutportFormat(format, blockBytes)
}

func (ibp *internalBlockProcessor) convertShardBlockBytesToInternalBlock(blockBytes []byte) (*block.Header, error) {
	blockHeader := &block.Header{}
	err := ibp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	return blockHeader, nil
}

func (ibp *internalBlockProcessor) convertShardBlockBytesByOutportFormat(format common.OutportFormat, blockBytes []byte) (interface{}, error) {
	switch format {
	case common.Internal:
		return ibp.convertShardBlockBytesToInternalBlock(blockBytes)
	case common.Proto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutportFormat
	}
}

// GetInternalMetaBlockByNonce wil return a meta block by nonce
func (ibp *internalBlockProcessor) GetInternalMetaBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error) {
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

	return ibp.convertMetaBlockBytesByOutportFormat(format, blockBytes)
}

// GetInternalMetaBlockByHash wil return a meta block by hash
func (ibp *internalBlockProcessor) GetInternalMetaBlockByHash(format common.OutportFormat, hash []byte) (interface{}, error) {
	blockBytes, err := ibp.getFromStorer(dataRetriever.MetaBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	return ibp.convertMetaBlockBytesByOutportFormat(format, blockBytes)
}

// GetInternalMetaBlockByRound wil return a meta block by round
func (ibp *internalBlockProcessor) GetInternalMetaBlockByRound(format common.OutportFormat, round uint64) (interface{}, error) {
	_, blockBytes, err := ibp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	return ibp.convertMetaBlockBytesByOutportFormat(format, blockBytes)
}

func (ibp *internalBlockProcessor) convertMetaBlockBytesToInternalBlock(blockBytes []byte) (*block.MetaBlock, error) {
	blockHeader := &block.MetaBlock{}
	err := ibp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	return blockHeader, nil
}

func (ibp *internalBlockProcessor) convertMetaBlockBytesByOutportFormat(format common.OutportFormat, blockBytes []byte) (interface{}, error) {
	switch format {
	case common.Internal:
		return ibp.convertMetaBlockBytesToInternalBlock(blockBytes)
	case common.Proto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutportFormat
	}
}
