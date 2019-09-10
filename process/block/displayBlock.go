package block

import (
    "fmt"
    "sync"

    "github.com/ElrondNetwork/elrond-go/core"
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

// getNumTxsFromPool returns the number of transactions from pool for a given shard
func (txc *transactionCounter) getNumTxsFromPool(shardId uint32, dataPool dataRetriever.PoolsHolder, nrShards uint32) int {
    txPool := dataPool.Transactions()
    if txPool == nil {
        return 0
    }

    sumTxs := 0

    strCache := process.ShardCacherIdentifier(shardId, shardId)
    txStore := txPool.ShardDataStore(strCache)
    if txStore != nil {
        sumTxs += txStore.Len()
    }

    for i := uint32(0); i < nrShards; i++ {
        if i == shardId {
            continue
        }

        strCache = process.ShardCacherIdentifier(i, shardId)
        txStore = txPool.ShardDataStore(strCache)
        if txStore != nil {
            sumTxs += txStore.Len()
        }

        strCache = process.ShardCacherIdentifier(shardId, i)
        txStore = txPool.ShardDataStore(strCache)
        if txStore != nil {
            sumTxs += txStore.Len()
        }
    }

    return sumTxs
}

// substractRestoredTxs updated the total processed txs in case of restore
func (txc *transactionCounter) substractRestoredTxs(txsNr int) {
    txc.mutex.Lock()
    txc.totalTxs = txc.totalTxs - txsNr
    txc.mutex.Unlock()
}

// displayLogInfo writes to the output information about the block and transactions
func (txc *transactionCounter) displayLogInfo(
    header *block.Header,
    body block.Body,
    headerHash []byte,
    numShards uint32,
    selfId uint32,
    dataPool dataRetriever.PoolsHolder,
) {
    dispHeader, dispLines := txc.createDisplayableShardHeaderAndBlockBody(header, body)

    tblString, err := display.CreateTableString(dispHeader, dispLines)
    if err != nil {
        log.Error(err.Error())
        return
    }

    txc.mutex.RLock()
    tblString = tblString + fmt.Sprintf("\nHeader hash: %s\n\n"+
        "Total txs processed until now: %d. Total txs processed for this block: %d. Total txs remained in pool: %d\n\n"+
        "Total shards: %d. Current shard id: %d\n",
        core.ToB64(headerHash),
        txc.totalTxs,
        txc.currentBlockTxs,
        txc.getNumTxsFromPool(selfId, dataPool, numShards),
        numShards,
        selfId)
    txc.mutex.RUnlock()
    log.Info(tblString)
}

func (txc *transactionCounter) createDisplayableShardHeaderAndBlockBody(
    header *block.Header,
    body block.Body,
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

    part := fmt.Sprintf("MetaBlockHashes")
    for i := 0; i < len(header.MetaBlockHashes); i++ {
        if i == 0 || i >= len(header.MetaBlockHashes)-1 {
            lines = append(lines, display.NewLineData(false, []string{
                part,
                fmt.Sprintf("MetaBlockHash_%d", i+1),
                core.ToB64(header.MetaBlockHashes[i])}))

            part = ""
        } else if i == 1 {
            lines = append(lines, display.NewLineData(false, []string{
                part,
                fmt.Sprintf("..."),
                fmt.Sprintf("...")}))

            part = ""
        }
    }

    lines[len(lines)-1].HorizontalRuleAfter = true

    return lines
}

func (txc *transactionCounter) displayTxBlockBody(lines []*display.LineData, body block.Body) []*display.LineData {
    currentBlockTxs := 0

    for i := 0; i < len(body); i++ {
        miniBlock := body[i]

        part := fmt.Sprintf("MiniBlock_%d_%d", miniBlock.SenderShardID, miniBlock.ReceiverShardID)

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

    txc.mutex.Lock()
    txc.currentBlockTxs = currentBlockTxs
    txc.totalTxs += currentBlockTxs
    txc.mutex.Unlock()

    return lines
}

func DisplayLastNotarized(
    marshalizer marshal.Marshalizer,
    hasher hashing.Hasher,
    lastNotarizedHdrForShard data.HeaderHandler,
    shardId uint32) {
    lastNotarizedHdrHashForShard, errNotCritical := core.CalculateHash(
        marshalizer,
        hasher,
        lastNotarizedHdrForShard)
    if errNotCritical != nil {
        log.Debug(errNotCritical.Error())
    }

    log.Info(fmt.Sprintf("last notarized block from shard %d has: round = %d, nonce = %d, hash = %s\n",
        shardId,
        lastNotarizedHdrForShard.GetRound(),
        lastNotarizedHdrForShard.GetNonce(),
        core.ToB64(lastNotarizedHdrHashForShard)))
}
