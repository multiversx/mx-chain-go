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
	miniBlocksAndHashes := make([]block.MiniblockAndHash, 0)
	numTransactions := uint32(0)
	if check.IfNil(hdr) {
		log.Warn("transactionCoordinator.CreateMbsCrossShardDstMe header is nil")

		return miniBlocksAndHashes, 0, false, nil
	}
	shouldSkipShard := make(map[uint32]bool)
	// TODO: init the gas estimation per added mbs counter

	// TODO: replace this request with the separate one used for proposals.
	finalCrossMiniBlockInfos := tc.blockDataRequester.GetFinalCrossMiniBlockInfoAndRequestMissing(hdr)
	defer func() {
		log.Debug("transactionCoordinator.CreateMbsCrossShardDstMe",
			"header round", hdr.GetRound(),
			"header nonce", hdr.GetNonce(),
			"num mini blocks to be processed", len(finalCrossMiniBlockInfos),
			"total gas provided", tc.gasHandler.TotalGasProvided()) // todo: the gas handler needs to be separate from
	}()

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

		// TODO: the request here needs to be replaced with the one used for proposals.
		preproc := tc.getPreProcessor(miniBlock.Type)
		if check.IfNil(preproc) {
			return nil, 0, false, fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, miniBlock.Type)
		}

		// TODO: The method here should be renamed, as it checks which transactions are not available locally (in the node pool)
		missingTxs := preproc.RequestTransactionsForMiniBlock(miniBlock)
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
			"total gas provided", tc.gasHandler.TotalGasProvided(), // todo: don't use the same instance used in processing, as it will be accessed in parallel
		)

		miniBlocksAndHashes = append(miniBlocksAndHashes, block.MiniblockAndHash{
			Miniblock: miniBlock,
			Hash:      miniBlockInfo.Hash,
		})
		numTransactions += uint32(len(miniBlock.TxHashes))
	}

	allMiniBlocksAdded := len(miniBlocksAndHashes) == len(finalCrossMiniBlockInfos)

	return miniBlocksAndHashes, numTransactions, allMiniBlocksAdded, nil
}

// SelectOutgoingTransactions returns transactions originating in the shard, for a block proposal
func (tc *transactionCoordinator) SelectOutgoingTransactions() [][]byte {
	txHashes := make([][]byte, 0)
	var err error
	for _, blockType := range tc.keysTxPreProcs {
		txPreProc := tc.getPreProcessor(blockType)
		if check.IfNil(txPreProc) {
			log.Warn("transactionCoordinator.SelectOutgoingTransactions: getPreProcessor returned nil for block type", "blockType", blockType)
			continue
		}

		txHashes, err = txPreProc.SelectOutgoingTransactions()
		if err != nil {
			log.Warn("transactionCoordinator.SelectOutgoingTransactions: SelectOutgoingTransactions returned error", "error", err)
			continue
		}
		txHashes = append(txHashes, txHashes...)
	}

	return txHashes
}
