package preprocess

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"sync"
)

type intermediateResultsProcessor struct {
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	adrConv          state.AddressConverter
	blockType        block.Type

	mutInterResultsForBlock sync.Mutex
	interResultsForBlock    map[string]*txInfo
}

// NewIntermediateResultsProcessor creates a new intermediate results processor
func NewIntermediateResultsProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	coordinator sharding.Coordinator,
	adrConv state.AddressConverter,
	blockType block.Type,
) (*intermediateResultsProcessor, error) {
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if coordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if adrConv == nil {
		return nil, process.ErrNilAddressConverter
	}

	irp := &intermediateResultsProcessor{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: coordinator,
		adrConv:          adrConv,
		blockType:        blockType,
	}

	irp.interResultsForBlock = make(map[string]*txInfo, 0)

	return irp, nil
}

// CreateAllInterMiniBlocks returns the cross shard miniblocks for the current round created from the smart contract results
func (irp *intermediateResultsProcessor) CreateAllInterMiniBlocks() []*block.MiniBlock {
	miniBlocks := make([]*block.MiniBlock, irp.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < irp.shardCoordinator.NumberOfShards(); i++ {
		miniBlocks[i] = &block.MiniBlock{}
	}

	irp.mutInterResultsForBlock.Lock()

	for key, value := range irp.interResultsForBlock {
		if value.receiverShardID != irp.shardCoordinator.SelfId() {
			miniBlocks[value.receiverShardID].TxHashes = append(miniBlocks[value.receiverShardID].TxHashes, []byte(key))
		}
	}

	for i := 0; i < len(miniBlocks); i++ {
		if len(miniBlocks[i].TxHashes) > 0 {
			miniBlocks[i].SenderShardID = irp.shardCoordinator.SelfId()
			miniBlocks[i].ReceiverShardID = uint32(i)
			miniBlocks[i].Type = irp.blockType
		}
	}

	irp.interResultsForBlock = make(map[string]*txInfo, 0)

	irp.mutInterResultsForBlock.Unlock()

	return miniBlocks
}

// VerifyInterMiniBlocks verifies if the smart contract results added to the block are valid
func (irp *intermediateResultsProcessor) VerifyInterMiniBlocks(body block.Body) error {
	scrMbs := irp.CreateAllInterMiniBlocks()

	for i := 0; i < len(body); i++ {
		mb := body[i]
		if mb.Type != irp.blockType {
			continue
		}
		if mb.ReceiverShardID == irp.shardCoordinator.SelfId() {
			continue
		}

		createdScrMb := scrMbs[mb.ReceiverShardID]

		createdHash, err := core.CalculateHash(irp.marshalizer, irp.hasher, createdScrMb)
		if err != nil {
			return err
		}

		receivedHash, err := core.CalculateHash(irp.marshalizer, irp.hasher, mb)
		if err != nil {
			return err
		}

		if !bytes.Equal(createdHash, receivedHash) {
			return process.ErrMiniBlockHashMismatch
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

		scrHash, err := core.CalculateHash(irp.marshalizer, irp.hasher, addScr)
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

	return shardForSrc, shardForDst, nil
}
