package mock

type Marshalizer struct {
}

func (j Marshalizer) Marshal(obj interface{}) ([]byte, error) {
	return nil, nil
}
func (j Marshalizer) Unmarshal(obj interface{}, buff []byte) error {
	return nil
}

type Hasher struct {
}

func (hash Hasher) Compute(s string) []byte {
	return nil
}
func (hash Hasher) EmptyHash() []byte {
	return nil
}
func (Hasher) Size() int {
	return 0
}
