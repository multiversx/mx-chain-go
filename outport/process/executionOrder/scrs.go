package executionOrder

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
)

func setOrderSmartContractResults(pool *outport.Pool, scheduledMbsFromPreviousBlock []*block.MiniBlock, scrsToMe map[string]data.TransactionHandlerWithGasUsedAndFee) []string {
	scheduledExecutedTxsPrevBlockMap := make(map[string]struct{})
	for _, mb := range scheduledMbsFromPreviousBlock {
		for _, txHash := range mb.TxHashes {
			scheduledExecutedTxsPrevBlockMap[string(txHash)] = struct{}{}
		}
	}

	scheduledExecutedSCRsPrevBlock := make([]string, 0)
	scrsWithNoTxInCurrentShard := make(map[string]map[string]data.TransactionHandlerWithGasUsedAndFee)
	for scrHash, scrHandler := range pool.Scrs {
		scr, ok := scrHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		_, originalTxWasScheduledExecuted := scheduledExecutedTxsPrevBlockMap[string(scr.OriginalTxHash)]
		if originalTxWasScheduledExecuted {
			scheduledExecutedSCRsPrevBlock = append(scheduledExecutedSCRsPrevBlock, scrHash)
		}

		tx, found := pool.Txs[string(scr.OriginalTxHash)]
		if !found {
			groupScrsWithNoTxInCurrentShard(scrsWithNoTxInCurrentShard, string(scr.OriginalTxHash), scrHandler, scrHash)
			continue
		}

		scrHandler.SetExecutionOrder(tx.GetExecutionOrder())
	}

	setExecutionOrderScrsWithNoTxInCurrentShard(scrsWithNoTxInCurrentShard, scrsToMe)

	return scheduledExecutedSCRsPrevBlock
}

func groupScrsWithNoTxInCurrentShard(scrsWithNoTxInCurrentShard map[string]map[string]data.TransactionHandlerWithGasUsedAndFee, originalTxHash string, scrHandler data.TransactionHandlerWithGasUsedAndFee, scrHash string) {
	_, ok := scrsWithNoTxInCurrentShard[originalTxHash]
	if !ok {
		scrsWithNoTxInCurrentShard[originalTxHash] = make(map[string]data.TransactionHandlerWithGasUsedAndFee, 0)
	}

	scrsWithNoTxInCurrentShard[originalTxHash][scrHash] = scrHandler
}

func setExecutionOrderScrsWithNoTxInCurrentShard(groupedScrsByOriginalTxHash map[string]map[string]data.TransactionHandlerWithGasUsedAndFee, scrsToMe map[string]data.TransactionHandlerWithGasUsedAndFee) {
	for _, scrsGrouped := range groupedScrsByOriginalTxHash {
		maxOrder := 0
		for _, scr := range scrsGrouped {
			if maxOrder < scr.GetExecutionOrder() {
				maxOrder = scr.GetExecutionOrder()
			}
		}

		for scrHash, scr := range scrsGrouped {
			_, isSCRToMe := scrsToMe[scrHash]
			if isSCRToMe {
				continue
			}

			scr.SetExecutionOrder(maxOrder)
		}
	}
}
