package mock

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

// PubkeyConverterMock -
type PubkeyConverterMock struct {
	addressLen int
}

// NewPubkeyConverterMock -
func NewPubkeyConverterMock(addressLen int) *PubkeyConverterMock {
	return &PubkeyConverterMock{
		addressLen: addressLen,
	}
}

// Bytes -
func (pcm *PubkeyConverterMock) Bytes(humanReadable string) ([]byte, error) {
	return hex.DecodeString(humanReadable)
}

// String -
func (pcm *PubkeyConverterMock) String(pkBytes []byte) (string, error) {
	return hex.EncodeToString(pkBytes), nil
}

// CreateAddressFromString -
func (pcm *PubkeyConverterMock) CreateAddressFromString(humanReadable string) (state.AddressContainer, error) {
	buff, err := pcm.Bytes(humanReadable)
	if err != nil {
		return nil, err
	}

	return state.NewAddress(buff), nil
}

// CreateAddressFromBytes -
func (pcm *PubkeyConverterMock) CreateAddressFromBytes(pkBytes []byte) (state.AddressContainer, error) {
	return state.NewAddress(pkBytes), nil
}

// AddressLen -
func (pcm *PubkeyConverterMock) AddressLen() int {
	return pcm.addressLen
}

// IsInterfaceNil -
func (pcm *PubkeyConverterMock) IsInterfaceNil() bool {
	return pcm == nil
}
