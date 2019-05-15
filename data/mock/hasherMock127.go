package mock

type HasherMock127 struct {
}

func (HasherMock127) Compute(s string) []byte {
	buff := make([]byte, 0)

	var i byte
	for i = 0; i < 127; i++ {
		buff = append(buff, i)
	}

	return buff
}

func (HasherMock127) EmptyHash() []byte {
	return nil
}

func (HasherMock127) Size() int {
	return 64
}
