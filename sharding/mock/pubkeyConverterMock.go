package mock

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/core"
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
func (pcm *PubkeyConverterMock) Encode(pkBytes []byte) (string, error) {
	return hex.EncodeToString(pkBytes), nil
}

// EncodeSlice -
func (pcm *PubkeyConverterMock) EncodeSlice(pkBytesSlice [][]byte) ([]string, error) {
	encodedSlice := make([]string, 0)

	for _, pkBytes := range pkBytesSlice {
		encodedSlice = append(encodedSlice, hex.EncodeToString(pkBytes))
	}

	return encodedSlice, nil
}

// SilentEncode -
func (pcm *PubkeyConverterMock) SilentEncode(pkBytes []byte, log core.Logger) string {
	return hex.EncodeToString(pkBytes)
}

// Len -
func (pcm *PubkeyConverterMock) Len() int {
	return pcm.len
}

// IsInterfaceNil -
func (pcm *PubkeyConverterMock) IsInterfaceNil() bool {
	return pcm == nil
}
