package block

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/process"
)

type transactionCounter struct {
	mutex           sync.RWMutex
	currentBlockTxs int
	totalTxs        int
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

// GetNumTxsWithDst returns the number of transactions for a certain destination shard
func GetNumTxsWithDst(dstShardId uint32, dataPool dataRetriever.PoolsHolder, nrShards uint32) int {
	txPool := dataPool.Transactions()
	if txPool == nil {
		return 0
	}

	sumTxs := 0

	for i := uint32(0); i < nrShards; i++ {
		strCache := process.ShardCacherIdentifier(i, dstShardId)
		txStore := txPool.ShardDataStore(strCache)
		if txStore == nil {
			continue
		}
		sumTxs += txStore.Len()
	}

	return sumTxs
}

// SubstractRestoredTxs updated the total processed txs in case of restore
func (txc *transactionCounter) SubstractRestoredTxs(txsNr int) {
	txc.mutex.Lock()
	txc.totalTxs = txc.totalTxs - txsNr
	txc.mutex.Unlock()
}

// DisplayLogInfo writes to the output information about the block and transactions
func DisplayLogInfo(
	header *block.Header,
	body block.Body,
	headerHash []byte,
	numShards uint32,
	selfId uint32,
	dataPool dataRetriever.PoolsHolder,
	txCounter *transactionCounter,
) {
	dispHeader, dispLines := createDisplayableShardHeaderAndBlockBody(header, body, txCounter)

	tblString, err := display.CreateTableString(dispHeader, dispLines)
	if err != nil {
		log.Error(err.Error())
		return
	}

	txCounter.mutex.RLock()
	tblString = tblString + fmt.Sprintf("\nHeader hash: %s\n\n"+
		"Total txs processed until now: %d. Total txs processed for this block: %d. Total txs remained in pool: %d\n\n"+
		"Total shards: %d. Current shard id: %d\n",
		core.ToB64(headerHash),
		txCounter.totalTxs,
		txCounter.currentBlockTxs,
		GetNumTxsWithDst(selfId, dataPool, numShards),
		numShards,
		selfId)
	txCounter.mutex.RUnlock()
	log.Info(tblString)
}

func createDisplayableShardHeaderAndBlockBody(
	header *block.Header,
	body block.Body,
	txCounter *transactionCounter,
) ([]string, []*display.LineData) {

	tableHeader := []string{"Part", "Parameter", "Value"}

	lines := displayHeader(header)

	shardLines := make([]*display.LineData, 0)
	shardLines = append(shardLines, display.NewLineData(false, []string{
		"Header",
		"Block type",
		"TxBlock"}))
	shardLines = append(shardLines, display.NewLineData(false, []string{
		"",
		"Shard",
		fmt.Sprintf("%d", header.ShardId)}))
	shardLines = append(shardLines, lines...)

	if header.BlockBodyType == block.TxBlock {
		shardLines = displayTxBlockBody(shardLines, body, txCounter)

		return tableHeader, shardLines
	}

	// TODO: implement the other block bodies

	shardLines = append(shardLines, display.NewLineData(false, []string{"Unknown", "", ""}))
	return tableHeader, shardLines
}

func displayTxBlockBody(lines []*display.LineData, body block.Body, txCounter *transactionCounter) []*display.LineData {
	txCounter.mutex.Lock()
	txCounter.currentBlockTxs = 0
	txCounter.mutex.Unlock()

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]

		part := fmt.Sprintf("MiniBlock_%d", miniBlock.ReceiverShardID)

		if miniBlock.TxHashes == nil || len(miniBlock.TxHashes) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				part, "", "<EMPTY>"}))
		}

		txCounter.mutex.Lock()
		txCounter.currentBlockTxs += len(miniBlock.TxHashes)
		txCounter.totalTxs += len(miniBlock.TxHashes)
		txCounter.mutex.Unlock()

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if j == 0 || j >= len(miniBlock.TxHashes)-1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					fmt.Sprintf("TxHash_%d", j+1),
					core.ToB64(miniBlock.TxHashes[j])}))

				part = ""
			} else if j == 1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					fmt.Sprintf("..."),
					fmt.Sprintf("...")}))

				part = ""
			}
		}

		lines[len(lines)-1].HorizontalRuleAfter = true
	}

	return lines
}
