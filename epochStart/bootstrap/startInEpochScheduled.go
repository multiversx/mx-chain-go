package bootstrap

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/update"
)

type startInEpochWithScheduledDataSyncer struct {
	scheduledTxsHandler       process.ScheduledTxsExecutionHandler
	scheduledHeadersSyncer    epochStart.HeadersByHashSyncer
	scheduledMiniBlocksSyncer epochStart.PendingMiniBlocksSyncHandler
	txSyncer                  update.TransactionsSyncHandler
	scheduledEnableEpoch      uint32
}

func newStartInEpochShardHeaderDataSyncerWithScheduled(
	scheduledTxsHandler process.ScheduledTxsExecutionHandler,
	headersSyncer epochStart.HeadersByHashSyncer,
	miniBlocksSyncer epochStart.PendingMiniBlocksSyncHandler,
	txSyncer update.TransactionsSyncHandler,
	scheduledEnableEpoch uint32,
) (*startInEpochWithScheduledDataSyncer, error) {

	if check.IfNil(scheduledTxsHandler) {
		return nil, epochStart.ErrNilScheduledTxsHandler
	}
	if check.IfNil(headersSyncer) {
		return nil, epochStart.ErrNilHeadersSyncer
	}
	if check.IfNil(miniBlocksSyncer) {
		return nil, epochStart.ErrNilMiniBlocksSyncer
	}
	if check.IfNil(txSyncer) {
		return nil, epochStart.ErrNilTransactionsSyncer
	}

	return &startInEpochWithScheduledDataSyncer{
		scheduledTxsHandler:       scheduledTxsHandler,
		scheduledMiniBlocksSyncer: miniBlocksSyncer,
		scheduledHeadersSyncer:    headersSyncer,
		txSyncer:                  txSyncer,
		scheduledEnableEpoch:      scheduledEnableEpoch,
	}, nil
}

// UpdateSyncDataIfNeeded checks if according to the header, there is additional or different data required to be synchronized
// and returns that data.
func (ses *startInEpochWithScheduledDataSyncer) UpdateSyncDataIfNeeded(
	notarizedShardHeader data.ShardHeaderHandler,
) (data.ShardHeaderHandler, map[string]data.HeaderHandler, map[string]*block.MiniBlock, error) {
	if ses.scheduledEnableEpoch > notarizedShardHeader.GetEpoch() {
		return notarizedShardHeader, nil, nil, nil
	}

	headerToBeProcessed, headers, err := ses.getRequiredHeaderByHash(notarizedShardHeader)
	if err != nil {
		return nil, nil, nil, err
	}

	allMiniBlocks, err := ses.getMiniBlocks(notarizedShardHeader)
	if err != nil {
		return nil, nil, nil, err
	}

	err = ses.prepareScheduledIntermediateTxs(headerToBeProcessed, notarizedShardHeader, allMiniBlocks)
	if err != nil {
		return nil, nil, nil, err
	}

	return headerToBeProcessed, headers, allMiniBlocks, nil
}

// IsInterfaceNil returns true if the receiver is nil, false otherwise
func (ses *startInEpochWithScheduledDataSyncer) IsInterfaceNil() bool {
	return ses == nil
}

func (ses *startInEpochWithScheduledDataSyncer) getRequiredHeaderByHash(
	notarizedShardHeader data.ShardHeaderHandler,
) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error) {
	shardIDs, hashesToRequest := getShardIDAndHashesForIncludedMetaBlocks(notarizedShardHeader)

	shardIDs = append(shardIDs, notarizedShardHeader.GetShardID())
	hashesToRequest = append(hashesToRequest, notarizedShardHeader.GetPrevHash())

	headers, err := ses.syncHeaders(shardIDs, hashesToRequest)
	if err != nil {
		return nil, nil, err
	}

	headerToBeProcessed, ok := headers[string(notarizedShardHeader.GetPrevHash())].(data.ShardHeaderHandler)
	if !ok {
		return nil, nil, fmt.Errorf("%w in getRequiredHeaderByHash: shard header hash: %s",
			epochStart.ErrMissingHeader,
			hex.EncodeToString(notarizedShardHeader.GetPrevHash()))
	}

	shardIDs, hashesToRequest = getShardIDAndHashesForIncludedMetaBlocks(headerToBeProcessed)
	additionalMetaHashToRequest := getPreviousToFirstReferencedMetaHeaderHash(notarizedShardHeader, headers)
	if len(additionalMetaHashToRequest) != 0 {
		shardIDs = append(shardIDs, core.MetachainShardId)
		hashesToRequest = append(hashesToRequest, additionalMetaHashToRequest)
	}

	prevHeaders, err := ses.syncHeaders(shardIDs, hashesToRequest)
	if err != nil {
		return nil, nil, err
	}

	for hash, hdr := range prevHeaders {
		headers[hash] = hdr
	}

	// get also the previous meta block to the first used meta block, which should be completed
	if len(hashesToRequest) > 0 {
		var prevPrevHeaders map[string]data.HeaderHandler
		shardIDs = []uint32{core.MetachainShardId}
		header := prevHeaders[string(hashesToRequest[0])]
		if header == nil {
			return nil, nil, fmt.Errorf("%w in getRequiredHeaderByHash: metaBlock hash: %s",
				epochStart.ErrMissingHeader,
				hex.EncodeToString(hashesToRequest[0]))
		}

		hashesToRequest = [][]byte{header.GetPrevHash()}
		prevPrevHeaders, err = ses.syncHeaders(shardIDs, hashesToRequest)
		if err != nil {
			return nil, nil, err
		}

		for hash, hdr := range prevPrevHeaders {
			headers[hash] = hdr
		}
	}

	return headerToBeProcessed, headers, nil
}

func getPreviousToFirstReferencedMetaHeaderHash(shardHeader data.ShardHeaderHandler, headers map[string]data.HeaderHandler) []byte {
	hashes := shardHeader.GetMetaBlockHashes()
	if len(hashes) == 0 {
		return nil
	}

	firstReferencedMetaHash := hashes[0]
	firstReferencedMetaHeader := headers[string(firstReferencedMetaHash)]
	if firstReferencedMetaHeader == nil {
		log.Error("getPreviousToFirstReferencedMetaHeaderHash", "hash", firstReferencedMetaHash, "error", epochStart.ErrMissingHeader)
		return nil
	}

	metaHeader, ok := firstReferencedMetaHeader.(data.MetaHeaderHandler)
	if !ok {
		log.Error("getPreviousToFirstReferencedMetaHeaderHash", "hash", firstReferencedMetaHash, "error", epochStart.ErrWrongTypeAssertion)
		return nil
	}

	return metaHeader.GetPrevHash()
}

func getShardIDAndHashesForIncludedMetaBlocks(notarizedShardHeader data.ShardHeaderHandler) ([]uint32, [][]byte) {
	shardIDs := make([]uint32, 0)
	hashesToRequest := make([][]byte, 0)

	// if there were processed metaBlocks in the notarized header, these need to be synced, so they can get again processed
	metaBlockHashes := notarizedShardHeader.GetMetaBlockHashes()
	for i := range metaBlockHashes {
		shardIDs = append(shardIDs, core.MetachainShardId)
		hashesToRequest = append(hashesToRequest, metaBlockHashes[i])
	}
	return shardIDs, hashesToRequest
}

func (ses *startInEpochWithScheduledDataSyncer) syncHeaders(
	shardIDs []uint32,
	hashesToRequest [][]byte,
) (map[string]data.HeaderHandler, error) {
	ses.scheduledHeadersSyncer.ClearFields()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := ses.scheduledHeadersSyncer.SyncMissingHeadersByHash(shardIDs, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	return ses.scheduledHeadersSyncer.GetHeaders()
}

func (ses *startInEpochWithScheduledDataSyncer) getMiniBlocks(
	notarizedShardHeader data.ShardHeaderHandler,
) (map[string]*block.MiniBlock, error) {
	return ses.getRequiredMiniBlocksByMbHeader(notarizedShardHeader.GetMiniBlockHeaderHandlers())
}

func (ses *startInEpochWithScheduledDataSyncer) getRequiredMiniBlocksByMbHeader(
	mbHeaders []data.MiniBlockHeaderHandler,
) (map[string]*block.MiniBlock, error) {
	ses.scheduledMiniBlocksSyncer.ClearFields()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := ses.scheduledMiniBlocksSyncer.SyncPendingMiniBlocks(mbHeaders, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	return ses.scheduledMiniBlocksSyncer.GetMiniBlocks()
}

// GetRootHashToSync checks according to the header what root hash is required to be synchronized
func (ses *startInEpochWithScheduledDataSyncer) GetRootHashToSync(notarizedShardHeader data.ShardHeaderHandler) []byte {
	if ses.scheduledEnableEpoch > notarizedShardHeader.GetEpoch() {
		return notarizedShardHeader.GetRootHash()
	}

	additionalData := notarizedShardHeader.GetAdditionalData()
	if additionalData != nil {
		return additionalData.GetScheduledRootHash()
	}

	return notarizedShardHeader.GetRootHash()
}

func (ses *startInEpochWithScheduledDataSyncer) prepareScheduledIntermediateTxs(
	prevHeader data.HeaderHandler,
	header data.HeaderHandler,
	miniBlocks map[string]*block.MiniBlock,
) error {
	scheduledTxHashes, prevHeaderMiniblocks, err := ses.getScheduledTransactionHashes(prevHeader)
	if err != nil {
		return err
	}

	allTxs, err := ses.getAllTransactionsForMiniBlocks(miniBlocks, header.GetEpoch())
	if err != nil {
		return err
	}

	scheduledIntermediateTxs, err := ses.filterScheduledIntermediateTxs(miniBlocks, scheduledTxHashes, allTxs, prevHeader.GetShardID())
	if err != nil {
		return err
	}

	additionalData := header.GetAdditionalData()
	if additionalData != nil {
		scheduledIntermediateTxsMap := getScheduledIntermediateTxsMapInOrder(
			header.GetMiniBlockHeaderHandlers(),
			miniBlocks,
			scheduledIntermediateTxs,
		)
		gasAndFees := scheduled.GasAndFees{
			AccumulatedFees: additionalData.GetScheduledAccumulatedFees(),
			DeveloperFees:   additionalData.GetScheduledDeveloperFees(),
			GasProvided:     additionalData.GetScheduledGasProvided(),
			GasPenalized:    additionalData.GetScheduledGasPenalized(),
			GasRefunded:     additionalData.GetScheduledGasRefunded(),
		}
		scheduledMiniBlocks := getScheduledMiniBlocks(header, miniBlocks)
		scheduledInfo := &process.ScheduledInfo{
			RootHash:        additionalData.GetScheduledRootHash(),
			IntermediateTxs: scheduledIntermediateTxsMap,
			GasAndFees:      gasAndFees,
			MiniBlocks:      scheduledMiniBlocks,
		}
		ses.saveScheduledInfo(header.GetPrevHash(), scheduledInfo)
	}

	if miniBlocks == nil {
		miniBlocks = make(map[string]*block.MiniBlock)
	}

	for hash, mb := range prevHeaderMiniblocks {
		miniBlocks[hash] = mb
	}

	return nil
}

func (ses *startInEpochWithScheduledDataSyncer) filterScheduledIntermediateTxs(
	miniBlocks map[string]*block.MiniBlock,
	scheduledTxHashes map[string]uint32,
	allTxs map[string]data.TransactionHandler,
	selfShardID uint32,
) (map[string]data.TransactionHandler, error) {
	scheduledIntermediateTxs := make(map[string]data.TransactionHandler)
	for txHash, txHandler := range allTxs {
		if isScheduledIntermediateTx(miniBlocks, scheduledTxHashes, []byte(txHash), txHandler, selfShardID) {
			scheduledIntermediateTxs[txHash] = txHandler
			log.Debug("startInEpochWithScheduledDataSyncer.filterScheduledIntermediateTxs",
				"intermediate tx hash", []byte(txHash),
				"intermediate tx nonce", txHandler.GetNonce(),
				"intermediate tx sender address", txHandler.GetSndAddr(),
				"intermediate tx receiver address", txHandler.GetRcvAddr(),
				"intermediate tx data", string(txHandler.GetData()),
			)
		}
	}

	return scheduledIntermediateTxs, nil
}

func isScheduledIntermediateTx(
	miniBlocks map[string]*block.MiniBlock,
	scheduledTxHashes map[string]uint32,
	txHash []byte,
	txHandler data.TransactionHandler,
	selfShardID uint32,
) bool {
	blockType := getBlockTypeOfTx(txHash, miniBlocks)
	if blockType != block.SmartContractResultBlock && blockType != block.InvalidBlock {
		log.Debug("isScheduledIntermediateTx", "blockType", blockType, "txHash", txHash, "ret", false)
		return false
	}

	var scheduledTxHash []byte

	if blockType == block.SmartContractResultBlock {
		scr, ok := txHandler.(*smartContractResult.SmartContractResult)
		if !ok {
			log.Error("isScheduledIntermediateTx", "error", epochStart.ErrWrongTypeAssertion)
			return false
		}
		scheduledTxHash = scr.PrevTxHash
	} else {
		scheduledTxHash = txHash
	}

	receiverShardID, isScheduledIntermediateTransaction := scheduledTxHashes[string(scheduledTxHash)]
	isTxExecutedInSelfShard := receiverShardID == selfShardID || blockType == block.InvalidBlock
	log.Debug("isScheduledIntermediateTx",
		"blockType", blockType,
		"txHash", txHash,
		"isScheduledIntermediateTransaction", isScheduledIntermediateTransaction,
		"isTxExecutedInSelfShard", isTxExecutedInSelfShard)

	return isScheduledIntermediateTransaction && isTxExecutedInSelfShard
}

func getScheduledIntermediateTxsMapInOrder(
	miniBlockHeaderHandlerList []data.MiniBlockHeaderHandler,
	miniBlocks map[string]*block.MiniBlock,
	intermediateTxs map[string]data.TransactionHandler,
) map[block.Type][]data.TransactionHandler {
	intermediateTxsMap := make(map[block.Type][]data.TransactionHandler)

	for _, mbHeader := range miniBlockHeaderHandlerList {
		miniBlock, ok := miniBlocks[string(mbHeader.GetHash())]
		if !ok {
			continue
		}

		for _, hash := range miniBlock.TxHashes {
			txHandler, ok := intermediateTxs[string(hash)]
			if !ok {
				continue
			}

			intermediateTxsMap[miniBlock.Type] = append(intermediateTxsMap[miniBlock.Type], txHandler)
		}
	}

	return intermediateTxsMap
}

func getBlockTypeOfTx(txHash []byte, miniBlocks map[string]*block.MiniBlock) block.Type {
	for _, miniBlock := range miniBlocks {
		for _, hash := range miniBlock.TxHashes {
			if bytes.Equal(hash, txHash) {
				return miniBlock.Type
			}
		}
	}

	log.Warn("getBlockTypeOfTx: tx not found in mini blocks", "tx hash", txHash)
	return block.SmartContractResultBlock
}

func getScheduledMiniBlocks(
	header data.HeaderHandler,
	miniBlocks map[string]*block.MiniBlock,
) block.MiniBlockSlice {

	scheduledMiniBlocks := make(block.MiniBlockSlice, 0)
	mbHeaders := header.GetMiniBlockHeaderHandlers()
	for _, mbHeader := range mbHeaders {
		if mbHeader.GetProcessingType() != int32(block.Processed) {
			continue
		}

		miniBlock, ok := miniBlocks[string(mbHeader.GetHash())]
		if !ok {
			log.Warn("getScheduledMiniBlocks: mini block was not found", "mb hash", mbHeader.GetHash())
			continue
		}

		scheduledMiniBlocks = append(scheduledMiniBlocks, miniBlock)
	}

	return scheduledMiniBlocks
}

func (ses *startInEpochWithScheduledDataSyncer) saveScheduledInfo(headerHash []byte, scheduledInfo *process.ScheduledInfo) {
	if scheduledInfo.RootHash == nil {
		return
	}

	log.Debug("startInEpochWithScheduledDataSyncer.saveScheduledInfo",
		"headerHash", headerHash,
		"scheduledRootHash", scheduledInfo.RootHash,
		"num of scheduled mbs", len(scheduledInfo.MiniBlocks),
		"num of scheduled intermediate txs", getNumScheduledIntermediateTxs(scheduledInfo.IntermediateTxs),
		"gasAndFees.AccumulatedFees", scheduledInfo.GasAndFees.AccumulatedFees.String(),
		"gasAndFees.DeveloperFees", scheduledInfo.GasAndFees.DeveloperFees.String(),
		"gasAndFees.GasProvided", scheduledInfo.GasAndFees.GasProvided,
		"gasAndFees.GasPenalized", scheduledInfo.GasAndFees.GasPenalized,
		"gasAndFees.GasRefunded", scheduledInfo.GasAndFees.GasRefunded,
	)

	ses.scheduledTxsHandler.SaveState(headerHash, scheduledInfo)
}

func (ses *startInEpochWithScheduledDataSyncer) getAllTransactionsForMiniBlocks(miniBlocks map[string]*block.MiniBlock, epoch uint32) (map[string]data.TransactionHandler, error) {
	ctx := context.Background()
	err := ses.txSyncer.SyncTransactionsFor(miniBlocks, epoch, ctx)
	if err != nil {
		return nil, err
	}

	return ses.txSyncer.GetTransactions()
}

func (ses *startInEpochWithScheduledDataSyncer) getScheduledMiniBlockHeaders(header data.HeaderHandler) []data.MiniBlockHeaderHandler {
	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()
	schMiniBlockHeaders := make([]data.MiniBlockHeaderHandler, 0)

	for i, mbh := range miniBlockHeaders {
		if mbh.GetProcessingType() == int32(block.Scheduled) {
			schMiniBlockHeaders = append(schMiniBlockHeaders, miniBlockHeaders[i])
		}
	}

	return schMiniBlockHeaders
}

func (ses *startInEpochWithScheduledDataSyncer) getScheduledTransactionHashes(header data.HeaderHandler) (map[string]uint32, map[string]*block.MiniBlock, error) {
	miniBlockHeaders := ses.getScheduledMiniBlockHeaders(header)
	miniBlocks, err := ses.getRequiredMiniBlocksByMbHeader(miniBlockHeaders)
	if err != nil {
		return nil, nil, err
	}

	scheduledTxsForShard := make(map[string]uint32)
	for _, miniBlockHeader := range miniBlockHeaders {
		pi, miniBlock, miniBlockHash, shouldSkip := getMiniBlockAndProcessedIndexes(miniBlockHeader, miniBlocks)
		if shouldSkip {
			continue
		}

		createScheduledTxsForShardMap(pi, miniBlock, miniBlockHash, scheduledTxsForShard)
	}

	return scheduledTxsForShard, miniBlocks, nil
}

func getMiniBlockAndProcessedIndexes(
	miniBlockHeader data.MiniBlockHeaderHandler,
	miniBlocks map[string]*block.MiniBlock,
) (*processedIndexes, *block.MiniBlock, []byte, bool) {

	pi := &processedIndexes{}

	miniBlockHash := miniBlockHeader.GetHash()
	miniBlock, ok := miniBlocks[string(miniBlockHash)]
	if !ok {
		log.Warn("startInEpochWithScheduledDataSyncer.getMiniBlockAndProcessedIndexes: mini block was not found", "mb hash", miniBlockHash)
		return nil, nil, nil, true
	}

	pi.firstIndex = miniBlockHeader.GetIndexOfFirstTxProcessed()
	pi.lastIndex = miniBlockHeader.GetIndexOfLastTxProcessed()

	if pi.firstIndex > pi.lastIndex {
		log.Warn("startInEpochWithScheduledDataSyncer.getMiniBlockAndProcessedIndexes: wrong first/last index",
			"mb hash", miniBlockHash,
			"index of first tx processed", pi.firstIndex,
			"index of last tx processed", pi.lastIndex,
			"num txs", len(miniBlock.TxHashes),
		)
		return nil, nil, nil, true
	}

	return pi, miniBlock, miniBlockHash, false
}

func createScheduledTxsForShardMap(
	pi *processedIndexes,
	miniBlock *block.MiniBlock,
	miniBlockHash []byte,
	scheduledTxsForShard map[string]uint32,
) {
	for index := pi.firstIndex; index <= pi.lastIndex; index++ {
		if index >= int32(len(miniBlock.TxHashes)) {
			log.Warn("startInEpochWithScheduledDataSyncer.createScheduledTxsForShardMap: index out of bound",
				"mb hash", miniBlockHash,
				"index", index,
				"num txs", len(miniBlock.TxHashes),
			)
			break
		}

		txHash := miniBlock.TxHashes[index]
		scheduledTxsForShard[string(txHash)] = miniBlock.GetReceiverShardID()
		log.Debug("startInEpochWithScheduledDataSyncer.createScheduledTxsForShardMap", "hash", txHash)
	}
}

func getNumScheduledIntermediateTxs(mapScheduledIntermediateTxs map[block.Type][]data.TransactionHandler) int {
	numScheduledIntermediateTxs := 0
	for _, scheduledIntermediateTxs := range mapScheduledIntermediateTxs {
		numScheduledIntermediateTxs += len(scheduledIntermediateTxs)
	}

	return numScheduledIntermediateTxs
}
