package block

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-logger-go"
)

type transactionCountersProvider interface {
	CurrentBlockTxs() uint64
	TotalTxs() uint64
	IsInterfaceNil() bool
}

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
	countersProvider transactionCountersProvider,
	header *block.MetaBlock,
	body *block.Body,
	headerHash []byte,
	numShardHeadersFromPool int,
	blockTracker process.BlockTracker,
) {
	if check.IfNil(countersProvider) {
		log.Warn("programming error in headersCounter.displayLogInfo - nil countersProvider")
		return
	}

	hc.calculateNumOfShardMBHeaders(header)

	dispHeader, dispLines := hc.createDisplayableMetaHeader(header)
	dispLines = hc.displayTxBlockBody(dispLines, header, body)

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

	log.Debug("metablock metrics info",
		"total txs processed", countersProvider.TotalTxs(),
		"block txs processed", countersProvider.CurrentBlockTxs(),
		"hash", headerHash,
	)

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

func (hc *headersCounter) displayTxBlockBody(
	lines []*display.LineData,
	header data.HeaderHandler,
	body *block.Body,
) []*display.LineData {
	currentBlockTxs := 0

	miniBlockHeaders := header.GetMiniBlockHeadersHashes()
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]

		scheduledModeInMiniBlock := miniBlock.IsScheduledMiniBlock()
		executionTypeInMiniBlockStr := ""
		if scheduledModeInMiniBlock {
			executionTypeInMiniBlockStr = "S_"
		}

		part := fmt.Sprintf("%s_MiniBlock_%s%d->%d",
			miniBlock.Type.String(),
			executionTypeInMiniBlockStr,
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID)

		if miniBlock.TxHashes == nil || len(miniBlock.TxHashes) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				part, "", "<EMPTY>"}))
		}

		if len(miniBlockHeaders) > i {
			lines = append(lines, display.NewLineData(false, []string{"", "MbHash", logger.DisplayByteSlice(miniBlockHeaders[i])}))
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
