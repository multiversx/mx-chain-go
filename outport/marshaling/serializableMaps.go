package marshaling

type SerializableMapStringBytes struct {
	Keys   [][]byte
	Values [][]byte
}

func NewSerializableMapStringBytes(data map[string][]byte) *SerializableMapStringBytes {
	s := &SerializableMapStringBytes{
		Keys:   make([][]byte, 0, len(data)),
		Values: make([][]byte, 0, len(data)),
	}

	for key, value := range data {
		s.Keys = append(s.Keys, []byte(key))
		s.Values = append(s.Values, value)
	}

	return s
}

func (s *SerializableMapStringBytes) ConvertToMap() map[string][]byte {
	m := make(map[string][]byte)

	for i := 0; i < len(s.Keys); i++ {
		key := string(s.Keys[i])
		value := s.Values[i]
		m[key] = value
	}

	return m
}
