package executionOrder

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	sort "github.com/ElrondNetwork/elrond-go-core/core/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

func (s *sorter) PutExecutionOrderInAPIMiniblocks(
	miniblocks []*api.MiniBlock,
	selfShardID uint32,
	prevRandSeed []byte,
) {
	// sorted
	transactionsToMe := extractNormalTransactionAndScrsToMeAPIMiniblocks(miniblocks, selfShardID, true)
	// sorted
	scheduledTransactionsToMe := extractNormalTransactionAndScrsToMeAPIMiniblocks(miniblocks, selfShardID, false)

	// have to be sorted
	transactionsFromMe := extractNormalTransactionsAndInvalidFromMeAPIMiniblocks(miniblocks, selfShardID, true)
	sort.SortTransactionsBySenderAndNonceWithFrontRunningProtectionAPITransactions(transactionsFromMe, s.hasher, prevRandSeed)

	rewardsTxs := getRewardsTxsFromMeAPIMiniblocks(miniblocks, selfShardID)

	// have to be sorted
	scheduledTransactionsFromMe := extractNormalTransactionsAndInvalidFromMeAPIMiniblocks(miniblocks, selfShardID, false)
	sort.SortTransactionsBySenderAndNonceWithFrontRunningProtectionAPITransactions(scheduledTransactionsFromMe, s.hasher, prevRandSeed)

	allTransactions := append(transactionsToMe, scheduledTransactionsToMe...)
	allTransactions = append(allTransactions, transactionsFromMe...)
	allTransactions = append(allTransactions, rewardsTxs...)
	allTransactions = append(allTransactions, scheduledTransactionsFromMe...)

	for index, tx := range allTransactions {
		tx.ExecutionOrder = uint32(index)
	}
	addExecutionOrderInSmartContractResults(allTransactions, miniblocks)

	return
}

func extractNormalTransactionAndScrsToMeAPIMiniblocks(miniblocks []*api.MiniBlock, selfShardID uint32, ignoreScheduled bool) []*transaction.ApiTransactionResult {
	grouped := make([]*transaction.ApiTransactionResult, 0)
	for _, mb := range miniblocks {
		isScheduled := mb.ProcessingType == block.Scheduled.String()
		if isScheduled == ignoreScheduled {
			continue
		}
		isToMeCross := mb.DestinationShard == selfShardID && mb.SourceShard != mb.DestinationShard
		if !isToMeCross {
			continue
		}

		correctType := mb.Type == block.TxBlock.String() || mb.Type == block.SmartContractResultBlock.String() || mb.Type == block.RewardsBlock.String()
		if correctType {
			grouped = append(grouped, mb.Transactions...)
		}
	}

	return grouped
}

func extractNormalTransactionsAndInvalidFromMeAPIMiniblocks(miniblocks []*api.MiniBlock, selfShardID uint32, ignoreScheduled bool) []*transaction.ApiTransactionResult {
	grouped := make([]*transaction.ApiTransactionResult, 0)
	for _, mb := range miniblocks {
		isScheduled := mb.ProcessingType == block.Scheduled.String()
		if isScheduled == ignoreScheduled {
			continue
		}

		isFromMe := mb.SourceShard == selfShardID
		if !isFromMe {
			continue
		}

		correctType := mb.Type == block.TxBlock.String() || mb.Type == block.InvalidBlock.String()
		if correctType {
			grouped = append(grouped, mb.Transactions...)
			continue
		}
	}

	return grouped
}

func getRewardsTxsFromMeAPIMiniblocks(miniblocks []*api.MiniBlock, selfShardID uint32) []*transaction.ApiTransactionResult {
	grouped := make([]*transaction.ApiTransactionResult, 0)
	if selfShardID != core.MetachainShardId {
		return grouped
	}

	for _, mb := range miniblocks {
		if mb.Type != block.RewardsBlock.String() {
			continue
		}
		isFromMe := mb.SourceShard == selfShardID
		if !isFromMe {
			continue
		}
		grouped = append(grouped, mb.Transactions...)
	}

	return grouped
}

func addExecutionOrderInSmartContractResults(allTransactions []*transaction.ApiTransactionResult, miniblocks []*api.MiniBlock) {
	txsMap := make(map[string]*transaction.ApiTransactionResult)
	for _, tx := range allTransactions {
		txsMap[tx.Hash] = tx
	}

	for _, mb := range miniblocks {
		if mb.Type != block.SmartContractResultBlock.String() {
			continue
		}

		for _, scr := range mb.Transactions {
			if scr.OriginalTransactionHash == "" {
				continue
			}
			initialTx, found := txsMap[scr.OriginalTransactionHash]
			if !found {
				continue
			}
			scr.ExecutionOrder = initialTx.ExecutionOrder
		}
	}
}
