package dataCodec

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// AbiSerializer is the interface to work with abi codec
type AbiSerializer interface {
	Serialize(inputValues []any) (string, error)
	Deserialize(data string, outputValues []any) error
}

type EventDataProcessor interface {
	SerializeEventData(eventData sovereign.EventData) ([]byte, error)
	DeserializeEventData(data []byte) (*sovereign.EventData, error)
}

type TokenDataProcessor interface {
	SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error)
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
}

type OperationDataProcessor interface {
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
}

// SovereignDataDecoder is the interface for serializing/deserializing data
type SovereignDataDecoder interface {
	SerializeEventData(eventData sovereign.EventData) ([]byte, error)
	DeserializeEventData(data []byte) (*sovereign.EventData, error)
	SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error)
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
	IsInterfaceNil() bool
}
