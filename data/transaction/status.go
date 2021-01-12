package transaction

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// TxStatus is the status of a transaction
type TxStatus string

const (
	// TxStatusPending = received and maybe executed on source shard, but not on destination shard
	TxStatusPending TxStatus = "pending"
	// TxStatusSuccess = received and executed
	TxStatusSuccess TxStatus = "success"
	// TxStatusFail = received and executed with error
	TxStatusFail TxStatus = "fail"
	// TxStatusInvalid = considered invalid
	TxStatusInvalid TxStatus = "invalid"
)

// String returns the string representation of the status
func (tx TxStatus) String() string {
	return string(tx)
}

// StatusComputer computes a transaction status
type StatusComputer struct {
	MiniblockType        block.Type
	IsMiniblockFinalized bool
	SourceShard          uint32
	DestinationShard     uint32
	Receiver             []byte
	TransactionData      []byte
	SelfShard            uint32
}

// ComputeStatusWhenInStorageKnowingMiniblock computes the transaction status for a historical transaction
func (params *StatusComputer) ComputeStatusWhenInStorageKnowingMiniblock() TxStatus {
	if params.isMiniblockInvalid() {
		return TxStatusInvalid
	}
	if params.IsMiniblockFinalized || params.isDestinationMe() || params.isContractDeploy() {
		return TxStatusSuccess
	}

	return TxStatusPending
}

// ComputeStatusWhenInStorageNotKnowingMiniblock computes the transaction status when transaction is in current epoch's storage
// Limitation: in this case, since we do not know the miniblock type, we cannot know if a transaction is actually, "invalid".
// However, when "dblookupext" indexing is enabled, this function is not used.
func (params *StatusComputer) ComputeStatusWhenInStorageNotKnowingMiniblock() TxStatus {
	if params.isDestinationMe() || params.isContractDeploy() {
		return TxStatusSuccess
	}

	// At least partially executed (since in source's storage)
	return TxStatusPending
}

func (params *StatusComputer) isMiniblockInvalid() bool {
	return params.MiniblockType == block.InvalidBlock
}

func (params *StatusComputer) isDestinationMe() bool {
	return params.SelfShard == params.DestinationShard
}

func (params *StatusComputer) isContractDeploy() bool {
	return core.IsEmptyAddress(params.Receiver) && len(params.TransactionData) > 0
}
