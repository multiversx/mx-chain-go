package block

import (
	"fmt"
	"math"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/counting"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type transactionCounter struct {
	mutex            sync.RWMutex
	currentBlockTxs  uint64
	totalTxs         uint64
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	appStatusHandler core.AppStatusHandler
	shardID          uint32
}

// ArgsTransactionCounter represents the arguments needed to create a new transaction counter
type ArgsTransactionCounter struct {
	AppStatusHandler core.AppStatusHandler
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
	ShardID          uint32
}

// NewTransactionCounter returns a new object that keeps track of how many transactions
// were executed in total, and in the current block
func NewTransactionCounter(args ArgsTransactionCounter) (*transactionCounter, error) {
	if check.IfNil(args.AppStatusHandler) {
		return nil, process.ErrNilAppStatusHandler
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	return &transactionCounter{
		mutex:            sync.RWMutex{},
		appStatusHandler: args.AppStatusHandler,
		currentBlockTxs:  0,
		totalTxs:         0,
		hasher:           args.Hasher,
		marshalizer:      args.Marshalizer,
		shardID:          args.ShardID,
	}, nil
}

func (txc *transactionCounter) getPoolCounts(poolsHolder dataRetriever.PoolsHolder) (txCounts counting.Counts, rewardCounts counting.Counts, unsignedCounts counting.Counts) {
	txCounts = poolsHolder.Transactions().GetCounts()
	rewardCounts = poolsHolder.RewardTransactions().GetCounts()
	unsignedCounts = poolsHolder.UnsignedTransactions().GetCounts()
	return
}

// headerReverted updates the total processed txs in case of restore. It also sets the current block txs to 0
func (txc *transactionCounter) headerReverted(hdr data.HeaderHandler) {
	if check.IfNil(hdr) {
		log.Warn("programming error: nil header in transactionCounter.headerReverted function")
		return
	}

	currentBlockTxs := txc.getProcessedTxCount(hdr)

	txc.mutex.Lock()
	txc.currentBlockTxs = 0
	txc.safeSubtractTotalTxs(uint64(currentBlockTxs))
	txc.appStatusHandler.SetUInt64Value(common.MetricNumProcessedTxs, txc.totalTxs)
	txc.mutex.Unlock()
}

func (txc *transactionCounter) safeSubtractTotalTxs(delta uint64) {
	if txc.totalTxs < delta {
		txc.totalTxs = 0
		return
	}

	txc.totalTxs -= delta
}

func (txc *transactionCounter) headerExecuted(hdr data.HeaderHandler) {
	if check.IfNil(hdr) {
		log.Warn("programming error: nil header in transactionCounter.headerExecuted function")
		return
	}

	currentBlockTxs := txc.getProcessedTxCount(hdr)

	txc.mutex.Lock()
	txc.currentBlockTxs = uint64(currentBlockTxs)
	txc.totalTxs += uint64(currentBlockTxs)
	txc.appStatusHandler.SetUInt64Value(common.MetricNumProcessedTxs, txc.totalTxs)
	txc.mutex.Unlock()
}

func (txc *transactionCounter) getProcessedTxCount(hdr data.HeaderHandler) int32 {
	currentBlockTxs := int32(0)
	for _, miniBlockHeaderHandler := range hdr.GetMiniBlockHeaderHandlers() {
		if miniBlockHeaderHandler.GetTypeInt32() == int32(block.PeerBlock) {
			continue
		}

		isMiniblockScheduledFromMe := miniBlockHeaderHandler.GetSenderShardID() == txc.shardID &&
			miniBlockHeaderHandler.GetProcessingType() == int32(block.Scheduled)
		if isMiniblockScheduledFromMe {
			continue
		}

		currentBlockTxs += miniBlockHeaderHandler.GetIndexOfLastTxProcessed() - miniBlockHeaderHandler.GetIndexOfFirstTxProcessed() + 1
	}

	return currentBlockTxs
}

// displayLogInfo writes to the output information about the block and transactions
func (txc *transactionCounter) displayLogInfo(
	header data.HeaderHandler,
	body *block.Body,
	headerHash []byte,
	numShards uint32,
	selfId uint32,
	_ dataRetriever.PoolsHolder,
	blockTracker process.BlockTracker,
) {
	dispHeader, dispLines := txc.createDisplayableShardHeaderAndBlockBody(header, body)

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
		"shard", getShardName(selfId),
	}
	txc.mutex.RUnlock()
	log.Debug(message, arguments...)

	blockTracker.DisplayTrackedHeaders()
}

func (txc *transactionCounter) createDisplayableShardHeaderAndBlockBody(
	header data.HeaderHandler,
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
			getShardName(header.GetShardID())}),
	}

	lines := displayHeader(header)

	shardLines := make([]*display.LineData, 0, len(lines)+len(headerLines))
	shardLines = append(shardLines, headerLines...)
	shardLines = append(shardLines, lines...)

	shardHeaderHashesGetter, ok := header.(extendedShardHeaderHashesGetter)
	if ok {
		shardLines = txc.displayExtendedShardHeaderHashesIncluded(shardLines, shardHeaderHashesGetter.GetExtendedShardHeaderHashes())
	}

	var varBlockBodyType int32 = math.MaxInt32
	shardHeader, ok := header.(data.ShardHeaderHandler)
	if ok {
		varBlockBodyType = shardHeader.GetBlockBodyTypeInt32()
	}

	if varBlockBodyType == int32(block.TxBlock) {
		shardLines = txc.displayMetaHashesIncluded(shardLines, shardHeader)
		shardLines = txc.displayTxBlockBody(shardLines, header, body)

		return tableHeader, shardLines
	}

	// TODO: implement the other block bodies

	shardLines = append(shardLines, display.NewLineData(false, []string{"Unknown", "", ""}))
	return tableHeader, shardLines
}

func (txc *transactionCounter) displayExtendedShardHeaderHashesIncluded(
	lines []*display.LineData,
	extendedShardHeaderHashes [][]byte,
) []*display.LineData {

	if len(extendedShardHeaderHashes) == 0 {
		return lines
	}

	part := "ExtendedShardHeaderHashes"
	for i := 0; i < len(extendedShardHeaderHashes); i++ {
		if i == 0 || i >= len(extendedShardHeaderHashes)-1 {
			lines = append(lines, display.NewLineData(false, []string{
				part,
				fmt.Sprintf("ExtendedShardHeaderHash_%d", i+1),
				logger.DisplayByteSlice(extendedShardHeaderHashes[i])}))

			part = ""
			continue
		}

		if i == 1 {
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

func (txc *transactionCounter) displayMetaHashesIncluded(
	lines []*display.LineData,
	header data.ShardHeaderHandler,
) []*display.LineData {

	if header.GetMetaBlockHashes() == nil || len(header.GetMetaBlockHashes()) == 0 {
		return lines
	}

	part := "MetaBlockHashes"
	for i := 0; i < len(header.GetMetaBlockHashes()); i++ {
		if i == 0 || i >= len(header.GetMetaBlockHashes())-1 {
			lines = append(lines, display.NewLineData(false, []string{
				part,
				fmt.Sprintf("MetaBlockHash_%d", i+1),
				logger.DisplayByteSlice(header.GetMetaBlockHashes()[i])}))

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

func (txc *transactionCounter) displayTxBlockBody(
	lines []*display.LineData,
	header data.HeaderHandler,
	body *block.Body,
) []*display.LineData {
	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]

		processingTypeInMiniBlockHeaderStr := ""
		if len(miniBlockHeaders) > i {
			processingTypeInMiniBlockHeaderStr = getProcessingTypeAsString(miniBlockHeaders[i])
		}

		constructionStateInMiniBlockHeaderStr := ""
		if len(miniBlockHeaders) > i {
			constructionStateInMiniBlockHeaderStr = getConstructionStateAsString(miniBlockHeaders[i])
		}

		processingTypeInMiniBlockStr := ""
		if miniBlock.IsScheduledMiniBlock() {
			processingTypeInMiniBlockStr = "S_"
		}

		senderShardStr := getShardName(miniBlock.SenderShardID)
		receiverShardStr := getShardName(miniBlock.ReceiverShardID)

		part := fmt.Sprintf("%s%s%s_MiniBlock_%s%s->%s",
			processingTypeInMiniBlockHeaderStr,
			constructionStateInMiniBlockHeaderStr,
			miniBlock.Type.String(),
			processingTypeInMiniBlockStr,
			senderShardStr,
			receiverShardStr)

		if miniBlock.TxHashes == nil || len(miniBlock.TxHashes) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				part, "", "<EMPTY>"}))
		}

		if len(miniBlockHeaders) > i {
			lines = append(lines, display.NewLineData(false, []string{"", "MbHash", logger.DisplayByteSlice(miniBlockHeaders[i].GetHash())}))
			strProcessedRange := fmt.Sprintf("%d-%d", miniBlockHeaders[i].GetIndexOfFirstTxProcessed(), miniBlockHeaders[i].GetIndexOfLastTxProcessed())
			lines = append(lines, display.NewLineData(false, []string{"", "TxsProcessedRange", strProcessedRange}))
		}

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

// TODO: Could move this in mx-chain-core-go
func getShardName(shardID uint32) string {
	var shardStr string

	switch shardID {
	case core.MetachainShardId:
		shardStr = "MetaChain"
	case core.MainChainShardId:
		shardStr = "MainChain"
	case core.SovereignChainShardId:
		shardStr = "SovereignChain"
	default:
		shardStr = fmt.Sprintf("%d", shardID)
	}

	return shardStr
}

func getProcessingTypeAsString(miniBlockHeader data.MiniBlockHeaderHandler) string {
	processingType := block.ProcessingType(miniBlockHeader.GetProcessingType())
	switch processingType {
	case block.Scheduled:
		return "Scheduled_"
	case block.Processed:
		return "Processed_"
	}

	return ""
}

func getConstructionStateAsString(miniBlockHeader data.MiniBlockHeaderHandler) string {
	constructionState := block.MiniBlockState(miniBlockHeader.GetConstructionState())
	switch constructionState {
	case block.Proposed:
		return "Proposed_"
	case block.PartialExecuted:
		return "Partial_"
	}

	return ""
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
		"shard", getShardName(shardId),
		"epoch", lastNotarizedHdrForShard.GetEpoch(),
		"round", lastNotarizedHdrForShard.GetRound(),
		"nonce", lastNotarizedHdrForShard.GetNonce(),
		"hash", lastNotarizedHdrHashForShard)
}

// CurrentBlockTxs returns the current block's number of processed transactions
func (txc *transactionCounter) CurrentBlockTxs() uint64 {
	txc.mutex.RLock()
	defer txc.mutex.RUnlock()

	return txc.currentBlockTxs
}

// TotalTxs returns the total number of processed transactions
func (txc *transactionCounter) TotalTxs() uint64 {
	txc.mutex.RLock()
	defer txc.mutex.RUnlock()

	return txc.totalTxs
}

// IsInterfaceNil returns true if there is no value under the interface
func (txc *transactionCounter) IsInterfaceNil() bool {
	return txc == nil
}
