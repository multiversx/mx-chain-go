package mock

import (
	"encoding/binary"

	"github.com/pkg/errors"
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

func (*nonceHashConverterMock) ToUint64(buff []byte) (uint64, error) {
	if buff == nil {
		return 0, errors.New("failure, nil slice")
	}

	if len(buff) != 8 {
		return 0, errors.New("failure, len not 8")
	}

	return binary.BigEndian.Uint64(buff), nil
}
