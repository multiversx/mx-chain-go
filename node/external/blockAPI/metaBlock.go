package blockAPI

import (
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
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
			apiTransactionHandler:    arg.APITransactionHandler,
			txStatusComputer:         arg.StatusComputer,
			hasher:                   arg.Hasher,
			addressPubKeyConverter:   arg.AddressPubkeyConverter,
			emptyReceiptsHash:        emptyReceiptsHash,
			logsFacade:               arg.LogsFacade,
			receiptsRepository:       arg.ReceiptsRepository,
		},
	}
}

// GetBlockByNonce wil return a meta APIBlock by nonce
func (mbp *metaAPIBlockProcessor) GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
	storerUnit := dataRetriever.MetaHdrNonceHashDataUnit

	nonceToByteSlice := mbp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := mbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	// if genesis block, get the nonce key corresponding to the altered block
	if nonce == 0 {
		nonceToByteSlice = append(nonceToByteSlice, []byte(common.GenesisStorageSuffix)...)
	}

	alteredHeaderHash, err := mbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	blockBytes, err := mbp.getFromStorer(dataRetriever.MetaBlockUnit, alteredHeaderHash)
	if err != nil {
		return nil, err
	}

	return mbp.convertMetaBlockBytesToAPIBlock(headerHash, blockBytes, options)
}

// GetBlockByHash will return a meta APIBlock by hash
func (mbp *metaAPIBlockProcessor) GetBlockByHash(hash []byte, options api.BlockQueryOptions) (*api.Block, error) {
	blockBytes, err := mbp.getFromStorer(dataRetriever.MetaBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	blockHeader := &block.MetaBlock{}
	err = mbp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	// if genesis block, get the altered block bytes
	if blockHeader.GetNonce() == 0 {
		alteredHash := createAlteredBlockHash(hash)
		blockBytes, err = mbp.getFromStorer(dataRetriever.MetaBlockUnit, alteredHash)
		if err != nil {
			return nil, err
		}
	}

	blockAPI, err := mbp.convertMetaBlockBytesToAPIBlock(hash, blockBytes, options)
	if err != nil {
		return nil, err
	}

	return mbp.computeStatusAndPutInBlock(blockAPI, dataRetriever.MetaHdrNonceHashDataUnit)
}

// GetBlockByRound will return a meta APIBlock by round
func (mbp *metaAPIBlockProcessor) GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error) {
	headerHash, blockBytes, err := mbp.getBlockHeaderHashAndBytesByRound(round, dataRetriever.MetaBlockUnit)
	if err != nil {
		return nil, err
	}

	return mbp.convertMetaBlockBytesToAPIBlock(headerHash, blockBytes, options)
}

func (mbp *metaAPIBlockProcessor) convertMetaBlockBytesToAPIBlock(hash []byte, blockBytes []byte, options api.BlockQueryOptions) (*api.Block, error) {
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
		if options.WithTransactions {
			miniBlockCopy := mb
			err = mbp.getAndAttachTxsToMb(&miniBlockCopy, headerEpoch, miniblockAPI, options)
			if err != nil {
				return nil, err
			}
		}

		miniblocks = append(miniblocks, miniblockAPI)
	}

	intraMb, err := mbp.getIntrashardMiniblocksFromReceiptsStorage(blockHeader, hash, options)
	if err != nil {
		return nil, err
	}

	if len(intraMb) > 0 {
		miniblocks = append(miniblocks, intraMb...)
	}

	miniblocks = filterOutDuplicatedMiniblocks(miniblocks)

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

	apiMetaBlock := &api.Block{
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
		StateRootHash:          hex.EncodeToString(blockHeader.RootHash),
		Status:                 BlockStatusOnChain,
	}

	addScheduledInfoInBlock(blockHeader, apiMetaBlock)
	addStartOfEpochInfoInBlock(blockHeader, apiMetaBlock)

	return apiMetaBlock, nil
}

func addStartOfEpochInfoInBlock(metaBlock *block.MetaBlock, apiBlock *api.Block) {
	if !metaBlock.IsStartOfEpochBlock() {
		return
	}

	epochStartEconomics := metaBlock.EpochStart.Economics

	apiBlock.EpochStartInfo = &api.EpochStartInfo{
		TotalSupply:                      epochStartEconomics.TotalSupply.String(),
		TotalToDistribute:                epochStartEconomics.TotalToDistribute.String(),
		TotalNewlyMinted:                 epochStartEconomics.TotalNewlyMinted.String(),
		RewardsPerBlock:                  epochStartEconomics.RewardsPerBlock.String(),
		RewardsForProtocolSustainability: epochStartEconomics.RewardsForProtocolSustainability.String(),
		NodePrice:                        epochStartEconomics.NodePrice.String(),
		PrevEpochStartRound:              epochStartEconomics.PrevEpochStartRound,
		PrevEpochStartHash:               hex.EncodeToString(epochStartEconomics.PrevEpochStartHash),
	}

	if len(metaBlock.EpochStart.LastFinalizedHeaders) == 0 {
		return
	}

	epochStartShardsData := metaBlock.EpochStart.LastFinalizedHeaders
	apiBlock.EpochStartShardsData = make([]*api.EpochStartShardData, 0, len(epochStartShardsData))
	for _, epochStartShardData := range epochStartShardsData {
		addEpochStartShardDataForMeta(epochStartShardData, apiBlock)
	}
}

func addEpochStartShardDataForMeta(epochStartShardData block.EpochStartShardData, apiBlock *api.Block) {
	shardData := &api.EpochStartShardData{
		ShardID:               epochStartShardData.ShardID,
		Epoch:                 epochStartShardData.Epoch,
		Round:                 epochStartShardData.Round,
		Nonce:                 epochStartShardData.Nonce,
		HeaderHash:            hex.EncodeToString(epochStartShardData.HeaderHash),
		RootHash:              hex.EncodeToString(epochStartShardData.RootHash),
		ScheduledRootHash:     hex.EncodeToString(epochStartShardData.ScheduledRootHash),
		FirstPendingMetaBlock: hex.EncodeToString(epochStartShardData.FirstPendingMetaBlock),
		LastFinishedMetaBlock: hex.EncodeToString(epochStartShardData.LastFinishedMetaBlock),
	}

	apiBlock.EpochStartShardsData = append(apiBlock.EpochStartShardsData, shardData)

	if len(epochStartShardData.PendingMiniBlockHeaders) == 0 {
		return
	}

	shardData.PendingMiniBlockHeaders = make([]*api.MiniBlock, 0, len(epochStartShardData.PendingMiniBlockHeaders))
	for _, pendingMb := range epochStartShardData.PendingMiniBlockHeaders {
		shardData.PendingMiniBlockHeaders = append(shardData.PendingMiniBlockHeaders, &api.MiniBlock{
			Hash:             hex.EncodeToString(pendingMb.Hash),
			SourceShard:      pendingMb.SenderShardID,
			DestinationShard: pendingMb.ReceiverShardID,
			Type:             pendingMb.Type.String(),
		})
	}
}

// IsInterfaceNil returns true if underlying object is nil
func (mbp *metaAPIBlockProcessor) IsInterfaceNil() bool {
	return mbp == nil
}
