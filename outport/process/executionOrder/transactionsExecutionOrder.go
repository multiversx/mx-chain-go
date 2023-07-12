package executionOrder

import (
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	txsSort "github.com/multiversx/mx-chain-core-go/core/transaction"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-go/common"
	processOut "github.com/multiversx/mx-chain-go/outport/process"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

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
	pool *outport.TransactionPool,
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
	resultsTxsToMe, err := extractNormalTransactionAndScrsToMe(pool, blockBody, header)
	if err != nil {
		return nil, nil, err
	}

	// need to be sorted
	resultTxsFromMe, err := s.extractTransactionsGroupedFromMe(pool, blockBody, header, scheduledMbsFromPreviousBlock)
	if err != nil {
		return nil, nil, err
	}

	s.sortTransactions(resultTxsFromMe.transactionsFromMe, header)

	rewardsTxs, err := getRewardsTxsFromMe(pool, blockBody, header)
	if err != nil {
		return nil, nil, err
	}

	// scheduled from me, need to be sorted
	s.sortTransactions(resultTxsFromMe.scheduledTransactionsFromMe, header)

	allTransaction := append(resultsTxsToMe.transactionsToMe, resultTxsFromMe.transactionsFromMe...)
	allTransaction = append(allTransaction, rewardsTxs...)
	allTransaction = append(allTransaction, resultsTxsToMe.scheduledTransactionsToMe...)
	allTransaction = append(allTransaction, resultTxsFromMe.scheduledTransactionsFromMe...)

	for idx, tx := range allTransaction {
		tx.SetExecutionOrder(uint32(idx))
	}

	scheduledExecutedSCRSHashesPrevBlock := setOrderSmartContractResults(pool, scheduledMbsFromPreviousBlock, resultsTxsToMe.scrsToMe)

	return scheduledExecutedSCRSHashesPrevBlock, resultTxsFromMe.scheduledExecutedInvalidTxsHashesPrevBlock, nil
}

func (s *sorter) sortTransactions(transactions []data.TxWithExecutionOrderHandler, header data.HeaderHandler) {
	currentEpoch := s.enableEpochsHandler.GetCurrentEpoch()
	if s.enableEpochsHandler.IsFrontRunningProtectionFlagEnabledInEpoch(currentEpoch) {
		txsSort.SortTransactionsBySenderAndNonceWithFrontRunningProtectionExtendedTransactions(transactions, s.hasher, header.GetPrevRandSeed())
	} else {
		txsSort.SortTransactionsBySenderAndNonceExtendedTransactions(transactions)
	}
}

func (s *sorter) extractTransactionsGroupedFromMe(
	pool *outport.TransactionPool, blockBody *block.Body, header data.HeaderHandler, scheduledMbsFromPreviousBlock []*block.MiniBlock,
) (*resultsTransactionsFromMe, error) {
	transactionsFromMe := make([]data.TxWithExecutionOrderHandler, 0)
	scheduledTransactionsFromMe := make([]data.TxWithExecutionOrderHandler, 0)

	scheduledExecutedInvalidTxsHashesPrevBlock := make([]string, 0)
	for mbIndex, mb := range blockBody.MiniBlocks {
		var txs []data.TxWithExecutionOrderHandler
		var err error
		if isScheduledMBProcessed(header, mbIndex) {
			continue
		}

		isFromMe := mb.SenderShardID == header.GetShardID()
		if !isFromMe {
			continue
		}

		if mb.Type == block.TxBlock {
			txs, err = extractTxsFromMap(mb.TxHashes, pool.Transactions)
		}
		if mb.Type == block.InvalidBlock {
			var scheduledExecutedInvalidTxsHashesCurrentMB []string
			txs, scheduledExecutedInvalidTxsHashesCurrentMB, err = s.getInvalidTxsExecutedInCurrentBlock(scheduledMbsFromPreviousBlock, mb, pool)
			scheduledExecutedInvalidTxsHashesPrevBlock = append(scheduledExecutedInvalidTxsHashesPrevBlock, scheduledExecutedInvalidTxsHashesCurrentMB...)
		}

		if err != nil {
			return nil, err
		}

		if isScheduledMBNotProcessed(header, mbIndex) {
			scheduledTransactionsFromMe = append(scheduledTransactionsFromMe, txs...)
		} else {
			transactionsFromMe = append(transactionsFromMe, txs...)
		}
	}

	return &resultsTransactionsFromMe{
		transactionsFromMe:                         transactionsFromMe,
		scheduledTransactionsFromMe:                scheduledTransactionsFromMe,
		scheduledExecutedInvalidTxsHashesPrevBlock: scheduledExecutedInvalidTxsHashesPrevBlock,
	}, nil
}

func (s *sorter) getInvalidTxsExecutedInCurrentBlock(scheduledMbsFromPreviousBlock []*block.MiniBlock, mb *block.MiniBlock, pool *outport.TransactionPool) ([]data.TxWithExecutionOrderHandler, []string, error) {
	if len(scheduledMbsFromPreviousBlock) == 0 {
		txs, err := extractTxsFromMap(mb.TxHashes, pool.InvalidTxs)
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
			scheduledExecutedInvalidTxsHashesPrevBlock = append(scheduledExecutedInvalidTxsHashesPrevBlock, hex.EncodeToString(hash))
			continue
		}
		invalidTxHashes = append(invalidTxHashes, hash)
	}

	txs, err := extractTxsFromMap(invalidTxHashes, pool.InvalidTxs)
	return txs, scheduledExecutedInvalidTxsHashesPrevBlock, err
}

func extractNormalTransactionAndScrsToMe(pool *outport.TransactionPool, blockBody *block.Body, header data.HeaderHandler) (*resultsTransactionsToMe, error) {
	transactionsToMe := make([]data.TxWithExecutionOrderHandler, 0)
	scheduledTransactionsToMe := make([]data.TxWithExecutionOrderHandler, 0)
	scrsToMe := make(map[string]data.TxWithExecutionOrderHandler)

	for mbIndex, mb := range blockBody.MiniBlocks {
		var err error
		var txs []data.TxWithExecutionOrderHandler
		if isScheduledMBProcessed(header, mbIndex) {
			continue
		}

		isToMeCross := mb.ReceiverShardID == header.GetShardID() && mb.SenderShardID != mb.ReceiverShardID
		if !isToMeCross {
			continue
		}

		executedTxsHashes := extractExecutedTxHashes(mbIndex, mb.TxHashes, header)
		if mb.Type == block.TxBlock {
			txs, err = extractTxsFromMap(executedTxsHashes, pool.Transactions)
		}
		if mb.Type == block.SmartContractResultBlock {
			txs, err = extractSCRsFromMap(executedTxsHashes, pool.SmartContractResults)
			extractAndPutScrsToDestinationMap(executedTxsHashes, pool.SmartContractResults, scrsToMe)
		}
		if mb.Type == block.RewardsBlock {
			txs, err = extractRewardsFromMap(executedTxsHashes, pool.Rewards)
		}
		if err != nil {
			return nil, err
		}

		if isScheduledMBNotProcessed(header, mbIndex) {
			scheduledTransactionsToMe = append(scheduledTransactionsToMe, txs...)
		} else {
			transactionsToMe = append(transactionsToMe, txs...)
		}
	}

	return &resultsTransactionsToMe{
		transactionsToMe:          transactionsToMe,
		scheduledTransactionsToMe: scheduledTransactionsToMe,
		scrsToMe:                  scrsToMe,
	}, nil
}

func getRewardsTxsFromMe(pool *outport.TransactionPool, blockBody *block.Body, header data.HeaderHandler) ([]data.TxWithExecutionOrderHandler, error) {
	rewardsTxsHashes := make([][]byte, 0)
	rewardsTxs := make([]data.TxWithExecutionOrderHandler, 0)
	if header.GetShardID() != core.MetachainShardId {
		return rewardsTxs, nil
	}

	for _, mb := range blockBody.MiniBlocks {
		if mb.Type != block.RewardsBlock {
			continue
		}
		rewardsTxsHashes = append(rewardsTxsHashes, mb.TxHashes...)
	}

	return extractRewardsFromMap(rewardsTxsHashes, pool.Rewards)
}

func extractTxsFromMap(txsHashes [][]byte, txs map[string]*outport.TxInfo) ([]data.TxWithExecutionOrderHandler, error) {
	result := make([]data.TxWithExecutionOrderHandler, 0, len(txsHashes))
	for _, txHash := range txsHashes {
		txHashHex := hex.EncodeToString(txHash)
		tx, found := txs[txHashHex]
		if !found {
			return nil, fmt.Errorf("cannot find transaction in pool, txHash: %s", txHashHex)
		}
		result = append(result, tx)
	}

	return result, nil
}

func extractSCRsFromMap(txsHashes [][]byte, scrs map[string]*outport.SCRInfo) ([]data.TxWithExecutionOrderHandler, error) {
	result := make([]data.TxWithExecutionOrderHandler, 0, len(txsHashes))
	for _, txHash := range txsHashes {
		txHashHex := hex.EncodeToString(txHash)
		scr, found := scrs[txHashHex]
		if !found {
			return nil, fmt.Errorf("cannot find scr in pool, txHash: %s", txHashHex)
		}
		result = append(result, scr)
	}

	return result, nil
}

func extractRewardsFromMap(txsHashes [][]byte, rewards map[string]*outport.RewardInfo) ([]data.TxWithExecutionOrderHandler, error) {
	result := make([]data.TxWithExecutionOrderHandler, 0, len(txsHashes))
	for _, txHash := range txsHashes {
		txHashHex := hex.EncodeToString(txHash)
		reward, found := rewards[txHashHex]
		if !found {
			return nil, fmt.Errorf("cannot find reward in pool, txHash: %s", txHashHex)
		}
		result = append(result, reward)
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

func extractAndPutScrsToDestinationMap(scrsHashes [][]byte, scrsMap map[string]*outport.SCRInfo, destinationMap map[string]data.TxWithExecutionOrderHandler) {
	for _, scrHash := range scrsHashes {
		scrHashHex := hex.EncodeToString(scrHash)
		scr, found := scrsMap[scrHashHex]
		if !found {
			continue
		}
		destinationMap[string(scrHash)] = scr
	}
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
