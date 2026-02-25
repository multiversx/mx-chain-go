package coordinator

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
)

// CreateMbsCrossShardDstMe creates cross-shard miniblocks for the current shard
func (tc *transactionCoordinator) CreateMbsCrossShardDstMe(
	hdr data.HeaderHandler,
	processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) (addedMiniBlocksAndHashes []block.MiniblockAndHash, pendingMiniBlocksAndHashes []block.MiniblockAndHash, numTransactions uint32, allMiniBlocksAdded bool, hasMissingData bool, err error) {
	if check.IfNil(hdr) {
		return nil, nil, 0, false, false, process.ErrNilHeaderHandler
	}

	numMiniBlocksAlreadyProcessed := 0
	miniBlocksAndHashes := make([]block.MiniblockAndHash, 0)
	numTransactions = uint32(0)
	shouldSkipShard := make(map[uint32]bool)

	finalCrossMiniBlockInfos := tc.blockDataRequesterProposal.GetFinalCrossMiniBlockInfoAndRequestMissing(hdr)
	defer func() {
		log.Debug("transactionCoordinator.CreateMbsCrossShardDstMe",
			"header round", hdr.GetRound(),
			"header nonce", hdr.GetNonce(),
			"num mini blocks to be processed", len(finalCrossMiniBlockInfos),
			"total gas consumed in self shard", tc.gasComputation.TotalGasConsumedInSelfShard())
	}()

	txsForMbs := make(map[string][]data.TransactionHandler, 0)
	mbsSlice := make([]data.MiniBlockHeaderHandler, 0)
	for _, miniBlockInfo := range finalCrossMiniBlockInfos {
		isAsyncExecutionEnabled := hdr.IsHeaderV3()
		if !isAsyncExecutionEnabled && tc.blockSizeComputation.IsMaxBlockSizeReached(0, 0) {
			log.Debug("transactionCoordinator.CreateMbsCrossShardDstMe",
				"stop creating", "max block size has been reached")
			break
		}

		if shouldSkipShard[miniBlockInfo.SenderShardID] {
			log.Trace("transactionCoordinator.CreateMbsCrossShardDstMe: should skip shard",
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
			)
			continue
		}

		processedMbInfo := getProcessedMiniBlockInfo(processedMiniBlocksInfo, miniBlockInfo.Hash)
		if processedMbInfo.FullyProcessed {
			numMiniBlocksAlreadyProcessed++
			log.Trace("transactionCoordinator.CreateMbsCrossShardDstMe: mini block already processed",
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
			)
			continue
		}

		miniVal, _ := tc.dataPool.MiniBlocks().Peek(miniBlockInfo.Hash)
		if miniVal == nil {
			shouldSkipShard[miniBlockInfo.SenderShardID] = true
			log.Trace("transactionCoordinator.CreateMbsCrossShardDstMe: mini block not found and was skipped",
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
			)
			continue
		}

		miniBlock, ok := miniVal.(*block.MiniBlock)
		if !ok {
			shouldSkipShard[miniBlockInfo.SenderShardID] = true
			log.Error("transactionCoordinator.CreateMbsCrossShardDstMe: mini block assertion type failed",
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
			)
			continue
		}

		preproc := tc.preProcProposal.getPreProcessor(miniBlock.Type)
		if check.IfNil(preproc) {
			return nil, nil, 0, false, false, fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, miniBlock.Type)
		}

		existingTxsForMb, missingTxs := preproc.GetTransactionsAndRequestMissingForMiniBlock(miniBlock)
		if missingTxs > 0 {
			shouldSkipShard[miniBlockInfo.SenderShardID] = true
			log.Trace("transactionCoordinator.CreateMbsCrossShardDstMe: transactions not found",
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
				"missing txs", missingTxs,
			)
			continue
		}
		log.Debug("transactionCoordinator.CreateMbsCrossShardDstMe: selected mini block",
			"sender shard", miniBlockInfo.SenderShardID,
			"hash", miniBlockInfo.Hash,
			"type", miniBlock.Type,
			"round", miniBlockInfo.Round,
			"num txs", len(miniBlock.TxHashes),
		)

		miniBlocksAndHashes = append(miniBlocksAndHashes, block.MiniblockAndHash{
			Miniblock: miniBlock,
			Hash:      miniBlockInfo.Hash,
		})
		numTransactions += uint32(len(miniBlock.TxHashes))

		txsForMbs[string(miniBlockInfo.Hash)] = existingTxsForMb
		mbsSlice = append(mbsSlice, &block.MiniBlockHeader{
			Hash:            miniBlockInfo.Hash,
			SenderShardID:   miniBlock.SenderShardID,
			ReceiverShardID: miniBlock.ReceiverShardID,
			TxCount:         uint32(len(miniBlock.TxHashes)),
			Type:            miniBlock.Type,
			Reserved:        miniBlock.Reserved,
		})
	}

	lastMBIndex, _, err := tc.gasComputation.AddIncomingMiniBlocks(mbsSlice, txsForMbs)
	if err != nil {
		return nil, nil, 0, false, false, err
	}

	// if not all mini blocks were included, remove them from the miniBlocksAndHashes slice
	// but add them into pendingMiniBlocksAndHashes
	if lastMBIndex < len(mbsSlice)-1 {
		log.Debug("transactionCoordinator.CreateMbsCrossShardDstMe: could not select all mini blocks, saving them as pending", "lastMBIndex", lastMBIndex)

		for _, mbAndHash := range miniBlocksAndHashes[lastMBIndex+1:] {
			numTransactions -= uint32(len(mbAndHash.Miniblock.TxHashes))
		}

		pendingMiniBlocksAndHashes = miniBlocksAndHashes[lastMBIndex+1:]
		miniBlocksAndHashes = miniBlocksAndHashes[:lastMBIndex+1]
	}

	allMiniBlocksAdded = len(miniBlocksAndHashes)+numMiniBlocksAlreadyProcessed == len(finalCrossMiniBlockInfos)
	hasMissingData = !allMiniBlocksAdded && len(shouldSkipShard) > 0

	return miniBlocksAndHashes, pendingMiniBlocksAndHashes, numTransactions, allMiniBlocksAdded, hasMissingData, nil
}

func (tc *transactionCoordinator) getAOTSelection(nonce uint64) ([][]byte, []data.TransactionHandler) {
	if check.IfNil(tc.aotSelector) {
		return [][]byte{}, []data.TransactionHandler{}
	}

	// cancel any ongoing AOT selection
	tc.aotSelector.CancelOngoingSelection()
	aotResult, found := tc.aotSelector.GetPreSelectedTransactions(nonce)
	if !found || aotResult == nil {
		log.Trace("GetAOTSelection: no AOT pre-selected transactions found for nonce", "nonce", nonce)
		return [][]byte{}, []data.TransactionHandler{}
	}

	if len(aotResult.TxHashes) == 0 {
		return [][]byte{}, []data.TransactionHandler{}
	}

	// Retrieve transaction handlers from the data pool using cached hashes
	selectedTxHashes, selectedTxs := tc.getTxHandlersFromHashes(aotResult.TxHashes)
	if len(selectedTxs) == 0 {
		log.Warn("AOT selection abandoned, some pre-selected txs unavailable in pool, will re-select", "nonce", nonce, "numHashes", len(aotResult.TxHashes))
		return [][]byte{}, []data.TransactionHandler{}
	}

	log.Info("SelectOutgoingTransactions: using AOT pre-selected transactions",
		"nonce", nonce,
		"numTxs", len(aotResult.TxHashes))

	return selectedTxHashes, selectedTxs
}

// SelectOutgoingTransactions returns transactions originating in the shard, for a block proposal
func (tc *transactionCoordinator) SelectOutgoingTransactions(
	nonce uint64,
	haveTimeForSelection func() bool,
) (selectedTxHashes [][]byte, selectedPendingIncomingMiniBlocks []data.MiniBlockHeaderHandler) {
	selectedTxHashes = make([][]byte, 0)
	selectedTxs := make([]data.TransactionHandler, 0)

	selectedTxHashes, selectedTxs = tc.getAOTSelection(nonce)
	// if no tx returned from AOT selection, fallback to regular selection from pre-processors
	if len(selectedTxs) == 0 {
		for _, blockType := range tc.preProcProposal.keysTxPreProcs {
			txPreProc := tc.preProcProposal.getPreProcessor(blockType)
			if check.IfNil(txPreProc) {
				log.Warn("transactionCoordinator.SelectOutgoingTransactions: getPreProcessor returned nil for block type", "blockType", blockType)
				continue
			}

			gasBandwidth := tc.gasComputation.GetBandwidthForTransactions()
			txHashes, txs, err := txPreProc.SelectOutgoingTransactions(gasBandwidth, nonce, haveTimeForSelection)
			if err != nil {
				log.Warn("transactionCoordinator.SelectOutgoingTransactions: SelectOutgoingTransactions returned error", "error", err)
				continue
			}
			selectedTxHashes = append(selectedTxHashes, txHashes...)
			selectedTxs = append(selectedTxs, txs...)
		}
	}

	selectedTxHashes, pendingMiniBlocksAdded, err := tc.gasComputation.AddOutgoingTransactions(selectedTxHashes, selectedTxs)
	if err != nil {
		log.Warn("transactionCoordinator.AddOutgoingTransactions: AddOutgoingTransactions returned error", "error", err)
	}

	return selectedTxHashes, pendingMiniBlocksAdded
}

// getTxHandlersFromHashes retrieves transaction handlers from the data pool using the provided hashes
func (tc *transactionCoordinator) getTxHandlersFromHashes(txHashes [][]byte) ([][]byte, []data.TransactionHandler) {
	validHashes := make([][]byte, 0, len(txHashes))
	txs := make([]data.TransactionHandler, 0, len(txHashes))

	txPool := tc.dataPool.Transactions()
	for _, txHash := range txHashes {
		val, ok := txPool.SearchFirstData(txHash)
		if !ok {
			log.Trace("getTxHandlersFromHashes: transaction not found in pool", "hash", txHash)
			return [][]byte{}, []data.TransactionHandler{}
		}

		tx, isTx := val.(data.TransactionHandler)
		if !isTx {
			log.Warn("getTxHandlersFromHashes: value is not a TransactionHandler", "hash", txHash)
			return [][]byte{}, []data.TransactionHandler{}
		}

		validHashes = append(validHashes, txHash)
		txs = append(txs, tx)
	}

	return validHashes, txs
}

func (tc *transactionCoordinator) verifyCreatedMiniBlocksSanity(body *block.Body) error {
	miniblocksFromSelf := make([]*block.MiniBlock, 0)
	for _, mb := range body.MiniBlocks {
		if mb.SenderShardID == tc.shardCoordinator.SelfId() {
			miniblocksFromSelf = append(miniblocksFromSelf, mb)
		}
	}

	collectedMbsAfterExecution := tc.GetCreatedMiniBlocksFromMe()
	unExecutableTransactions := tc.getUnExecutableTransactions()
	invalidTxsInterimProc := tc.getInterimProcessor(block.InvalidBlock)
	invalidTransactions := invalidTxsInterimProc.GetAllCurrentFinishedTxs()

	allProposedOutgoingTxsInBody, err := collectTransactionsFromMiniBlocks(miniblocksFromSelf)
	if err != nil {
		return fmt.Errorf("%w: for body miniBlocks", err)
	}

	allCollectedTxs, err := collectTransactionsFromMiniBlocks(collectedMbsAfterExecution)
	if err != nil {
		return fmt.Errorf("%w: for created miniBlocks", err)
	}

	// add the un-executable transactions to the collected transactions
	for txHash := range unExecutableTransactions {
		if _, exists := allCollectedTxs[txHash]; exists {
			return fmt.Errorf("%w: for collected unexecutable transactions", process.ErrDuplicatedTransaction)
		}
		allCollectedTxs[txHash] = struct{}{}
	}

	// check that invalid transactions are not part of the collected transactions
	for txHash := range invalidTransactions {
		if _, exists := allCollectedTxs[txHash]; exists {
			return fmt.Errorf("%w: for collected invalid transactions", process.ErrDuplicatedTransaction)
		}
		allCollectedTxs[txHash] = struct{}{}
	}

	// check that allProposedOutgoingTxsInBody are part of the collected transactions
	// the collected transactions may contain also extra items (rewards/peer changes/scrs)
	for txHash := range allProposedOutgoingTxsInBody {
		if _, exists := allCollectedTxs[txHash]; !exists {
			return process.ErrTransactionsMismatch
		}
	}

	return nil
}

func collectTransactionsFromMiniBlocks(intraShardMbs []*block.MiniBlock) (map[string]struct{}, error) {
	allTxsInBody := make(map[string]struct{})
	for _, mb := range intraShardMbs {
		for _, txHash := range mb.TxHashes {
			if _, exists := allTxsInBody[string(txHash)]; exists {
				return nil, process.ErrDuplicatedTransaction
			}
			allTxsInBody[string(txHash)] = struct{}{}
		}
	}
	return allTxsInBody, nil
}
