package blockAPI

import (
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type metaAPIBlockProcessor struct {
	*baseAPIBockProcessor
}

// NewMetaApiBlockProcessor will create a new instance of meta api block processor
func NewMetaApiBlockProcessor(arg *APIBlockProcessorArg) *metaAPIBlockProcessor {
	hasDbLookupExtensions := arg.HistoryRepo.IsEnabled()

	return &metaAPIBlockProcessor{
		baseAPIBockProcessor: &baseAPIBockProcessor{
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

// GetBlockByHash will return a shard APIBlock by hash
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
			miniblockAPI.Transactions = mbp.getTxsByMb(&miniBlockCopy, headerEpoch)
		}

		miniblocks = append(miniblocks, miniblockAPI)
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
