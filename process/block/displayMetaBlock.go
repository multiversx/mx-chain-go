package block

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/process"
)

type headersCounter struct {
	shardMBHeaderCounterMutex           sync.RWMutex
	shardMBHeadersCurrentBlockProcessed uint64
	shardMBHeadersTotalProcessed        uint64
	peakTPS                             uint64
}

// NewHeaderCounter returns a new object that keeps track of how many headers
// were processed in total, and in the current block
func NewHeaderCounter() *headersCounter {
	return &headersCounter{
		shardMBHeaderCounterMutex:           sync.RWMutex{},
		shardMBHeadersCurrentBlockProcessed: 0,
		shardMBHeadersTotalProcessed:        0,
		peakTPS:                             0,
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
	body *block.Body,
	headerHash []byte,
	numShardHeadersFromPool int,
	blockTracker process.BlockTracker,
	roundDuration uint64,
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
	message := fmt.Sprintf("header hash: %s\n%s", logger.DisplayByteSlice(headerHash), tblString)
	arguments := []interface{}{
		"total MB processed", hc.shardMBHeadersTotalProcessed,
		"block MB processed", hc.shardMBHeadersCurrentBlockProcessed,
		"shard headers in pool", numShardHeadersFromPool,
	}
	hc.shardMBHeaderCounterMutex.RUnlock()

	log.Debug(message, arguments...)

	numTxs := getNumTxs(header, body)
	tps := numTxs / roundDuration
	if tps > hc.peakTPS {
		hc.peakTPS = tps
	}

	log.Debug("tps info",
		"shard", header.GetShardID(),
		"round", header.GetRound(),
		"nonce", header.GetNonce(),
		"num txs", numTxs,
		"tps", tps,
		"peak tps", hc.peakTPS)

	blockTracker.DisplayTrackedHeaders()
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

	var lines []*display.LineData
	if header.IsStartOfEpochBlock() {
		lines = displayEpochStartMetaBlock(header)
	} else {
		lines = displayHeader(header)
	}

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
			logger.DisplayByteSlice(shardData.HeaderHash)}))

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
					logger.DisplayByteSlice(shardData.ShardMiniBlockHeaders[j].Hash)}))
			} else if j == 1 {
				lines = append(lines, display.NewLineData(false, []string{
					"",
					"...",
					"...",
				}))
			}
		}

		lines[len(lines)-1].HorizontalRuleAfter = true
	}

	return lines
}

func (hc *headersCounter) displayTxBlockBody(lines []*display.LineData, body *block.Body) []*display.LineData {
	currentBlockTxs := 0

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]

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

	return lines
}

func (hc *headersCounter) getNumShardMBHeadersTotalProcessed() uint64 {
	hc.shardMBHeaderCounterMutex.Lock()
	defer hc.shardMBHeaderCounterMutex.Unlock()

	return hc.shardMBHeadersTotalProcessed
}

func displayEpochStartMetaBlock(block *block.MetaBlock) []*display.LineData {
	lines := displayHeader(block)
	economicsLines := displayEconomicsData(block.EpochStart.Economics)
	lines = append(lines, economicsLines...)

	for _, shardData := range block.EpochStart.LastFinalizedHeaders {
		shardDataLines := displayEpochStartShardData(shardData)
		lines = append(lines, shardDataLines...)
	}

	return lines
}

func displayEpochStartShardData(shardData block.EpochStartShardData) []*display.LineData {
	lines := []*display.LineData{
		display.NewLineData(false, []string{
			fmt.Sprintf("EpochStartShardData - Shard %d", shardData.ShardID),
			"Round",
			fmt.Sprintf("%d", shardData.Round)}),
		display.NewLineData(false, []string{
			"",
			"Nonce",
			fmt.Sprintf("%d", shardData.Nonce)}),
		display.NewLineData(false, []string{
			"",
			"FirstPendingMetaBlock",
			logger.DisplayByteSlice(shardData.FirstPendingMetaBlock)}),
		display.NewLineData(false, []string{
			"",
			"LastFinishedMetaBlock",
			logger.DisplayByteSlice(shardData.LastFinishedMetaBlock)}),
		display.NewLineData(false, []string{
			"",
			"HeaderHash",
			logger.DisplayByteSlice(shardData.HeaderHash)}),
		display.NewLineData(false, []string{
			"",
			"RootHash",
			logger.DisplayByteSlice(shardData.RootHash)}),
		display.NewLineData(false, []string{
			"",
			"PendingMiniBlockHeaders count",
			fmt.Sprintf("%d", len(shardData.PendingMiniBlockHeaders))}),
	}

	lines[len(lines)-1].HorizontalRuleAfter = true

	return lines
}

func displayEconomicsData(economics block.Economics) []*display.LineData {
	return []*display.LineData{
		display.NewLineData(false, []string{
			"Epoch Start - Economics",
			"PrevEpochStartHash",
			logger.DisplayByteSlice(economics.PrevEpochStartHash)}),
		display.NewLineData(false, []string{
			"",
			"NodePrice",
			economics.NodePrice.String()}),
		display.NewLineData(false, []string{
			"",
			"RewardsPerBlock",
			economics.RewardsPerBlock.String()}),
		display.NewLineData(false, []string{
			"",
			"GenesisTotalSupply",
			economics.TotalSupply.String()}),
		display.NewLineData(false, []string{
			"",
			"TotalNewlyMinted",
			economics.TotalNewlyMinted.String()}),
		display.NewLineData(false, []string{
			"",
			"TotalToDistribute",
			economics.TotalToDistribute.String()}),
		display.NewLineData(true, []string{
			"",
			"PrevEpochStartRound",
			fmt.Sprintf("%d", economics.PrevEpochStartRound)}),
	}
}

func getNumTxs(metaBlock *block.MetaBlock, body *block.Body) uint64 {
	shardInfo := metaBlock.ShardInfo
	numTxs := uint64(0)
	for i := 0; i < len(shardInfo); i++ {
		shardMiniBlockHeaders := shardInfo[i].ShardMiniBlockHeaders
		numTxsInShardHeader := uint64(0)
		for j := 0; j < len(shardMiniBlockHeaders); j++ {
			numTxsInShardHeader += uint64(shardMiniBlockHeaders[j].TxCount)
		}

		log.Trace("txs info",
			"shard", shardInfo[i].GetShardID(),
			"round", shardInfo[i].GetRound(),
			"nonce", shardInfo[i].GetNonce(),
			"num txs", numTxsInShardHeader)

		numTxs += numTxsInShardHeader
	}

	numTxsInMetaBlock := uint64(0)
	for i := 0; i < len(body.MiniBlocks); i++ {
		numTxsInMetaBlock += uint64(len(body.MiniBlocks[i].TxHashes))
	}

	log.Trace("txs info",
		"shard", metaBlock.GetShardID(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
		"num txs", numTxsInMetaBlock)

	numTxs += numTxsInMetaBlock

	return numTxs
}
