package executionOrder

import (
	"encoding/hex"

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
	startIndexTransactionsToMe := 0
	putOrderInTransactions(transactionsToMe, startIndexTransactionsToMe)

	// scheduled to me, already sorted
	scheduledTransactionsToMe := extractNormalTransactionAndScrsToMe(pool, blockBody, header, false)
	startIndexScheduledTransactionsToMe := len(transactionsToMe)
	putOrderInTransactions(scheduledTransactionsToMe, startIndexScheduledTransactionsToMe)

	// need to be sorted
	transactionsFromMe := extractNormalTransactionsAndInvalidFromMe(pool, blockBody, header, true)

	// scheduled from me, need to be sorted
	scheduledTransactionsFromMe := extractNormalTransactionsAndInvalidFromMe(pool, blockBody, header, false)

	txsSort.SortTransactionsBySenderAndNonceWithFrontRunningProtectionExtendedTransactions(transactionsFromMe, s.hasher, header.GetPrevRandSeed())
	startIndexTransactionsFromMe := startIndexScheduledTransactionsToMe + len(scheduledTransactionsToMe)
	putOrderInTransactions(transactionsFromMe, startIndexTransactionsFromMe)

	txsSort.SortTransactionsBySenderAndNonceWithFrontRunningProtectionExtendedTransactions(scheduledTransactionsFromMe, s.hasher, header.GetPrevRandSeed())
	startIndexScheduledTransactionFromMe := startIndexTransactionsFromMe + len(transactionsFromMe)
	putOrderInTransactions(scheduledTransactionsFromMe, startIndexScheduledTransactionFromMe)

	setOrderSmartContractResults(pool)

	return nil
}

func putOrderInTransactions(txs []data.TransactionHandlerWithGasUsedAndFee, startIndex int) {
	for idx, tx := range txs {
		tx.SetExecutionOrder(idx + startIndex)
	}
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
		if mb.IsScheduledMiniBlock() == ignoreScheduled || shouldIgnoreProcessedMBScheduled(header, mbIndex) {
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
	scrsHashes := make([][]byte, 0)
	normalTxsHashes := make([][]byte, 0)
	for mbIndex, mb := range blockBody.MiniBlocks {
		if mb.IsScheduledMiniBlock() == ignoreScheduled || shouldIgnoreProcessedMBScheduled(header, mbIndex) {
			continue
		}

		isToMeCross := mb.ReceiverShardID == header.GetShardID() && mb.SenderShardID != mb.ReceiverShardID
		if !isToMeCross {
			continue
		}

		executedTxsHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, header)
		if mb.Type == block.TxBlock {
			normalTxsHashes = append(normalTxsHashes, executedTxsHashes...)
			continue
		}
		if mb.Type == block.SmartContractResultBlock {
			scrsHashes = append(scrsHashes, executedTxsHashes...)
			continue
		}
	}

	normalTxs := extractTxsFromMap(normalTxsHashes, pool.Txs)
	smartContractResults := extractTxsFromMap(scrsHashes, pool.Scrs)

	return append(normalTxs, smartContractResults...)
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
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniblockHeaders) <= mbIndex {
		return false
	}

	processingType := miniblockHeaders[mbIndex].GetProcessingType()

	return processingType == int32(block.Processed)
}
