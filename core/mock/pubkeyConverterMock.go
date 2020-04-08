package mock

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

// PubkeyConverterMock -
type PubkeyConverterMock struct {
	len int
}

// NewPubkeyConverterMock -
func NewPubkeyConverterMock(addressLen int) *PubkeyConverterMock {
	return &PubkeyConverterMock{
		len: addressLen,
	}
}

// Decode -
func (pcm *PubkeyConverterMock) Decode(humanReadable string) ([]byte, error) {
	return hex.DecodeString(humanReadable)
}

// Encode -
func (pcm *PubkeyConverterMock) Encode(pkBytes []byte) string {
	return hex.EncodeToString(pkBytes)
}

// CreateAddressFromString -
func (pcm *PubkeyConverterMock) CreateAddressFromString(humanReadable string) (state.AddressContainer, error) {
	buff, err := pcm.Decode(humanReadable)
	if err != nil {
		return nil, err
	}

	return state.NewAddress(buff), nil
}

// CreateAddressFromBytes -
func (pcm *PubkeyConverterMock) CreateAddressFromBytes(pkBytes []byte) (state.AddressContainer, error) {
	return state.NewAddress(pkBytes), nil
}

// Len -
func (pcm *PubkeyConverterMock) Len() int {
	return pcm.len
}

// IsInterfaceNil -
func (pcm *PubkeyConverterMock) IsInterfaceNil() bool {
	return pcm == nil
}
