package mock

// Uint64ByteSliceConverterMock converts byte slice to/from uint64
type Uint64ByteSliceConverterMock struct {
	ToByteSliceCalled func(uint64) []byte
	ToUint64Called    func([]byte) (uint64, error)
}

// ToByteSlice is a mock implementation for Uint64ByteSliceConverter
func (u *Uint64ByteSliceConverterMock) ToByteSlice(p uint64) []byte {
	return u.ToByteSliceCalled(p)
}

// ToUint64 is a mock implementation for Uint64ByteSliceConverter
func (u *Uint64ByteSliceConverterMock) ToUint64(p []byte) (uint64, error) {
	return u.ToUint64Called(p)
}

// IsInterfaceNil returns true if there is no value under the interface
func (u *Uint64ByteSliceConverterMock) IsInterfaceNil() bool {
	return u == nil
}
