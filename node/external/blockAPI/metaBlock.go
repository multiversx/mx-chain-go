package blockAPI

import (
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type metaAPIBlockProcessor struct {
	*baseAPIBlockProcessor
}

// newMetaApiBlockProcessor will create a new instance of meta api block processor
func newMetaApiBlockProcessor(arg *ArgAPIBlockProcessor, emptyReceiptsHash []byte) *metaAPIBlockProcessor {
	hasDbLookupExtensions := arg.HistoryRepo.IsEnabled()

	return &metaAPIBlockProcessor{
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

// GetBlockByNonce wil return a meta APIBlock by nonce
func (mbp *metaAPIBlockProcessor) GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error) {
	storerUnit := dataRetriever.MetaHdrNonceHashDataUnit

	nonceToByteSlice := mbp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := mbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	blockBytes, err := mbp.getFromStorer(dataRetriever.MetaBlockUnit, headerHash)
	if err != nil {
		return nil, err
	}

	return mbp.convertMetaBlockBytesToAPIBlock(headerHash, blockBytes, withTxs)
}

// GetBlockByHash will return a meta APIBlock by hash
func (mbp *metaAPIBlockProcessor) GetBlockByHash(hash []byte, withTxs bool) (*api.Block, error) {
	blockBytes, err := mbp.getFromStorer(dataRetriever.MetaBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	blockAPI, err := mbp.convertMetaBlockBytesToAPIBlock(hash, blockBytes, withTxs)
	if err != nil {
		return nil, err
	}

	return mbp.computeStatusAndPutInBlock(blockAPI, dataRetriever.MetaHdrNonceHashDataUnit)
}

// GetBlockByRound will return a meta APIBlock by round
func (mbp *metaAPIBlockProcessor) GetBlockByRound(round uint64, withTxs bool) (*api.Block, error) {
	headerHash, blockBytes, err := mbp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	return mbp.convertMetaBlockBytesToAPIBlock(headerHash, blockBytes, withTxs)
}

func (mbp *metaAPIBlockProcessor) convertMetaBlockBytesToAPIBlock(hash []byte, blockBytes []byte, withTxs bool) (*api.Block, error) {
	blockHeader := &block.MetaBlock{}
	err := mbp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	headerEpoch := blockHeader.Epoch

	numOfTxs := uint32(0)
	miniblocks := make([]*api.MiniBlock, 0)
	for _, mb := range blockHeader.MiniBlockHeaders {
		if mb.Type == block.PeerBlock {
			continue
		}

		numOfTxs += mb.TxCount

		miniblockAPI := &api.MiniBlock{
			Hash:             hex.EncodeToString(mb.Hash),
			Type:             mb.Type.String(),
			SourceShard:      mb.SenderShardID,
			DestinationShard: mb.ReceiverShardID,
		}
		if withTxs {
			miniBlockCopy := mb
			mbp.getAndAttachTxsToMb(&miniBlockCopy, headerEpoch, miniblockAPI)
		}

		miniblocks = append(miniblocks, miniblockAPI)
	}

	intraMb := mbp.getIntraMiniblocks(blockHeader.GetReceiptsHash(), headerEpoch, withTxs)
	if len(intraMb) > 0 {
		miniblocks = append(miniblocks, intraMb...)
	}

	notarizedBlocks := make([]*api.NotarizedBlock, 0, len(blockHeader.ShardInfo))
	for _, shardData := range blockHeader.ShardInfo {
		notarizedBlock := &api.NotarizedBlock{
			Hash:  hex.EncodeToString(shardData.HeaderHash),
			Nonce: shardData.Nonce,
			Round: shardData.Round,
			Shard: shardData.ShardID,
		}

		notarizedBlocks = append(notarizedBlocks, notarizedBlock)
	}

	metaBlock := &api.Block{
		Nonce:                  blockHeader.Nonce,
		Round:                  blockHeader.Round,
		Epoch:                  blockHeader.Epoch,
		Shard:                  core.MetachainShardId,
		Hash:                   hex.EncodeToString(hash),
		PrevBlockHash:          hex.EncodeToString(blockHeader.PrevHash),
		NumTxs:                 numOfTxs,
		NotarizedBlocks:        notarizedBlocks,
		MiniBlocks:             miniblocks,
		AccumulatedFees:        blockHeader.AccumulatedFees.String(),
		DeveloperFees:          blockHeader.DeveloperFees.String(),
		AccumulatedFeesInEpoch: blockHeader.AccumulatedFeesInEpoch.String(),
		DeveloperFeesInEpoch:   blockHeader.DevFeesInEpoch.String(),
		Timestamp:              time.Duration(blockHeader.GetTimeStamp()),
		Status:                 BlockStatusOnChain,
	}

	if blockHeader.IsStartOfEpochBlock() {
		epochStartEconomics := blockHeader.EpochStart.Economics

		metaBlock.EpochStartInfo = &api.EpochStartInfo{
			TotalSupply:                      epochStartEconomics.TotalSupply.String(),
			TotalToDistribute:                epochStartEconomics.TotalToDistribute.String(),
			TotalNewlyMinted:                 epochStartEconomics.TotalNewlyMinted.String(),
			RewardsPerBlock:                  epochStartEconomics.RewardsPerBlock.String(),
			RewardsForProtocolSustainability: epochStartEconomics.RewardsForProtocolSustainability.String(),
			NodePrice:                        epochStartEconomics.NodePrice.String(),
			PrevEpochStartRound:              epochStartEconomics.PrevEpochStartRound,
			PrevEpochStartHash:               hex.EncodeToString(epochStartEconomics.PrevEpochStartHash),
		}
	}

	return metaBlock, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (mbp *metaAPIBlockProcessor) IsInterfaceNil() bool {
	return mbp == nil
}
