package postprocess

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-logger-go"
)

var _ process.DataMarshalizer = (*basePostProcessor)(nil)

type txShardInfo struct {
	senderShardID   uint32
	receiverShardID uint32
}

type txInfo struct {
	tx    data.TransactionHandler
	index uint32
	*txShardInfo
}

var log = logger.GetOrCreate("process/block/postprocess")

type basePostProcessor struct {
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	store            dataRetriever.StorageService
	shardCoordinator sharding.Coordinator
	storageType      dataRetriever.UnitType

	mutInterResultsForBlock sync.Mutex
	interResultsForBlock    map[string]*txInfo
	mapProcessedResult      map[string][][]byte
	intraShardMiniBlock     *block.MiniBlock
	economicsFee            process.FeeHandler
	index                   uint32
}

// SaveCurrentIntermediateTxToStorage saves all current intermediate results to the provided storage unit
func (bpp *basePostProcessor) SaveCurrentIntermediateTxToStorage() {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	for _, txInfoValue := range bpp.interResultsForBlock {
		bpp.saveIntermediateTxToStorage(txInfoValue.tx)
	}
}

func (bpp *basePostProcessor) saveIntermediateTxToStorage(tx data.TransactionHandler) {
	if check.IfNil(tx) {
		log.Warn("basePostProcessor.saveIntermediateTxToStorage", "error", process.ErrMissingTransaction.Error())
		return
	}

	buff, err := bpp.marshalizer.Marshal(tx)
	if err != nil {
		log.Warn("basePostProcessor.saveIntermediateTxToStorage", "error", err.Error())
		return
	}

	errNotCritical := bpp.store.Put(bpp.storageType, bpp.hasher.Compute(string(buff)), buff)
	if errNotCritical != nil {
		log.Debug("SaveCurrentIntermediateTxToStorage put", "type", bpp.storageType, "error", errNotCritical.Error())
	}
}

// CreateBlockStarted cleans the local cache map for processed/created intermediate transactions at this round
func (bpp *basePostProcessor) CreateBlockStarted() {
	bpp.mutInterResultsForBlock.Lock()
	bpp.interResultsForBlock = make(map[string]*txInfo)
	bpp.intraShardMiniBlock = nil
	bpp.mapProcessedResult = make(map[string][][]byte)
	bpp.index = 0
	bpp.mutInterResultsForBlock.Unlock()
}

// CreateMarshalledData creates the marshalled data for broadcasting purposes
func (bpp *basePostProcessor) CreateMarshalledData(txHashes [][]byte) ([][]byte, error) {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	mrsTxs := make([][]byte, 0, len(txHashes))
	for _, txHash := range txHashes {
		txInfoObject := bpp.interResultsForBlock[string(txHash)]
		if txInfoObject == nil || check.IfNil(txInfoObject.tx) {
			log.Warn("basePostProcessor.CreateMarshalledData: tx not found", "hash", txHash)
			continue
		}

		txMrs, err := bpp.marshalizer.Marshal(txInfoObject.tx)
		if err != nil {
			return nil, process.ErrMarshalWithoutSuccess
		}
		mrsTxs = append(mrsTxs, txMrs)
	}

	return mrsTxs, nil
}

// GetAllCurrentFinishedTxs returns the cached finalized transactions for current round
func (bpp *basePostProcessor) GetAllCurrentFinishedTxs() map[string]data.TransactionHandler {
	bpp.mutInterResultsForBlock.Lock()

	scrPool := make(map[string]data.TransactionHandler)
	for txHash, txInfoInterResult := range bpp.interResultsForBlock {
		scrPool[txHash] = txInfoInterResult.tx
	}
	bpp.mutInterResultsForBlock.Unlock()

	return scrPool
}

func (bpp *basePostProcessor) verifyMiniBlock(createMBs map[uint32][]*block.MiniBlock, mb *block.MiniBlock) error {
	createdScrMbs, ok := createMBs[mb.ReceiverShardID]
	if !ok {
		log.Debug("missing mini block", "type", mb.Type, "sender", mb.SenderShardID, "receiver", mb.ReceiverShardID, "numTxs", len(mb.TxHashes))
		return process.ErrNilMiniBlocks
	}

	mapCreatedHashes := make(map[string]struct{})
	for _, createdScrMb := range createdScrMbs {
		createdHash, err := core.CalculateHash(bpp.marshalizer, bpp.hasher, createdScrMb)
		if err != nil {
			return err
		}

		mapCreatedHashes[string(createdHash)] = struct{}{}
	}

	receivedHash, err := core.CalculateHash(bpp.marshalizer, bpp.hasher, mb)
	if err != nil {
		return err
	}

	_, existsHash := mapCreatedHashes[string(receivedHash)]
	if !existsHash {
		log.Debug("received mini block", "type", mb.Type, "sender", mb.SenderShardID, "receiver", mb.ReceiverShardID, "numTxs", len(mb.TxHashes))
		for _, createdScrMb := range createdScrMbs {
			log.Debug("created mini block", "type", createdScrMb.Type, "sender", createdScrMb.SenderShardID, "receiver", createdScrMb.ReceiverShardID, "numTxs", len(createdScrMb.TxHashes))
		}
		return process.ErrMiniBlockHashMismatch
	}

	return nil
}

// GetCreatedInShardMiniBlock returns a clone of the intra shard mini block
func (bpp *basePostProcessor) GetCreatedInShardMiniBlock() *block.MiniBlock {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	if bpp.intraShardMiniBlock == nil {
		return nil
	}

	return bpp.intraShardMiniBlock.DeepClone()
}

// RemoveProcessedResults will remove the processed results since the last init
func (bpp *basePostProcessor) RemoveProcessedResults(key []byte) [][]byte {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	txHashes, ok := bpp.mapProcessedResult[string(key)]
	if !ok {
		return nil
	}

	for _, txHash := range txHashes {
		delete(bpp.interResultsForBlock, string(txHash))
	}

	return txHashes
}

// InitProcessedResults will initialize the processed results
func (bpp *basePostProcessor) InitProcessedResults(key []byte) {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	bpp.mapProcessedResult[string(key)] = make([][]byte, 0)
}

func (bpp *basePostProcessor) splitMiniBlocksIfNeeded(miniBlocks []*block.MiniBlock) []*block.MiniBlock {
	splitMiniBlocks := make([]*block.MiniBlock, 0)
	for _, miniBlock := range miniBlocks {
		if miniBlock.ReceiverShardID == bpp.shardCoordinator.SelfId() {
			splitMiniBlocks = append(splitMiniBlocks, miniBlock)
			continue
		}

		splitMiniBlocks = append(splitMiniBlocks, bpp.splitMiniBlockIfNeeded(miniBlock)...)
	}
	return splitMiniBlocks
}

func (bpp *basePostProcessor) splitMiniBlockIfNeeded(miniBlock *block.MiniBlock) []*block.MiniBlock {
	splitMiniBlocks := make([]*block.MiniBlock, 0)
	currentMiniBlock := createEmptyMiniBlock(miniBlock)
	gasLimitInReceiverShard := uint64(0)

	for _, txHash := range miniBlock.TxHashes {
		interResult, ok := bpp.interResultsForBlock[string(txHash)]
		if !ok {
			log.Warn("basePostProcessor.splitMiniBlockIfNeeded: missing tx", "hash", txHash)
			currentMiniBlock.TxHashes = append(currentMiniBlock.TxHashes, txHash)
			continue
		}

		isGasLimitExceeded := gasLimitInReceiverShard+interResult.tx.GetGasLimit() >
			bpp.economicsFee.MaxGasLimitPerMiniBlockForSafeCrossShard()
		if isGasLimitExceeded {
			log.Debug("basePostProcessor.splitMiniBlockIfNeeded: gas limit exceeded",
				"mb type", currentMiniBlock.Type,
				"sender shard", currentMiniBlock.SenderShardID,
				"receiver shard", currentMiniBlock.ReceiverShardID,
				"initial num txs", len(miniBlock.TxHashes),
				"adjusted num txs", len(currentMiniBlock.TxHashes),
			)

			if len(currentMiniBlock.TxHashes) > 0 {
				splitMiniBlocks = append(splitMiniBlocks, currentMiniBlock)
			}

			currentMiniBlock = createEmptyMiniBlock(miniBlock)
			gasLimitInReceiverShard = 0
		}

		gasLimitInReceiverShard += interResult.tx.GetGasLimit()
		currentMiniBlock.TxHashes = append(currentMiniBlock.TxHashes, txHash)
	}

	if len(currentMiniBlock.TxHashes) > 0 {
		splitMiniBlocks = append(splitMiniBlocks, currentMiniBlock)
	}

	return splitMiniBlocks
}

func createEmptyMiniBlock(miniBlock *block.MiniBlock) *block.MiniBlock {
	return &block.MiniBlock{
		SenderShardID:   miniBlock.SenderShardID,
		ReceiverShardID: miniBlock.ReceiverShardID,
		Type:            miniBlock.Type,
		Reserved:        miniBlock.Reserved,
		TxHashes:        make([][]byte, 0),
	}
}

func createMiniBlocksMap(scrMbs []*block.MiniBlock) map[uint32][]*block.MiniBlock {
	createdMapMbs := make(map[uint32][]*block.MiniBlock)
	for _, mb := range scrMbs {
		_, ok := createdMapMbs[mb.ReceiverShardID]
		if !ok {
			createdMapMbs[mb.ReceiverShardID] = make([]*block.MiniBlock, 0)
		}

		createdMapMbs[mb.ReceiverShardID] = append(createdMapMbs[mb.ReceiverShardID], mb)
	}

	return createdMapMbs
}

func (bpp *basePostProcessor) addIntermediateTxToResultsForBlock(
	txHandler data.TransactionHandler,
	txHash []byte,
	sndShardID uint32,
	rcvShardID uint32,
) {
	addScrShardInfo := &txShardInfo{receiverShardID: rcvShardID, senderShardID: sndShardID}
	scrInfo := &txInfo{tx: txHandler, txShardInfo: addScrShardInfo, index: bpp.index}
	bpp.index++
	bpp.interResultsForBlock[string(txHash)] = scrInfo

	for key := range bpp.mapProcessedResult {
		bpp.mapProcessedResult[key] = append(bpp.mapProcessedResult[key], txHash)
	}
}
