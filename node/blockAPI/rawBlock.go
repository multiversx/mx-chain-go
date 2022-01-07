package blockAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

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

// GetRawShardBlockByNonce wil return a shard block by nonce as raw data
func (rbp *rawBlockProcessor) GetRawShardBlockByNonce(nonce uint64) ([]byte, error) {
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

	return blockBytes, nil
}

// GetRawShardBlockByHash wil return a shard block by hash as raw data
func (rbp *rawBlockProcessor) GetRawShardBlockByHash(hash []byte) ([]byte, error) {
	blockBytes, err := rbp.getFromStorer(dataRetriever.BlockHeaderUnit, hash)
	if err != nil {
		return nil, err
	}

	return blockBytes, nil
}

// GetRawShardBlockByRoud wil return a shard block by round as raw data
func (rbp *rawBlockProcessor) GetRawShardBlockByRound(round uint64) ([]byte, error) {
	_, blockBytes, err := rbp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	return blockBytes, nil
}

// GetRawMetaBlockByNonce wil return a meta block by nonce as raw data
func (rbp *rawBlockProcessor) GetRawMetaBlockByNonce(nonce uint64) ([]byte, error) {
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

	return blockBytes, nil
}

// GetRawMetaBlockByHash wil return a meta block by hash as raw data
func (rbp *rawBlockProcessor) GetRawMetaBlockByHash(hash []byte) ([]byte, error) {
	blockBytes, err := rbp.getFromStorer(dataRetriever.MetaBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	return blockBytes, nil
}

// GetRawMetaBlockByRound wil return a meta block by round as raw data
func (rbp *rawBlockProcessor) GetRawMetaBlockByRound(round uint64) ([]byte, error) {
	_, blockBytes, err := rbp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	return blockBytes, nil
}

// GetInternalShardBlockByNonce wil return a shard block by nonce
func (rbp *rawBlockProcessor) GetInternalShardBlockByNonce(nonce uint64) (*block.Header, error) {
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

	return rbp.convertShardBlockBytesToInternalBlock(blockBytes)
}

// GetInternalShardBlockByHash wil return a shard block by hash
func (rbp *rawBlockProcessor) GetInternalShardBlockByHash(hash []byte) (*block.Header, error) {
	blockBytes, err := rbp.getFromStorer(dataRetriever.BlockHeaderUnit, hash)
	if err != nil {
		return nil, err
	}

	return rbp.convertShardBlockBytesToInternalBlock(blockBytes)
}

// GetInternalShardBlockByRound wil return a shard block by round
func (rbp *rawBlockProcessor) GetInternalShardBlockByRound(round uint64) (*block.Header, error) {
	_, blockBytes, err := rbp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.BlockHeaderUnit)
	if err != nil {
		return nil, err
	}

	return rbp.convertShardBlockBytesToInternalBlock(blockBytes)
}

func (rbp *rawBlockProcessor) convertShardBlockBytesToInternalBlock(blockBytes []byte) (*block.Header, error) {
	blockHeader := &block.Header{}
	err := rbp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	return blockHeader, nil
}

// GetInternalMetaBlockByNonce wil return a meta block by nonce
func (rbp *rawBlockProcessor) GetInternalMetaBlockByNonce(nonce uint64) (*block.MetaBlock, error) {
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

	return rbp.convertMetaBlockBytesToInternalBlock(blockBytes)
}

// GetInternalMetaBlockByHash wil return a meta block by hash
func (rbp *rawBlockProcessor) GetInternalMetaBlockByHash(hash []byte) (*block.MetaBlock, error) {
	blockBytes, err := rbp.getFromStorer(dataRetriever.MetaBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	return rbp.convertMetaBlockBytesToInternalBlock(blockBytes)
}

// GetInternalMetaBlockByRound wil return a meta block by round
func (rbp *rawBlockProcessor) GetInternalMetaBlockByRound(round uint64) (*block.MetaBlock, error) {
	_, blockBytes, err := rbp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	return rbp.convertMetaBlockBytesToInternalBlock(blockBytes)
}

func (rbp *rawBlockProcessor) convertMetaBlockBytesToInternalBlock(blockBytes []byte) (*block.MetaBlock, error) {
	blockHeader := &block.MetaBlock{}
	err := rbp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	return blockHeader, nil
}
