package postprocess

import (
	"bytes"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type intermediateResultsProcessor struct {
	adrConv   state.AddressConverter
	store     dataRetriever.StorageService
	blockType block.Type
	currTxs   dataRetriever.TransactionCacher

	*basePostProcessor
}

// NewIntermediateResultsProcessor creates a new intermediate results processor
func NewIntermediateResultsProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	coordinator sharding.Coordinator,
	adrConv state.AddressConverter,
	store dataRetriever.StorageService,
	blockType block.Type,
	currTxs dataRetriever.TransactionCacher,
) (*intermediateResultsProcessor, error) {
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(adrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(currTxs) {
		return nil, process.ErrNilTxForCurrentBlockHandler
	}

	base := &basePostProcessor{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: coordinator,
		store:            store,
		storageType:      dataRetriever.UnsignedTransactionUnit,
	}

	irp := &intermediateResultsProcessor{
		basePostProcessor: base,
		adrConv:           adrConv,
		blockType:         blockType,
		store:             store,
		currTxs:           currTxs,
	}

	irp.interResultsForBlock = make(map[string]*txInfo, 0)

	return irp, nil
}

// CreateAllInterMiniBlocks returns the cross shard miniblocks for the current round created from the smart contract results
func (irp *intermediateResultsProcessor) CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock {
	miniBlocks := make(map[uint32]*block.MiniBlock, int(irp.shardCoordinator.NumberOfShards())+1)
	for i := uint32(0); i < irp.shardCoordinator.NumberOfShards(); i++ {
		miniBlocks[i] = &block.MiniBlock{}
	}
	miniBlocks[sharding.MetachainShardId] = &block.MiniBlock{}

	irp.currTxs.Clean()
	irp.mutInterResultsForBlock.Lock()

	for key, value := range irp.interResultsForBlock {
		recvShId := value.receiverShardID
		miniBlocks[recvShId].TxHashes = append(miniBlocks[recvShId].TxHashes, []byte(key))
		irp.currTxs.AddTx([]byte(key), value.tx)
	}

	finalMBs := make(map[uint32]*block.MiniBlock, 0)
	for shId, miniblock := range miniBlocks {
		if len(miniblock.TxHashes) > 0 {
			miniblock.SenderShardID = irp.shardCoordinator.SelfId()
			miniblock.ReceiverShardID = shId
			miniblock.Type = irp.blockType

			sort.Slice(miniblock.TxHashes, func(a, b int) bool {
				return bytes.Compare(miniblock.TxHashes[a], miniblock.TxHashes[b]) < 0
			})

			finalMBs[shId] = miniblock
		}
	}

	irp.mutInterResultsForBlock.Unlock()

	return finalMBs
}

// VerifyInterMiniBlocks verifies if the smart contract results added to the block are valid
func (irp *intermediateResultsProcessor) VerifyInterMiniBlocks(body block.Body) error {
	scrMbs := irp.CreateAllInterMiniBlocks()

	for i := 0; i < len(body); i++ {
		mb := body[i]
		if mb.Type != irp.blockType {
			continue
		}
		if mb.ReceiverShardID == irp.shardCoordinator.SelfId() &&
			mb.SenderShardID != irp.shardCoordinator.SelfId() {
			continue
		}

		err := irp.verifyMiniBlock(scrMbs, mb)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddIntermediateTransactions adds smart contract results from smart contract processing for cross-shard calls
func (irp *intermediateResultsProcessor) AddIntermediateTransactions(txs []data.TransactionHandler) error {
	irp.mutInterResultsForBlock.Lock()
	defer irp.mutInterResultsForBlock.Unlock()

	for i := 0; i < len(txs); i++ {
		addScr, ok := txs[i].(*smartContractResult.SmartContractResult)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		scrHash, err := core.CalculateHash(irp.marshalizer, irp.hasher, txs[i])
		if err != nil {
			return err
		}

		sndShId, dstShId, err := irp.getShardIdsFromAddresses(addScr.SndAddr, addScr.RcvAddr)
		if err != nil {
			return err
		}

		addScrShardInfo := &txShardInfo{receiverShardID: dstShId, senderShardID: sndShId}
		scrInfo := &txInfo{tx: addScr, txShardInfo: addScrShardInfo}
		irp.interResultsForBlock[string(scrHash)] = scrInfo
	}

	return nil
}

func (irp *intermediateResultsProcessor) getShardIdsFromAddresses(sndAddr []byte, rcvAddr []byte) (uint32, uint32, error) {
	adrSrc, err := irp.adrConv.CreateAddressFromPublicKeyBytes(sndAddr)
	if err != nil {
		return irp.shardCoordinator.NumberOfShards(), irp.shardCoordinator.NumberOfShards(), err
	}
	adrDst, err := irp.adrConv.CreateAddressFromPublicKeyBytes(rcvAddr)
	if err != nil {
		return irp.shardCoordinator.NumberOfShards(), irp.shardCoordinator.NumberOfShards(), err
	}

	shardForSrc := irp.shardCoordinator.ComputeId(adrSrc)
	shardForDst := irp.shardCoordinator.ComputeId(adrDst)

	isEmptyAddress := bytes.Equal(sndAddr, make([]byte, irp.adrConv.AddressLen()))
	if isEmptyAddress {
		shardForSrc = irp.shardCoordinator.SelfId()
	}

	isEmptyAddress = bytes.Equal(rcvAddr, make([]byte, irp.adrConv.AddressLen()))
	if isEmptyAddress {
		shardForDst = irp.shardCoordinator.SelfId()
	}

	return shardForSrc, shardForDst, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (irp *intermediateResultsProcessor) IsInterfaceNil() bool {
	if irp == nil {
		return true
	}
	return false
}
