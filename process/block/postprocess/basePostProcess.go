package postprocess

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/prometheus/common/log"
)

type txShardInfo struct {
	senderShardID   uint32
	receiverShardID uint32
}

type txInfo struct {
	tx data.TransactionHandler
	*txShardInfo
}

type basePostProcessor struct {
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	store            dataRetriever.StorageService
	shardCoordinator sharding.Coordinator
	storageType      dataRetriever.UnitType

	mutInterResultsForBlock sync.Mutex
	interResultsForBlock    map[string]*txInfo
}

// SaveCurrentIntermediateTxToStorage saves all current intermediate results to the provided storage unit
func (bpp *basePostProcessor) SaveCurrentIntermediateTxToStorage() error {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	for _, txInfoValue := range bpp.interResultsForBlock {
		if txInfoValue.tx == nil {
			return process.ErrMissingTransaction
		}

		buff, err := bpp.marshalizer.Marshal(txInfoValue.tx)
		if err != nil {
			return err
		}

		errNotCritical := bpp.store.Put(bpp.storageType, bpp.hasher.Compute(string(buff)), buff)
		if errNotCritical != nil {
			log.Debug("UnsignedTransactionUnit.Put", "error", errNotCritical.Error())
		}
	}

	return nil
}

// CreateBlockStarted cleans the local cache map for processed/created intermediate transactions at this round
func (bpp *basePostProcessor) CreateBlockStarted() {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()
	bpp.interResultsForBlock = make(map[string]*txInfo, 0)
}

// CreateMarshalizedData creates the marshalized data for broadcasting purposes
func (bpp *basePostProcessor) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	bpp.mutInterResultsForBlock.Lock()
	defer bpp.mutInterResultsForBlock.Unlock()

	mrsTxs := make([][]byte, 0, len(txHashes))
	for _, txHash := range txHashes {
		txInfo := bpp.interResultsForBlock[string(txHash)]

		if txInfo == nil || txInfo.tx == nil {
			continue
		}

		txMrs, err := bpp.marshalizer.Marshal(txInfo.tx)
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
	for txHash, txInfo := range bpp.interResultsForBlock {
		if txInfo.receiverShardID != bpp.shardCoordinator.SelfId() {
			continue
		}
		if txInfo.senderShardID != bpp.shardCoordinator.SelfId() {
			continue
		}
		scrPool[txHash] = txInfo.tx
	}
	bpp.mutInterResultsForBlock.Unlock()

	return scrPool
}

func (bpp *basePostProcessor) verifyMiniBlock(createMBs map[uint32]*block.MiniBlock, mb *block.MiniBlock) error {
	createdScrMb, ok := createMBs[mb.ReceiverShardID]
	if createdScrMb == nil || !ok {
		return process.ErrNilMiniBlocks
	}

	createdHash, err := core.CalculateHash(bpp.marshalizer, bpp.hasher, createdScrMb)
	if err != nil {
		return err
	}

	receivedHash, err := core.CalculateHash(bpp.marshalizer, bpp.hasher, mb)
	if err != nil {
		return err
	}

	if !bytes.Equal(createdHash, receivedHash) {
		return process.ErrMiniBlockHashMismatch
	}

	return nil
}
