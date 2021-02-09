package transaction

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// Encoder represents a byte slice to string encoder
type Encoder interface {
	Encode(buff []byte) string
	IsInterfaceNil() bool
}

// Marshalizer is able to encode an object to its byte slice representation
type Marshalizer interface {
	Marshal(obj interface{}) ([]byte, error)
	IsInterfaceNil() bool
}

// StatusComputerHandler computes a transaction status
type StatusComputerHandler interface {
	ComputeStatusWhenInStorageKnowingMiniblock(miniblockType block.Type, tx *ApiTransactionResult) TxStatus
	ComputeStatusWhenInStorageNotKnowingMiniblock(destinationShard uint32, tx *ApiTransactionResult) TxStatus
	SetStatusIfIsRewardReverted(
		tx *ApiTransactionResult,
		miniblockType block.Type,
		headerNonce uint64,
		headerHash []byte,
) bool
}
