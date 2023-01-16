package txstatus

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("storage/txstatus")

// statusComputer computes a transaction status
type statusComputer struct {
	selfShardID              uint32
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	store                    dataRetriever.StorageService
}

// NewStatusComputer will create a new instance of statusComputer
func NewStatusComputer(
	selfShardID uint32,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
	store dataRetriever.StorageService,
) (*statusComputer, error) {
	if check.IfNil(uint64ByteSliceConverter) {
		return nil, ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(store) {
		return nil, ErrNiStorageService
	}

	return &statusComputer{
		selfShardID:              selfShardID,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		store:                    store,
	}, nil
}

// ComputeStatusWhenInStorageKnowingMiniblock computes the transaction status for a historical transaction
func (sc *statusComputer) ComputeStatusWhenInStorageKnowingMiniblock(
	miniblockType block.Type,
	tx *transaction.ApiTransactionResult,
) (transaction.TxStatus, error) {

	if tx == nil {
		return "", ErrNilApiTransactionResult
	}

	if sc.isMiniblockInvalid(miniblockType) {
		return transaction.TxStatusInvalid, nil
	}

	receiver := tx.Tx.GetRcvAddr()
	isMiniblockFinalized := tx.NotarizedAtDestinationInMetaNonce > 0
	isSuccess := isMiniblockFinalized || sc.isDestinationMe(tx.DestinationShard) || sc.isContractDeploy(receiver, tx.Data)
	if isSuccess {
		return transaction.TxStatusSuccess, nil
	}

	return transaction.TxStatusPending, nil
}

// ComputeStatusWhenInStorageNotKnowingMiniblock computes the transaction status when transaction is in current epoch's storage
// Limitation: in this case, since we do not know the miniblock type, we cannot know if a transaction is actually, "invalid".
// However, when "dblookupext" indexing is enabled, this function is not used.
func (sc *statusComputer) ComputeStatusWhenInStorageNotKnowingMiniblock(
	destinationShard uint32,
	tx *transaction.ApiTransactionResult,
) (transaction.TxStatus, error) {
	if tx == nil {
		return "", ErrNilApiTransactionResult
	}

	receiver := tx.Tx.GetRcvAddr()
	isSuccess := sc.isDestinationMe(destinationShard) || sc.isContractDeploy(receiver, tx.Data)
	if isSuccess {
		return transaction.TxStatusSuccess, nil
	}

	// At least partially executed (since in source's storage)
	return transaction.TxStatusPending, nil
}

func (sc *statusComputer) isMiniblockInvalid(miniblockType block.Type) bool {
	return miniblockType == block.InvalidBlock
}

func (sc *statusComputer) isDestinationMe(destinationShard uint32) bool {
	return sc.selfShardID == destinationShard
}

func (sc *statusComputer) isContractDeploy(receiver []byte, transactionData []byte) bool {
	return core.IsEmptyAddress(receiver) && len(transactionData) > 0
}

// SetStatusIfIsRewardReverted will compute and set status for a reverted reward transaction
func (sc *statusComputer) SetStatusIfIsRewardReverted(
	tx *transaction.ApiTransactionResult,
	miniblockType block.Type,
	headerNonce uint64,
	headerHash []byte,
) (bool, error) {

	if tx == nil {
		return false, ErrNilApiTransactionResult
	}

	if miniblockType != block.RewardsBlock {
		return false, nil
	}

	var storerUnit dataRetriever.UnitType

	selfShardID := sc.selfShardID
	if selfShardID == core.MetachainShardId {
		storerUnit = dataRetriever.MetaHdrNonceHashDataUnit
	} else {
		storerUnit = dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(selfShardID)
	}

	nonceToByteSlice := sc.uint64ByteSliceConverter.ToByteSlice(headerNonce)
	headerHashFromStorage, err := sc.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		log.Warn("cannot get header hash by nonce", "error", err.Error())
		return false, nil
	}

	if bytes.Equal(headerHashFromStorage, headerHash) {
		return false, nil
	}

	tx.Status = transaction.TxStatusRewardReverted
	return true, nil
}
