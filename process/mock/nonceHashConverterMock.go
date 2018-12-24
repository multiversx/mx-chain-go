package mock

import (
	"encoding/binary"
)

type nonceHashConverterMock struct {
}

func NewNonceHashConverterMock() *nonceHashConverterMock {
	return &nonceHashConverterMock{}
}

func (*nonceHashConverterMock) ToByteSlice(value uint64) []byte {
	buff := make([]byte, 8)

	binary.BigEndian.PutUint64(buff, value)

	return buff
}

func (*nonceHashConverterMock) ToUint64(buff []byte) *uint64 {
	if buff == nil {
		return nil
	}

	if len(buff) != 8 {
		return nil
	}

	value := binary.BigEndian.Uint64(buff)
	return &value
}
