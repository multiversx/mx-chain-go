package postprocess

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.DataMarshalizer = (*basePostProcessor)(nil)

type txShardInfo struct {
	senderShardID   uint32
	receiverShardID uint32
}

type txInfo struct {
	tx data.TransactionHandler
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
	mapTxToResult           map[string][]string
	intraShardMiniBlock     *block.MiniBlock
	economicsFee            process.FeeHandler
}

// SaveCurrentIntermediateTxToStorage saves all current intermediate results to the provided storage unit
func (bpp *basePostProcessor) SaveCurrentIntermediateTxToStorage() error {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	for _, txInfoValue := range bpp.interResultsForBlock {
		if check.IfNil(txInfoValue.tx) {
			return process.ErrMissingTransaction
		}

		buff, err := bpp.marshalizer.Marshal(txInfoValue.tx)
		if err != nil {
			return err
		}

		errNotCritical := bpp.store.Put(bpp.storageType, bpp.hasher.Compute(string(buff)), buff)
		if errNotCritical != nil {
			log.Debug("SaveCurrentIntermediateTxToStorage put", "type", bpp.storageType, "error", errNotCritical.Error())
		}
	}

	return nil
}

// CreateBlockStarted cleans the local cache map for processed/created intermediate transactions at this round
func (bpp *basePostProcessor) CreateBlockStarted() {
	bpp.mutInterResultsForBlock.Lock()
	bpp.interResultsForBlock = make(map[string]*txInfo)
	bpp.intraShardMiniBlock = nil
	bpp.mapTxToResult = make(map[string][]string)
	bpp.mutInterResultsForBlock.Unlock()
}

// CreateMarshalizedData creates the marshalized data for broadcasting purposes
func (bpp *basePostProcessor) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	mrsTxs := make([][]byte, 0, len(txHashes))
	for _, txHash := range txHashes {
		txInfoObject := bpp.interResultsForBlock[string(txHash)]
		if txInfoObject == nil || check.IfNil(txInfoObject.tx) {
			log.Warn("basePostProcessor.CreateMarshalizedData: tx not found", "hash", txHash)
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

	return bpp.intraShardMiniBlock.Clone()
}

// RemoveProcessedResultsFor will remove the created results for the transactions which were reverted
func (bpp *basePostProcessor) RemoveProcessedResultsFor(txHashes [][]byte) {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	if len(bpp.mapTxToResult) == 0 {
		return
	}

	for _, txHash := range txHashes {
		resultHashes, ok := bpp.mapTxToResult[string(txHash)]
		if !ok {
			continue
		}

		for _, resultHash := range resultHashes {
			delete(bpp.interResultsForBlock, resultHash)
		}
		delete(bpp.mapTxToResult, string(txHash))
	}
}

func (bpp *basePostProcessor) splitMiniBlocksIfNeeded(miniBlocks []*block.MiniBlock) []*block.MiniBlock {
	splitMiniBlocks := make([]*block.MiniBlock, 0)
	for _, miniBlock := range miniBlocks {
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
			continue
		}

		isGasLimitExceeded := gasLimitInReceiverShard+interResult.tx.GetGasLimit() >
			bpp.economicsFee.MaxGasLimitPerBlock(currentMiniBlock.ReceiverShardID)
		if isGasLimitExceeded {
			log.Debug("basePostProcessor.splitMiniBlockIfNeeded: gas limit exceeded",
				"mb type", currentMiniBlock.Type,
				"sender shard", currentMiniBlock.SenderShardID,
				"receiver shard", currentMiniBlock.ReceiverShardID,
				"initial num txs", len(miniBlock.TxHashes),
				"adjusted num txs", len(currentMiniBlock.TxHashes),
			)
			splitMiniBlocks = append(splitMiniBlocks, currentMiniBlock)
			currentMiniBlock = createEmptyMiniBlock(miniBlock)
			gasLimitInReceiverShard = 0
		}

		gasLimitInReceiverShard += interResult.tx.GetGasLimit()
		currentMiniBlock.TxHashes = append(currentMiniBlock.TxHashes, txHash)
	}

	splitMiniBlocks = append(splitMiniBlocks, currentMiniBlock)
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
