package mock

import "encoding/binary"

// Uint64ByteSliceConverterMock converts byte slice to/from uint64
type Uint64ByteSliceConverterMock struct {
}

// ToByteSlice is a mock implementation for Uint64ByteSliceConverter
func (u *Uint64ByteSliceConverterMock) ToByteSlice(p uint64) []byte {
	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, p)

	return buff
}

// ToUint64 is a mock implementation for Uint64ByteSliceConverter
func (u *Uint64ByteSliceConverterMock) ToUint64(p []byte) (uint64, error) {
	return binary.BigEndian.Uint64(p), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (u *Uint64ByteSliceConverterMock) IsInterfaceNil() bool {
	return u == nil
}
