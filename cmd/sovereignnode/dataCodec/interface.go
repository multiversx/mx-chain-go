package dataCodec

import "github.com/multiversx/mx-chain-core-go/data/sovereign"

// DataCodecProcessor is the interface for serializing/deserializing data
type DataCodecProcessor interface {
	SerializeEventData(eventData sovereign.EventData) ([]byte, error)
	DeserializeEventData(data []byte) (*sovereign.EventData, error)
	SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error)
	DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error)
	GetTokenDataBytes(tokenNonce []byte, tokenData []byte) ([]byte, error)
	SerializeOperation(operation sovereign.Operation) ([]byte, error)
	IsInterfaceNil() bool
}
