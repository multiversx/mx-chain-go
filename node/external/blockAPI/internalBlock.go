package blockAPI

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
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
			apiTransactionHandler:    arg.APITransactionHandler,
			txStatusComputer:         arg.StatusComputer,
			hasher:                   arg.Hasher,
			addressPubKeyConverter:   arg.AddressPubkeyConverter,
			emptyReceiptsHash:        emptyReceiptsHash,
			enableEpochsHandler:      arg.EnableEpochsHandler,
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
	return process.UnmarshalShardHeader(ibp.marshalizer, blockBytes)
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

	storer, err := ibp.store.GetStorer(dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	blockBytes, err := storer.GetFromEpoch([]byte(epochStartIdentifier), epoch)
	if err != nil {
		return nil, err
	}

	return ibp.convertMetaBlockBytesByOutputFormat(format, blockBytes)
}

// GetInternalStartOfEpochValidatorsInfo wil return the epoch start validators info for the provided epoch
func (ibp *internalBlockProcessor) GetInternalStartOfEpochValidatorsInfo(epoch uint32) ([]*state.ShardValidatorInfo, error) {
	if ibp.selfShardID != core.MetachainShardId {
		return nil, ErrMetachainOnlyEndpoint
	}

	storer, err := ibp.store.GetStorer(dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	blockBytes, err := storer.GetFromEpoch([]byte(epochStartIdentifier), epoch)
	if err != nil {
		return nil, err
	}

	metaBlock := &block.MetaBlock{}
	err = ibp.marshalizer.Unmarshal(metaBlock, blockBytes)
	if err != nil {
		return nil, err
	}

	validatorsInfo, err := ibp.getAllValidatorsInfo(metaBlock)
	if err != nil {
		return nil, err
	}

	return validatorsInfo, nil
}

func (ibp *internalBlockProcessor) getAllValidatorsInfo(metaBlock data.HeaderHandler) ([]*state.ShardValidatorInfo, error) {
	allValidatorInfo := make([]*state.ShardValidatorInfo, 0)
	for _, miniBlockHeader := range metaBlock.GetMiniBlockHeaderHandlers() {
		hash := miniBlockHeader.GetHash()

		miniBlock, err := ibp.getMiniBlockByHash(hash, metaBlock.GetEpoch())
		if err != nil {
			return nil, err
		}

		if miniBlock.Type != block.PeerBlock {
			continue
		}

		validatorInfo, err := ibp.getValidatorsInfo(miniBlock, metaBlock.GetEpoch())
		if err != nil {
			return nil, err
		}

		allValidatorInfo = append(allValidatorInfo, validatorInfo...)
	}

	return allValidatorInfo, nil
}

func (ibp *internalBlockProcessor) getValidatorsInfo(
	miniBlock *block.MiniBlock,
	epoch uint32,
) ([]*state.ShardValidatorInfo, error) {
	validatorsInfoBytes := make([][]byte, 0)
	if epoch >= ibp.enableEpochsHandler.RefactorPeersMiniBlocksEnableEpoch() {
		validatorsInfoBuff, err := ibp.store.GetAll(dataRetriever.UnsignedTransactionUnit, miniBlock.TxHashes)
		if err != nil {
			return nil, err
		}

		for _, validatorInfoBuff := range validatorsInfoBuff {
			validatorsInfoBytes = append(validatorsInfoBytes, validatorInfoBuff)
		}
	} else {
		validatorsInfoBytes = miniBlock.TxHashes
	}

	validatorsInfo := make([]*state.ShardValidatorInfo, 0)
	for _, validatorInfoBytes := range validatorsInfoBytes {
		shardValidatorInfo := &state.ShardValidatorInfo{}
		err := ibp.marshalizer.Unmarshal(shardValidatorInfo, validatorInfoBytes)
		if err != nil {
			return nil, err
		}

		validatorsInfo = append(validatorsInfo, shardValidatorInfo)
	}

	return validatorsInfo, nil
}

func (ibp *internalBlockProcessor) getMiniBlockByHash(hash []byte, epoch uint32) (*block.MiniBlock, error) {
	storer, err := ibp.store.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return nil, err
	}

	blockBytes, err := storer.GetFromEpoch(hash, epoch)
	if err != nil {
		return nil, err
	}

	miniBlock := &block.MiniBlock{}
	err = ibp.marshalizer.Unmarshal(miniBlock, blockBytes)
	if err != nil {
		return nil, err
	}

	return miniBlock, nil
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
	storer, err := ibp.store.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return nil, err
	}

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
