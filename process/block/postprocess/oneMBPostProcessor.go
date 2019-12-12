package postprocess

import (
	"bytes"
	"sort"

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

type oneMBPostProcessor struct {
	blockType block.Type
	*basePostProcessor
}

// NewOneMiniBlockPostProcessor creates a new intermediate results processor
func NewOneMiniBlockPostProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	coordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	blockType block.Type,
	storageType dataRetriever.UnitType,
) (*oneMBPostProcessor, error) {
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStorage
	}

	base := &basePostProcessor{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: coordinator,
		store:            store,
		storageType:      storageType,
	}

	irp := &oneMBPostProcessor{
		basePostProcessor: base,
		blockType:         blockType,
	}

	irp.interResultsForBlock = make(map[string]*txInfo, 0)

	return irp, nil
}

// CreateAllInterMiniBlocks returns the miniblock for the current round created from the receipts/bad transactions
func (irp *oneMBPostProcessor) CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock {
	selfId := irp.shardCoordinator.SelfId()

	miniBlocks := make(map[uint32]*block.MiniBlock, 0)
	irp.mutInterResultsForBlock.Lock()
	defer irp.mutInterResultsForBlock.Unlock()

	if len(irp.interResultsForBlock) == 0 {
		return miniBlocks
	}

	miniBlocks[selfId] = &block.MiniBlock{
		Type:            irp.blockType,
		ReceiverShardID: selfId,
		SenderShardID:   selfId,
	}

	for key := range irp.interResultsForBlock {
		miniBlocks[selfId].TxHashes = append(miniBlocks[selfId].TxHashes, []byte(key))
	}

	sort.Slice(miniBlocks[selfId].TxHashes, func(a, b int) bool {
		return bytes.Compare(miniBlocks[selfId].TxHashes[a], miniBlocks[selfId].TxHashes[b]) < 0
	})

	return miniBlocks
}

// VerifyInterMiniBlocks verifies if the receipts/bad transactions added to the block are valid
func (irp *oneMBPostProcessor) VerifyInterMiniBlocks(body block.Body) error {
	scrMbs := irp.CreateAllInterMiniBlocks()

	verifiedOne := false
	for i := 0; i < len(body); i++ {
		mb := body[i]
		if mb.Type != irp.blockType {
			continue
		}

		if verifiedOne {
			return process.ErrTooManyReceiptsMiniBlocks
		}

		err := irp.verifyMiniBlock(scrMbs, mb)
		if err != nil {
			return err
		}

		verifiedOne = true
	}

	return nil
}

// AddIntermediateTransactions adds receipts/bad transactions resulting from transaction processor
func (irp *oneMBPostProcessor) AddIntermediateTransactions(txs []data.TransactionHandler) error {
	irp.mutInterResultsForBlock.Lock()
	defer irp.mutInterResultsForBlock.Unlock()

	selfId := irp.shardCoordinator.SelfId()

	for i := 0; i < len(txs); i++ {
		receiptHash, err := core.CalculateHash(irp.marshalizer, irp.hasher, txs[i])
		if err != nil {
			return err
		}

		addReceiptShardInfo := &txShardInfo{receiverShardID: selfId, senderShardID: selfId}
		scrInfo := &txInfo{tx: txs[i], txShardInfo: addReceiptShardInfo}
		irp.interResultsForBlock[string(receiptHash)] = scrInfo
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (irp *oneMBPostProcessor) IsInterfaceNil() bool {
	return irp == nil
}
