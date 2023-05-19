package executionOrder

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

func setOrderSmartContractResults(pool *outport.TransactionPool, scheduledMbsFromPreviousBlock []*block.MiniBlock, scrsToMe map[string]data.TxWithExecutionOrderHandler) []string {
	scheduledExecutedTxsPrevBlockMap := make(map[string]struct{})
	for _, mb := range scheduledMbsFromPreviousBlock {
		for _, txHash := range mb.TxHashes {
			scheduledExecutedTxsPrevBlockMap[string(txHash)] = struct{}{}
		}
	}

	scheduledExecutedSCRsPrevBlock := make([]string, 0)
	scrsWithNoTxInCurrentShard := make(map[string]map[string]data.TxWithExecutionOrderHandler)
	for scrHash, scrHandler := range pool.SmartContractResults {
		scr := scrHandler.SmartContractResult

		_, originalTxWasScheduledExecuted := scheduledExecutedTxsPrevBlockMap[string(scr.OriginalTxHash)]
		if originalTxWasScheduledExecuted {
			scheduledExecutedSCRsPrevBlock = append(scheduledExecutedSCRsPrevBlock, scrHash)
		}

		tx, found := pool.Transactions[string(scr.OriginalTxHash)]
		if !found {
			groupScrsWithNoTxInCurrentShard(scrsWithNoTxInCurrentShard, string(scr.OriginalTxHash), scrHandler, scrHash)
			continue
		}

		scrHandler.ExecutionOrder = tx.GetExecutionOrder()
	}

	setExecutionOrderScrsWithNoTxInCurrentShard(scrsWithNoTxInCurrentShard, scrsToMe)

	return scheduledExecutedSCRsPrevBlock
}

func groupScrsWithNoTxInCurrentShard(
	scrsWithNoTxInCurrentShard map[string]map[string]data.TxWithExecutionOrderHandler,
	originalTxHash string,
	scrHandler data.TxWithExecutionOrderHandler,
	scrHash string,
) {
	_, ok := scrsWithNoTxInCurrentShard[originalTxHash]
	if !ok {
		scrsWithNoTxInCurrentShard[originalTxHash] = make(map[string]data.TxWithExecutionOrderHandler, 0)
	}

	scrsWithNoTxInCurrentShard[originalTxHash][scrHash] = scrHandler
}

func setExecutionOrderScrsWithNoTxInCurrentShard(groupedScrsByOriginalTxHash map[string]map[string]data.TxWithExecutionOrderHandler, scrsToMe map[string]data.TxWithExecutionOrderHandler) {
	for _, scrsGrouped := range groupedScrsByOriginalTxHash {
		maxOrder := uint32(0)
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
