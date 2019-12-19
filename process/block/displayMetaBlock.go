package block

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/display"
)

type headersCounter struct {
	shardMBHeaderCounterMutex           sync.RWMutex
	shardMBHeadersCurrentBlockProcessed uint64
	shardMBHeadersTotalProcessed        uint64
}

// NewHeaderCounter returns a new object that keeps track of how many headers
// were processed in total, and in the current block
func NewHeaderCounter() *headersCounter {
	return &headersCounter{
		shardMBHeaderCounterMutex:           sync.RWMutex{},
		shardMBHeadersCurrentBlockProcessed: 0,
		shardMBHeadersTotalProcessed:        0,
	}
}

func (hc *headersCounter) subtractRestoredMBHeaders(numMiniBlockHeaders int) {
	hc.shardMBHeaderCounterMutex.Lock()
	defer hc.shardMBHeaderCounterMutex.Unlock()
	if hc.shardMBHeadersTotalProcessed < uint64(numMiniBlockHeaders) {
		hc.shardMBHeadersTotalProcessed = 0
		return
	}

	hc.shardMBHeadersTotalProcessed -= uint64(numMiniBlockHeaders)
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
		log.Debug("CreateTableString", "error", err.Error())
		return
	}

	hc.shardMBHeaderCounterMutex.RLock()
	message := fmt.Sprintf("header hash: %s\n%s", display.DisplayByteSlice(headerHash), tblString)
	arguments := []interface{}{
		"total MB processed", hc.shardMBHeadersTotalProcessed,
		"block MB processed", hc.shardMBHeadersCurrentBlockProcessed,
		"shard headers in pool", numHeadersFromPool,
	}
	hc.shardMBHeaderCounterMutex.RUnlock()

	log.Debug(message, arguments...)
}

func (hc *headersCounter) createDisplayableMetaHeader(
	header *block.MetaBlock,
) ([]string, []*display.LineData) {

	tableHeader := []string{"Part", "Parameter", "Value"}

	metaLinesHeader := []*display.LineData{
		display.NewLineData(false, []string{
			"Header",
			"Block type",
			"MetaBlock"}),
	}

	lines := displayHeader(header)

	metaLines := make([]*display.LineData, 0, len(lines)+len(metaLinesHeader))
	metaLines = append(metaLines, metaLinesHeader...)
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
			display.DisplayByteSlice(shardData.HeaderHash)}))

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
					display.DisplayByteSlice(shardData.ShardMiniBlockHeaders[j].Hash)}))
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
					display.DisplayByteSlice(miniBlock.TxHashes[j])}))

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
