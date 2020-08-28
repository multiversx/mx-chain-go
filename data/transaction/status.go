package transaction

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// TxStatus is the status of a transaction
type TxStatus string

const (
	// TxStatusReceived = received but not yet executed
	TxStatusReceived TxStatus = "received"
	// TxStatusPartiallyExecuted = received and executed on source shard
	TxStatusPartiallyExecuted TxStatus = "partially-executed"
	// TxStatusExecuted = received and executed
	TxStatusExecuted TxStatus = "executed"
	// TxStatusNotExecuted = received and executed with error
	TxStatusNotExecuted TxStatus = "not-executed"
	// TxStatusInvalid = considered invalid
	TxStatusInvalid TxStatus = "invalid"
)

// StatusComputer computes a transaction status
type StatusComputer struct {
	MiniblockType    block.Type
	SourceShard      uint32
	DestinationShard uint32
	Receiver         []byte
	TransactionData  []byte
	SelfShard        uint32
}

// ComputeStatusWhenInPool computes the transaction status when transaction is in pool
func (params *StatusComputer) ComputeStatusWhenInPool() TxStatus {
	if params.isContractDeploy() {
		return TxStatusReceived
	}
	if params.isDestinationMe() && params.isCrossShard() {
		return TxStatusPartiallyExecuted
	}

	return TxStatusReceived
}

// ComputeStatusWhenInStorageKnowingMiniblock computes the transaction status for a historical transaction
func (params *StatusComputer) ComputeStatusWhenInStorageKnowingMiniblock() TxStatus {
	if params.isMiniblockInvalid() {
		return TxStatusInvalid
	}
	if params.isDestinationMe() || params.isContractDeploy() {
		return TxStatusExecuted
	}

	return TxStatusPartiallyExecuted
}

// ComputeStatusWhenInStorageNotKnowingMiniblock computes the transaction status when transaction is in current epoch's storage
// Limitation: in this case, since we do not know the miniblock type, we cannot know if a transaction is actually, "invalid".
// However, when "dblookupext" indexing is enabled, this function is not used.
func (params *StatusComputer) ComputeStatusWhenInStorageNotKnowingMiniblock() TxStatus {
	if params.isDestinationMe() || params.isContractDeploy() {
		return TxStatusExecuted
	}

	// At least partially executed (since in source's storage)
	return TxStatusPartiallyExecuted
}

func (params *StatusComputer) isMiniblockInvalid() bool {
	return params.MiniblockType == block.InvalidBlock
}

func (params *StatusComputer) isDestinationMe() bool {
	return params.SelfShard == params.DestinationShard
}

func (params *StatusComputer) isCrossShard() bool {
	return params.SourceShard != params.DestinationShard
}

func (params *StatusComputer) isContractDeploy() bool {
	return core.IsEmptyAddress(params.Receiver) && len(params.TransactionData) > 0
}
