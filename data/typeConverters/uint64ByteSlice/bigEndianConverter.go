package uint64ByteSlice

import (
	"encoding/binary"

	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
)

type bigEndianConverter struct {
}

// NewBigEndianConverter creates a big endian uint64-byte slice converter
func NewBigEndianConverter() *bigEndianConverter {
	return &bigEndianConverter{}
}

// ToByteSlice converts uint64 to its byte slice representation
func (*bigEndianConverter) ToByteSlice(value uint64) []byte {
	buff := make([]byte, 8)

	binary.BigEndian.PutUint64(buff, value)

	return buff
}

// ToUint64 converts a byte slice to uint64. Errors if something went wrong
func (*bigEndianConverter) ToUint64(buff []byte) (uint64, error) {
	if buff == nil {
		return 0, typeConverters.ErrNilByteSlice
	}

	if len(buff) != 8 {
		return 0, typeConverters.ErrByteSliceLenShouldHaveBeen8
	}

	return binary.BigEndian.Uint64(buff), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bec *bigEndianConverter) IsInterfaceNil() bool {
	if bec == nil {
		return true
	}
	return false
}
