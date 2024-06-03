package dataCodec

import (
	abiSerializer "github.com/multiversx/mx-chain-go/process/block/sovereign"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

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
	abiSerializer.AbiSerializer
	SerializeEventData(eventData sovereign.EventData) ([]byte, error)
	DeserializeEventData(data []byte) (*sovereign.EventData, error)
	SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error)
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
	IsInterfaceNil() bool
}
