package executionOrder

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	txsSort "github.com/ElrondNetwork/elrond-go-core/core/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	processOut "github.com/ElrondNetwork/elrond-go/outport/process"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgSorter holds the arguments needed for creating a new instance of sorter
type ArgSorter struct {
	Hasher              hashing.Hasher
	Marshaller          marshal.Marshalizer
	MbsStorer           storage.Storer
	EnableEpochsHandler common.EnableEpochsHandler
}

var log = logger.GetOrCreate("outport/process/executionOrder")

type sorter struct {
	mbsGetter           mbsGetter
	hasher              hashing.Hasher
	enableEpochsHandler common.EnableEpochsHandler
}

// NewSorter will create a new instance of sorter
func NewSorter(arg ArgSorter) (*sorter, error) {
	if check.IfNil(arg.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(arg.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(arg.MbsStorer) {
		return nil, processOut.ErrNilStorer
	}
	if check.IfNil(arg.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}

	return &sorter{
		mbsGetter:           newMiniblocksGetter(arg.MbsStorer, arg.Marshaller),
		hasher:              arg.Hasher,
		enableEpochsHandler: arg.EnableEpochsHandler,
	}, nil
}

// PutExecutionOrderInTransactionPool will put the execution order for every transaction and smart contract result
func (s *sorter) PutExecutionOrderInTransactionPool(
	pool *outport.Pool,
	header data.HeaderHandler,
	body data.BodyHandler,
	prevHeader data.HeaderHandler,
) ([]string, []string, error) {
	blockBody, ok := body.(*block.Body)
	if !ok {
		log.Warn("s.PutExecutionOrderInTransactionPool cannot cast BodyHandler to *Body")
		return nil, nil, nil
	}

	scheduledMbsFromPreviousBlock, err := s.mbsGetter.GetScheduledMBs(header, prevHeader)
	if err != nil {
		return nil, nil, err
	}

	// already sorted
	transactionsToMe, scheduledTransactionsToMe, err := extractNormalTransactionAndScrsToMe(pool, blockBody, header)
	if err != nil {
		return nil, nil, err
	}

	// need to be sorted
	transactionsFromMe, scheduledTransactionsFromMe, scheduledExecutedInvalidTxsHashesPrevBlock, err := s.extractNormalTransactionsAndInvalidFromMe(pool, blockBody, header, scheduledMbsFromPreviousBlock)
	if err != nil {
		return nil, nil, err
	}

	s.sortTransactions(transactionsFromMe, header)

	rewardsTxs, err := getRewardsTxsFromMe(pool, blockBody, header)
	if err != nil {
		return nil, nil, err
	}

	// scheduled from me, need to be sorted
	s.sortTransactions(scheduledTransactionsFromMe, header)

	allTransaction := append(transactionsToMe, transactionsFromMe...)
	allTransaction = append(allTransaction, rewardsTxs...)
	allTransaction = append(allTransaction, scheduledTransactionsToMe...)
	allTransaction = append(allTransaction, scheduledTransactionsFromMe...)

	for idx, tx := range allTransaction {
		tx.SetExecutionOrder(idx)
	}

	scheduledExecutedSCRSHashesPrevBlock := setOrderSmartContractResults(pool, scheduledMbsFromPreviousBlock)

	printPool(pool)

	return scheduledExecutedSCRSHashesPrevBlock, scheduledExecutedInvalidTxsHashesPrevBlock, nil
}

func (s *sorter) sortTransactions(transactions []data.TransactionHandlerWithGasUsedAndFee, header data.HeaderHandler) {
	if s.enableEpochsHandler.IsFrontRunningProtectionFlagEnabled() {
		txsSort.SortTransactionsBySenderAndNonceWithFrontRunningProtectionExtendedTransactions(transactions, s.hasher, header.GetPrevRandSeed())
	} else {
		txsSort.SortTransactionsBySenderAndNonceExtendedTransactions(transactions)
	}
}

func setOrderSmartContractResults(pool *outport.Pool, scheduledMbsFromPreviousBlock []*block.MiniBlock) []string {
	scheduledExecutedTxsPrevBlockMap := make(map[string]struct{})
	for _, mb := range scheduledMbsFromPreviousBlock {
		for _, txHash := range mb.TxHashes {
			scheduledExecutedTxsPrevBlockMap[string(txHash)] = struct{}{}
		}
	}

	scheduledExecutedSCRsPrevBlock := make([]string, 0)
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
			continue
		}

		scrHandler.SetExecutionOrder(tx.GetExecutionOrder())
	}

	return scheduledExecutedSCRsPrevBlock
}

func (s *sorter) extractNormalTransactionsAndInvalidFromMe(
	pool *outport.Pool, blockBody *block.Body, header data.HeaderHandler, scheduledMbsFromPreviousBlock []*block.MiniBlock,
) ([]data.TransactionHandlerWithGasUsedAndFee, []data.TransactionHandlerWithGasUsedAndFee, []string, error) {
	transactionsFromMe := make([]data.TransactionHandlerWithGasUsedAndFee, 0)
	scheduledTransactionsFromMe := make([]data.TransactionHandlerWithGasUsedAndFee, 0)

	scheduledExecutedInvalidTxsHashesPrevBlock := make([]string, 0)
	for mbIndex, mb := range blockBody.MiniBlocks {
		var txs []data.TransactionHandlerWithGasUsedAndFee
		var err error
		if isScheduledMBProcessed(header, mbIndex) {
			continue
		}

		isFromMe := mb.SenderShardID == header.GetShardID()
		if !isFromMe {
			continue
		}

		if mb.Type == block.TxBlock {
			txs, err = extractTxsFromMap(mb.TxHashes, pool.Txs)
		}
		if mb.Type == block.InvalidBlock {
			var scheduledExecutedInvalidTxsHashesCurrentMB []string
			txs, scheduledExecutedInvalidTxsHashesCurrentMB, err = s.getInvalidTxsExecutedInCurrentBlock(scheduledMbsFromPreviousBlock, mb, pool)
			scheduledExecutedInvalidTxsHashesPrevBlock = append(scheduledExecutedInvalidTxsHashesPrevBlock, scheduledExecutedInvalidTxsHashesCurrentMB...)
		}

		if err != nil {
			return nil, nil, nil, err
		}

		if isScheduledMBNotProcessed(header, mbIndex) {
			scheduledTransactionsFromMe = append(scheduledTransactionsFromMe, txs...)
		} else {
			transactionsFromMe = append(transactionsFromMe, txs...)
		}
	}

	return transactionsFromMe, scheduledTransactionsFromMe, scheduledExecutedInvalidTxsHashesPrevBlock, nil
}

func (s *sorter) getInvalidTxsExecutedInCurrentBlock(scheduledMbsFromPreviousBlock []*block.MiniBlock, mb *block.MiniBlock, pool *outport.Pool) ([]data.TransactionHandlerWithGasUsedAndFee, []string, error) {
	if len(scheduledMbsFromPreviousBlock) == 0 {
		txs, err := extractTxsFromMap(mb.TxHashes, pool.Invalid)
		return txs, []string{}, err
	}

	allScheduledTxs := make(map[string]struct{})
	for _, scheduledMb := range scheduledMbsFromPreviousBlock {
		for _, txHash := range scheduledMb.TxHashes {
			allScheduledTxs[string(txHash)] = struct{}{}
		}
	}

	scheduledExecutedInvalidTxsHashesPrevBlock := make([]string, 0)
	invalidTxHashes := make([][]byte, 0)
	for _, hash := range mb.TxHashes {
		_, found := allScheduledTxs[string(hash)]
		if found {
			scheduledExecutedInvalidTxsHashesPrevBlock = append(scheduledExecutedInvalidTxsHashesPrevBlock, string(hash))
			continue
		}
		invalidTxHashes = append(invalidTxHashes, hash)
	}

	txs, err := extractTxsFromMap(invalidTxHashes, pool.Invalid)
	return txs, scheduledExecutedInvalidTxsHashesPrevBlock, err
}

func extractNormalTransactionAndScrsToMe(pool *outport.Pool, blockBody *block.Body, header data.HeaderHandler) ([]data.TransactionHandlerWithGasUsedAndFee, []data.TransactionHandlerWithGasUsedAndFee, error) {
	transactionsToMe := make([]data.TransactionHandlerWithGasUsedAndFee, 0)
	scheduledTransactionsToMe := make([]data.TransactionHandlerWithGasUsedAndFee, 0)

	for mbIndex, mb := range blockBody.MiniBlocks {
		var err error
		var txs []data.TransactionHandlerWithGasUsedAndFee
		if isScheduledMBProcessed(header, mbIndex) {
			continue
		}

		isToMeCross := mb.ReceiverShardID == header.GetShardID() && mb.SenderShardID != mb.ReceiverShardID
		if !isToMeCross {
			continue
		}

		executedTxsHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, header)
		if mb.Type == block.TxBlock {
			txs, err = extractTxsFromMap(executedTxsHashes, pool.Txs)
		}
		if mb.Type == block.SmartContractResultBlock {
			txs, err = extractTxsFromMap(executedTxsHashes, pool.Scrs)
		}
		if mb.Type == block.RewardsBlock {
			txs, err = extractTxsFromMap(executedTxsHashes, pool.Rewards)
		}
		if err != nil {
			return nil, nil, err
		}

		if isScheduledMBNotProcessed(header, mbIndex) {
			scheduledTransactionsToMe = append(scheduledTransactionsToMe, txs...)
		} else {
			transactionsToMe = append(transactionsToMe, txs...)
		}
	}

	return transactionsToMe, scheduledTransactionsToMe, nil
}

func getRewardsTxsFromMe(pool *outport.Pool, blockBody *block.Body, header data.HeaderHandler) ([]data.TransactionHandlerWithGasUsedAndFee, error) {
	rewardsTxsHashes := make([][]byte, 0)
	rewardsTxs := make([]data.TransactionHandlerWithGasUsedAndFee, 0)
	if header.GetShardID() != core.MetachainShardId {
		return rewardsTxs, nil
	}

	for _, mb := range blockBody.MiniBlocks {
		if mb.Type != block.RewardsBlock {
			continue
		}
		rewardsTxsHashes = append(rewardsTxsHashes, mb.TxHashes...)
	}

	return extractTxsFromMap(rewardsTxsHashes, pool.Rewards)
}

func extractTxsFromMap(txsHashes [][]byte, txs map[string]data.TransactionHandlerWithGasUsedAndFee) ([]data.TransactionHandlerWithGasUsedAndFee, error) {
	result := make([]data.TransactionHandlerWithGasUsedAndFee, 0, len(txsHashes))
	for _, txHash := range txsHashes {
		tx, found := txs[string(txHash)]
		if !found {
			return nil, fmt.Errorf("cannot find transaction in pool, txHash: %s", hex.EncodeToString(txHash))
		}
		result = append(result, tx)
	}

	return result, nil
}

func extractExecutedTxHashes(mbIndex int, mbTxHashes [][]byte, header data.HeaderHandler) [][]byte {
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniblockHeaders) <= mbIndex {
		return mbTxHashes
	}

	firstProcessed := miniblockHeaders[mbIndex].GetIndexOfFirstTxProcessed()
	lastProcessed := miniblockHeaders[mbIndex].GetIndexOfLastTxProcessed()

	return mbTxHashes[firstProcessed : lastProcessed+1]
}

func isScheduledMBProcessed(header data.HeaderHandler, mbIndex int) bool {
	return getProcessingType(header, mbIndex) == int32(block.Processed)
}

func isScheduledMBNotProcessed(header data.HeaderHandler, mbIndex int) bool {
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
func printPool(pool *outport.Pool) {
	printMapTxs := func(txs map[string]data.TransactionHandlerWithGasUsedAndFee) {
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
		printMapTxs(pool.Txs)
	}
	if len(pool.Invalid) > 0 {
		log.Warn("############### INVALID ####################")
		printMapTxs(pool.Invalid)
	}

	if len(pool.Scrs) > 0 {
		log.Warn("############### SCRS ####################")
		printMapTxs(pool.Scrs)
	}

	if len(pool.Rewards) > 0 {
		log.Warn("############### REWARDS ####################")
		printMapTxs(pool.Rewards)
	}
	if total > 0 {
		log.Warn("###################################")
	}
}
