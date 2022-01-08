package blockAPI

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ErrInvalidOutportFormat signals that the outport format type is not valid
var ErrInvalidOutportFormat = errors.New("the outport format type is invalid")

type rawBlockProcessor struct {
	*baseAPIBlockProcessor
}

// NewRawBlockProcessor will create a new instance of raw block processor
func NewRawBlockProcessor(arg *APIBlockProcessorArg) *rawBlockProcessor {
	hasDbLookupExtensions := arg.HistoryRepo.IsEnabled()

	return &rawBlockProcessor{
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
func (rbp *rawBlockProcessor) GetInternalShardBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error) {
	storerUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(rbp.selfShardID)

	nonceToByteSlice := rbp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := rbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	blockBytes, err := rbp.getFromStorer(dataRetriever.BlockHeaderUnit, headerHash)
	if err != nil {
		return nil, err
	}

	return rbp.convertShardBlockBytesByOutportFormat(format, blockBytes)
}

// GetInternalShardBlockByHash wil return a shard block by hash
func (rbp *rawBlockProcessor) GetInternalShardBlockByHash(format common.OutportFormat, hash []byte) (interface{}, error) {
	blockBytes, err := rbp.getFromStorer(dataRetriever.BlockHeaderUnit, hash)
	if err != nil {
		return nil, err
	}

	return rbp.convertShardBlockBytesByOutportFormat(format, blockBytes)
}

// GetInternalShardBlockByRound wil return a shard block by round
func (rbp *rawBlockProcessor) GetInternalShardBlockByRound(format common.OutportFormat, round uint64) (interface{}, error) {
	_, blockBytes, err := rbp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	return rbp.convertShardBlockBytesByOutportFormat(format, blockBytes)
}

func (rbp *rawBlockProcessor) convertShardBlockBytesToInternalBlock(blockBytes []byte) (*block.Header, error) {
	blockHeader := &block.Header{}
	err := rbp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	return blockHeader, nil
}

func (rbp *rawBlockProcessor) convertShardBlockBytesByOutportFormat(format common.OutportFormat, blockBytes []byte) (interface{}, error) {
	switch format {
	case common.Internal:
		return rbp.convertShardBlockBytesToInternalBlock(blockBytes)
	case common.Proto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutportFormat
	}
}

// GetInternalMetaBlockByNonce wil return a meta block by nonce
func (rbp *rawBlockProcessor) GetInternalMetaBlockByNonce(format common.OutportFormat, nonce uint64) (interface{}, error) {
	storerUnit := dataRetriever.MetaHdrNonceHashDataUnit

	nonceToByteSlice := rbp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := rbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	blockBytes, err := rbp.getFromStorer(dataRetriever.MetaBlockUnit, headerHash)
	if err != nil {
		return nil, err
	}

	return rbp.convertMetaBlockBytesByOutportFormat(format, blockBytes)
}

// GetInternalMetaBlockByHash wil return a meta block by hash
func (rbp *rawBlockProcessor) GetInternalMetaBlockByHash(format common.OutportFormat, hash []byte) (interface{}, error) {
	blockBytes, err := rbp.getFromStorer(dataRetriever.MetaBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	return rbp.convertMetaBlockBytesByOutportFormat(format, blockBytes)
}

// GetInternalMetaBlockByRound wil return a meta block by round
func (rbp *rawBlockProcessor) GetInternalMetaBlockByRound(format common.OutportFormat, round uint64) (interface{}, error) {
	_, blockBytes, err := rbp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	return rbp.convertMetaBlockBytesByOutportFormat(format, blockBytes)
}

func (rbp *rawBlockProcessor) convertMetaBlockBytesToInternalBlock(blockBytes []byte) (*block.MetaBlock, error) {
	blockHeader := &block.MetaBlock{}
	err := rbp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	return blockHeader, nil
}

func (rbp *rawBlockProcessor) convertMetaBlockBytesByOutportFormat(format common.OutportFormat, blockBytes []byte) (interface{}, error) {
	switch format {
	case common.Internal:
		return rbp.convertMetaBlockBytesToInternalBlock(blockBytes)
	case common.Proto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutportFormat
	}
}
