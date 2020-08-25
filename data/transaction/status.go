package transaction

import (
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
	// Question for review: not executed means "executed with error", correct?
	TxStatusNotExecuted TxStatus = "not-executed"
	// TxStatusInvalid = considered invalid
	TxStatusInvalid TxStatus = "invalid"
)

// ComputeStatusWhenHistoricalTransaction computes the transaction status for a historical transaction
func ComputeStatusKnowingMiniblock(miniblockType block.Type, destinationShard uint32, selfShard uint32) TxStatus {
	if miniblockType == block.InvalidBlock {
		return TxStatusInvalid
	}
	if destinationShard == selfShard {
		return TxStatusExecuted
	}

	return TxStatusPartiallyExecuted
}

// ComputeStatusWhenInPool computes the transaction status when transaction is in pool
func ComputeStatusWhenInPool(sourceShard uint32, destinationShard uint32, selfShard uint32) TxStatus {
	isDestinationMe := selfShard == destinationShard
	isCrossShard := sourceShard != destinationShard
	if isDestinationMe && isCrossShard {
		return TxStatusPartiallyExecuted
	}

	return TxStatusReceived
}

// ComputeStatusWhenInCurrentEpochStorage computes the transaction status when transaction is in current epoch's storage
func ComputeStatusWhenInCurrentEpochStorage(sourceShard uint32, destinationShard uint32, selfShard uint32) TxStatus {
	// Question for review: here we cannot know if the transaction is, actually, "invalid".
	// However, when "fullHistory" indexing is enabled, this function is not used.

	isDestinationMe := selfShard == destinationShard
	if isDestinationMe {
		return TxStatusExecuted
	}

	// At least partially executed (since in source's storage)
	return TxStatusPartiallyExecuted
}
