package incomingHeader

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// HeadersPool should be able to add new headers in pool
type HeadersPool interface {
	AddHeaderInShard(headerHash []byte, header data.HeaderHandler, shardID uint32)
	IsInterfaceNil() bool
}

// TransactionPool should be able to add a new transaction in the pool
type TransactionPool interface {
	AddData(key []byte, data interface{}, sizeInBytes int, cacheId string)
	IsInterfaceNil() bool
}

// TopicsChecker should be able to check the topics validity
type TopicsChecker interface {
	CheckValidity(topics [][]byte) error
	IsInterfaceNil() bool
}

// AbiSerializer is the interface to work with abi codec
type AbiSerializer interface {
	Serialize(inputValues []any) (string, error)
	Deserialize(data string, outputValues []any) error
}

// EventDataEncoder is the interface for serializing/deserializing event data
type EventDataEncoder interface {
	SerializeEventData(eventData sovereign.EventData) ([]byte, error)
	DeserializeEventData(data []byte) (*sovereign.EventData, error)
}

// TokenDataEncoder is the interface for serializing/deserializing token data
type TokenDataEncoder interface {
	SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error)
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
}

// OperationDataEncoder is the interface for serializing operations
type OperationDataEncoder interface {
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
}

// SovereignDataCodec is the interface for serializing/deserializing data
type SovereignDataCodec interface {
	SerializeEventData(eventData sovereign.EventData) ([]byte, error)
	DeserializeEventData(data []byte) (*sovereign.EventData, error)
	SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error)
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
	IsInterfaceNil() bool
}
