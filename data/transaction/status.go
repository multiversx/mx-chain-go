package transaction

import (
	"bytes"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// TxStatus is the status of a transaction
type TxStatus string

var log = logger.GetOrCreate("data/transaction")

const (
	// TxStatusPending = received and maybe executed on source shard, but not on destination shard
	TxStatusPending TxStatus = "pending"
	// TxStatusSuccess = received and executed
	TxStatusSuccess TxStatus = "success"
	// TxStatusFail = received and executed with error
	TxStatusFail TxStatus = "fail"
	// TxStatusInvalid = considered invalid
	TxStatusInvalid TxStatus = "invalid"
	// TxStatusRewardReverted represents the identifier for a reverted reward transaction
	TxStatusRewardReverted TxStatus = "reward-reverted"
)

// String returns the string representation of the status
func (tx TxStatus) String() string {
	return string(tx)
}

// statusComputer computes a transaction status
type statusComputer struct {
	selfShardId              uint32
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	store                    dataRetriever.StorageService
}

// Create a new instance of statusComputer
func NewStatusComputer(
	selfShardId uint32,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
	store dataRetriever.StorageService,
) *statusComputer {
	statusComputer := &statusComputer{
		selfShardId:              selfShardId,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		store:                    store,
	}

	return statusComputer
}

// ComputeStatusWhenInStorageKnowingMiniblock computes the transaction status for a historical transaction
func (sc *statusComputer) ComputeStatusWhenInStorageKnowingMiniblock(
	miniblockType block.Type,
	tx *ApiTransactionResult,
) (TxStatus, error) {

	if tx == nil {
		return "", ErrNilApiTransactionResult
	}

	isMiniblockFinalized := tx.NotarizedAtDestinationInMetaNonce > 0
	receiver := tx.Tx.GetRcvAddr()

	if sc.isMiniblockInvalid(miniblockType) {
		return TxStatusInvalid, nil
	}
	if isMiniblockFinalized || sc.isDestinationMe(tx.DestinationShard) || sc.isContractDeploy(receiver, tx.Data) {
		return TxStatusSuccess, nil
	}

	return TxStatusPending, nil
}

// ComputeStatusWhenInStorageNotKnowingMiniblock computes the transaction status when transaction is in current epoch's storage
// Limitation: in this case, since we do not know the miniblock type, we cannot know if a transaction is actually, "invalid".
// However, when "dblookupext" indexing is enabled, this function is not used.
func (sc *statusComputer) ComputeStatusWhenInStorageNotKnowingMiniblock(
	destinationShard uint32,
	tx *ApiTransactionResult,
) (TxStatus, error) {
	receiver := tx.Tx.GetRcvAddr()

	if tx == nil {
		return "", ErrNilApiTransactionResult
	}

	if sc.isDestinationMe(destinationShard) || sc.isContractDeploy(receiver, tx.Data) {
		return TxStatusSuccess, nil
	}

	// At least partially executed (since in source's storage)
	return TxStatusPending, nil
}

func (sc *statusComputer) isMiniblockInvalid(miniblockType block.Type) bool {
	return miniblockType == block.InvalidBlock
}

func (sc *statusComputer) isDestinationMe(destinationShard uint32) bool {
	return sc.selfShardId == destinationShard
}

func (sc *statusComputer) isContractDeploy(receiver []byte, transactionData []byte) bool {
	return core.IsEmptyAddress(receiver) && len(transactionData) > 0
}

// SetStatusIfIsRewardReverted will compute and set status for a reverted reward transaction
func (sc *statusComputer) SetStatusIfIsRewardReverted(
	tx *ApiTransactionResult,
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

	selfShardID := sc.selfShardId
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

	tx.Status = TxStatusRewardReverted
	return true, nil
}
