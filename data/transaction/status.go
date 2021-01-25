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
	MiniblockType            block.Type
	IsMiniblockFinalized     bool
	SourceShard              uint32
	DestinationShard         uint32
	Receiver                 []byte
	TransactionData          []byte
	SelfShard                uint32
	HeaderNonce              uint64
	HeaderHash               []byte
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	Store                    dataRetriever.StorageService
}

// ComputeStatusWhenInStorageKnowingMiniblock computes the transaction status for a historical transaction
func (sc *StatusComputer) ComputeStatusWhenInStorageKnowingMiniblock() TxStatus {
	if sc.isMiniblockInvalid() {
		return TxStatusInvalid
	}
	if sc.IsMiniblockFinalized || sc.isDestinationMe() || sc.isContractDeploy() {
		return TxStatusSuccess
	}

	return TxStatusPending
}

// ComputeStatusWhenInStorageNotKnowingMiniblock computes the transaction status when transaction is in current epoch's storage
// Limitation: in this case, since we do not know the miniblock type, we cannot know if a transaction is actually, "invalid".
// However, when "dblookupext" indexing is enabled, this function is not used.
func (sc *StatusComputer) ComputeStatusWhenInStorageNotKnowingMiniblock() TxStatus {
	if sc.isDestinationMe() || sc.isContractDeploy() {
		return TxStatusSuccess
	}

	// At least partially executed (since in source's storage)
	return TxStatusPending
}

func (sc *StatusComputer) isMiniblockInvalid() bool {
	return sc.MiniblockType == block.InvalidBlock
}

func (sc *StatusComputer) isDestinationMe() bool {
	return sc.SelfShard == sc.DestinationShard
}

func (sc *StatusComputer) isContractDeploy() bool {
	return core.IsEmptyAddress(sc.Receiver) && len(sc.TransactionData) > 0
}

// SetStatusIfIsRewardReverted will compute and set status for a reverted reward transaction
func (sc *StatusComputer) SetStatusIfIsRewardReverted(tx *ApiTransactionResult) bool {
	if sc.MiniblockType != block.RewardsBlock {
		return false
	}

	var storerUnit dataRetriever.UnitType

	selfShardID := sc.SelfShard
	if selfShardID == core.MetachainShardId {
		storerUnit = dataRetriever.MetaHdrNonceHashDataUnit
	} else {
		storerUnit = dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(selfShardID)
	}

	nonceToByteSlice := sc.Uint64ByteSliceConverter.ToByteSlice(sc.HeaderNonce)
	headerHash, err := sc.Store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		log.Warn("cannot get header hash by nonce", "error", err.Error())
		return false
	}

	if bytes.Equal(headerHash, sc.HeaderHash) {
		return false
	}

	tx.Status = TxStatusRewardReverted
	return true
}
