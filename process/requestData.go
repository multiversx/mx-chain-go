package process

import (
	"fmt"
)

// RequestDataType represents the data type for the requested data
type RequestDataType byte

func (rdt RequestDataType) String() string {
	switch rdt {
	case HashType:
		return "hash type"
	case NonceType:
		return "nonce type"
	}

	return fmt.Sprintf("unknown type %d", rdt)
}

const (
	// HashType indicates that the request data object is of type hash
	HashType RequestDataType = 0
	// NonceType indicates that the request data object is of type nonce (uint64)
	NonceType RequestDataType = 1
)

// RequestData holds the requested data
// This struct will be serialized and sent to the other peers
type RequestData struct {
	Type  RequestDataType
	Value []byte
}
