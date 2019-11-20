package preprocess

import (
	"bytes"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
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
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	adrConv          state.AddressConverter
	store            dataRetriever.StorageService
	blockType        block.Type
	currTxs          dataRetriever.TransactionCacher

	mutInterResultsForBlock sync.Mutex
	interResultsForBlock    map[string]*txInfo
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
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if coordinator == nil || coordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if adrConv == nil || adrConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if store == nil || store.IsInterfaceNil() {
		return nil, process.ErrNilStorage
	}
	if currTxs == nil || currTxs.IsInterfaceNil() {
		return nil, process.ErrNilTxForCurrentBlockHandler
	}

	irp := &intermediateResultsProcessor{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: coordinator,
		adrConv:          adrConv,
		blockType:        blockType,
		store:            store,
		currTxs:          currTxs,
	}

	irp.interResultsForBlock = make(map[string]*txInfo, 0)

	return irp, nil
}

// CreateAllInterMiniBlocks returns the cross shard miniblocks for the current round created from the smart contract results
func (irp *intermediateResultsProcessor) CreateAllInterMiniBlocks() map[uint32]*block.MiniBlock {
	miniBlocks := make(map[uint32]*block.MiniBlock)
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
		if mb.ReceiverShardID == irp.shardCoordinator.SelfId() {
			continue
		}

		createdScrMb, ok := scrMbs[mb.ReceiverShardID]
		if createdScrMb == nil || !ok {
			return process.ErrNilMiniBlocks
		}

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

// SaveCurrentIntermediateTxToStorage saves all current intermediate results to the provided storage unit
func (irp *intermediateResultsProcessor) SaveCurrentIntermediateTxToStorage() error {
	irp.mutInterResultsForBlock.Lock()
	defer irp.mutInterResultsForBlock.Unlock()

	for _, txInfoValue := range irp.interResultsForBlock {
		if txInfoValue.tx == nil {
			return process.ErrMissingTransaction
		}

		buff, err := irp.marshalizer.Marshal(txInfoValue.tx)
		if err != nil {
			return err
		}

		errNotCritical := irp.store.Put(dataRetriever.UnsignedTransactionUnit, irp.hasher.Compute(string(buff)), buff)
		if errNotCritical != nil {
			log.Error(errNotCritical.Error())
		}
	}

	return nil
}

// CreateBlockStarted cleans the local cache map for processed/created intermediate transactions at this round
func (irp *intermediateResultsProcessor) CreateBlockStarted() {
	irp.mutInterResultsForBlock.Lock()
	defer irp.mutInterResultsForBlock.Unlock()
	irp.interResultsForBlock = make(map[string]*txInfo, 0)
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

// CreateMarshalizedData creates the marshalized data for broadcasting purposes
func (irp *intermediateResultsProcessor) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	irp.mutInterResultsForBlock.Lock()
	defer irp.mutInterResultsForBlock.Unlock()

	mrsTxs := make([][]byte, 0)
	for _, txHash := range txHashes {
		txInfo := irp.interResultsForBlock[string(txHash)]

		if txInfo == nil || txInfo.tx == nil {
			continue
		}

		txMrs, err := irp.marshalizer.Marshal(txInfo.tx)
		if err != nil {
			return nil, process.ErrMarshalWithoutSuccess
		}
		mrsTxs = append(mrsTxs, txMrs)
	}

	return mrsTxs, nil
}

// GetAllCurrentFinishedTxs returns the cached finalized transactions for current round
func (irp *intermediateResultsProcessor) GetAllCurrentFinishedTxs() map[string]data.TransactionHandler {
	irp.mutInterResultsForBlock.Lock()

	scrPool := make(map[string]data.TransactionHandler)
	for txHash, txInfo := range irp.interResultsForBlock {
		if txInfo.receiverShardID != irp.shardCoordinator.SelfId() {
			continue
		}
		if txInfo.senderShardID != irp.shardCoordinator.SelfId() {
			continue
		}
		scrPool[txHash] = txInfo.tx
	}
	irp.mutInterResultsForBlock.Unlock()

	return scrPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (irp *intermediateResultsProcessor) IsInterfaceNil() bool {
	if irp == nil {
		return true
	}
	return false
}
