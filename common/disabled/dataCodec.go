package disabled

import "github.com/multiversx/mx-chain-core-go/data/sovereign"

type dataCodec struct {
}

// NewDisabledDataCodec -
func NewDisabledDataCodec() *dataCodec {
	return &dataCodec{}
}

// SerializeEventData returns nothing
func (dc *dataCodec) SerializeEventData(_ sovereign.EventData) ([]byte, error) {
	return make([]byte, 0), nil
}

// DeserializeEventData returns nothing
func (dc *dataCodec) DeserializeEventData(_ []byte) (*sovereign.EventData, error) {
	return &sovereign.EventData{}, nil
}

// SerializeTokenData returns nothing
func (dc *dataCodec) SerializeTokenData(_ sovereign.EsdtTokenData) ([]byte, error) {
	return make([]byte, 0), nil
}

// DeserializeTokenData returns nothing
func (dc *dataCodec) DeserializeTokenData(_ []byte) (*sovereign.EsdtTokenData, error) {
	return &sovereign.EsdtTokenData{}, nil
}

// GetTokenDataBytes returns nothing
func (dc *dataCodec) GetTokenDataBytes(_ []byte, _ []byte) ([]byte, error) {
	return make([]byte, 0), nil
}

// SerializeOperation returns nothing
func (dc *dataCodec) SerializeOperation(_ sovereign.Operation) ([]byte, error) {
	return make([]byte, 0), nil
}

// IsInterfaceNil - returns true if there is no value under the interface
func (dc *dataCodec) IsInterfaceNil() bool {
	return dc == nil
}
