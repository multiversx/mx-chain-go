package block

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

type headersCounter struct {
	shardMBHeaderCounterMutex           sync.RWMutex
	shardMBHeadersCurrentBlockProcessed uint64
	shardMBHeadersTotalProcessed        uint64
}

// NewHeaderCounter returns a new object that keeps track of how many headers
// were processed in total, and in the current block
func NewHeaderCounter(store dataRetriever.StorageService, marshalizer marshal.Marshalizer) *headersCounter {
	shardMBHeadersTotalProcessed := getNumObjFromStorage(store, marshalizer, core.MetricNumShardHeadersProcessed)

	return &headersCounter{
		shardMBHeaderCounterMutex:           sync.RWMutex{},
		shardMBHeadersCurrentBlockProcessed: 0,
		shardMBHeadersTotalProcessed:        shardMBHeadersTotalProcessed,
	}
}

func (hc *headersCounter) subtractRestoredMBHeaders(numMiniBlockHeaders int) {
	hc.shardMBHeaderCounterMutex.Lock()
	hc.shardMBHeadersTotalProcessed -= uint64(numMiniBlockHeaders)
	hc.shardMBHeaderCounterMutex.Unlock()
}

func (hc *headersCounter) countShardMBHeaders(numShardMBHeaders int) {
	hc.shardMBHeaderCounterMutex.Lock()
	hc.shardMBHeadersCurrentBlockProcessed += uint64(numShardMBHeaders)
	hc.shardMBHeadersTotalProcessed += uint64(numShardMBHeaders)
	hc.shardMBHeaderCounterMutex.Unlock()
}

func (hc *headersCounter) calculateNumOfShardMBHeaders(header *block.MetaBlock) {
	hc.shardMBHeaderCounterMutex.Lock()
	hc.shardMBHeadersCurrentBlockProcessed = 0
	hc.shardMBHeaderCounterMutex.Unlock()

	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		hc.countShardMBHeaders(len(shardData.ShardMiniBlockHeaders))
	}
}

func (hc *headersCounter) displayLogInfo(
	header *block.MetaBlock,
	body block.Body,
	headerHash []byte,
	numHeadersFromPool int,
) {
	hc.calculateNumOfShardMBHeaders(header)

	dispHeader, dispLines := hc.createDisplayableMetaHeader(header)
	dispLines = hc.displayTxBlockBody(dispLines, body)

	tblString, err := display.CreateTableString(dispHeader, dispLines)
	if err != nil {
		log.Error(err.Error())
		return
	}

	hc.shardMBHeaderCounterMutex.RLock()
	tblString = tblString + fmt.Sprintf("\nHeader hash: %s\n\nTotal shard MB headers "+
		"processed until now: %d. Total shard MB headers processed for this block: %d. Total shard headers remained in pool: %d\n",
		core.ToB64(headerHash),
		hc.shardMBHeadersTotalProcessed,
		hc.shardMBHeadersCurrentBlockProcessed,
		numHeadersFromPool)
	hc.shardMBHeaderCounterMutex.RUnlock()

	log.Info(tblString)
}

func (hc *headersCounter) createDisplayableMetaHeader(
	header *block.MetaBlock,
) ([]string, []*display.LineData) {

	tableHeader := []string{"Part", "Parameter", "Value"}

	lines := displayHeader(header)

	metaLines := make([]*display.LineData, 0)
	metaLines = append(metaLines, display.NewLineData(false, []string{
		"Header",
		"Block type",
		"MetaBlock"}))
	metaLines = append(metaLines, lines...)
	metaLines = hc.displayShardInfo(metaLines, header)

	return tableHeader, metaLines
}

func (hc *headersCounter) displayShardInfo(lines []*display.LineData, header *block.MetaBlock) []*display.LineData {
	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]

		lines = append(lines, display.NewLineData(false, []string{
			fmt.Sprintf("ShardData_%d", shardData.ShardID),
			"Header hash",
			base64.StdEncoding.EncodeToString(shardData.HeaderHash)}))

		if shardData.ShardMiniBlockHeaders == nil || len(shardData.ShardMiniBlockHeaders) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				"", "ShardMiniBlockHeaders", "<EMPTY>"}))
		}

		for j := 0; j < len(shardData.ShardMiniBlockHeaders); j++ {
			if j == 0 || j >= len(shardData.ShardMiniBlockHeaders)-1 {
				senderShard := shardData.ShardMiniBlockHeaders[j].SenderShardID
				receiverShard := shardData.ShardMiniBlockHeaders[j].ReceiverShardID

				lines = append(lines, display.NewLineData(false, []string{
					"",
					fmt.Sprintf("%d ShardMiniBlockHeaderHash_%d_%d", j+1, senderShard, receiverShard),
					core.ToB64(shardData.ShardMiniBlockHeaders[j].Hash)}))
			} else if j == 1 {
				lines = append(lines, display.NewLineData(false, []string{
					"",
					fmt.Sprintf("..."),
					fmt.Sprintf("...")}))
			}
		}

		lines[len(lines)-1].HorizontalRuleAfter = true
	}

	return lines
}

func (hc *headersCounter) displayTxBlockBody(lines []*display.LineData, body block.Body) []*display.LineData {
	currentBlockTxs := 0

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]

		part := fmt.Sprintf("%s_MiniBlock_%d->%d",
			miniBlock.Type.String(),
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

func (hc *headersCounter) getNumShardMBHeadersTotalProcessed() uint64 {
	hc.shardMBHeaderCounterMutex.Lock()
	defer hc.shardMBHeaderCounterMutex.Unlock()

	return hc.shardMBHeadersTotalProcessed
}
