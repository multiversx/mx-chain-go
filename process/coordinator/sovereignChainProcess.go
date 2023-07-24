package coordinator

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
)

type sovereignChainTransactionCoordinator struct {
	*transactionCoordinator
}

// NewSovereignChainTransactionCoordinator creates a new sovereign chain transaction coordinator
func NewSovereignChainTransactionCoordinator(
	trasnsactionCoordinator process.TransactionCoordinator,
) (*sovereignChainTransactionCoordinator, error) {
	if trasnsactionCoordinator == nil {
		return nil, process.ErrNilTransactionCoordinator
	}

	tc, ok := trasnsactionCoordinator.(*transactionCoordinator)
	if !ok {
		return nil, errorsMx.ErrWrongTypeAssertion
	}

	sctc := &sovereignChainTransactionCoordinator{
		transactionCoordinator: tc,
	}

	return sctc, nil
}

// CreateMbsAndProcessCrossShardTransactionsDstMe creates mini blocks and processes cross shard transactions with destination in self shard
func (sctc *sovereignChainTransactionCoordinator) CreateMbsAndProcessCrossShardTransactionsDstMe(
	hdr data.HeaderHandler,
	processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
) (block.MiniBlockSlice, uint32, bool, error) {
	createMBDestMeExecutionInfo := initMiniBlockDestMeExecutionInfo()

	if check.IfNil(hdr) {
		log.Warn("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: header is nil")

		// we return the nil error here as to allow the proposer execute as much as it can, even if it ends up in a
		// totally unlikely situation in which it needs to process a nil block.

		return createMBDestMeExecutionInfo.miniBlocks, createMBDestMeExecutionInfo.numTxAdded, false, nil
	}

	shardHeaderExtendedHanlder, isShardHeaderExtendedHandler := hdr.(data.ShardHeaderExtendedHandler)
	if !isShardHeaderExtendedHandler {
		log.Warn("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: wrong type assertion from data.HeaderHandler to data.ShardHeaderExtendedHandler")
		return createMBDestMeExecutionInfo.miniBlocks, createMBDestMeExecutionInfo.numTxAdded, false, fmt.Errorf("%w from data.HeaderHandler to data.ShardHeaderExtendedHandler", errorsMx.ErrWrongTypeAssertion)
	}

	shouldSkipShard := make(map[uint32]bool)

	defer func() {
		log.Debug("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: gas provided, refunded and penalized info",
			"header round", hdr.GetRound(),
			"num mini blocks to be processed", len(shardHeaderExtendedHanlder.GetIncomingMiniBlockHandlers()),
			"num already mini blocks processed", createMBDestMeExecutionInfo.numAlreadyMiniBlocksProcessed,
			"num new mini blocks processed", createMBDestMeExecutionInfo.numNewMiniBlocksProcessed)
	}()

	for _, mbh := range shardHeaderExtendedHanlder.GetIncomingMiniBlockHandlers() {
		miniBlock, isMiniBlock := mbh.(*block.MiniBlock)
		if !isMiniBlock {
			return nil, 0, false, process.ErrWrongTypeAssertion
		}

		miniBlockHash, err := core.CalculateHash(sctc.marshalizer, sctc.hasher, miniBlock)
		if err != nil {
			return nil, 0, false, err
		}

		if !haveTime() && !haveAdditionalTime() {
			log.Debug("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe",
				"scheduled mode", scheduledMode,
				"stop creating", "time is out")
			break
		}

		if sctc.blockSizeComputation.IsMaxBlockSizeReached(0, 0) {
			log.Debug("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe",
				"scheduled mode", scheduledMode,
				"stop creating", "max block size has been reached")
			break
		}

		if shouldSkipShard[miniBlock.SenderShardID] {
			log.Trace("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: should skip shard",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlock.SenderShardID,
				"hash", miniBlockHash,
				"round", hdr.GetRound(),
			)
			continue
		}

		processedMbInfo := getProcessedMiniBlockInfo(processedMiniBlocksInfo, miniBlockHash)
		if processedMbInfo.FullyProcessed {
			createMBDestMeExecutionInfo.numAlreadyMiniBlocksProcessed++
			log.Trace("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: mini block already processed",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlock.SenderShardID,
				"hash", miniBlockHash,
				"round", hdr.GetRound(),
			)
			continue
		}

		preproc := sctc.getPreProcessor(miniBlock.Type)
		if check.IfNil(preproc) {
			return nil, 0, false, fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, miniBlock.Type)
		}

		requestedTxs := preproc.RequestTransactionsForMiniBlock(miniBlock)
		if requestedTxs > 0 {
			shouldSkipShard[miniBlock.SenderShardID] = true
			log.Trace("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: transactions not found and were requested",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlock.SenderShardID,
				"hash", miniBlockHash,
				"round", hdr.GetRound(),
				"requested txs", requestedTxs,
			)
			continue
		}

		oldIndexOfLastTxProcessed := processedMbInfo.IndexOfLastTxProcessed

		errProc := sctc.processCompleteMiniBlock(preproc, miniBlock, miniBlockHash, haveTime, haveAdditionalTime, scheduledMode, processedMbInfo)
		sctc.handleProcessMiniBlockExecution(oldIndexOfLastTxProcessed, miniBlock, processedMbInfo, createMBDestMeExecutionInfo)

		log.Debug("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: processing",
			"scheduled mode", scheduledMode,
			"sender shard", miniBlock.SenderShardID,
			"hash", miniBlockHash,
			"type", miniBlock.Type,
			"round", hdr.GetRound(),
			"num txs", len(miniBlock.TxHashes),
			"num all txs processed", processedMbInfo.IndexOfLastTxProcessed+1,
			"num current txs processed", processedMbInfo.IndexOfLastTxProcessed-oldIndexOfLastTxProcessed,
			"fully processed", processedMbInfo.FullyProcessed,
			"total gas provided", sctc.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", sctc.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", sctc.gasHandler.TotalGasRefunded(),
			"total gas penalized", sctc.gasHandler.TotalGasPenalized(),
		)

		if errProc != nil {
			shouldSkipShard[miniBlock.SenderShardID] = true
			log.Debug("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: processed complete mini block failed",
				"error", errProc,
			)
			continue
		}

		log.Debug("sovereignChainTransactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: processed complete mini block succeeded")
	}

	numTotalMiniBlocksProcessed := createMBDestMeExecutionInfo.numAlreadyMiniBlocksProcessed + createMBDestMeExecutionInfo.numNewMiniBlocksProcessed
	allMBsProcessed := numTotalMiniBlocksProcessed == len(shardHeaderExtendedHanlder.GetIncomingMiniBlockHandlers())

	return createMBDestMeExecutionInfo.miniBlocks, createMBDestMeExecutionInfo.numTxAdded, allMBsProcessed, nil
}
