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

// StatusComputer computes a transaction status
type StatusComputer struct {
	SelfShardId uint32
}

// Create a new instance of StatusComputer
func NewStatusComputer(selfShardId uint32) *StatusComputer {
	statusComputer := &StatusComputer{
		SelfShardId: selfShardId}

	return statusComputer
}

// ComputeStatusWhenInStorageKnowingMiniblock computes the transaction status for a historical transaction
func (sc *StatusComputer) ComputeStatusWhenInStorageKnowingMiniblock(
	miniblockType block.Type,
	tx *ApiTransactionResult,
) TxStatus {
	isMiniblockFinalized := tx.NotarizedAtDestinationInMetaNonce > 0
	receiver := tx.Tx.GetRcvAddr()

	if sc.isMiniblockInvalid(miniblockType) {
		return TxStatusInvalid
	}
	if isMiniblockFinalized || sc.isDestinationMe(tx.DestinationShard) || sc.isContractDeploy(receiver,tx.Data) {
		return TxStatusSuccess
	}

	return TxStatusPending
}

// ComputeStatusWhenInStorageNotKnowingMiniblock computes the transaction status when transaction is in current epoch's storage
// Limitation: in this case, since we do not know the miniblock type, we cannot know if a transaction is actually, "invalid".
// However, when "dblookupext" indexing is enabled, this function is not used.
func (sc *StatusComputer) ComputeStatusWhenInStorageNotKnowingMiniblock(
	destinationShard uint32,
	tx *ApiTransactionResult,
) TxStatus {
	receiver := tx.Tx.GetRcvAddr()
	transactionData := tx.Data

	if sc.isDestinationMe(destinationShard) || sc.isContractDeploy(receiver,transactionData) {
		return TxStatusSuccess
	}

	// At least partially executed (since in source's storage)
	return TxStatusPending
}

func (sc *StatusComputer) isMiniblockInvalid(miniblockType block.Type) bool {
	return miniblockType == block.InvalidBlock
}

func (sc *StatusComputer) isDestinationMe(destinationShard uint32) bool {
	return sc.SelfShardId == destinationShard
}

func (sc *StatusComputer) isContractDeploy(receiver []byte, transactionData []byte) bool {
	return core.IsEmptyAddress(receiver) && len(transactionData) > 0
}

// SetStatusIfIsRewardReverted will compute and set status for a reverted reward transaction
func (sc *StatusComputer) SetStatusIfIsRewardReverted(
	tx *ApiTransactionResult,
	miniblockType block.Type,
	headerNonce uint64,
	headerHash []byte,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
	store dataRetriever.StorageService,
) bool {

	if miniblockType != block.RewardsBlock {
		return false
	}

	var storerUnit dataRetriever.UnitType

	selfShardID := sc.SelfShardId
	if selfShardID == core.MetachainShardId {
		storerUnit = dataRetriever.MetaHdrNonceHashDataUnit
	} else {
		storerUnit = dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(selfShardID)
	}

	nonceToByteSlice := uint64ByteSliceConverter.ToByteSlice(headerNonce)
	headerHashFromStorage, err := store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		log.Warn("cannot get header hash by nonce", "error", err.Error())
		return false
	}

	if bytes.Equal(headerHashFromStorage, headerHash) {
		return false
	}

	tx.Status = TxStatusRewardReverted
	return true
}
