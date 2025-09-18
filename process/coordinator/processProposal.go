package coordinator

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
)

func (tc *transactionCoordinator) CreateMbsCrossShardDstMe(
	hdr data.HeaderHandler,
	processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) ([]block.MiniblockAndHash, uint32, bool, error) {
	if check.IfNil(hdr) {
		return nil, 0, false, process.ErrNilHeaderHandler
	}

	numMiniBlocksAlreadyProcessed := 0
	miniBlocksAndHashes := make([]block.MiniblockAndHash, 0)
	numTransactions := uint32(0)
	shouldSkipShard := make(map[uint32]bool)

	tc.gasComputation.Reset()

	finalCrossMiniBlockInfos := tc.blockDataRequesterProposal.GetFinalCrossMiniBlockInfoAndRequestMissing(hdr)
	defer func() {
		log.Debug("transactionCoordinator.CreateMbsCrossShardDstMe",
			"header round", hdr.GetRound(),
			"header nonce", hdr.GetNonce(),
			"num mini blocks to be processed", len(finalCrossMiniBlockInfos),
			"total gas provided", tc.gasComputation.TotalGasConsumed())
	}()

	txsForMbs := make(map[string][]data.TransactionHandler, 0)
	mbsSlice := make([]data.MiniBlockHeaderHandler, 0)
	for _, miniBlockInfo := range finalCrossMiniBlockInfos {
		if tc.blockSizeComputation.IsMaxBlockSizeReached(0, 0) {
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

		miniVal, _ := tc.miniBlockPool.Peek(miniBlockInfo.Hash)
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
			return nil, 0, false, fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, miniBlock.Type)
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

	lastMBIndex, _, err := tc.gasComputation.CheckIncomingMiniBlocks(mbsSlice, txsForMbs)
	if err != nil {
		return nil, 0, false, err
	}

	// if not all mini blocks were included, remove them from the miniBlocksAndHashes slice
	if lastMBIndex != len(mbsSlice) {
		miniBlocksAndHashes = miniBlocksAndHashes[:lastMBIndex+1]
	}

	allMiniBlocksAdded := len(miniBlocksAndHashes)+numMiniBlocksAlreadyProcessed == len(finalCrossMiniBlockInfos)

	return miniBlocksAndHashes, numTransactions, allMiniBlocksAdded, nil
}

// SelectOutgoingTransactions returns transactions originating in the shard, for a block proposal
func (tc *transactionCoordinator) SelectOutgoingTransactions() [][]byte {
	selectedTxHashes := make([][]byte, 0)
	selectedTxs := make([]data.TransactionHandler, 0)
	for _, blockType := range tc.preProcProposal.keysTxPreProcs {
		txPreProc := tc.preProcProposal.getPreProcessor(blockType)
		if check.IfNil(txPreProc) {
			log.Warn("transactionCoordinator.SelectOutgoingTransactions: getPreProcessor returned nil for block type", "blockType", blockType)
			continue
		}

		txHashes, txs, err := txPreProc.SelectOutgoingTransactions()
		if err != nil {
			log.Warn("transactionCoordinator.SelectOutgoingTransactions: SelectOutgoingTransactions returned error", "error", err)
			continue
		}
		selectedTxHashes = append(selectedTxHashes, txHashes...)
		selectedTxs = append(selectedTxs, txs...)
	}

	selectedTxHashes, err := tc.gasComputation.CheckOutgoingTransactions(selectedTxHashes, selectedTxs)
	if err != nil {
		log.Warn("transactionCoordinator.CheckOutgoingTransactions: CheckOutgoingTransactions returned error", "error", err)
	}

	return selectedTxHashes
}
