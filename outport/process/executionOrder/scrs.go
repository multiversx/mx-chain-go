package executionOrder

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
)

func setOrderSmartContractResults(pool *outport.Pool, scheduledMbsFromPreviousBlock []*block.MiniBlock) []string {
	scheduledExecutedTxsPrevBlockMap := make(map[string]struct{})
	for _, mb := range scheduledMbsFromPreviousBlock {
		for _, txHash := range mb.TxHashes {
			scheduledExecutedTxsPrevBlockMap[string(txHash)] = struct{}{}
		}
	}

	scheduledExecutedSCRsPrevBlock := make([]string, 0)
	scrsWithNoTxInCurrentShard := make(map[string][]data.TransactionHandlerWithGasUsedAndFee)
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
			groupScrsWithNoTxInCurrentShard(scrsWithNoTxInCurrentShard, string(scr.OriginalTxHash), scrHandler)
			continue
		}

		scrHandler.SetExecutionOrder(tx.GetExecutionOrder())
	}

	setExecutionOrderScrsWithNoTxInCurrentShard(scrsWithNoTxInCurrentShard)

	return scheduledExecutedSCRsPrevBlock
}

func groupScrsWithNoTxInCurrentShard(scrsWithNoTxInCurrentShard map[string][]data.TransactionHandlerWithGasUsedAndFee, originalTxHash string, scrHandler data.TransactionHandlerWithGasUsedAndFee) {
	_, ok := scrsWithNoTxInCurrentShard[originalTxHash]
	if !ok {
		scrsWithNoTxInCurrentShard[originalTxHash] = make([]data.TransactionHandlerWithGasUsedAndFee, 0)
	}

	scrsWithNoTxInCurrentShard[originalTxHash] = append(scrsWithNoTxInCurrentShard[originalTxHash], scrHandler)
}

func setExecutionOrderScrsWithNoTxInCurrentShard(scrs map[string][]data.TransactionHandlerWithGasUsedAndFee) {
	for _, scrsGrouped := range scrs {
		maxOrder := 0
		for _, scr := range scrsGrouped {
			if maxOrder < scr.GetExecutionOrder() {
				maxOrder = scr.GetExecutionOrder()
			}
		}

		for _, scr := range scrsGrouped {
			scr.SetExecutionOrder(maxOrder)
		}
	}
}
