package dataRetriever

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// RequestDataType represents the data type for the requested data
type RequestDataType byte

func (rdt RequestDataType) String() string {
	switch rdt {
	case HashType:
		return "hash type"
	case HashArrayType:
		return "hash array type"
	case NonceType:
		return "nonce type"
	default:
		return fmt.Sprintf("unknown type %d", rdt)
	}
}

const (
	// HashType indicates that the request data object is of type hash
	HashType RequestDataType = iota + 1
	// HashArrayType that the request data object contains a serialised array of hashes
	HashArrayType
	// NonceType indicates that the request data object is of type nonce (uint64)
	NonceType
)

// RequestData holds the requested data
// This struct will be serialized and sent to the other peers
type RequestData struct {
	Type  RequestDataType
	Value []byte
}

// Unmarshal sets the fields according to p2p.MessageP2P.Data() contents
// Errors if something went wrong
func (rd *RequestData) Unmarshal(marshalizer marshal.Marshalizer, message p2p.MessageP2P) error {
	if marshalizer == nil {
		return ErrNilMarshalizer
	}

	if message == nil {
		return ErrNilMessage
	}

	if message.Data() == nil {
		return ErrNilDataToProcess
	}

	err := marshalizer.Unmarshal(rd, message.Data())
	if err != nil {
		return err
	}

	return nil
}
