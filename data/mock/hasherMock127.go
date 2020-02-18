package mock

// HasherMock127 -
type HasherMock127 struct {
}

// Compute -
func (HasherMock127) Compute(s string) []byte {
	buff := make([]byte, 0)

	var i byte
	for i = 0; i < 127; i++ {
		buff = append(buff, i)
	}

	return buff
}

// EmptyHash -
func (HasherMock127) EmptyHash() []byte {
	return nil
}

// Size -
func (HasherMock127) Size() int {
	return 64
}

// IsInterfaceNil returns true if there is no value under the interface
func (hash *HasherMock127) IsInterfaceNil() bool {
	return hash == nil
}
