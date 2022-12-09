package executionOrder

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	txsSort "github.com/ElrondNetwork/elrond-go-core/core/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.GetOrCreate("outport/process/executionOrder")

type sorter struct {
	hasher hashing.Hasher
}

func NewSorter(hasher hashing.Hasher) (*sorter, error) {
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}

	return &sorter{
		hasher: hasher,
	}, nil
}

func (s *sorter) PutExecutionOrderInTransactionPool(
	pool *outport.Pool,
	header data.HeaderHandler,
	body data.BodyHandler,
) error {
	blockBody, ok := body.(*block.Body)
	if !ok {
		log.Warn("s.PutExecutionOrderInTransactionPool cannot cast BodyHandler to *Body")
		return nil
	}

	// already sorted
	transactionsToMe := extractNormalTransactionAndScrsToMe(pool, blockBody, header, true)

	// scheduled to me, already sorted
	scheduledTransactionsToMe := extractNormalTransactionAndScrsToMe(pool, blockBody, header, false)

	// need to be sorted
	transactionsFromMe := extractNormalTransactionsAndInvalidFromMe(pool, blockBody, header, true)
	txsSort.SortTransactionsBySenderAndNonceWithFrontRunningProtectionExtendedTransactions(transactionsFromMe, s.hasher, header.GetPrevRandSeed())

	rewardsTxs := getRewardsTxsFromMe(pool, blockBody, header)

	// scheduled from me, need to be sorted
	scheduledTransactionsFromMe := extractNormalTransactionsAndInvalidFromMe(pool, blockBody, header, false)
	txsSort.SortTransactionsBySenderAndNonceWithFrontRunningProtectionExtendedTransactions(scheduledTransactionsFromMe, s.hasher, header.GetPrevRandSeed())

	allTransaction := append(transactionsToMe, scheduledTransactionsToMe...)
	allTransaction = append(allTransaction, transactionsFromMe...)
	allTransaction = append(allTransaction, rewardsTxs...)
	allTransaction = append(allTransaction, scheduledTransactionsFromMe...)

	for idx, tx := range allTransaction {
		tx.SetExecutionOrder(idx)
	}

	setOrderSmartContractResults(pool)

	prinPool(pool)
	return nil
}

func setOrderSmartContractResults(pool *outport.Pool) {
	for _, scrHandler := range pool.Scrs {
		scr, ok := scrHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		tx, found := pool.Txs[string(scr.OriginalTxHash)]
		if !found {
			continue
		}

		scrHandler.SetExecutionOrder(tx.GetExecutionOrder())
	}
}

func extractNormalTransactionsAndInvalidFromMe(pool *outport.Pool, blockBody *block.Body, header data.HeaderHandler, ignoreScheduled bool) []data.TransactionHandlerWithGasUsedAndFee {
	normalTxsHashes := make([][]byte, 0)
	invalidTxsHashes := make([][]byte, 0)
	for mbIndex, mb := range blockBody.MiniBlocks {
		if isMBScheduled(header, mbIndex) == ignoreScheduled || shouldIgnoreProcessedMBScheduled(header, mbIndex) {
			continue
		}

		isFromMe := mb.SenderShardID == header.GetShardID()
		if !isFromMe {
			continue
		}

		if mb.Type == block.TxBlock {
			normalTxsHashes = append(normalTxsHashes, mb.TxHashes...)
			continue
		}
		if mb.Type == block.InvalidBlock {
			invalidTxsHashes = append(invalidTxsHashes, mb.TxHashes...)
			continue
		}
	}

	normalTxs := extractTxsFromMap(normalTxsHashes, pool.Txs)
	invalidTxs := extractTxsFromMap(invalidTxsHashes, pool.Invalid)

	return append(normalTxs, invalidTxs...)
}

func extractNormalTransactionAndScrsToMe(pool *outport.Pool, blockBody *block.Body, header data.HeaderHandler, ignoreScheduled bool) []data.TransactionHandlerWithGasUsedAndFee {
	grouped := make([]data.TransactionHandlerWithGasUsedAndFee, 0)
	for mbIndex, mb := range blockBody.MiniBlocks {
		if isMBScheduled(header, mbIndex) == ignoreScheduled || shouldIgnoreProcessedMBScheduled(header, mbIndex) {
			continue
		}

		isToMeCross := mb.ReceiverShardID == header.GetShardID() && mb.SenderShardID != mb.ReceiverShardID
		if !isToMeCross {
			continue
		}

		executedTxsHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, header)
		if mb.Type == block.TxBlock {
			grouped = append(grouped, extractTxsFromMap(executedTxsHashes, pool.Txs)...)
			continue
		}
		if mb.Type == block.SmartContractResultBlock {
			grouped = append(grouped, extractTxsFromMap(executedTxsHashes, pool.Scrs)...)
			continue
		}
		if mb.Type == block.RewardsBlock {
			grouped = append(grouped, extractTxsFromMap(executedTxsHashes, pool.Rewards)...)
			continue
		}
	}

	return grouped
}

func getRewardsTxsFromMe(pool *outport.Pool, blockBody *block.Body, header data.HeaderHandler) []data.TransactionHandlerWithGasUsedAndFee {
	rewardsTxsHashes := make([][]byte, 0)
	rewardsTxs := make([]data.TransactionHandlerWithGasUsedAndFee, 0)
	if header.GetShardID() != core.MetachainShardId {
		return rewardsTxs
	}

	for _, mb := range blockBody.MiniBlocks {
		if mb.Type != block.RewardsBlock {
			continue
		}
		isFromMe := mb.SenderShardID == header.GetShardID()
		if !isFromMe {
			continue
		}
		rewardsTxsHashes = append(rewardsTxsHashes, mb.TxHashes...)
	}

	return extractTxsFromMap(rewardsTxsHashes, pool.Rewards)
}

func extractTxsFromMap(txsHashes [][]byte, txs map[string]data.TransactionHandlerWithGasUsedAndFee) []data.TransactionHandlerWithGasUsedAndFee {
	result := make([]data.TransactionHandlerWithGasUsedAndFee, 0, len(txsHashes))
	for _, txHash := range txsHashes {
		tx, found := txs[string(txHash)]
		if !found {
			log.Warn("extractTxsFromMap cannot find transaction in pool", "txHash", hex.EncodeToString(txHash))
			continue
		}
		result = append(result, tx)
	}

	return result
}

func extractExecutedTxHashes(mbIndex int, mbTxHashes [][]byte, header data.HeaderHandler) [][]byte {
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniblockHeaders) <= mbIndex {
		return mbTxHashes
	}

	firstProcessed := miniblockHeaders[mbIndex].GetIndexOfFirstTxProcessed()
	lastProcessed := miniblockHeaders[mbIndex].GetIndexOfLastTxProcessed()

	executedTxHashes := make([][]byte, 0)
	for txIndex, txHash := range mbTxHashes {
		if int32(txIndex) < firstProcessed || int32(txIndex) > lastProcessed {
			continue
		}

		executedTxHashes = append(executedTxHashes, txHash)
	}

	return executedTxHashes
}

func shouldIgnoreProcessedMBScheduled(header data.HeaderHandler, mbIndex int) bool {
	return getProcessingType(header, mbIndex) == int32(block.Processed)
}

func isMBScheduled(header data.HeaderHandler, mbIndex int) bool {
	return getProcessingType(header, mbIndex) == int32(block.Scheduled)
}

func getProcessingType(header data.HeaderHandler, mbIndex int) int32 {
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniblockHeaders) <= mbIndex {
		return int32(block.Normal)
	}

	return miniblockHeaders[mbIndex].GetProcessingType()
}

// TODO remove this after system test will pass
func prinPool(pool *outport.Pool) {
	prinMapTxs := func(txs map[string]data.TransactionHandlerWithGasUsedAndFee) {
		for hash, tx := range txs {
			log.Warn(hex.EncodeToString([]byte(hash)), "order", tx.GetExecutionOrder())
		}
	}

	total := len(pool.Txs) + len(pool.Invalid) + len(pool.Scrs) + len(pool.Rewards)
	if total > 0 {
		log.Warn("###################################")
	}

	if len(pool.Txs) > 0 {
		log.Warn("############### NORMAL TXS ####################")
		prinMapTxs(pool.Txs)
	}
	if len(pool.Invalid) > 0 {
		log.Warn("############### INVALID ####################")
		prinMapTxs(pool.Invalid)
	}

	if len(pool.Scrs) > 0 {
		log.Warn("############### SCRS ####################")
		prinMapTxs(pool.Scrs)
	}

	if len(pool.Rewards) > 0 {
		log.Warn("############### REWARDS ####################")
		prinMapTxs(pool.Rewards)
	}
	if total > 0 {
		log.Warn("###################################")
	}
}
