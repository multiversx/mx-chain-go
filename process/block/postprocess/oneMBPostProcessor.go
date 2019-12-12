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

	opp := &oneMBPostProcessor{
		basePostProcessor: base,
		blockType:         blockType,
	}

	opp.interResultsForBlock = make(map[string]*txInfo)

	return opp, nil
}

// CreateAllInterMiniBlocks returns the miniblock for the current round created from the receipts/bad transactions
func (opp *oneMBPostProcessor) CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock {
	selfId := opp.shardCoordinator.SelfId()

	miniBlocks := make(map[uint32]*block.MiniBlock)
	opp.mutInterResultsForBlock.Lock()
	defer opp.mutInterResultsForBlock.Unlock()

	if len(opp.interResultsForBlock) == 0 {
		return miniBlocks
	}

	miniBlocks[selfId] = &block.MiniBlock{
		Type:            opp.blockType,
		ReceiverShardID: selfId,
		SenderShardID:   selfId,
	}

	for key := range opp.interResultsForBlock {
		miniBlocks[selfId].TxHashes = append(miniBlocks[selfId].TxHashes, []byte(key))
	}

	sort.Slice(miniBlocks[selfId].TxHashes, func(a, b int) bool {
		return bytes.Compare(miniBlocks[selfId].TxHashes[a], miniBlocks[selfId].TxHashes[b]) < 0
	})

	return miniBlocks
}

// VerifyInterMiniBlocks verifies if the receipts/bad transactions added to the block are valid
func (opp *oneMBPostProcessor) VerifyInterMiniBlocks(body block.Body) error {
	scrMbs := opp.CreateAllInterMiniBlocks()

	verifiedOne := false
	for i := 0; i < len(body); i++ {
		mb := body[i]
		if mb.Type != opp.blockType {
			continue
		}

		if verifiedOne {
			return process.ErrTooManyReceiptsMiniBlocks
		}

		err := opp.verifyMiniBlock(scrMbs, mb)
		if err != nil {
			return err
		}

		verifiedOne = true
	}

	return nil
}

// AddIntermediateTransactions adds receipts/bad transactions resulting from transaction processor
func (opp *oneMBPostProcessor) AddIntermediateTransactions(txs []data.TransactionHandler) error {
	opp.mutInterResultsForBlock.Lock()
	defer opp.mutInterResultsForBlock.Unlock()

	selfId := opp.shardCoordinator.SelfId()

	for i := 0; i < len(txs); i++ {
		receiptHash, err := core.CalculateHash(opp.marshalizer, opp.hasher, txs[i])
		if err != nil {
			return err
		}

		addReceiptShardInfo := &txShardInfo{receiverShardID: selfId, senderShardID: selfId}
		scrInfo := &txInfo{tx: txs[i], txShardInfo: addReceiptShardInfo}
		opp.interResultsForBlock[string(receiptHash)] = scrInfo
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (opp *oneMBPostProcessor) IsInterfaceNil() bool {
	return opp == nil
}
