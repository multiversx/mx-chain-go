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

// ComputeStatusKnowingMiniblock computes the transaction status knowing the parent miniblock
func ComputeStatusKnowingMiniblock(miniblock *block.MiniBlock, selfShard uint32) TxStatus {
	if miniblock.Type == block.InvalidBlock {
		return TxStatusInvalid
	}
	if miniblock.ReceiverShardID == selfShard {
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

// ComputeStatusWhenInStorage computes the transaction status when transaction is in storage
func ComputeStatusWhenInStorage(sourceShard uint32, destinationShard uint32, selfShard uint32) TxStatus {
	isDestinationMe := selfShard == destinationShard
	if isDestinationMe {
		return TxStatusExecuted
	}

	// At least partially executed (since in source's storage)
	return TxStatusPartiallyExecuted
}
