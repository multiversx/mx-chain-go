package block

import (
	"fmt"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/counting"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

type transactionCounter struct {
	mutex           sync.RWMutex
	currentBlockTxs uint64
	totalTxs        uint64
}

// NewTransactionCounter returns a new object that keeps track of how many transactions
// were executed in total, and in the current block
func NewTransactionCounter() *transactionCounter {
	return &transactionCounter{
		mutex:           sync.RWMutex{},
		currentBlockTxs: 0,
		totalTxs:        0,
	}
}

func (txc *transactionCounter) getPoolCounts(poolsHolder dataRetriever.PoolsHolder) (txCounts counting.Counts, rewardCounts counting.Counts, unsignedCounts counting.Counts) {
	txCounts = poolsHolder.Transactions().GetCounts()
	rewardCounts = poolsHolder.RewardTransactions().GetCounts()
	unsignedCounts = poolsHolder.UnsignedTransactions().GetCounts()
	return
}

// subtractRestoredTxs updated the total processed txs in case of restore
func (txc *transactionCounter) subtractRestoredTxs(txsNr int) {
	txc.mutex.Lock()
	defer txc.mutex.Unlock()
	if txc.totalTxs < uint64(txsNr) {
		txc.totalTxs = 0
		return
	}

	txc.totalTxs -= uint64(txsNr)
}

// displayLogInfo writes to the output information about the block and transactions
func (txc *transactionCounter) displayLogInfo(
	header *block.Header,
	body *block.Body,
	headerHash []byte,
	numShards uint32,
	selfId uint32,
	dataPool dataRetriever.PoolsHolder,
	appStatusHandler core.AppStatusHandler,
	blockTracker process.BlockTracker,
) {
	dispHeader, dispLines := txc.createDisplayableShardHeaderAndBlockBody(header, body)

	txc.mutex.RLock()
	appStatusHandler.SetUInt64Value(core.MetricNumProcessedTxs, txc.totalTxs)
	txc.mutex.RUnlock()

	tblString, err := display.CreateTableString(dispHeader, dispLines)
	if err != nil {
		log.Debug("CreateTableString", "error", err.Error())
		return
	}

	txc.mutex.RLock()
	message := fmt.Sprintf("header hash: %s\n%s", logger.DisplayByteSlice(headerHash), tblString)
	arguments := []interface{}{
		"total txs processed", txc.totalTxs,
		"block txs processed", txc.currentBlockTxs,
		"num shards", numShards,
		"shard", selfId,
	}
	txc.mutex.RUnlock()
	log.Debug(message, arguments...)

	blockTracker.DisplayTrackedHeaders()
}

func (txc *transactionCounter) createDisplayableShardHeaderAndBlockBody(
	header *block.Header,
	body *block.Body,
) ([]string, []*display.LineData) {

	tableHeader := []string{"Part", "Parameter", "Value"}

	headerLines := []*display.LineData{
		display.NewLineData(false, []string{
			"Header",
			"Block type",
			"TxBlock"}),
		display.NewLineData(false, []string{
			"",
			"Shard",
			fmt.Sprintf("%d", header.ShardID)}),
	}

	lines := displayHeader(header)

	shardLines := make([]*display.LineData, 0, len(lines)+len(headerLines))
	shardLines = append(shardLines, headerLines...)
	shardLines = append(shardLines, lines...)

	if header.BlockBodyType == block.TxBlock {
		shardLines = txc.displayMetaHashesIncluded(shardLines, header)
		shardLines = txc.displayTxBlockBody(shardLines, body)

		return tableHeader, shardLines
	}

	// TODO: implement the other block bodies

	shardLines = append(shardLines, display.NewLineData(false, []string{"Unknown", "", ""}))
	return tableHeader, shardLines
}

func (txc *transactionCounter) displayMetaHashesIncluded(
	lines []*display.LineData,
	header *block.Header,
) []*display.LineData {

	if header.MetaBlockHashes == nil || len(header.MetaBlockHashes) == 0 {
		return lines
	}

	part := "MetaBlockHashes"
	for i := 0; i < len(header.MetaBlockHashes); i++ {
		if i == 0 || i >= len(header.MetaBlockHashes)-1 {
			lines = append(lines, display.NewLineData(false, []string{
				part,
				fmt.Sprintf("MetaBlockHash_%d", i+1),
				logger.DisplayByteSlice(header.MetaBlockHashes[i])}))

			part = ""
		} else if i == 1 {
			lines = append(lines, display.NewLineData(false, []string{
				part,
				"...",
				"...",
			}))

			part = ""
		}
	}

	lines[len(lines)-1].HorizontalRuleAfter = true

	return lines
}

func (txc *transactionCounter) displayTxBlockBody(lines []*display.LineData, body *block.Body) []*display.LineData {
	currentBlockTxs := 0

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]

		mbTypeStr := miniBlock.Type.String()
		if isScheduledMiniBlock(miniBlock) {
			//mbTypeStr = block.ScheduledBlock.String()
			mbTypeStr = "ScheduledBlock"
		}

		part := fmt.Sprintf("%s_MiniBlock_%d->%d",
			mbTypeStr,
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID)

		if miniBlock.TxHashes == nil || len(miniBlock.TxHashes) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				part, "", "<EMPTY>"}))
		}

		currentBlockTxs += len(miniBlock.TxHashes)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if j == 0 || j >= len(miniBlock.TxHashes)-1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					fmt.Sprintf("TxHash_%d", j+1),
					logger.DisplayByteSlice(miniBlock.TxHashes[j])}))

				part = ""
			} else if j == 1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					"...",
					"...",
				}))

				part = ""
			}
		}

		lines[len(lines)-1].HorizontalRuleAfter = true
	}

	txc.mutex.Lock()
	txc.currentBlockTxs = uint64(currentBlockTxs)
	txc.totalTxs += uint64(currentBlockTxs)
	txc.mutex.Unlock()

	return lines
}

// DisplayLastNotarized will display information about last notarized block
func DisplayLastNotarized(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	lastNotarizedHdrForShard data.HeaderHandler,
	shardId uint32) {

	if lastNotarizedHdrForShard == nil || lastNotarizedHdrForShard.IsInterfaceNil() {
		log.Debug("last notarized header for shard is nil")
		return
	}

	lastNotarizedHdrHashForShard, errNotCritical := core.CalculateHash(
		marshalizer,
		hasher,
		lastNotarizedHdrForShard)
	if errNotCritical != nil {
		log.Trace("CalculateHash", "error", errNotCritical.Error())
	}

	log.Debug("last notarized block from shard",
		"shard", shardId,
		"round", lastNotarizedHdrForShard.GetRound(),
		"nonce", lastNotarizedHdrForShard.GetNonce(),
		"hash", lastNotarizedHdrHashForShard)
}
