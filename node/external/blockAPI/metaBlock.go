package blockAPI

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

type metaAPIBlockProcessor struct {
	*baseAPIBlockProcessor
}

// newMetaApiBlockProcessor will create a new instance of meta api block processor
func newMetaApiBlockProcessor(arg *ArgAPIBlockProcessor, emptyReceiptsHash []byte) *metaAPIBlockProcessor {
	hasDbLookupExtensions := arg.HistoryRepo.IsEnabled()

	return &metaAPIBlockProcessor{
		baseAPIBlockProcessor: &baseAPIBlockProcessor{
			hasDbLookupExtensions:        hasDbLookupExtensions,
			selfShardID:                  arg.SelfShardID,
			store:                        arg.Store,
			marshalizer:                  arg.Marshalizer,
			uint64ByteSliceConverter:     arg.Uint64ByteSliceConverter,
			historyRepo:                  arg.HistoryRepo,
			apiTransactionHandler:        arg.APITransactionHandler,
			txStatusComputer:             arg.StatusComputer,
			hasher:                       arg.Hasher,
			addressPubKeyConverter:       arg.AddressPubkeyConverter,
			emptyReceiptsHash:            emptyReceiptsHash,
			logsFacade:                   arg.LogsFacade,
			receiptsRepository:           arg.ReceiptsRepository,
			alteredAccountsProvider:      arg.AlteredAccountsProvider,
			accountsRepository:           arg.AccountsRepository,
			scheduledTxsExecutionHandler: arg.ScheduledTxsExecutionHandler,
			enableEpochsHandler:          arg.EnableEpochsHandler,
			proofsPool:                   arg.ProofsPool,
			blockchain:                   arg.BlockChain,
		},
	}
}

// GetBlockByNonce wil return a meta APIBlock by nonce
func (mbp *metaAPIBlockProcessor) GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
	if !mbp.isBlockNonceInStorage(nonce) {
		return nil, errBlockNotFound
	}

	headerHash, blockBytes, err := mbp.getBlockHashAndBytesByNonce(nonce)
	if err != nil {
		return nil, err
	}

	return mbp.convertMetaBlockBytesToAPIBlock(headerHash, blockBytes, options)
}

func (mbp *metaAPIBlockProcessor) getBlockHashAndBytesByNonce(nonce uint64) ([]byte, []byte, error) {
	storerUnit := dataRetriever.MetaHdrNonceHashDataUnit

	nonceToByteSlice := mbp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := mbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	// if genesis block, get the nonce key corresponding to the altered block
	if nonce == 0 {
		nonceToByteSlice = append(nonceToByteSlice, []byte(common.GenesisStorageSuffix)...)
	}

	alteredHeaderHash, err := mbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	blockBytes, err := mbp.getFromStorer(dataRetriever.MetaBlockUnit, alteredHeaderHash)
	if err != nil {
		return nil, nil, err
	}

	return headerHash, blockBytes, nil
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

// GetAlteredAccountsForBlock returns the altered accounts for the desired meta block
func (mbp *metaAPIBlockProcessor) GetAlteredAccountsForBlock(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
	headerHash, blockBytes, err := mbp.getHashAndBlockBytesFromStorer(options.GetBlockParameters)
	if err != nil {
		return nil, err
	}

	apiBlock, err := mbp.convertMetaBlockBytesToAPIBlock(headerHash, blockBytes, api.BlockQueryOptions{WithTransactions: true, WithLogs: true})
	if err != nil {
		return nil, err
	}

	return mbp.apiBlockToAlteredAccounts(apiBlock, options)
}

func (mbp *metaAPIBlockProcessor) getHashAndBlockBytesFromStorer(params api.GetBlockParameters) ([]byte, []byte, error) {
	switch params.RequestType {
	case api.BlockFetchTypeByHash:
		return mbp.getHashAndBlockBytesFromStorerByHash(params)
	case api.BlockFetchTypeByNonce:
		return mbp.getHashAndBlockBytesFromStorerByNonce(params)
	default:
		return nil, nil, errUnknownBlockRequestType
	}
}

func (mbp *metaAPIBlockProcessor) getHashAndBlockBytesFromStorerByHash(params api.GetBlockParameters) ([]byte, []byte, error) {
	headerBytes, err := mbp.getFromStorer(dataRetriever.MetaBlockUnit, params.Hash)
	return params.Hash, headerBytes, err
}

func (mbp *metaAPIBlockProcessor) getHashAndBlockBytesFromStorerByNonce(params api.GetBlockParameters) ([]byte, []byte, error) {
	storerUnit := dataRetriever.MetaHdrNonceHashDataUnit

	nonceToByteSlice := mbp.uint64ByteSliceConverter.ToByteSlice(params.Nonce)
	headerHash, err := mbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	headerBytes, err := mbp.getFromStorer(dataRetriever.MetaBlockUnit, headerHash)
	return headerHash, headerBytes, err
}

func (mbp *metaAPIBlockProcessor) convertMetaBlockBytesToAPIBlock(hash []byte, blockBytes []byte, options api.BlockQueryOptions) (*api.Block, error) {
	blockHeader, err := process.UnmarshalMetaHeader(mbp.marshalizer, blockBytes)
	if err != nil {
		return nil, err
	}

	notarizedBlocks := make([]*api.NotarizedBlock, 0, len(blockHeader.GetShardInfoHandlers()))
	for _, shardData := range blockHeader.GetShardInfoHandlers() {
		notarizedBlock := &api.NotarizedBlock{
			Hash:  hex.EncodeToString(shardData.GetHeaderHash()),
			Nonce: shardData.GetNonce(),
			Round: shardData.GetRound(),
			Shard: shardData.GetShardID(),
		}

		notarizedBlocks = append(notarizedBlocks, notarizedBlock)
	}

	apiMetaBlock := &api.Block{
		Nonce:           blockHeader.GetNonce(),
		Round:           blockHeader.GetRound(),
		Epoch:           blockHeader.GetEpoch(),
		Shard:           core.MetachainShardId,
		Hash:            hex.EncodeToString(hash),
		PrevBlockHash:   hex.EncodeToString(blockHeader.GetPrevHash()),
		NotarizedBlocks: notarizedBlocks,
		Timestamp:       int64(blockHeader.GetTimeStamp()),
		TimestampMs:     int64(common.ConvertTimeStampSecToMs(blockHeader.GetTimeStamp())),
		Status:          BlockStatusOnChain,
		PubKeyBitmap:    hex.EncodeToString(blockHeader.GetPubKeysBitmap()),
		Signature:       hex.EncodeToString(blockHeader.GetSignature()),
		LeaderSignature: hex.EncodeToString(blockHeader.GetLeaderSignature()),
		ChainID:         string(blockHeader.GetChainID()),
		SoftwareVersion: hex.EncodeToString(blockHeader.GetSoftwareVersion()),
		ReceiptsHash:    hex.EncodeToString(blockHeader.GetReceiptsHash()),
		Reserved:        blockHeader.GetReserved(),
		RandSeed:        hex.EncodeToString(blockHeader.GetRandSeed()),
		PrevRandSeed:    hex.EncodeToString(blockHeader.GetPrevRandSeed()),
	}

	if !blockHeader.IsHeaderV3() {
		err = mbp.addMbsAndNumTxsV1(apiMetaBlock, blockHeader, hash, options)
	} else {
		//async execution
		err = mbp.addMbsAndNumTxsAsyncExecutionBasedOnExecutionResult(apiMetaBlock, blockHeader, hash, options)
	}
	if err != nil {
		return nil, err
	}

	addExtraFieldsLastExecutionResultMeta(blockHeader, apiMetaBlock)
	addScheduledInfoInBlock(blockHeader, apiMetaBlock)
	addStartOfEpochInfoInBlock(blockHeader, apiMetaBlock)

	addExecutionResultsAndLastExecutionResults(blockHeader, apiMetaBlock)

	err = mbp.addProof(hash, blockHeader, apiMetaBlock)
	if err != nil {
		return nil, err
	}

	return apiMetaBlock, nil
}

func (mbp *metaAPIBlockProcessor) getMbsAndNumTxs(blockHeader data.MetaHeaderHandler, options api.BlockQueryOptions) ([]*api.MiniBlock, uint32, error) {
	numOfTxs := uint32(0)
	miniblocks := make([]*api.MiniBlock, 0)
	for _, mb := range blockHeader.GetMiniBlockHeaderHandlers() {
		mbType := block.Type(mb.GetTypeInt32())
		if mbType == block.PeerBlock {
			continue
		}

		numOfTxs += mb.GetTxCount()

		miniblockAPI := &api.MiniBlock{
			Hash:             hex.EncodeToString(mb.GetHash()),
			Type:             mbType.String(),
			SourceShard:      mb.GetSenderShardID(),
			DestinationShard: mb.GetReceiverShardID(),
		}
		if options.WithTransactions {
			miniBlockCopy := mb
			err := mbp.getAndAttachTxsToMb(miniBlockCopy, blockHeader, miniblockAPI, options)
			if err != nil {
				return nil, 0, err
			}
		}

		miniblocks = append(miniblocks, miniblockAPI)
	}

	return miniblocks, numOfTxs, nil
}

func (mbp *metaAPIBlockProcessor) addMbsAndNumTxsV1(apiBlock *api.Block, blockHeader data.MetaHeaderHandler, headerHash []byte, options api.BlockQueryOptions) error {
	miniblocks, numOfTxs, err := mbp.getMbsAndNumTxs(blockHeader, options)
	if err != nil {
		return err
	}

	intraMb, err := mbp.getIntrashardMiniblocksFromReceiptsStorage(blockHeader, headerHash, options)
	if err != nil {
		return err
	}

	if len(intraMb) > 0 {
		miniblocks = append(miniblocks, intraMb...)
	}

	miniblocks = filterOutDuplicatedMiniblocks(miniblocks)

	apiBlock.MiniBlocks = miniblocks
	apiBlock.NumTxs = numOfTxs

	return nil
}

func addExtraFieldsLastExecutionResultMeta(metaHeader data.MetaHeaderHandler, apiBlock *api.Block) {
	if !metaHeader.IsHeaderV3() {
		apiBlock.AccumulatedFees = metaHeader.GetAccumulatedFees().String()
		apiBlock.DeveloperFees = metaHeader.GetDeveloperFees().String()

		apiBlock.AccumulatedFeesInEpoch = metaHeader.GetAccumulatedFeesInEpoch().String()
		apiBlock.DeveloperFeesInEpoch = metaHeader.GetDevFeesInEpoch().String()
		apiBlock.StateRootHash = hex.EncodeToString(metaHeader.GetRootHash())

		return
	}

	metaExecutionResult, ok := metaHeader.GetLastExecutionResultHandler().(data.LastMetaExecutionResultHandler)
	if !ok {
		log.Warn("mbp.addExtraFields cannot cast last execution result handler to data.LastMetaExecutionResultHandler")
		return
	}

	apiBlock.AccumulatedFeesInEpoch = metaExecutionResult.GetExecutionResultHandler().GetAccumulatedFeesInEpoch().String()
	apiBlock.DeveloperFeesInEpoch = metaExecutionResult.GetExecutionResultHandler().GetDevFeesInEpoch().String()
	apiBlock.StateRootHash = hex.EncodeToString(metaExecutionResult.GetExecutionResultHandler().GetRootHash())
}

func addStartOfEpochInfoInBlock(metaHeader data.MetaHeaderHandler, apiBlock *api.Block) {
	if !metaHeader.IsStartOfEpochBlock() {
		return
	}

	epochStartEconomics := metaHeader.GetEpochStartHandler().GetEconomicsHandler()

	apiBlock.EpochStartInfo = &api.EpochStartInfo{
		TotalSupply:                      epochStartEconomics.GetTotalSupply().String(),
		TotalToDistribute:                epochStartEconomics.GetTotalToDistribute().String(),
		TotalNewlyMinted:                 epochStartEconomics.GetTotalNewlyMinted().String(),
		RewardsPerBlock:                  epochStartEconomics.GetRewardsPerBlock().String(),
		RewardsForProtocolSustainability: epochStartEconomics.GetRewardsForProtocolSustainability().String(),
		NodePrice:                        epochStartEconomics.GetNodePrice().String(),
		PrevEpochStartRound:              epochStartEconomics.GetPrevEpochStartRound(),
		PrevEpochStartHash:               hex.EncodeToString(epochStartEconomics.GetPrevEpochStartHash()),
	}

	lastFinalizedHeaderHandlers := metaHeader.GetEpochStartHandler().GetLastFinalizedHeaderHandlers()
	if len(lastFinalizedHeaderHandlers) == 0 {
		return
	}

	apiBlock.EpochStartShardsData = make([]*api.EpochStartShardData, 0, len(lastFinalizedHeaderHandlers))
	for _, epochStartShardData := range lastFinalizedHeaderHandlers {
		addEpochStartShardDataForMeta(epochStartShardData, apiBlock)
	}
}

func addEpochStartShardDataForMeta(epochStartShardData data.EpochStartShardDataHandler, apiBlock *api.Block) {
	shardData := &api.EpochStartShardData{
		ShardID:               epochStartShardData.GetShardID(),
		Epoch:                 epochStartShardData.GetEpoch(),
		Round:                 epochStartShardData.GetRound(),
		Nonce:                 epochStartShardData.GetNonce(),
		HeaderHash:            hex.EncodeToString(epochStartShardData.GetHeaderHash()),
		RootHash:              hex.EncodeToString(epochStartShardData.GetRootHash()),
		FirstPendingMetaBlock: hex.EncodeToString(epochStartShardData.GetFirstPendingMetaBlock()),
		LastFinishedMetaBlock: hex.EncodeToString(epochStartShardData.GetLastFinishedMetaBlock()),
	}

	epochStartData, ok := epochStartShardData.(*block.EpochStartShardData)
	if ok {
		shardData.ScheduledRootHash = hex.EncodeToString(epochStartData.ScheduledRootHash)
	}

	apiBlock.EpochStartShardsData = append(apiBlock.EpochStartShardsData, shardData)

	pendingMiniBlockHeaders := epochStartShardData.GetPendingMiniBlockHeaderHandlers()
	if len(pendingMiniBlockHeaders) == 0 {
		return
	}

	shardData.PendingMiniBlockHeaders = make([]*api.MiniBlock, 0, len(pendingMiniBlockHeaders))
	for _, pendingMb := range pendingMiniBlockHeaders {
		pendingMbType := block.Type(pendingMb.GetTypeInt32())

		shardData.PendingMiniBlockHeaders = append(shardData.PendingMiniBlockHeaders, &api.MiniBlock{
			Hash:             hex.EncodeToString(pendingMb.GetHash()),
			SourceShard:      pendingMb.GetSenderShardID(),
			DestinationShard: pendingMb.GetReceiverShardID(),
			Type:             pendingMbType.String(),
		})
	}
}

// IsInterfaceNil returns true if underlying object is nil
func (mbp *metaAPIBlockProcessor) IsInterfaceNil() bool {
	return mbp == nil
}
