package uint64ByteSlice

import (
	"encoding/binary"
)

type bigEndianConverter struct {
}

func NewBigEndianConverter() *bigEndianConverter {
	return &bigEndianConverter{}
}

func (*bigEndianConverter) ToByteSlice(value uint64) []byte {
	buff := make([]byte, 8)

	binary.BigEndian.PutUint64(buff, value)

	return buff
}

func (*bigEndianConverter) ToUint64(buff []byte) *uint64 {
	if buff == nil {
		return nil
	}

	if len(buff) != 8 {
		return nil
	}

	value := binary.BigEndian.Uint64(buff)
	return &value
}
